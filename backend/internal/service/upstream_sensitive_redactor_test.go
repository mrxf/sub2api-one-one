//go:build unit

package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestUpstreamSensitiveRedactor_BlocksEmailsTokensAndAccountIdentifiers(t *testing.T) {
	rawLeak := `Upstream 429 for account 2车-poolmember02@example.com (encoded: poolmember02%40example.com, unicode: poolmember02\u0040example.com), ` +
		`org-SyntheticOrg0123456789AB, proj_9876543210abcdef, user-SyntheticUsr0123456789AB, auth0|SyntheticSub0123456789, ` +
		`"chatgpt_account_id":"00000000-1111-4222-8333-444444444444", proxy http://proxyuser:secretpass@198.51.100.8:8080, ` +
		`jwt eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJhdXRoMHxTeW50aGV0aWNTdWIwMTIzNDU2Nzg5In0.c3ludGhldGljLXNpZ25hdHVyZTEyMzQ1Ng, ` +
		`refresh rt.1.SyntheticRefresh0123456789abcdef_synthetic, key sk-proj-abcdefghijklmnopqrstuvwxyz123456`

	require.True(t, ContainsUpstreamAccountSensitiveInfo(rawLeak))

	redacted := RedactUpstreamSensitiveText(rawLeak)
	require.NotContains(t, redacted, "poolmember02")
	require.NotContains(t, redacted, "example.com")
	require.NotContains(t, redacted, "SyntheticOrg0123456789AB")
	require.NotContains(t, redacted, "proj_9876543210abcdef")
	require.NotContains(t, redacted, "SyntheticUsr0123456789AB")
	require.NotContains(t, redacted, "SyntheticSub0123456789")
	require.NotContains(t, redacted, "00000000-1111-4222-8333-444444444444")
	require.NotContains(t, redacted, "proxyuser")
	require.NotContains(t, redacted, "secretpass")
	require.NotContains(t, redacted, "eyJhbGci")
	require.NotContains(t, redacted, "SyntheticRefresh0123456789abcdef")
	require.NotContains(t, redacted, "sk-proj-")

	// Safe non-sensitive model/error messages must remain untouched
	safeMsg := "Rate limit reached for model gpt-5-codex: please retry after 30s (org-chart and user-friendly docs unaffected)"
	require.False(t, ContainsUpstreamAccountSensitiveInfo(safeMsg))
	require.Equal(t, safeMsg, RedactUpstreamSensitiveText(safeMsg))
}

func TestAccountPublicIsolationAndValidation(t *testing.T) {
	acc := &Account{
		ID:       77,
		Name:     "2车-poolmember02@example.com",
		Platform: PlatformOpenAI,
		Credentials: map[string]any{
			"email":              "poolmember02@example.com",
			"chatgpt_account_id": "00000000-1111-4222-8333-444444444444",
			"chatgpt_user_id":    "user-SyntheticUsr0123456789AB",
			"organization_id":    "org-SyntheticOrg0123456789AB",
		},
		Extra: map[string]any{
			"email": "poolmember02@example.com",
		},
	}

	// If admin accidentally sets DisplayAlias to email or internal name, PublicDisplayName falls back to safe alias
	acc.DisplayAlias = strPtrDedicatedTest("2车-poolmember02@example.com")
	require.Equal(t, "OpenAI 独享 #77", acc.PublicDisplayName())

	acc.DisplayAlias = strPtrDedicatedTest("00000000-1111-4222-8333-444444444444")
	require.Equal(t, "OpenAI 独享 #77", acc.PublicDisplayName())

	acc.DisplayAlias = strPtrDedicatedTest("OpenAI ProLite 独享专线 02号")
	require.Equal(t, "OpenAI ProLite 独享专线 02号", acc.PublicDisplayName())

	// Account.SanitizePublicText scrubs internal account name and account-specific UUIDs
	leakedDesc := "账号 2车-poolmember02@example.com (workspace 00000000-1111-4222-8333-444444444444) 满血可跑"
	sanitizedDesc := acc.SanitizePublicText(leakedDesc)
	require.NotContains(t, sanitizedDesc, "poolmember02")
	require.NotContains(t, sanitizedDesc, "00000000-1111-4222-8333-444444444444")

	// Admin input validation blocks sensitive identifiers in public alias / sale description
	require.ErrorIs(t, ValidateAccountPublicText(acc, strPtrDedicatedTest("poolmember02@example.com"), nil), ErrDedicatedPublicSensitiveForbidden)
	require.ErrorIs(t, ValidateAccountPublicText(acc, nil, strPtrDedicatedTest("org-SyntheticOrg0123456789AB 独享")), ErrDedicatedPublicSensitiveForbidden)
	require.ErrorIs(t, ValidateAccountPublicText(acc, nil, strPtrDedicatedTest("00000000-1111-4222-8333-444444444444")), ErrDedicatedPublicSensitiveForbidden)
	require.NoError(t, ValidateAccountPublicText(acc, strPtrDedicatedTest("Claude Max 20x 独享 01号"), strPtrDedicatedTest("全新开通 30 天官方原生订阅")))

	// Buyout config validation
	require.NoError(t, ValidateAccountBuyoutConfig(0, nil))
	require.NoError(t, ValidateAccountBuyoutConfig(580, strPtrDedicatedTest("邮箱: poolmember02@example.com\n密码: MySecretPassword\n2FA: JBSWY3DPEHPK3PXP")))
	require.ErrorIs(t, ValidateAccountBuyoutConfig(-1, strPtrDedicatedTest("info")), ErrDedicatedSaleConfigInvalid)
	require.ErrorIs(t, ValidateAccountBuyoutConfig(580, strPtrDedicatedTest("   ")), ErrDedicatedBuyoutInfoRequired)
	require.ErrorIs(t, ValidateAccountBuyoutConfig(580, nil), ErrDedicatedBuyoutInfoRequired)
}

