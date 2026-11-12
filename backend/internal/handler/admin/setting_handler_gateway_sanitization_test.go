package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSettingHandler_GatewaySanitizationRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tt := range []struct {
		name           string
		initial        map[string]string
		body           string
		storedTimezone string
		storedSanitize string
		wantTimezone   any
		wantSanitize   any
		status         int
	}{
		{
			name:           "defaults when absent",
			initial:        nil,
			body:           `{}`,
			storedTimezone: "",
			storedSanitize: "true",
			wantTimezone:   "",
			wantSanitize:   true,
			status:         http.StatusOK,
		},
		{
			name:           "set US timezone and disable accept-language sanitization",
			initial:        nil,
			body:           `{"client_dateline_timezone":"  America/Los_Angeles  ","enable_accept_language_sanitization":false}`,
			storedTimezone: "America/Los_Angeles",
			storedSanitize: "false",
			wantTimezone:   "America/Los_Angeles",
			wantSanitize:   false,
			status:         http.StatusOK,
		},
		{
			name: "omitted preserves existing values",
			initial: map[string]string{
				service.SettingKeyClientDatelineTimezone:           "America/New_York",
				service.SettingKeyEnableAcceptLanguageSanitization: "false",
			},
			body:           `{}`,
			storedTimezone: "America/New_York",
			storedSanitize: "false",
			wantTimezone:   "America/New_York",
			wantSanitize:   false,
			status:         http.StatusOK,
		},
		{
			name: "clear timezone back to empty string",
			initial: map[string]string{
				service.SettingKeyClientDatelineTimezone: "America/Los_Angeles",
			},
			body:           `{"client_dateline_timezone":"   "}`,
			storedTimezone: "",
			storedSanitize: "true",
			wantTimezone:   "",
			wantSanitize:   true,
			status:         http.StatusOK,
		},
		{
			name:    "rejects Local timezone",
			initial: nil,
			body:    `{"client_dateline_timezone":"Local"}`,
			status:  http.StatusBadRequest,
		},
		{
			name:    "rejects invalid IANA timezone",
			initial: nil,
			body:    `{"client_dateline_timezone":"Mars/Olympus_Mons"}`,
			status:  http.StatusBadRequest,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			repo := &settingHandlerRepoStub{values: tt.initial}
			svc := service.NewSettingService(repo, &config.Config{})
			handler := NewSettingHandler(svc, nil, nil, nil, nil, nil, nil)

			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			handler.UpdateSettings(c)

			require.Equal(t, tt.status, rec.Code, rec.Body.String())
			if tt.status != http.StatusOK {
				require.Empty(t, repo.lastUpdates)
				return
			}

			require.Equal(t, tt.storedTimezone, repo.values[service.SettingKeyClientDatelineTimezone])
			require.Equal(t, tt.storedSanitize, repo.values[service.SettingKeyEnableAcceptLanguageSanitization])
			require.Equal(t, tt.wantTimezone, svc.GetClientDatelineTimezone(context.Background()))
			require.Equal(t, tt.wantSanitize, svc.IsAcceptLanguageSanitizationEnabled(context.Background()))

			var body struct {
				Data map[string]any `json:"data"`
			}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			require.Equal(t, tt.wantTimezone, body.Data[service.SettingKeyClientDatelineTimezone])
			require.Equal(t, tt.wantSanitize, body.Data[service.SettingKeyEnableAcceptLanguageSanitization])

			getRec := httptest.NewRecorder()
			getContext, _ := gin.CreateTestContext(getRec)
			getContext.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/settings", nil)
			handler.GetSettings(getContext)
			require.Equal(t, http.StatusOK, getRec.Code, getRec.Body.String())
			require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &body))
			require.Equal(t, tt.wantTimezone, body.Data[service.SettingKeyClientDatelineTimezone])
			require.Equal(t, tt.wantSanitize, body.Data[service.SettingKeyEnableAcceptLanguageSanitization])
		})
	}
}

func TestSettingsAuditTracksGatewaySanitizationFields(t *testing.T) {
	before := &service.SystemSettings{
		ClientDatelineTimezone:           "",
		EnableAcceptLanguageSanitization: true,
	}
	after := &service.SystemSettings{
		ClientDatelineTimezone:           "",
		EnableAcceptLanguageSanitization: true,
	}
	require.NotContains(t, diffSettings(before, after, nil, nil, UpdateSettingsRequest{}), service.SettingKeyClientDatelineTimezone)
	require.NotContains(t, diffSettings(before, after, nil, nil, UpdateSettingsRequest{}), service.SettingKeyEnableAcceptLanguageSanitization)

	after.ClientDatelineTimezone = "America/Los_Angeles"
	after.EnableAcceptLanguageSanitization = false
	diff := diffSettings(before, after, nil, nil, UpdateSettingsRequest{})
	require.Contains(t, diff, service.SettingKeyClientDatelineTimezone)
	require.Contains(t, diff, service.SettingKeyEnableAcceptLanguageSanitization)
}
