package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/anthropicfp"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestGatewayClientDatelineNormalization_Scope covers the account/switch matrix
// for the shouldNormalizeClientDateline gate: Anthropic OAuth/SetupToken pass
// only when the switch is on; API-Key and non-Anthropic platforms are excluded
// unconditionally.
func TestGatewayClientDatelineNormalization_Scope(t *testing.T) {
	repo := &gatewayTTLSettingRepo{data: map[string]string{}}
	gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{})
	svc := &GatewayService{
		settingService: NewSettingService(repo, &config.Config{}),
	}
	ctx := context.Background()

	// Default (missing key): fallback in parseSettings/cache loader is true.
	require.True(t, svc.shouldNormalizeClientDateline(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}))
	require.True(t, svc.shouldNormalizeClientDateline(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeSetupToken}))
	require.False(t, svc.shouldNormalizeClientDateline(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeAPIKey}))
	require.False(t, svc.shouldNormalizeClientDateline(ctx, &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth}))

	// Switch off: no account qualifies.
	repo.data[SettingKeyEnableClientDatelineNormalization] = "false"
	gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{})
	require.False(t, svc.shouldNormalizeClientDateline(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}))
	require.False(t, svc.shouldNormalizeClientDateline(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeSetupToken}))

	// Switch back on: OAuth qualifies again.
	repo.data[SettingKeyEnableClientDatelineNormalization] = "true"
	gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{})
	require.True(t, svc.shouldNormalizeClientDateline(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}))
}

// TestGatewayClientDatelineNormalization_HelperNoRewrite exercises the code
// path used by Forward: the helper must return ok=false when the switch is
// off, when the account is API-Key, when the account is nil, and when the
// body carries no fingerprinted dateline. It must return ok=true and a
// rewritten body when both the switch is on and the account is Anthropic
// OAuth/SetupToken and a rewrite actually happened.
func TestGatewayClientDatelineNormalization_HelperNoRewrite(t *testing.T) {
	repo := &gatewayTTLSettingRepo{data: map[string]string{
		SettingKeyEnableClientDatelineNormalization: "true",
	}}
	gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{})
	svc := &GatewayService{
		settingService: NewSettingService(repo, &config.Config{}),
	}
	ctx := context.Background()

	dirty := []byte(`{"messages":[{"role":"user","content":"<system-reminder>\nToday’s date is 2026/07/01.\n</system-reminder>"}]}`)
	clean := []byte(`{"messages":[{"role":"user","content":"just hello"}]}`)

	// API-Key account: never rewrites, even with dirty payload.
	next, ok := svc.normalizeClientDatelineIfEnabled(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeAPIKey}, dirty)
	require.False(t, ok)
	require.Nil(t, next)

	// Nil account: safe no-op.
	next, ok = svc.normalizeClientDatelineIfEnabled(ctx, nil, dirty)
	require.False(t, ok)
	require.Nil(t, next)

	// OAuth account + clean body: no changes, ok=false.
	next, ok = svc.normalizeClientDatelineIfEnabled(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}, clean)
	require.False(t, ok)
	require.Nil(t, next)

	// OAuth account + dirty body: rewritten, ok=true.
	next, ok = svc.normalizeClientDatelineIfEnabled(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}, dirty)
	require.True(t, ok)
	require.NotNil(t, next)
	require.Contains(t, string(next), "Today's date is 2026-07-01.")
	require.NotContains(t, string(next), "2026/07/01")
	require.NotContains(t, string(next), "Today’s date is")

	// SetupToken account + dirty body: rewritten, ok=true.
	next, ok = svc.normalizeClientDatelineIfEnabled(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeSetupToken}, dirty)
	require.True(t, ok)
	require.Contains(t, string(next), "Today's date is 2026-07-01.")

	// Switch off: even OAuth account is not rewritten.
	repo.data[SettingKeyEnableClientDatelineNormalization] = "false"
	gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{})
	next, ok = svc.normalizeClientDatelineIfEnabled(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}, dirty)
	require.False(t, ok)
	require.Nil(t, next)
}