func TestUserErrorPathsScrubSensitiveIdentifiers(t *testing.T) {
	acc := &Account{
		ID:       88,
		Name:     "2车-poolmember02@example.com",
		Platform: PlatformOpenAI,
		Credentials: map[string]any{
			"chatgpt_account_id": "00000000-1111-4222-8333-444444444444",
		},
	}

	// 1. Ops user error detail
	opsLog := &OpsErrorLogDetail{
		OpsErrorLog: OpsErrorLog{
			ID:         1,
			StatusCode: 429,
			Message:    "Account poolmember02@example.com in org-SyntheticOrg0123456789AB exceeded quota",
			APIKeyName: "key-for-poolmember02@example.com",
		},
		ErrorBody: `{"error":{"message":"user-SyntheticUsr0123456789AB rate limited"}}`,
	}
	detail := ToUserErrorRequestDetail(opsLog)
	require.NotNil(t, detail)
	require.NotContains(t, detail.Message, "poolmember02")
	require.NotContains(t, detail.Message, "SyntheticOrg0123456789AB")
	require.NotContains(t, detail.KeyName, "poolmember02")
	require.NotContains(t, detail.ErrorBody, "SyntheticUsr0123456789AB")

	// 2. Upstream error message extractor + sanitizer
	rawBody := []byte(`{"error":{"message":"Quota exceeded for poolmember02@example.com in org-SyntheticOrg0123456789AB"}}`)
	sanitizedUpstream := SanitizeUpstreamErrorMessage(ExtractUpstreamErrorMessage(rawBody))
	require.NotContains(t, sanitizedUpstream, "poolmember02")
	require.NotContains(t, sanitizedUpstream, "SyntheticOrg0123456789AB")

	// 3. Dedicated reset credit error sanitizer
	appErr := infraerrors.BadRequest("UPSTREAM_ERROR", "Cannot redeem credit for 2车-poolmember02@example.com (00000000-1111-4222-8333-444444444444)")
	safeErr := sanitizeDedicatedResetErrorForUser(acc, appErr)
	require.NotContains(t, safeErr.Error(), "poolmember02")
	require.NotContains(t, safeErr.Error(), "00000000-1111-4222-8333-444444444444")

	rawUnexpectedErr := errors.New("dial tcp http://u:p@1.2.3.4:8080 failed for poolmember02@example.com")
	shieldedErr := sanitizeDedicatedResetErrorForUser(acc, rawUnexpectedErr)
	require.Equal(t, 502, infraerrors.Code(shieldedErr))
	require.NotContains(t, shieldedErr.Error(), "poolmember02")
	require.NotContains(t, shieldedErr.Error(), "1.2.3.4")
}

func (r *memDedicatedAccountRepo) ListMarketAccounts(_ context.Context, platform string) ([]Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []Account
	for _, acc := range r.accounts {
		if acc.SaleStatus == AccountSaleStatusListed && acc.OwnerUserID == nil {
			if platform == "" || acc.Platform == platform {
				out = append(out, *acc)
			}
		}
	}
	return out, nil
}

