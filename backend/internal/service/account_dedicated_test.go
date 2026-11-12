//go:build unit

package service

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func int64PtrDedicatedTest(v int64) *int64 { return &v }
func strPtrDedicatedTest(v string) *string { return &v }

func TestValidateAccountSaleConfig(t *testing.T) {
	ownerID := int64(42)

	require.NoError(t, ValidateAccountSaleConfig("", nil, 0, 0, 0))
	require.NoError(t, ValidateAccountSaleConfig(AccountSaleStatusNone, nil, 0, 0, 0))
	require.NoError(t, ValidateAccountSaleConfig(AccountSaleStatusSold, &ownerID, 200, 0, 0))
	require.NoError(t, ValidateAccountSaleConfig(AccountSaleStatusListed, nil, 200, 5, 50))
	require.NoError(t, ValidateAccountSaleConfig(AccountSaleStatusListed, nil, 0.01, 0, 0))

	// Invalid statuses and numeric anomalies
	require.ErrorIs(t, ValidateAccountSaleConfig("unknown_status", nil, 100, 0, 0), ErrDedicatedSaleConfigInvalid)
	require.ErrorIs(t, ValidateAccountSaleConfig(AccountSaleStatusListed, nil, math.NaN(), 0, 0), ErrDedicatedSaleConfigInvalid)
	require.ErrorIs(t, ValidateAccountSaleConfig(AccountSaleStatusListed, nil, math.Inf(1), 0, 0), ErrDedicatedSaleConfigInvalid)
	require.ErrorIs(t, ValidateAccountSaleConfig(AccountSaleStatusListed, nil, -10, 0, 0), ErrDedicatedSaleConfigInvalid)
	require.ErrorIs(t, ValidateAccountSaleConfig(AccountSaleStatusListed, nil, 200, -1, 0), ErrDedicatedSaleConfigInvalid)
	require.ErrorIs(t, ValidateAccountSaleConfig(AccountSaleStatusListed, nil, 200, 5, -0.01), ErrDedicatedSaleConfigInvalid)

	// Zero-yuan / owner+shelf conflict vulnerabilities must be rejected
	require.ErrorIs(t, ValidateAccountSaleConfig(AccountSaleStatusListed, nil, 0, 0, 0), ErrDedicatedSaleConfigInvalid)
	require.ErrorIs(t, ValidateAccountSaleConfig(AccountSaleStatusListed, &ownerID, 200, 5, 100), ErrDedicatedSaleConfigInvalid)
	require.ErrorIs(t, ValidateAccountSaleConfig(AccountSaleStatusLocked, &ownerID, 200, 5, 100), ErrDedicatedSaleConfigInvalid)
	require.ErrorIs(t, ValidateAccountSaleConfig(AccountSaleStatusListed, nil, 100, 5, 120), ErrDedicatedSaleConfigInvalid)
}

func TestAccountCurrentSalePriceAndDailyDecay(t *testing.T) {
	listedAt := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	acc := &Account{
		ID:              101,
		Platform:        PlatformAnthropic,
		SaleStatus:      AccountSaleStatusListed,
		SalePrice:       200.00,
		DailyPriceDecay: 6.50,
		MinSalePrice:    150.00,
		ListedAt:        &listedAt,
	}

	// Before or at listing time -> full price
	require.Equal(t, 0, acc.ElapsedListedDays(listedAt.Add(-time.Hour)))
	require.InDelta(t, 200.00, acc.CurrentSalePrice(listedAt), 1e-6)

	// 23h59m59s elapsed -> still 0 full days
	justBeforeDay1 := listedAt.Add(24*time.Hour - time.Second)
	require.Equal(t, 0, acc.ElapsedListedDays(justBeforeDay1))
	require.InDelta(t, 200.00, acc.CurrentSalePrice(justBeforeDay1), 1e-6)
	nextDecay := acc.NextPriceDecayAt(justBeforeDay1)
	require.NotNil(t, nextDecay)
	require.True(t, nextDecay.Equal(listedAt.Add(24*time.Hour)))

	// Exactly 24h -> 1 day decay (200 - 6.50 = 193.50)
	day1 := listedAt.Add(24 * time.Hour)
	require.Equal(t, 1, acc.ElapsedListedDays(day1))
	require.InDelta(t, 193.50, acc.CurrentSalePrice(day1), 1e-6)

	// 2 days elapsed -> 200 - 13.00 = 187.00
	day2 := listedAt.Add(48*time.Hour + 30*time.Minute)
	require.Equal(t, 2, acc.ElapsedListedDays(day2))
	require.InDelta(t, 187.00, acc.CurrentSalePrice(day2), 1e-6)

	// 20 days elapsed -> clamps strictly to MinSalePrice (150.00) and stops NextPriceDecayAt
	day20 := listedAt.Add(20 * 24 * time.Hour)
	require.InDelta(t, 150.00, acc.CurrentSalePrice(day20), 1e-6)
	require.Nil(t, acc.NextPriceDecayAt(day20))

	// When MinSalePrice is 0 (unspecified), EffectiveMinSalePrice defaults to 0.01 CNY to prevent 0-yuan exploit
	noFloorAcc := &Account{
		ID:              102,
		SaleStatus:      AccountSaleStatusListed,
		SalePrice:       10.00,
		DailyPriceDecay: 5.00,
		MinSalePrice:    0,
		ListedAt:        &listedAt,
	}
	require.InDelta(t, 0.01, noFloorAcc.EffectiveMinSalePrice(), 1e-6)
	require.InDelta(t, 0.01, noFloorAcc.CurrentSalePrice(listedAt.Add(5*24*time.Hour)), 1e-6)
}