// TestGatewayClientDatelineNormalization_LeavesUserProseUntouched double-checks
// that the pure normalizer never touches content outside <system-reminder>
// blocks. This is an integration guard between the switch-gated helper and
// the pkg/anthropicfp scope contract, tripped by anyone who broadens scope.
func TestGatewayClientDatelineNormalization_LeavesUserProseUntouched(t *testing.T) {
	repo := &gatewayTTLSettingRepo{data: map[string]string{
		SettingKeyEnableClientDatelineNormalization: "true",
	}}
	gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{})
	svc := &GatewayService{
		settingService: NewSettingService(repo, &config.Config{}),
	}
	ctx := context.Background()

	// User prose that happens to include a fingerprint-looking sentence
	// (outside <system-reminder>) must be preserved byte-for-byte.
	body := []byte(`{"messages":[{"role":"user","content":"I wrote: Today’s date is 2026/07/01. What do you think?"}]}`)
	next, ok := svc.normalizeClientDatelineIfEnabled(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}, body)
	require.False(t, ok, "must not rewrite user prose outside <system-reminder>")
	require.Nil(t, next)

	// Direct pure-fn check for redundancy.
	out, hits, changed := anthropicfp.NormalizeDateline(body)
	require.False(t, changed)
	require.Empty(t, hits)
	require.Equal(t, body, out)
}

func TestGatewayClientDatelineNormalization_TimezoneOverride(t *testing.T) {
	// Fix clock to 2026-11-13 02:30:00 UTC:
	// - Asia/Shanghai (UTC+8)      -> 2026-11-13
	// - America/Los_Angeles (UTC-8)-> 2026-11-12
	prevNow := gatewayDatelineNowFunc
	t.Cleanup(func() { gatewayDatelineNowFunc = prevNow })
	gatewayDatelineNowFunc = func() time.Time {
		return time.Date(2026, 11, 13, 2, 30, 0, 0, time.UTC)
	}

	repo := &gatewayTTLSettingRepo{data: map[string]string{
		SettingKeyEnableClientDatelineNormalization: "true",
		SettingKeyClientDatelineTimezone:            "America/Los_Angeles",
	}}
	gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{})
	svc := &GatewayService{
		settingService: NewSettingService(repo, &config.Config{}),
	}
	ctx := context.Background()

	// Canonical ASCII dateline from an Asia/Shanghai client (2026-11-13) gets aligned to US Pacific (2026-11-12),
	// while user prose outside <system-reminder> is untouched.
	body := []byte(`{"system":"Today's date is 2026-11-13.","messages":[{"role":"user","content":"<system-reminder>\n# currentDate\nToday's date is 2026-11-13.\n</system-reminder>\nNote: Today's date is 2026-11-13."}]}`)
	next, ok := svc.normalizeClientDatelineIfEnabled(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}, body)
	require.True(t, ok)
	require.Equal(t, "Today's date is 2026-11-12.", gjson.GetBytes(next, "system").String())
	msgContent := gjson.GetBytes(next, "messages.0.content").String()
	require.Contains(t, msgContent, "<system-reminder>\n# currentDate\nToday's date is 2026-11-12.\n</system-reminder>")
	require.Contains(t, msgContent, "Note: Today's date is 2026-11-13.")

	// API-Key account remains untouched even when timezone is configured.
	nextAPIKey, okAPIKey := svc.normalizeClientDatelineIfEnabled(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeAPIKey}, body)
	require.False(t, okAPIKey)
	require.Nil(t, nextAPIKey)

	// Empty timezone preserves original client date when already ASCII canonical.
	repo.data[SettingKeyClientDatelineTimezone] = ""
	gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{})
	nextEmptyTZ, okEmptyTZ := svc.normalizeClientDatelineIfEnabled(ctx, &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}, body)
	require.False(t, okEmptyTZ)
	require.Nil(t, nextEmptyTZ)
}

func TestValidateClientDatelineTimezone(t *testing.T) {
	require.NoError(t, ValidateClientDatelineTimezone(""))
	require.NoError(t, ValidateClientDatelineTimezone("  "))
	require.NoError(t, ValidateClientDatelineTimezone("America/Los_Angeles"))
	require.NoError(t, ValidateClientDatelineTimezone("UTC"))
	require.Error(t, ValidateClientDatelineTimezone("Local"))
	require.Error(t, ValidateClientDatelineTimezone("local"))
	require.Error(t, ValidateClientDatelineTimezone("Mars/Olympus_Mons"))
	require.Equal(t, "America/Los_Angeles", NormalizeClientDatelineTimezone(" America/Los_Angeles "))
	require.Equal(t, "", NormalizeClientDatelineTimezone("Local"))
	require.Equal(t, "", NormalizeClientDatelineTimezone("Invalid/Zone"))
}