func (r *memDedicatedAccountRepo) BuyoutWithBalance(_ context.Context, buyerUserID int64, accountID int64, balanceMultiplier float64, generateRawKey func() (string, error)) (*DedicatedBalanceBuyoutOutcome, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	acc, ok := r.accounts[accountID]
	if !ok {
		return nil, ErrAccountNotFound
	}
	if balanceMultiplier <= 0 {
		balanceMultiplier = 1
	}
	if acc.BoughtOutAt != nil {
		return nil, ErrDedicatedAlreadyBoughtOut
	}
	if !acc.IsBuyoutEnabled() {
		return nil, ErrDedicatedBuyoutUnavailable
	}
	now := time.Now().UTC()
	buyoutText := ""
	if acc.BuyoutInfo != nil {
		buyoutText = *acc.BuyoutInfo
	}
	if acc.SaleStatus == AccountSaleStatusSold && acc.OwnerUserID != nil {
		if *acc.OwnerUserID != buyerUserID {
			return nil, ErrDedicatedAccountForbidden
		}
		acc.BoughtOutAt = &now
		cp := *acc
		return &DedicatedBalanceBuyoutOutcome{
			Account:          &cp,
			BuyoutInfo:       buyoutText,
			PriceCNY:         acc.BuyoutPrice,
			BalanceDeducted:  acc.BuyoutPrice / balanceMultiplier,
			RemainingBalance: 100.0,
		}, nil
	}
	if acc.SaleStatus != AccountSaleStatusListed || acc.OwnerUserID != nil {
		return nil, ErrDedicatedAccountUnavailable
	}
	rawKey := "sk-dedicated-buyout-key"
	if generateRawKey != nil {
		if k, err := generateRawKey(); err == nil {
			rawKey = k
		}
	}
	acc.SaleStatus = AccountSaleStatusSold
	acc.OwnerUserID = &buyerUserID
	acc.SoldAt = &now
	acc.BoughtOutAt = &now
	cp := *acc
	return &DedicatedBalanceBuyoutOutcome{
		Account: &cp,
		APIKey: &APIKey{
			ID:                 901,
			UserID:             buyerUserID,
			Name:               cp.PublicDisplayName(),
			Key:                rawKey,
			Status:             StatusActive,
			DedicatedAccountID: &cp.ID,
		},
		BuyoutInfo:       buyoutText,
		PriceCNY:         acc.BuyoutPrice,
		BalanceDeducted:  acc.BuyoutPrice / balanceMultiplier,
		RemainingBalance: 50.0,
	}, nil
}

