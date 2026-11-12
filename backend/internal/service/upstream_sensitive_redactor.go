package service

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	// ErrDedicatedPublicSensitiveForbidden 当管理员误把真实邮箱、令牌或上游组织/用户标识填进对外公开字段时拦截。
	ErrDedicatedPublicSensitiveForbidden = infraerrors.New(
		http.StatusBadRequest,
		"DEDICATED_PUBLIC_SENSITIVE_FORBIDDEN",
		"对外展示别名或商城卖点说明中禁止包含邮箱、密钥令牌或上游账号标识等敏感信息（如需交付原始账号信息，请填写在「买断交付信息」中）",
	)

	upstreamEmailRegex        = regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	upstreamEncodedEmailRegex = regexp.MustCompile(`(?i)[A-Za-z0-9._%+\-]+(?:%40|\\u0040)[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`)
	upstreamURLBasicAuthRegex = regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+.\-]*://)([^/\s:@]+):([^/\s@]+)@`)
	upstreamJWTRegex          = regexp.MustCompile(`\beyJ[A-Za-z0-9_\-]{10,}\.[A-Za-z0-9_\-]{10,}\.[A-Za-z0-9_\-]*`)
	upstreamSecretTokenRegex  = regexp.MustCompile(`\b(?:rt\.[A-Za-z0-9._\-]{16,}|sk-(?:ant-|proj-|svcacct-)?[A-Za-z0-9_\-]{12,}|AIza[0-9A-Za-z_\-]{20,}|authsess_[A-Za-z0-9_\-]{8,}|sess-[A-Za-z0-9_\-]{12,})`)
	upstreamAuth0SubRegex     = regexp.MustCompile(`\b(?:auth0|google-oauth2|windowslive|apple)\|[A-Za-z0-9._\-]+`)
	upstreamOrgCandidateRegex = regexp.MustCompile(`\borg-[A-Za-z0-9]{12,}\b`)
	upstreamPrjCandidateRegex = regexp.MustCompile(`\bproj_[A-Za-z0-9_\-]{12,}\b`)
	upstreamUsrCandidateRegex = regexp.MustCompile(`\buser-[A-Za-z0-9_\-]{16,}`)

	upstreamJSONSensitiveFieldRegex = regexp.MustCompile(
		`(?i)("(?:chatgpt_account_id|chatgpt_account_user_id|chatgpt_user_id|organization_id|anthropic_organization_id|account_email|user_email|parent_email)"\s*:\s*)"[^"]*"`,
	)
)

func hasUppercaseOrDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) || unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

// ContainsUpstreamAccountSensitiveInfo 检测文本中是否包含上游账号邮箱、令牌、组织/项目/用户 ID 或代理认证凭据。
func ContainsUpstreamAccountSensitiveInfo(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if upstreamEmailRegex.MatchString(s) ||
		upstreamEncodedEmailRegex.MatchString(s) ||
		upstreamURLBasicAuthRegex.MatchString(s) ||
		upstreamJWTRegex.MatchString(s) ||
		upstreamSecretTokenRegex.MatchString(s) ||
		upstreamAuth0SubRegex.MatchString(s) ||
		upstreamJSONSensitiveFieldRegex.MatchString(s) {
		return true
	}
	for _, m := range upstreamOrgCandidateRegex.FindAllString(s, -1) {
		if hasUppercaseOrDigit(strings.TrimPrefix(m, "org-")) {
			return true
		}
	}
	for _, m := range upstreamPrjCandidateRegex.FindAllString(s, -1) {
		if hasUppercaseOrDigit(strings.TrimPrefix(m, "proj_")) {
			return true
		}
	}
	for _, m := range upstreamUsrCandidateRegex.FindAllString(s, -1) {
		if hasUppercaseOrDigit(strings.TrimPrefix(m, "user-")) {
			return true
		}
	}
	return false
}

