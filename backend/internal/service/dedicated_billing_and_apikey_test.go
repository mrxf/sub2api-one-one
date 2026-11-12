//go:build unit

package service

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type memDedicatedAccountRepo struct {
	DedicatedAccountRepository
	mu       sync.Mutex
	accounts map[int64]*Account
}

func newMemDedicatedAccountRepo(items ...*Account) *memDedicatedAccountRepo {
	m := make(map[int64]*Account, len(items))
	for _, item := range items {
		cp := *item
		if item.Extra != nil {
			extraCopy := make(map[string]any, len(item.Extra))
			for k, v := range item.Extra {
				extraCopy[k] = v
			}
			cp.Extra = extraCopy
		}
		m[item.ID] = &cp
	}
	return &memDedicatedAccountRepo{accounts: m}
}

func (r *memDedicatedAccountRepo) ReleaseExpiredLocks(_ context.Context, _ time.Time) (int, error) {
	return 0, nil
}

func (r *memDedicatedAccountRepo) GetByID(_ context.Context, id int64) (*Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	acc, ok := r.accounts[id]
	if !ok {
		return nil, ErrAccountNotFound
	}
	cp := *acc
	return &cp, nil
}

func (r *memDedicatedAccountRepo) ListUserDedicatedAccounts(_ context.Context, userID int64) ([]Account, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []Account
	for _, acc := range r.accounts {
		if acc.OwnerUserID != nil && *acc.OwnerUserID == userID && acc.ParentAccountID == nil {
			out = append(out, *acc)
		}
	}
	return out, nil
}

func (r *memDedicatedAccountRepo) EnsureAccountHasGroup(_ context.Context, accountID int64, _ string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if acc, ok := r.accounts[accountID]; ok {
		if len(acc.GroupIDs) > 0 {
			return acc.GroupIDs[0], nil
		}
		acc.GroupIDs = []int64{99}
		return 99, nil
	}
	return 0, ErrAccountNotFound
}

type memAccountRepoForDedicated struct {
	AccountRepository
	backing *memDedicatedAccountRepo
}

func (r *memAccountRepoForDedicated) GetByID(ctx context.Context, id int64) (*Account, error) {
	return r.backing.GetByID(ctx, id)
}

func (r *memAccountRepoForDedicated) GetByIDs(ctx context.Context, ids []int64) ([]*Account, error) {
	out := make([]*Account, 0, len(ids))
	for _, id := range ids {
		if acc, err := r.backing.GetByID(ctx, id); err == nil && acc != nil {
			out = append(out, acc)
		}
	}
	return out, nil
}

func (r *memAccountRepoForDedicated) UpdateExtra(_ context.Context, id int64, updates map[string]any) error {
	r.backing.mu.Lock()
	defer r.backing.mu.Unlock()
	acc, ok := r.backing.accounts[id]
	if !ok {
		return ErrAccountNotFound
	}
	if acc.Extra == nil {
		acc.Extra = make(map[string]any)
	}
	for k, v := range updates {
		acc.Extra[k] = v
	}
	return nil
}

func (r *memAccountRepoForDedicated) UpdateSessionWindow(_ context.Context, _ int64, _, _ *time.Time, _ string) error {
	return nil
}

type memAPIKeyRepoForDedicated struct {
	APIKeyRepository
	keys []APIKey
}

func (r *memAPIKeyRepoForDedicated) ListByUserID(_ context.Context, userID int64, _ pagination.PaginationParams, _ APIKeyListFilters) ([]APIKey, *pagination.PaginationResult, error) {
	var out []APIKey
	for _, k := range r.keys {
		if k.UserID == userID {
			out = append(out, k)
		}
	}
	return out, &pagination.PaginationResult{Total: int64(len(out)), Page: 1, PageSize: 200, Pages: 1}, nil
}