func TestAccountIsSchedulable_ShelfGuard(t *testing.T) {
	listedAcc := &Account{
		ID:          1,
		Status:      StatusActive,
		Schedulable: true,
		SaleStatus:  AccountSaleStatusListed,
	}
	lockedAcc := &Account{
		ID:          2,
		Status:      StatusActive,
		Schedulable: true,
		SaleStatus:  AccountSaleStatusLocked,
	}
	soldAcc := &Account{
		ID:          3,
		Status:      StatusActive,
		Schedulable: true,
		SaleStatus:  AccountSaleStatusSold,
		OwnerUserID: int64PtrDedicatedTest(10),
	}
	publicAcc := &Account{
		ID:          4,
		Status:      StatusActive,
		Schedulable: true,
		SaleStatus:  AccountSaleStatusNone,
	}

	require.False(t, listedAcc.IsSchedulable(), "listed shelf account must never be schedulable")
	require.False(t, lockedAcc.IsSchedulable(), "locked checkout account must never be schedulable")
	require.True(t, soldAcc.IsSchedulable(), "sold active dedicated account is schedulable for its owner")
	require.True(t, publicAcc.IsSchedulable(), "normal active public account is schedulable")
}

func TestIsAccountAllowedForRequestContext_BidirectionalIsolation(t *testing.T) {
	user1 := int64(1001)
	user2 := int64(2002)

	accUser1A := &Account{
		ID:          11,
		Platform:    PlatformOpenAI,
		Status:      StatusActive,
		Schedulable: true,
		SaleStatus:  AccountSaleStatusSold,
		OwnerUserID: int64PtrDedicatedTest(user1),
	}
	accUser1AShadow := &Account{
		ID:              110,
		Platform:        PlatformOpenAI,
		Status:          StatusActive,
		Schedulable:     true,
		SaleStatus:      AccountSaleStatusSold,
		OwnerUserID:     int64PtrDedicatedTest(user1),
		ParentAccountID: int64PtrDedicatedTest(accUser1A.ID),
		QuotaDimension:  QuotaDimensionSpark,
	}
	accUser1B := &Account{
		ID:          12,
		Platform:    PlatformOpenAI,
		Status:      StatusActive,
		Schedulable: true,
		SaleStatus:  AccountSaleStatusSold,
		OwnerUserID: int64PtrDedicatedTest(user1),
	}
	accUser2C := &Account{
		ID:          21,
		Platform:    PlatformOpenAI,
		Status:      StatusActive,
		Schedulable: true,
		SaleStatus:  AccountSaleStatusSold,
		OwnerUserID: int64PtrDedicatedTest(user2),
	}
	accPublic := &Account{
		ID:          31,
		Platform:    PlatformOpenAI,
		Status:      StatusActive,
		Schedulable: true,
		SaleStatus:  AccountSaleStatusNone,
		OwnerUserID: nil,
	}
	accListedShelf := &Account{
		ID:          41,
		Platform:    PlatformOpenAI,
		Status:      StatusActive,
		Schedulable: true,
		SaleStatus:  AccountSaleStatusListed,
		OwnerUserID: nil,
	}
	accLockedShelf := &Account{
		ID:          42,
		Platform:    PlatformOpenAI,
		Status:      StatusActive,
		Schedulable: true,
		SaleStatus:  AccountSaleStatusLocked,
		OwnerUserID: nil,
	}

	// Case 1: Public API Key (User 1 or User 2) -> ONLY public accounts, NEVER owned dedicated or shelf accounts
	publicCtx := WithDedicatedRoutingContext(context.Background(), user1, nil, false)
	require.True(t, IsAccountAllowedForRequestContext(publicCtx, accPublic))
	require.False(t, IsAccountAllowedForRequestContext(publicCtx, accUser1A), "public key must never consume user's dedicated account")
	require.False(t, IsAccountAllowedForRequestContext(publicCtx, accUser1AShadow))
	require.False(t, IsAccountAllowedForRequestContext(publicCtx, accUser2C))
	require.False(t, IsAccountAllowedForRequestContext(publicCtx, accListedShelf))
	require.False(t, IsAccountAllowedForRequestContext(publicCtx, accLockedShelf))

	// Case 2: User 1's 1-to-1 Pinned Dedicated Key -> ONLY Account 11 (and its Spark shadow 110), ZERO overflow to 12, 21, or Public
	pinnedCtx := WithDedicatedRoutingContext(context.Background(), user1, int64PtrDedicatedTest(accUser1A.ID), false)
	require.True(t, IsAccountAllowedForRequestContext(pinnedCtx, accUser1A))
	require.True(t, IsAccountAllowedForRequestContext(pinnedCtx, accUser1AShadow), "Spark shadow of pinned parent account should be allowed")
	require.False(t, IsAccountAllowedForRequestContext(pinnedCtx, accUser1B), "1:1 pinned key must not cross over to user's other dedicated account")
	require.False(t, IsAccountAllowedForRequestContext(pinnedCtx, accUser2C), "must never route to another user's dedicated account")
	require.False(t, IsAccountAllowedForRequestContext(pinnedCtx, accPublic), "dedicated key must NEVER overflow to public shared pool")
	require.False(t, IsAccountAllowedForRequestContext(pinnedCtx, accListedShelf))

	// Case 3: User 1's Multi-Dedicated Pool Key (dedicated_pool_mode = true) -> allows User 1's A & B, ZERO overflow to User 2 or Public
	poolCtx := WithDedicatedRoutingContext(context.Background(), user1, nil, true)
	require.True(t, IsAccountAllowedForRequestContext(poolCtx, accUser1A))
	require.True(t, IsAccountAllowedForRequestContext(poolCtx, accUser1AShadow))
	require.True(t, IsAccountAllowedForRequestContext(poolCtx, accUser1B))
	require.False(t, IsAccountAllowedForRequestContext(poolCtx, accUser2C))
	require.False(t, IsAccountAllowedForRequestContext(poolCtx, accPublic), "dedicated pool mode must NEVER overflow to public shared pool")
	require.False(t, IsAccountAllowedForRequestContext(poolCtx, accListedShelf))

	// Case 4: Tampered context where User 2 tries to pin User 1's Account 11 -> strictly rejected
	crossTenantCtx := WithDedicatedRoutingContext(context.Background(), user2, int64PtrDedicatedTest(accUser1A.ID), false)
	require.False(t, IsAccountAllowedForRequestContext(crossTenantCtx, accUser1A))
	require.False(t, IsAccountAllowedForRequestContext(crossTenantCtx, accPublic))

	// Public display alias hides internal admin account name
	accWithInternalName := &Account{
		ID:           88,
		Name:         "internal-supplier-batch-09-secret@gmail.com",
		Platform:     PlatformAnthropic,
		DisplayAlias: strPtrDedicatedTest("Claude Max 20x 独享专线 #88"),
	}
	require.Equal(t, "Claude Max 20x 独享专线 #88", accWithInternalName.PublicDisplayName())
	accDefaultAlias := &Account{ID: 89, Name: "secret-internal-name", Platform: PlatformAnthropic}
	require.Equal(t, "Claude 独享 #89", accDefaultAlias.PublicDisplayName())
}