// RedactUpstreamSensitiveText 面向普通用户端 API 与网关错误响应的深度敏感信息清洗器：
// 彻底脱敏邮箱（含 URL/Unicode 编码）、代理认证串、JWT/Refresh/API Token、OpenAI/Anthropic 组织/项目/用户 ID 等。
func RedactUpstreamSensitiveText(s string) string {
	if s == "" {
		return s
	}
	out := upstreamURLBasicAuthRegex.ReplaceAllString(s, `${1}***:***@`)
	out = upstreamJSONSensitiveFieldRegex.ReplaceAllString(out, `${1}"[redacted]"`)
	out = upstreamEmailRegex.ReplaceAllString(out, "[redacted-email]")
	out = upstreamEncodedEmailRegex.ReplaceAllString(out, "[redacted-email]")
	out = upstreamJWTRegex.ReplaceAllString(out, "[redacted-token]")
	out = upstreamSecretTokenRegex.ReplaceAllString(out, "[redacted-token]")
	out = upstreamAuth0SubRegex.ReplaceAllString(out, "[redacted-user]")
	out = upstreamOrgCandidateRegex.ReplaceAllStringFunc(out, func(m string) string {
		if hasUppercaseOrDigit(strings.TrimPrefix(m, "org-")) {
			return "[redacted-org]"
		}
		return m
	})
	out = upstreamPrjCandidateRegex.ReplaceAllStringFunc(out, func(m string) string {
		if hasUppercaseOrDigit(strings.TrimPrefix(m, "proj_")) {
			return "[redacted-project]"
		}
		return m
	})
	out = upstreamUsrCandidateRegex.ReplaceAllStringFunc(out, func(m string) string {
		if hasUppercaseOrDigit(strings.TrimPrefix(m, "user-")) {
			return "[redacted-user]"
		}
		return m
	})
	return out
}

// RedactUpstreamSensitiveBytes 对上游原始错误响应字节流执行账号敏感标识清洗。
func RedactUpstreamSensitiveBytes(b []byte) []byte {
	if len(b) == 0 {
		return b
	}
	s := string(b)
	redacted := RedactUpstreamSensitiveText(s)
	if redacted == s {
		return b
	}
	return []byte(redacted)
}

func (a *Account) collectOwnSensitiveIdentifiers() []string {
	if a == nil {
		return nil
	}
	keys := []string{
		"email",
		"account_email",
		"chatgpt_account_id",
		"chatgpt_user_id",
		"chatgpt_account_user_id",
		"organization_id",
		"anthropic_organization_id",
		"project_id",
	}
	var out []string
	seen := make(map[string]struct{})
	addVal := func(v any) {
		str, ok := v.(string)
		if !ok {
			return
		}
		str = strings.TrimSpace(str)
		if len(str) < 6 {
			return
		}
		lower := strings.ToLower(str)
		if _, exists := seen[lower]; exists {
			return
		}
		seen[lower] = struct{}{}
		out = append(out, str)
	}
	for _, k := range keys {
		if len(a.Credentials) > 0 {
			addVal(a.Credentials[k])
		}
		if len(a.Extra) > 0 {
			addVal(a.Extra[k])
		}
	}
	return out
}

func (a *Account) containsOwnSensitiveIdentifier(text string) bool {
	if a == nil || strings.TrimSpace(text) == "" {
		return false
	}
	lowerText := strings.ToLower(text)
	for _, secret := range a.collectOwnSensitiveIdentifiers() {
		if strings.Contains(lowerText, strings.ToLower(secret)) {
			return true
		}
	}
	return false
}

// SanitizePublicText 对独享账号面向未买断用户的公开展示文本（如商城卖点说明）做双重脱敏。
func (a *Account) SanitizePublicText(text string) string {
	if text == "" {
		return ""
	}
	out := RedactUpstreamSensitiveText(text)
	if a != nil {
		for _, secret := range a.collectOwnSensitiveIdentifiers() {
			if secret != "" {
				ci := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(secret))
				out = ci.ReplaceAllString(out, "[redacted]")
			}
		}
	}
	return out
}

// ValidateAccountPublicText 校验管理员填写的公开别名与商城描述，阻止误粘贴账号邮箱、Token 或组织标识。
func ValidateAccountPublicText(acc *Account, displayAlias, saleDescription *string) error {
	for _, fieldPtr := range []*string{displayAlias, saleDescription} {
		if fieldPtr == nil {
			continue
		}
		val := strings.TrimSpace(*fieldPtr)
		if val == "" {
			continue
		}
		if ContainsUpstreamAccountSensitiveInfo(val) || (acc != nil && acc.containsOwnSensitiveIdentifier(val)) {
			return fmt.Errorf("%w", ErrDedicatedPublicSensitiveForbidden)
		}
	}
	return nil
}