func TestDedicatedAccountService_ResolveUserDedicatedAccountGroup_SecurityGuards(t *testing.T) {
	user1 := int64(101)
	user2 := int64(202)
	otherUser := int64(303)

	ownedAcc := &Account{
		ID:          1,
		Platform:    PlatformAnthropic,
		Status:      StatusActive,
		SaleStatus:  AccountSaleStatusSold,
		OwnerUserID: int64PtrDedicatedTest(user1),
		GroupIDs:    []int64{7},
	}
	otherOwnedAcc := &Account{
		ID:          2,
		Platform:    PlatformAnthropic,
		Status:      StatusActive,
		SaleStatus:  AccountSaleStatusSold,
		OwnerUserID: int64PtrDedicatedTest(user2),
		GroupIDs:    []int64{7},
	}
	shelfAcc := &Account{
		ID:         3,
		Platform:   PlatformAnthropic,
		Status:     StatusActive,
		SaleStatus: AccountSaleStatusListed,
		SalePrice:  200,
		GroupIDs:   []int64{7},
	}
	shadowAcc := &Account{
		ID:              4,
		Platform:        PlatformOpenAI,
		Status:          StatusActive,
		SaleStatus:      AccountSaleStatusSold,
		OwnerUserID:     int64PtrDedicatedTest(user1),
		ParentAccountID: int64PtrDedicatedTest(1),
	}

	repo := newMemDedicatedAccountRepo(ownedAcc, otherOwnedAcc, shelfAcc, shadowAcc)
	svc := NewDedicatedAccountService(repo, &memAccountRepoForDedicated{backing: repo}, nil, nil, nil, nil, nil, nil, nil, nil)

	// 1. Owner binding their own dedicated account succeeds and resolves group 7
	gid, err := svc.ResolveUserDedicatedAccountGroup(context.Background(), user1, int64PtrDedicatedTest(ownedAcc.ID), false, nil)
	require.NoError(t, err)
	require.NotNil(t, gid)
	require.Equal(t, int64(7), *gid)

	// 2. Privilege escalation: User 1 trying to bind User 2's dedicated account -> Forbidden
	_, err = svc.ResolveUserDedicatedAccountGroup(context.Background(), user1, int64PtrDedicatedTest(otherOwnedAcc.ID), false, nil)
	require.ErrorIs(t, err, ErrDedicatedAccountForbidden)

	// 3. Privilege escalation: User 1 trying to bind an unpurchased shelf account -> Forbidden
	_, err = svc.ResolveUserDedicatedAccountGroup(context.Background(), user1, int64PtrDedicatedTest(shelfAcc.ID), false, nil)
	require.ErrorIs(t, err, ErrDedicatedAccountForbidden)

	// 4. Binding a shadow child account directly -> Forbidden
	_, err = svc.ResolveUserDedicatedAccountGroup(context.Background(), user1, int64PtrDedicatedTest(shadowAcc.ID), false, nil)
	require.ErrorIs(t, err, ErrDedicatedAccountForbidden)

	// 5. User with 0 dedicated accounts trying to enable dedicated_pool_mode -> Forbidden (blocks zero-balance exploit!)
	_, err = svc.ResolveUserDedicatedAccountGroup(context.Background(), otherUser, nil, true, int64PtrDedicatedTest(7))
	require.ErrorIs(t, err, ErrDedicatedAccountForbidden)

	// 6. User 1 with owned dedicated account in group 7 enabling dedicated_pool_mode -> succeeds
	poolGroupID, err := svc.ResolveUserDedicatedAccountGroup(context.Background(), user1, nil, true, int64PtrDedicatedTest(7))
	require.NoError(t, err)
	require.NotNil(t, poolGroupID)
	require.Equal(t, int64(7), *poolGroupID)
}