func TestDedicatedAccountService_BuyoutFlowAndStrictIsolation(t *testing.T) {
	ctx := context.Background()
	ownerUser := int64(501)
	shelfBuyer := int64(502)
	attackerUser := int64(999)
	now := time.Now().UTC()

	rawBuyoutCredentials := "登录邮箱: poolmember02@example.com\n登录密码: P@ssw0rd_2026!\n2FA密钥: JBSWY3DPEHPK3PXP\nRefreshToken: rt.1.SyntheticRefresh0123456789abcdef_synthetic"
	shelfBuyoutCredentials := "Claude邮箱: shelfsecret@example.com\n密码: ClaudeBuyoutSecret!"

	ownedDedicated := &Account{
		ID:          10,
		Name:        "2车-poolmember02@example.com",
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Status:      StatusActive,
		Schedulable: true,
		SaleStatus:  AccountSaleStatusSold,
		OwnerUserID: int64PtrDedicatedTest(ownerUser),
		SoldAt:      &now,
		BuyoutPrice: 580.00,
		BuyoutInfo:  strPtrDedicatedTest(rawBuyoutCredentials),
		GroupIDs:    []int64{7},
		Credentials: map[string]any{
			"email":              "poolmember02@example.com",
			"chatgpt_account_id": "00000000-1111-4222-8333-444444444444",
			"organization_id":    "org-SyntheticOrg0123456789AB",
		},
		Extra: map[string]any{
			"email":                 "poolmember02@example.com",
			"codex_5h_used_percent": 15.0,
			"codex_7d_used_percent": 66.0,
		},
	}

	shelfDedicated := &Account{
		ID:              20,
		Name:            "3车-shelfsecret@example.com",
		Platform:        PlatformAnthropic,
		Type:            AccountTypeOAuth,
		Status:          StatusActive,
		Schedulable:     true,
		SaleStatus:      AccountSaleStatusListed,
		DisplayAlias:    strPtrDedicatedTest("Claude Max 20x 独享 03号"),
		SaleDescription: strPtrDedicatedTest("含隐藏邮箱 shelfsecret@example.com 也会被自动脱敏"),
		SalePrice:       200.00,
		ListedAt:        &now,
		BuyoutPrice:     699.00,
		BuyoutInfo:      strPtrDedicatedTest(shelfBuyoutCredentials),
		GroupIDs:        []int64{8},
		Credentials: map[string]any{
			"email": "shelfsecret@example.com",
		},
	}

	repo := newMemDedicatedAccountRepo(ownedDedicated, shelfDedicated)
	svc := NewDedicatedAccountService(
		repo,
		&memAccountRepoForDedicated{backing: repo},
		&memAPIKeyRepoForDedicated{},
		nil, nil, nil, nil, nil, nil, nil,
	)

	// 1. Marketplace payload MUST expose buyout_available & buyout_price_cny, and MUST NEVER leak emails or buyout_info!
	marketItems, err := svc.ListMarketAccounts(ctx, "")
	require.NoError(t, err)
	require.Len(t, marketItems, 1)
	require.Equal(t, int64(20), marketItems[0].ID)
	require.True(t, marketItems[0].BuyoutAvailable)
	require.InDelta(t, 699.00, marketItems[0].BuyoutPriceCNY, 1e-6)

	marketJSON, err := json.Marshal(marketItems)
	require.NoError(t, err)
	require.NotContains(t, string(marketJSON), "shelfsecret@example.com")
	require.NotContains(t, string(marketJSON), "ClaudeBuyoutSecret")
	require.NotContains(t, string(marketJSON), "buyout_info")

	// 2. Owner's dashboard list BEFORE buyout MUST NOT leak email, org, account UUID, or buyout_info!
	myAccountsBefore, err := svc.ListUserDedicatedAccounts(ctx, ownerUser)
	require.NoError(t, err)
	require.Len(t, myAccountsBefore, 1)
	require.Equal(t, "OpenAI 独享 #10", myAccountsBefore[0].DisplayName)
	require.False(t, myAccountsBefore[0].BoughtOut)
	require.True(t, myAccountsBefore[0].BuyoutAvailable)
	require.InDelta(t, 580.00, myAccountsBefore[0].BuyoutPriceCNY, 1e-6)

	myJSONBefore, err := json.Marshal(myAccountsBefore)
	require.NoError(t, err)
	require.NotContains(t, string(myJSONBefore), "poolmember02")
	require.NotContains(t, string(myJSONBefore), "P@ssw0rd_2026")
	require.NotContains(t, string(myJSONBefore), "buyout_info")

	// 3. Before paying buyout, even the owner CANNOT fetch buyout-info!
	_, err = svc.GetUserBoughtOutAccountInfo(ctx, ownerUser, ownedDedicated.ID)
	require.Error(t, err)
	require.Equal(t, 403, infraerrors.Code(err))
	require.Equal(t, "DEDICATED_NOT_BOUGHT_OUT", infraerrors.Reason(err))

	// 4. Attacker trying to buy out ownerUser's dedicated account -> Forbidden!
	_, err = svc.BuyoutWithBalance(ctx, attackerUser, ownedDedicated.ID)
	require.ErrorIs(t, err, ErrDedicatedAccountForbidden)

	// 5. Owner buys out their owned dedicated account with balance -> succeeds and returns exact plaintext buyout_info!
	buyoutRes, err := svc.BuyoutWithBalance(ctx, ownerUser, ownedDedicated.ID)
	require.NoError(t, err)
	require.NotNil(t, buyoutRes)
	require.True(t, buyoutRes.Account.BoughtOut)
	require.False(t, buyoutRes.Account.BuyoutAvailable, "already bought-out account cannot be bought out again")
	require.Equal(t, rawBuyoutCredentials, buyoutRes.BuyoutInfo)

	// 6. Duplicate buyout attempt -> rejected with ErrDedicatedAlreadyBoughtOut!
	_, err = svc.BuyoutWithBalance(ctx, ownerUser, ownedDedicated.ID)
	require.ErrorIs(t, err, ErrDedicatedAlreadyBoughtOut)

	// 7. After buyout: verified owner CAN fetch GetUserBoughtOutAccountInfo, while attacker STILL gets Forbidden,
	//    and ListUserDedicatedAccounts STILL keeps buyout_info out of routine list responses!
	infoResp, err := svc.GetUserBoughtOutAccountInfo(ctx, ownerUser, ownedDedicated.ID)
	require.NoError(t, err)
	require.Equal(t, rawBuyoutCredentials, infoResp.BuyoutInfo)

	_, err = svc.GetUserBoughtOutAccountInfo(ctx, attackerUser, ownedDedicated.ID)
	require.ErrorIs(t, err, ErrDedicatedAccountForbidden)

	myAccountsAfter, err := svc.ListUserDedicatedAccounts(ctx, ownerUser)
	require.NoError(t, err)
	require.Len(t, myAccountsAfter, 1)
	require.True(t, myAccountsAfter[0].BoughtOut)
	myJSONAfter, err := json.Marshal(myAccountsAfter)
	require.NoError(t, err)
	require.NotContains(t, string(myJSONAfter), "P@ssw0rd_2026")

	// 8. Direct shelf buyout by shelfBuyer -> atomically assigns ownership + creates 1:1 API key + marks BoughtOut + returns buyout_info!
	shelfBuyoutRes, err := svc.BuyoutWithBalance(ctx, shelfBuyer, shelfDedicated.ID)
	require.NoError(t, err)
	require.True(t, shelfBuyoutRes.Account.BoughtOut)
	require.NotEmpty(t, shelfBuyoutRes.APIKey)
	require.Contains(t, shelfBuyoutRes.BuyoutInfo, "shelfsecret@example.com")
}
