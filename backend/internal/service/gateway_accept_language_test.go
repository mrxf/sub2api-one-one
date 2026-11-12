package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGatewayService_AcceptLanguageSanitization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newCtxWithHeader := func(key, val string) *gin.Context {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
		if key != "" {
			c.Request.Header[key] = []string{val}
		}
		return c
	}

	t.Run("OAuth and SetupToken sanitize regional Accept-Language to * on messages and count_tokens by default", func(t *testing.T) {
		repo := &gatewayTTLSettingRepo{data: map[string]string{}}
		gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{})
		svc := &GatewayService{
			cfg:            &config.Config{},
			settingService: NewSettingService(repo, &config.Config{}),
		}

		for _, accountType := range []string{AccountTypeOAuth, AccountTypeSetupToken} {
			account := &Account{Platform: PlatformAnthropic, Type: accountType}
			for _, mimic := range []bool{false, true} {
				// Client sends regional zh-CN locale
				c := newCtxWithHeader("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
				req, _, err := svc.buildUpstreamRequest(
					context.Background(), c, account,
					[]byte(`{"model":"claude-sonnet-4-6","messages":[]}`),
					"oauth-token", "oauth", "claude-sonnet-4-6", true, mimic,
				)
				require.NoError(t, err)
				require.Equal(t, []string{"*"}, req.Header["accept-language"], "accountType=%s mimic=%v", accountType, mimic)
				require.Empty(t, req.Header["Accept-Language"], "canonical Accept-Language must not coexist with wire-cased key")

				// CountTokens path (which passes allowedHeaders even when mimic=true)
				ctCtx := newCtxWithHeader("Accept-Language", "zh-CN,zh;q=0.9")
				ctReq, _, err := svc.buildCountTokensRequest(
					context.Background(), ctCtx, account,
					[]byte(`{"model":"claude-sonnet-4-6","messages":[]}`),
					"oauth-token", "oauth", "claude-sonnet-4-6", mimic,
				)
				require.NoError(t, err)
				require.Equal(t, []string{"*"}, ctReq.Header["accept-language"], "count_tokens accountType=%s mimic=%v", accountType, mimic)
				require.Empty(t, ctReq.Header["Accept-Language"])
			}

			// Omitted Accept-Language on OAuth also normalizes to "*" (matching Node.js/undici Claude Code CLI)
			cOmitted := newCtxWithHeader("", "")
			reqOmitted, _, err := svc.buildUpstreamRequest(
				context.Background(), cOmitted, account,
				[]byte(`{"model":"claude-sonnet-4-6","messages":[]}`),
				"oauth-token", "oauth", "claude-sonnet-4-6", true, true,
			)
			require.NoError(t, err)
			require.Equal(t, []string{"*"}, reqOmitted.Header["accept-language"])
		}
	})

	t.Run("Switch disabled preserves client Accept-Language on passthrough", func(t *testing.T) {
		repo := &gatewayTTLSettingRepo{data: map[string]string{
			SettingKeyEnableAcceptLanguageSanitization: "false",
		}}
		gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{})
		svc := &GatewayService{
			cfg:            &config.Config{},
			settingService: NewSettingService(repo, &config.Config{}),
		}
		account := &Account{Platform: PlatformAnthropic, Type: AccountTypeOAuth}

		c := newCtxWithHeader("Accept-Language", "zh-CN,zh;q=0.9")
		req, _, err := svc.buildUpstreamRequest(
			context.Background(), c, account,
			[]byte(`{"model":"claude-sonnet-4-6","messages":[]}`),
			"oauth-token", "oauth", "claude-sonnet-4-6", true, false,
		)
		require.NoError(t, err)
		require.Equal(t, "zh-CN,zh;q=0.9", getHeaderRaw(req.Header, "accept-language"))

		ctCtx := newCtxWithHeader("Accept-Language", "zh-CN,zh;q=0.9")
		ctReq, _, err := svc.buildCountTokensRequest(
			context.Background(), ctCtx, account,
			[]byte(`{"model":"claude-sonnet-4-6","messages":[]}`),
			"oauth-token", "oauth", "claude-sonnet-4-6", false,
		)
		require.NoError(t, err)
		require.Equal(t, "zh-CN,zh;q=0.9", getHeaderRaw(ctReq.Header, "accept-language"))
	})

	t.Run("APIKey accounts never rewrite Accept-Language even when switch is enabled", func(t *testing.T) {
		repo := &gatewayTTLSettingRepo{data: map[string]string{
			SettingKeyEnableAcceptLanguageSanitization: "true",
		}}
		gatewayForwardingCache.Store(&cachedGatewayForwardingSettings{})
		svc := &GatewayService{
			cfg:            &config.Config{},
			settingService: NewSettingService(repo, &config.Config{}),
		}
		account := &Account{Platform: PlatformAnthropic, Type: AccountTypeAPIKey}

		c := newCtxWithHeader("Accept-Language", "zh-CN,zh;q=0.9")
		req, _, err := svc.buildUpstreamRequest(
			context.Background(), c, account,
			[]byte(`{"model":"claude-sonnet-4-6","messages":[]}`),
			"sk-ant-api-key", "apikey", "claude-sonnet-4-6", true, false,
		)
		require.NoError(t, err)
		require.Equal(t, "zh-CN,zh;q=0.9", getHeaderRaw(req.Header, "accept-language"))

		cOmitted := newCtxWithHeader("", "")
		reqOmitted, _, err := svc.buildUpstreamRequest(
			context.Background(), cOmitted, account,
			[]byte(`{"model":"claude-sonnet-4-6","messages":[]}`),
			"sk-ant-api-key", "apikey", "claude-sonnet-4-6", true, false,
		)
		require.NoError(t, err)
		require.Equal(t, "", getHeaderRaw(reqOmitted.Header, "accept-language"))
	})
}