func TestDedicatedAccountService_ListMyAccounts_StrictReadOnlyRedactionAndSyntheticResetCredits(t *testing.T) {
	user1 := int64(501)
	user2 := int64(502)
	now := time.Now().UTC()
	exp := now.Add(28 * 24 * time.Hour)
	internalNote := "SUPPLIER_INTERNAL_SECRET_NOTE_DO_NOT_LEAK"

	openAIDedicated := &Account{
		ID:           55,
		Name:         "supplier-raw-email-secret@openai.example",
		DisplayAlias: strPtrDedicatedTest("OpenAI Codex Pro 独享 #55"),
		Notes:        &internalNote,
		Platform:     PlatformOpenAI,
		Type:         AccountTypeOAuth,
		Status:       StatusActive,
		Schedulable:  true,
		Concurrency:  10,
		SaleStatus:   AccountSaleStatusSold,
		OwnerUserID:  int64PtrDedicatedTest(user1),
		ExpiresAt:    &exp,
		ProxyID:      int64PtrDedicatedTest(888),
		GroupIDs:     []int64{12},
		Credentials: map[string]any{
			"access_token":  "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.TOP_SECRET_TOKEN",
			"refresh_token": "rt_TOP_SECRET_REFRESH_TOKEN",
		},
		Extra: map[string]any{
			"synthetic_ui_test":     true,
			"codex_5h_used_percent": float64(82),
			"codex_5h_reset_at":     now.Add(3 * time.Hour).Format(time.RFC3339),
			"codex_7d_used_percent": float64(25),
			"codex_7d_reset_at":     now.Add(5 * 24 * time.Hour).Format(time.RFC3339),
			openaiQuotaResetCreditsKey: OpenAIRateLimitResetCredits{
				AvailableCount: 1,
				Credits: []OpenAIRateLimitResetCreditDetail{
					{ExpiresAt: now.Add(15 * 24 * time.Hour).Format(time.RFC3339)},
				},
			},
		},
	}

	repo := newMemDedicatedAccountRepo(openAIDedicated)
	accRepo := &memAccountRepoForDedicated{backing: repo}
	keyRepo := &memAPIKeyRepoForDedicated{
		keys: []APIKey{
			{
				ID:                 901,
				UserID:             user1,
				Name:               "独享 · OpenAI Codex Pro 独享 #55",
				Key:                "sk-dedicated-pinned-55",
				Status:             StatusAPIKeyActive,
				GroupID:            int64PtrDedicatedTest(12),
				DedicatedAccountID: int64PtrDedicatedTest(55),
			},
		},
	}

	svc := NewDedicatedAccountService(repo, accRepo, keyRepo, nil, nil, nil, nil, nil, nil, nil)

	views, err := svc.ListUserDedicatedAccounts(context.Background(), user1)
	require.NoError(t, err)
	require.Len(t, views, 1)

	rawJSON, err := json.Marshal(views[0])
	require.NoError(t, err)
	jsonStr := string(rawJSON)

	// Verify ZERO sensitive fields leak in user API payload
	require.NotContains(t, jsonStr, "TOP_SECRET_TOKEN")
	require.NotContains(t, jsonStr, "rt_TOP_SECRET_REFRESH_TOKEN")
	require.NotContains(t, jsonStr, "SUPPLIER_INTERNAL_SECRET_NOTE")
	require.NotContains(t, jsonStr, "supplier-raw-email-secret")
	require.NotContains(t, jsonStr, "proxy_id")

	// Verify expected user-facing fields (alias, 5H=18% remaining, 7D=75% remaining, 1 reset credit, bound API key)
	require.Equal(t, "OpenAI Codex Pro 独享 #55", views[0].DisplayName)
	require.NotNil(t, views[0].Quota)
	require.NotNil(t, views[0].Quota.FiveHour)
	require.InDelta(t, 82.0, views[0].Quota.FiveHour.UsedPercent, 1e-6)
	require.InDelta(t, 18.0, views[0].Quota.FiveHour.RemainingPercent, 1e-6)
	require.NotNil(t, views[0].Quota.SevenDay)
	require.InDelta(t, 75.0, views[0].Quota.SevenDay.RemainingPercent, 1e-6)
	require.NotNil(t, views[0].ResetCredits)
	require.Equal(t, 1, views[0].ResetCredits.AvailableCount)
	require.True(t, views[0].ResetCredits.Redeemable)
	require.Len(t, views[0].BoundAPIKeys, 1)
	require.Equal(t, "sk-dedicated-pinned-55", views[0].BoundAPIKeys[0].Key)

	// Cross-user reset credit redemption must be forbidden
	_, err = svc.RedeemUserAccountResetCredit(context.Background(), user2, 55, "idem-cross-user")
	require.ErrorIs(t, err, ErrDedicatedAccountForbidden)

	// Legitimate owner redeems 1 reset credit -> 5H quota restored to 100%, available_count becomes 0
	redeemRes, err := svc.RedeemUserAccountResetCredit(context.Background(), user1, 55, "idem-redeem-1")
	require.NoError(t, err)
	require.True(t, redeemRes.Success)
	require.NotNil(t, redeemRes.ResetCredits)
	require.Equal(t, 0, redeemRes.ResetCredits.AvailableCount)
	require.False(t, redeemRes.ResetCredits.Redeemable)
	require.NotNil(t, redeemRes.QuotaSnapshot)
	require.NotNil(t, redeemRes.QuotaSnapshot.FiveHour)
	require.InDelta(t, 0.0, redeemRes.QuotaSnapshot.FiveHour.UsedPercent, 1e-6)
	require.InDelta(t, 100.0, redeemRes.QuotaSnapshot.FiveHour.RemainingPercent, 1e-6)

	// Redeeming again when 0 credits left must fail cleanly
	_, err = svc.RedeemUserAccountResetCredit(context.Background(), user1, 55, "idem-redeem-2")
	require.Error(t, err)
}

func TestDedicatedAccountBilling_ZeroBalanceAndNoWalletDeduction(t *testing.T) {
	user := &User{ID: 777, Balance: 0} // Balance is 0.00!
	pinnedKey := &APIKey{
		ID:                 88,
		UserID:             user.ID,
		DedicatedAccountID: int64PtrDedicatedTest(55),
		Quota:              100,
		RateLimit5h:        20,
	}
	dedicatedAcc := &Account{
		ID:          55,
		OwnerUserID: int64PtrDedicatedTest(user.ID),
		SaleStatus:  AccountSaleStatusSold,
	}

	require.True(t, pinnedKey.IsDedicatedMode())
	require.True(t, isDedicatedAccountBilling(pinnedKey, user, dedicatedAcc))
	require.Equal(t, int8(2), BillingTypeDedicated)

	params := &postUsageBillingParams{
		Cost:                  &CostBreakdown{TotalCost: 1.25, ActualCost: 0},
		User:                  user,
		APIKey:                pinnedKey,
		Account:               dedicatedAcc,
		IsDedicatedBuyoutBill: true,
	}
	require.False(t, params.shouldDeductAPIKeyQuota(), "dedicated buyout billing must not consume USD key quota")
	require.False(t, params.shouldUpdateRateLimits(), "dedicated buyout billing relies on upstream 5H/7D account window, not USD key rate limit")
}
