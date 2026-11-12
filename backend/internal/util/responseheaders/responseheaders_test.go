package responseheaders

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func TestFilterHeadersDisabledUsesDefaultAllowlist(t *testing.T) {
	src := http.Header{}
	src.Add("Content-Type", "application/json")
	src.Add("X-Request-Id", "req-123")
	src.Add("X-Test", "ok")
	src.Add("Connection", "keep-alive")
	src.Add("Content-Length", "123")

	cfg := config.ResponseHeaderConfig{
		Enabled:     false,
		ForceRemove: []string{"x-request-id"},
	}

	filtered := FilterHeaders(src, CompileHeaderFilter(cfg))
	if filtered.Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type passthrough, got %q", filtered.Get("Content-Type"))
	}
	if filtered.Get("X-Request-Id") != "req-123" {
		t.Fatalf("expected X-Request-Id allowed, got %q", filtered.Get("X-Request-Id"))
	}
	if filtered.Get("X-Test") != "" {
		t.Fatalf("expected X-Test removed, got %q", filtered.Get("X-Test"))
	}
	if filtered.Get("Connection") != "" {
		t.Fatalf("expected Connection to be removed, got %q", filtered.Get("Connection"))
	}
	if filtered.Get("Content-Length") != "" {
		t.Fatalf("expected Content-Length to be removed, got %q", filtered.Get("Content-Length"))
	}
}

func TestFilterHeadersAllowsReasoningIncludedByDefault(t *testing.T) {
	src := http.Header{}
	src.Set("X-Reasoning-Included", "1")

	filtered := FilterHeaders(src, CompileHeaderFilter(config.ResponseHeaderConfig{}))
	if got := filtered.Get("X-Reasoning-Included"); got != "1" {
		t.Fatalf("expected X-Reasoning-Included passthrough, got %q", got)
	}
}

func TestFilterHeadersForceRemoveOverridesReasoningIncluded(t *testing.T) {
	src := http.Header{}
	src.Set("X-Reasoning-Included", "1")

	filtered := FilterHeaders(src, CompileHeaderFilter(config.ResponseHeaderConfig{
		Enabled:     true,
		ForceRemove: []string{"x-reasoning-included"},
	}))
	if got := filtered.Get("X-Reasoning-Included"); got != "" {
		t.Fatalf("expected X-Reasoning-Included removal, got %q", got)
	}
}

func TestFilterHeadersEnabledUsesAllowlist(t *testing.T) {
	src := http.Header{}
	src.Add("Content-Type", "application/json")
	src.Add("X-Extra", "ok")
	src.Add("X-Remove", "nope")
	src.Add("X-Blocked", "nope")

	cfg := config.ResponseHeaderConfig{
		Enabled:           true,
		AdditionalAllowed: []string{"x-extra"},
		ForceRemove:       []string{"x-remove"},
	}

	filtered := FilterHeaders(src, CompileHeaderFilter(cfg))
	if filtered.Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type allowed, got %q", filtered.Get("Content-Type"))
	}
	if filtered.Get("X-Extra") != "ok" {
		t.Fatalf("expected X-Extra allowed, got %q", filtered.Get("X-Extra"))
	}
	if filtered.Get("X-Remove") != "" {
		t.Fatalf("expected X-Remove removed, got %q", filtered.Get("X-Remove"))
	}
	if filtered.Get("X-Blocked") != "" {
		t.Fatalf("expected X-Blocked removed, got %q", filtered.Get("X-Blocked"))
	}
}

func TestFilterHeadersNeverAllowsUpstreamAccountIdentityHeaders(t *testing.T) {
	sensitiveHeaders := map[string]string{
		"OpenAI-Organization":       "org-SyntheticOrg0123456789AB",
		"OpenAI-Project":            "proj_secret12345678",
		"ChatGPT-Account-Id":        "00000000-1111-4222-8333-444444444444",
		"Anthropic-Organization-Id": "org_anthropic_secret",
		"X-Account-Email":           "poolmember02@example.com",
		"X-User-Email":              "poolmember02@example.com",
		"X-Organization-Id":         "org-SyntheticOrg0123456789AB",
	}

	src := http.Header{}
	src.Set("Content-Type", "application/json")
	allowAttempt := make([]string, 0, len(sensitiveHeaders))
	for k, v := range sensitiveHeaders {
		src.Set(k, v)
		allowAttempt = append(allowAttempt, k)
	}

	// Even when admin config explicitly puts sensitive headers into AdditionalAllowed, they must be stripped.
	compiled := CompileHeaderFilter(config.ResponseHeaderConfig{
		Enabled:           true,
		AdditionalAllowed: allowAttempt,
	})
	filtered := FilterHeaders(src, compiled)
	if filtered.Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type preserved, got %q", filtered.Get("Content-Type"))
	}
	for k := range sensitiveHeaders {
		if got := filtered.Get(k); got != "" {
			t.Fatalf("expected sensitive header %s to be stripped, got %q", k, got)
		}
	}

	// Also verify nil compiled filter (default allowlist mode) strips them.
	filteredDefault := FilterHeaders(src, nil)
	for k := range sensitiveHeaders {
		if got := filteredDefault.Get(k); got != "" {
			t.Fatalf("expected sensitive header %s to be stripped in default mode, got %q", k, got)
		}
	}
}

