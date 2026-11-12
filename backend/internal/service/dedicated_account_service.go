package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"sync"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/shopspring/decimal"
)

const dedicatedUsageForceCooldown = 60 * time.Second

// DedicatedAccountRepository 定义独享账号商城、原子锁单、履约分配、买断提号与只读归属查询所需的持久化能力。
type DedicatedAccountRepository interface {
	ReleaseExpiredLocks(ctx context.Context, now time.Time) (int, error)
	ListMarketAccounts(ctx context.Context, platform string) ([]Account, error)
	GetByID(ctx context.Context, id int64) (*Account, error)
	ListUserDedicatedAccounts(ctx context.Context, userID int64) ([]Account, error)
	EnsureAccountHasGroup(ctx context.Context, accountID int64, platform string) (int64, error)
	LockAccountForOrderInTx(ctx context.Context, tx *dbent.Tx, accountID int64, orderID int64, lockUntil time.Time) (*Account, error)
	LockAccountForBuyoutOrderInTx(ctx context.Context, tx *dbent.Tx, accountID int64, buyerUserID int64, orderID int64, lockUntil time.Time) (*Account, error)
	UnlockAccountForOrder(ctx context.Context, accountID int64, orderID int64) error
	FulfillOrderInTx(ctx context.Context, tx *dbent.Tx, accountID int64, buyerUserID int64, orderID int64, keyGen func() (string, error)) (*Account, *APIKey, error)
	FulfillBuyoutOrderInTx(ctx context.Context, tx *dbent.Tx, accountID int64, buyerUserID int64, orderID int64, keyGen func() (string, error)) (*Account, *APIKey, string, error)
	PurchaseWithBalance(ctx context.Context, buyerUserID int64, accountID int64, balanceMultiplier float64, keyGen func() (string, error)) (*DedicatedBalancePurchaseOutcome, error)
	BuyoutWithBalance(ctx context.Context, buyerUserID int64, accountID int64, balanceMultiplier float64, keyGen func() (string, error)) (*DedicatedBalanceBuyoutOutcome, error)
	RevokeOwnershipOnRefund(ctx context.Context, accountID int64, expectedOwnerUserID int64) ([]string, error)
}

type DedicatedBalancePurchaseOutcome struct {
	Account          *Account
	APIKey           *APIKey
	PriceCNY         float64
	BalanceDeducted  float64
	RemainingBalance float64
}

type DedicatedBalanceBuyoutOutcome struct {
	Account          *Account
	APIKey           *APIKey
	BuyoutInfo       string
	PriceCNY         float64
	BalanceDeducted  float64
	RemainingBalance float64
}

type DedicatedWindowProgress struct {
	UsedPercent      float64    `json:"used_percent"`
	RemainingPercent float64    `json:"remaining_percent"`
	ResetsAt         *time.Time `json:"resets_at,omitempty"`
	RemainingSeconds int        `json:"remaining_seconds"`
	Requests         int64      `json:"requests,omitempty"`
	Tokens           int64      `json:"tokens,omitempty"`
}

type DedicatedQuotaSnapshot struct {
	Source         string                   `json:"source,omitempty"`
	UpdatedAt      *time.Time               `json:"updated_at,omitempty"`
	FiveHour       *DedicatedWindowProgress `json:"five_hour,omitempty"`
	SevenDay       *DedicatedWindowProgress `json:"seven_day,omitempty"`
	SevenDaySonnet *DedicatedWindowProgress `json:"seven_day_sonnet,omitempty"`
	StatusMessage  string                   `json:"status_message,omitempty"`
}

type DedicatedResetCreditStatus struct {
	Supported      bool       `json:"supported"`
	Eligible       bool       `json:"eligible"`
	AvailableCount int        `json:"available_count"`
	Redeemable     bool       `json:"redeemable"`
	CooldownUntil  *time.Time `json:"cooldown_until,omitempty"`
	NextExpiresAt  *string    `json:"next_expires_at,omitempty"`
	FetchedAt      *time.Time `json:"fetched_at,omitempty"`
}

type DedicatedResetCreditRedeemResult struct {
	Success       bool                        `json:"success"`
	Message       string                      `json:"message"`
	ResetCredits  *DedicatedResetCreditStatus `json:"reset_credits,omitempty"`
	QuotaSnapshot *DedicatedQuotaSnapshot     `json:"quota_snapshot,omitempty"`
}

type DedicatedBoundAPIKeySummary struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	Key               string `json:"key"`
	Status            string `json:"status"`
	GroupID           *int64 `json:"group_id,omitempty"`
	DedicatedPoolMode bool   `json:"dedicated_pool_mode"`
}

type UserDedicatedAccountView struct {
	ID                        int64                         `json:"id"`
	DisplayName               string                        `json:"display_name"`
	Platform                  string                        `json:"platform"`
	Type                      string                        `json:"type"`
	HealthStatus              string                        `json:"health_status"`
	StatusLabel               string                        `json:"status_label"`
	Concurrency               int                           `json:"concurrency"`
	GroupID                   *int64                        `json:"group_id,omitempty"`
	GroupName                 string                        `json:"group_name,omitempty"`
	SoldAt                    *time.Time                    `json:"sold_at,omitempty"`
	ExpiresAt                 *time.Time                    `json:"expires_at,omitempty"`
	RemainingDays             *int                          `json:"remaining_days,omitempty"`
	RateLimitResetAt          *time.Time                    `json:"rate_limit_reset_at,omitempty"`
	LastUsedAt                *time.Time                    `json:"last_used_at,omitempty"`
	Quota                     *DedicatedQuotaSnapshot       `json:"quota,omitempty"`
	ResetCredits              *DedicatedResetCreditStatus   `json:"reset_credits"`
	BoundAPIKeys              []DedicatedBoundAPIKeySummary `json:"bound_api_keys"`
	BoughtOut                 bool                          `json:"bought_out"`
	BoughtOutAt               *time.Time                    `json:"bought_out_at,omitempty"`
	BuyoutAvailable           bool                          `json:"buyout_available"`
	BuyoutPriceCNY            float64                       `json:"buyout_price_cny,omitempty"`
	BuyoutPriceBalance        float64                       `json:"buyout_price_balance,omitempty"`
	BalanceCurrencyMultiplier float64                       `json:"balance_currency_multiplier,omitempty"`
}

type DedicatedMarketItem struct {
	ID                        int64                   `json:"id"`
	DisplayName               string                  `json:"display_name"`
	Platform                  string                  `json:"platform"`
	Type                      string                  `json:"type"`
	SaleStatus                string                  `json:"sale_status"`
	LockedUntil               *time.Time              `json:"locked_until,omitempty"`
	Concurrency               int                     `json:"concurrency"`
	GroupName                 string                  `json:"group_name,omitempty"`
	ExpiresAt                 *time.Time              `json:"expires_at,omitempty"`
	RemainingDays             *int                    `json:"remaining_days,omitempty"`
	ListedAt                  *time.Time              `json:"listed_at,omitempty"`
	ElapsedDays               int                     `json:"elapsed_days"`
	BasePriceCNY              float64                 `json:"base_price_cny"`
	DailyPriceDecayCNY        float64                 `json:"daily_price_decay_cny"`
	MinSalePriceCNY           float64                 `json:"min_sale_price_cny"`
	CurrentPriceCNY           float64                 `json:"current_price_cny"`
	CurrentPriceBalance       float64                 `json:"current_price_balance"`
	BalanceCurrencyMultiplier float64                 `json:"balance_currency_multiplier"`
	NextPriceDecayAt          *time.Time              `json:"next_price_decay_at,omitempty"`
	Description               string                  `json:"description,omitempty"`
	QuotaPreview              *DedicatedQuotaSnapshot `json:"quota_preview,omitempty"`
	BuyoutAvailable           bool                    `json:"buyout_available"`
	BuyoutPriceCNY            float64                 `json:"buyout_price_cny,omitempty"`
	BuyoutPriceBalance        float64                 `json:"buyout_price_balance,omitempty"`
}

type DedicatedPurchaseResult struct {
	Account          *UserDedicatedAccountView `json:"account"`
	APIKeyID         int64                     `json:"api_key_id"`
	APIKey           string                    `json:"api_key"`
	APIKeyName       string                    `json:"api_key_name"`
	PriceCNY         float64                   `json:"price_cny"`
	BalanceDeducted  float64                   `json:"balance_deducted"`
	RemainingBalance *float64                  `json:"remaining_balance,omitempty"`
}

type DedicatedBuyoutResult struct {
	Account          *UserDedicatedAccountView `json:"account"`
	BuyoutInfo       string                    `json:"buyout_info"`
	APIKeyID         int64                     `json:"api_key_id,omitempty"`
	APIKey           string                    `json:"api_key,omitempty"`
	APIKeyName       string                    `json:"api_key_name,omitempty"`
	PriceCNY         float64                   `json:"price_cny"`
	BalanceDeducted  float64                   `json:"balance_deducted"`
	RemainingBalance *float64                  `json:"remaining_balance,omitempty"`
}

type DedicatedBuyoutInfoResponse struct {
	AccountID   int64      `json:"account_id"`
	DisplayName string     `json:"display_name"`
	Platform    string     `json:"platform"`
	BoughtOutAt *time.Time `json:"bought_out_at,omitempty"`
	BuyoutInfo  string     `json:"buyout_info"`
}

type DedicatedAccountService struct {
	repo                 DedicatedAccountRepository
	accountRepo          AccountRepository
	apiKeyRepo           APIKeyRepository
	apiKeyService        *APIKeyService
	accountUsageService  *AccountUsageService
	claudeResetCredits   *ClaudeResetCreditService
	openAIQuotaService   *OpenAIQuotaService
	rateLimitService     *RateLimitService
	billingCacheService  *BillingCacheService
	paymentConfigService *PaymentConfigService

	forceRefreshCooldown sync.Map // accountID(int64) -> time.Time
}

func NewDedicatedAccountService(
	repo DedicatedAccountRepository,
	accountRepo AccountRepository,
	apiKeyRepo APIKeyRepository,
	apiKeyService *APIKeyService,
	accountUsageService *AccountUsageService,
	claudeResetCredits *ClaudeResetCreditService,
	openAIQuotaService *OpenAIQuotaService,
	rateLimitService *RateLimitService,
	billingCacheService *BillingCacheService,
	paymentConfigService *PaymentConfigService,
) *DedicatedAccountService {
	return &DedicatedAccountService{
		repo:                 repo,
		accountRepo:          accountRepo,
		apiKeyRepo:           apiKeyRepo,
		apiKeyService:        apiKeyService,
		accountUsageService:  accountUsageService,
		claudeResetCredits:   claudeResetCredits,
		openAIQuotaService:   openAIQuotaService,
		rateLimitService:     rateLimitService,
		billingCacheService:  billingCacheService,
		paymentConfigService: paymentConfigService,
	}
}

// ResolveUserDedicatedAccountGroup 实现 DedicatedAccountResolver，供 APIKeyService 在创建/编辑独享 Key 时校验归属并解析分组。
func (s *DedicatedAccountService) ResolveUserDedicatedAccountGroup(
	ctx context.Context,
	userID int64,
	dedicatedAccountID *int64,
	dedicatedPoolMode bool,
	requestedGroupID *int64,
) (*int64, error) {
	if s == nil || s.repo == nil || userID <= 0 {
		return nil, ErrDedicatedAccountForbidden
	}

	if dedicatedAccountID != nil && *dedicatedAccountID > 0 {
		acc, err := s.repo.GetByID(ctx, *dedicatedAccountID)
		if err != nil || acc == nil {
			return nil, ErrDedicatedAccountNotFound
		}
		if acc.ParentAccountID != nil || acc.IsOnSaleShelf() || acc.OwnerUserID == nil || *acc.OwnerUserID != userID {
			return nil, ErrDedicatedAccountForbidden
		}
		groupID, err := s.resolveOrEnsureAccountGroup(ctx, acc, requestedGroupID)
		if err != nil {
			return nil, err
		}
		return &groupID, nil
	}

	if dedicatedPoolMode {
		owned, err := s.repo.ListUserDedicatedAccounts(ctx, userID)
		if err != nil {
			return nil, err
		}
		if len(owned) == 0 {
			return nil, fmt.Errorf("%w: 您名下暂无已购独享账号，无法开启多号轮询模式", ErrDedicatedAccountForbidden)
		}
		if requestedGroupID != nil && *requestedGroupID > 0 {
			for i := range owned {
				if dedicatedAccountBelongsToGroup(&owned[i], *requestedGroupID) {
					gid := *requestedGroupID
					return &gid, nil
				}
			}
			return nil, fmt.Errorf("%w: 所选分组下没有归属于您的独享账号", ErrDedicatedAccountForbidden)
		}
		groupID, err := s.resolveOrEnsureAccountGroup(ctx, &owned[0], nil)
		if err != nil {
			return nil, err
		}
		return &groupID, nil
	}

	return requestedGroupID, nil
}

// GetDedicatedAccountAlias 返回独享账号面向用户的安全展示名称。
func (s *DedicatedAccountService) GetDedicatedAccountAlias(ctx context.Context, accountID int64) string {
	if s == nil || s.repo == nil || accountID <= 0 {
		return ""
	}
	acc, err := s.repo.GetByID(ctx, accountID)
	if err != nil || acc == nil {
		return ""
	}
	return acc.PublicDisplayName()
}

// PopulateAPIKeysDedicatedAliases 批量填充 API Key 关联独享账号的安全展示别名，避免 N+1 查询。
func (s *DedicatedAccountService) PopulateAPIKeysDedicatedAliases(ctx context.Context, keys []APIKey) {
	if s == nil || len(keys) == 0 {
		return
	}
	seen := make(map[int64]struct{})
	ids := make([]int64, 0)
	for i := range keys {
		if keys[i].DedicatedAccountID != nil && *keys[i].DedicatedAccountID > 0 {
			id := *keys[i].DedicatedAccountID
			if _, ok := seen[id]; !ok {
				seen[id] = struct{}{}
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		return
	}

	aliasByID := make(map[int64]string, len(ids))
	if s.accountRepo != nil {
		if accounts, err := s.accountRepo.GetByIDs(ctx, ids); err == nil {
			for _, acc := range accounts {
				if acc != nil {
					aliasByID[acc.ID] = acc.PublicDisplayName()
				}
			}
		}
	} else if s.repo != nil {
		for _, id := range ids {
			if acc, err := s.repo.GetByID(ctx, id); err == nil && acc != nil {
				aliasByID[id] = acc.PublicDisplayName()
			}
		}
	}

	for i := range keys {
		if keys[i].DedicatedAccountID != nil {
			keys[i].DedicatedAccountAlias = aliasByID[*keys[i].DedicatedAccountID]
		}
	}
}

func (s *DedicatedAccountService) resolveOrEnsureAccountGroup(ctx context.Context, acc *Account, requestedGroupID *int64) (int64, error) {
	if requestedGroupID != nil && *requestedGroupID > 0 {
		if dedicatedAccountBelongsToGroup(acc, *requestedGroupID) {
			return *requestedGroupID, nil
		}
		return 0, fmt.Errorf("%w: 该独享账号未关联所选分组", ErrDedicatedAccountForbidden)
	}
	for _, g := range acc.Groups {
		if g != nil && g.Status == StatusActive {
			return g.ID, nil
		}
	}
	if len(acc.GroupIDs) > 0 && acc.GroupIDs[0] > 0 {
		return acc.GroupIDs[0], nil
	}
	return s.repo.EnsureAccountHasGroup(ctx, acc.ID, acc.Platform)
}

func dedicatedAccountBelongsToGroup(acc *Account, groupID int64) bool {
	if acc == nil || groupID <= 0 {
		return false
	}
	for _, gid := range acc.GroupIDs {
		if gid == groupID {
			return true
		}
	}
	for _, g := range acc.Groups {
		if g != nil && g.ID == groupID {
			return true
		}
	}
	return false
}

func (s *DedicatedAccountService) currentBalanceMultiplier(ctx context.Context) float64 {
	if s != nil && s.paymentConfigService != nil {
		if cfg, err := s.paymentConfigService.GetPaymentConfig(ctx); err == nil && cfg != nil {
			return normalizeBalanceRechargeMultiplier(cfg.BalanceRechargeMultiplier)
		}
	}
	return defaultBalanceRechargeMultiplier
}

// ListMarketAccounts 获取当前上架中的独享账号商城列表（自动释放已超时锁单）。
// 注意：严禁在此接口暴露任何账号邮箱、凭据或买断纯文本（buyout_info）。
func (s *DedicatedAccountService) ListMarketAccounts(ctx context.Context, platform string) ([]DedicatedMarketItem, error) {
	if s == nil || s.repo == nil {
		return []DedicatedMarketItem{}, nil
	}
	now := time.Now()
	accounts, err := s.repo.ListMarketAccounts(ctx, strings.TrimSpace(platform))
	if err != nil {
		return nil, err
	}
	multiplier := s.currentBalanceMultiplier(ctx)
	items := make([]DedicatedMarketItem, 0, len(accounts))
	for i := range accounts {
		acc := &accounts[i]
		if acc.SalePrice <= 0 {
			continue
		}
		currentCNY := acc.CurrentSalePrice(now)
		currentBal := calculateCreditedBalance(currentCNY, multiplier)
		effExpiresAt := effectiveDedicatedExpiresAt(acc)
		buyoutEnabled := acc.IsBuyoutEnabled()
		var buyoutCNY, buyoutBal float64
		if buyoutEnabled {
			buyoutCNY = decimal.NewFromFloat(acc.BuyoutPrice).Round(2).InexactFloat64()
			buyoutBal = calculateCreditedBalance(buyoutCNY, multiplier)
		}
		item := DedicatedMarketItem{
			ID:                        acc.ID,
			DisplayName:               acc.PublicDisplayName(),
			Platform:                  acc.Platform,
			Type:                      acc.Type,
			SaleStatus:                NormalizeAccountSaleStatus(acc.SaleStatus),
			LockedUntil:               acc.LockedUntil,
			Concurrency:               acc.Concurrency,
			GroupName:                 acc.SanitizePublicText(firstAccountGroupName(acc)),
			ExpiresAt:                 effExpiresAt,
			RemainingDays:             computeRemainingDays(effExpiresAt, now),
			ListedAt:                  acc.ListedAt,
			ElapsedDays:               acc.ElapsedListedDays(now),
			BasePriceCNY:              acc.SalePrice,
			DailyPriceDecayCNY:        acc.DailyPriceDecay,
			MinSalePriceCNY:           acc.EffectiveMinSalePrice(),
			CurrentPriceCNY:           currentCNY,
			CurrentPriceBalance:       currentBal,
			BalanceCurrencyMultiplier: multiplier,
			NextPriceDecayAt:          acc.NextPriceDecayAt(now),
			Description:               acc.SanitizePublicText(derefOptionalString(acc.SaleDescription)),
			QuotaPreview:              s.passiveQuotaSnapshot(ctx, acc),
			BuyoutAvailable:           buyoutEnabled,
			BuyoutPriceCNY:            buyoutCNY,
			BuyoutPriceBalance:        buyoutBal,
		}
		items = append(items, item)
	}
	return items, nil
}

// QuoteMarketAccountForOrder 在创建在线支付订单前校验货架账号可用性并返回权威 CNY 实付价格。
func (s *DedicatedAccountService) QuoteMarketAccountForOrder(ctx context.Context, accountID int64) (float64, string, error) {
	if s == nil || s.repo == nil || accountID <= 0 {
		return 0, "", ErrDedicatedAccountNotFound
	}
	now := time.Now()
	_, _ = s.repo.ReleaseExpiredLocks(ctx, now)
	acc, err := s.repo.GetByID(ctx, accountID)
	if err != nil || acc == nil {
		return 0, "", ErrDedicatedAccountNotFound
	}
	if acc.ParentAccountID != nil || acc.Status != StatusActive || acc.OwnerUserID != nil {
		return 0, "", ErrDedicatedAccountUnavailable
	}
	status := NormalizeAccountSaleStatus(acc.SaleStatus)
	if status != AccountSaleStatusListed && !(status == AccountSaleStatusLocked && acc.LockedUntil != nil && !acc.LockedUntil.After(now)) {
		return 0, "", ErrDedicatedAccountUnavailable
	}
	priceCNY := acc.CurrentSalePrice(now)
	if priceCNY < 0.01 {
		return 0, "", ErrDedicatedSaleConfigInvalid
	}
	return priceCNY, acc.PublicDisplayName(), nil
}

// QuoteBuyoutAccountForOrder 在创建在线支付买断订单前校验账号可买断资格（支持商城直接买断与已持有独享账号升级买断）并返回权威 CNY 买断价。
func (s *DedicatedAccountService) QuoteBuyoutAccountForOrder(ctx context.Context, userID int64, accountID int64) (float64, string, error) {
	if s == nil || s.repo == nil || accountID <= 0 || userID <= 0 {
		return 0, "", ErrDedicatedAccountNotFound
	}
	now := time.Now()
	_, _ = s.repo.ReleaseExpiredLocks(ctx, now)
	acc, err := s.repo.GetByID(ctx, accountID)
	if err != nil || acc == nil {
		return 0, "", ErrDedicatedAccountNotFound
	}
	if acc.ParentAccountID != nil {
		return 0, "", ErrDedicatedAccountUnavailable
	}
	isOwnedByBuyer := acc.OwnerUserID != nil && *acc.OwnerUserID == userID
	if acc.OwnerUserID != nil && !isOwnedByBuyer {
		return 0, "", ErrDedicatedAccountForbidden
	}
	if acc.IsBoughtOut() {
		if isOwnedByBuyer {
			return 0, "", ErrDedicatedAlreadyBoughtOut
		}
		return 0, "", ErrDedicatedAccountUnavailable
	}
	if !acc.IsBuyoutEnabled() {
		return 0, "", ErrDedicatedBuyoutUnavailable
	}
	if isOwnedByBuyer {
		if acc.LockedByOrderID != nil && acc.LockedUntil != nil && acc.LockedUntil.After(now) {
			return 0, "", infraerrors.Conflict("DEDICATED_BUYOUT_ORDER_PENDING", "该账号存在待支付的买断订单，请先完成支付或在订单记录中取消后再试")
		}
	} else {
		if acc.Status != StatusActive {
			return 0, "", ErrDedicatedAccountUnavailable
		}
		status := NormalizeAccountSaleStatus(acc.SaleStatus)
		if status != AccountSaleStatusListed && !(status == AccountSaleStatusLocked && acc.LockedUntil != nil && !acc.LockedUntil.After(now)) {
			return 0, "", ErrDedicatedAccountUnavailable
		}
	}
	priceCNY := decimal.NewFromFloat(acc.BuyoutPrice).Round(2).InexactFloat64()
	if priceCNY < 0.01 {
		return 0, "", ErrDedicatedBuyoutUnavailable
	}
	return priceCNY, acc.PublicDisplayName(), nil
}

func (s *DedicatedAccountService) LockAccountForOrderInTx(ctx context.Context, tx *dbent.Tx, accountID int64, orderID int64, lockUntil time.Time) error {
	if s == nil || s.repo == nil {
		return ErrDedicatedAccountUnavailable
	}
	_, err := s.repo.LockAccountForOrderInTx(ctx, tx, accountID, orderID, lockUntil)
	return err
}

func (s *DedicatedAccountService) LockAccountForBuyoutOrderInTx(ctx context.Context, tx *dbent.Tx, accountID int64, buyerUserID int64, orderID int64, lockUntil time.Time) error {
	if s == nil || s.repo == nil {
		return ErrDedicatedAccountUnavailable
	}
	_, err := s.repo.LockAccountForBuyoutOrderInTx(ctx, tx, accountID, buyerUserID, orderID, lockUntil)
	return err
}

func (s *DedicatedAccountService) UnlockAccountForOrder(ctx context.Context, accountID int64, orderID int64) {
	if s == nil || s.repo == nil || accountID <= 0 || orderID <= 0 {
		return
	}
	if err := s.repo.UnlockAccountForOrder(ctx, accountID, orderID); err != nil {
		slog.Warn("unlock dedicated account for order failed", "account_id", accountID, "order_id", orderID, "error", err)
	}
}

// FulfillDedicatedOrderInTx 在支付成功事务中完成账号归属绑定与默认 1 对 1 API Key 自动生成。
func (s *DedicatedAccountService) FulfillDedicatedOrderInTx(ctx context.Context, tx *dbent.Tx, accountID int64, buyerUserID int64, orderID int64) (*Account, *APIKey, error) {
	if s == nil || s.repo == nil {
		return nil, nil, ErrDedicatedAccountUnavailable
	}
	return s.repo.FulfillOrderInTx(ctx, tx, accountID, buyerUserID, orderID, s.generateAPIKeyString)
}

// FulfillDedicatedBuyoutOrderInTx 在线支付买断订单成功回调时完成买断标记（及商城首次归属分配）。
func (s *DedicatedAccountService) FulfillDedicatedBuyoutOrderInTx(ctx context.Context, tx *dbent.Tx, accountID int64, buyerUserID int64, orderID int64) (*Account, *APIKey, string, error) {
	if s == nil || s.repo == nil {
		return nil, nil, "", ErrDedicatedAccountUnavailable
	}
	return s.repo.FulfillBuyoutOrderInTx(ctx, tx, accountID, buyerUserID, orderID, s.generateAPIKeyString)
}

// NotifyOrderFulfilled 支付事务提交后刷新用户的 API Key 认证缓存。
func (s *DedicatedAccountService) NotifyOrderFulfilled(ctx context.Context, buyerUserID int64) {
	if s != nil && s.apiKeyService != nil && buyerUserID > 0 {
		s.apiKeyService.InvalidateAuthCacheByUserID(ctx, buyerUserID)
	}
}

// RevokeOwnershipOnRefund 在管理员退还独享订单款项时收回账号归属、停用对应专属 Key 并刷新缓存（不扣减用户钱包余额）。
func (s *DedicatedAccountService) RevokeOwnershipOnRefund(ctx context.Context, accountID int64, expectedOwnerUserID int64) error {
	if s == nil || s.repo == nil || accountID <= 0 || expectedOwnerUserID <= 0 {
		return nil
	}
	disabledKeys, err := s.repo.RevokeOwnershipOnRefund(ctx, accountID, expectedOwnerUserID)
	if err != nil {
		return err
	}
	if s.apiKeyService != nil {
		s.apiKeyService.InvalidateAuthCacheByUserID(ctx, expectedOwnerUserID)
		for _, k := range disabledKeys {
			s.apiKeyService.InvalidateAuthCacheByKey(ctx, k)
		}
	}
	return nil
}

// PurchaseWithBalance 使用平台钱包余额原子购买上架中的独享账号（平台托管 API 使用权，不含原始账号信息提取）。
func (s *DedicatedAccountService) PurchaseWithBalance(ctx context.Context, userID int64, accountID int64) (*DedicatedPurchaseResult, error) {
	if s == nil || s.repo == nil {
		return nil, ErrDedicatedAccountUnavailable
	}
	if userID <= 0 || accountID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "无效的用户或独享账号 ID")
	}

	multiplier := s.currentBalanceMultiplier(ctx)
	outcome, err := s.repo.PurchaseWithBalance(ctx, userID, accountID, multiplier, s.generateAPIKeyString)
	if err != nil {
		return nil, err
	}

	if s.billingCacheService != nil {
		_ = s.billingCacheService.InvalidateUserBalance(ctx, userID)
	}
	if s.apiKeyService != nil {
		s.apiKeyService.InvalidateAuthCacheByUserID(ctx, userID)
	}

	now := time.Now()
	boundKeys := []DedicatedBoundAPIKeySummary{}
	if outcome.APIKey != nil {
		boundKeys = append(boundKeys, DedicatedBoundAPIKeySummary{
			ID:                outcome.APIKey.ID,
			Name:              outcome.Account.SanitizePublicText(outcome.APIKey.Name),
			Key:               outcome.APIKey.Key,
			Status:            outcome.APIKey.Status,
			GroupID:           outcome.APIKey.GroupID,
			DedicatedPoolMode: outcome.APIKey.DedicatedPoolMode,
		})
	}
	view := s.buildUserAccountView(ctx, outcome.Account, boundKeys, multiplier, now)
	rem := outcome.RemainingBalance
	return &DedicatedPurchaseResult{
		Account:          &view,
		APIKeyID:         outcome.APIKey.ID,
		APIKey:           outcome.APIKey.Key,
		APIKeyName:       outcome.Account.SanitizePublicText(outcome.APIKey.Name),
		PriceCNY:         outcome.PriceCNY,
		BalanceDeducted:  outcome.BalanceDeducted,
		RemainingBalance: &rem,
	}, nil
}

// BuyoutWithBalance 使用平台钱包余额原子买断独享账号（支持商城直接买断或已持有独享账号升级买断），
// 买断成功后向归属客户交付纯文本账号信息，并永久标记为已买断（不再享受封号补偿等优惠策略）。
func (s *DedicatedAccountService) BuyoutWithBalance(ctx context.Context, userID int64, accountID int64) (*DedicatedBuyoutResult, error) {
	if s == nil || s.repo == nil {
		return nil, ErrDedicatedAccountUnavailable
	}
	if userID <= 0 || accountID <= 0 {
		return nil, infraerrors.BadRequest("INVALID_INPUT", "无效的用户或独享账号 ID")
	}

	multiplier := s.currentBalanceMultiplier(ctx)
	outcome, err := s.repo.BuyoutWithBalance(ctx, userID, accountID, multiplier, s.generateAPIKeyString)
	if err != nil {
		return nil, err
	}

	if s.billingCacheService != nil {
		_ = s.billingCacheService.InvalidateUserBalance(ctx, userID)
	}
	if s.apiKeyService != nil {
		s.apiKeyService.InvalidateAuthCacheByUserID(ctx, userID)
	}

	now := time.Now()
	boundKeys := []DedicatedBoundAPIKeySummary{}
	var apiKeyID int64
	var apiKeyStr, apiKeyName string
	if outcome.APIKey != nil {
		apiKeyID = outcome.APIKey.ID
		apiKeyStr = outcome.APIKey.Key
		apiKeyName = outcome.Account.SanitizePublicText(outcome.APIKey.Name)
		boundKeys = append(boundKeys, DedicatedBoundAPIKeySummary{
			ID:                outcome.APIKey.ID,
			Name:              apiKeyName,
			Key:               outcome.APIKey.Key,
			Status:            outcome.APIKey.Status,
			GroupID:           outcome.APIKey.GroupID,
			DedicatedPoolMode: outcome.APIKey.DedicatedPoolMode,
		})
	}
	view := s.buildUserAccountView(ctx, outcome.Account, boundKeys, multiplier, now)
	rem := outcome.RemainingBalance
	return &DedicatedBuyoutResult{
		Account:          &view,
		BuyoutInfo:       outcome.BuyoutInfo,
		APIKeyID:         apiKeyID,
		APIKey:           apiKeyStr,
		APIKeyName:       apiKeyName,
		PriceCNY:         outcome.PriceCNY,
		BalanceDeducted:  outcome.BalanceDeducted,
		RemainingBalance: &rem,
	}, nil
}

// GetUserBoughtOutAccountInfo 仅在确认当前用户是该独享账号归属人且已完成买断（bought_out_at != nil）时，返回纯文本账号交付信息。
func (s *DedicatedAccountService) GetUserBoughtOutAccountInfo(
	ctx context.Context,
	userID int64,
	accountID int64,
) (*DedicatedBuyoutInfoResponse, error) {
	acc, err := s.loadVerifiedOwnedAccount(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	if !acc.IsBoughtOut() {
		return nil, infraerrors.Forbidden("DEDICATED_NOT_BOUGHT_OUT", "该独享账号尚未完成买断，暂无法查看买断账号信息")
	}
	return &DedicatedBuyoutInfoResponse{
		AccountID:   acc.ID,
		DisplayName: acc.PublicDisplayName(),
		Platform:    acc.Platform,
		BoughtOutAt: acc.BoughtOutAt,
		BuyoutInfo:  derefRawString(acc.BuyoutInfo),
	}, nil
}

// ListUserDedicatedAccounts 返回用户名下所有独享账号的只读安全视图（严格隐藏邮箱、凭据及未买断时的 buyout_info）。
func (s *DedicatedAccountService) ListUserDedicatedAccounts(ctx context.Context, userID int64) ([]UserDedicatedAccountView, error) {
	if s == nil || s.repo == nil || userID <= 0 {
		return []UserDedicatedAccountView{}, nil
	}
	accounts, err := s.repo.ListUserDedicatedAccounts(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(accounts) == 0 {
		return []UserDedicatedAccountView{}, nil
	}

	keysByAccount := make(map[int64][]DedicatedBoundAPIKeySummary)
	poolKeysByGroup := make(map[int64][]DedicatedBoundAPIKeySummary)
	if s.apiKeyRepo != nil {
		keys, _, keyErr := s.apiKeyRepo.ListByUserID(ctx, userID, pagination.PaginationParams{Page: 1, PageSize: 200}, APIKeyListFilters{})
		if keyErr == nil {
			for i := range keys {
				k := &keys[i]
				summary := DedicatedBoundAPIKeySummary{
					ID:                k.ID,
					Name:              RedactUpstreamSensitiveText(k.Name),
					Key:               k.Key,
					Status:            k.Status,
					GroupID:           k.GroupID,
					DedicatedPoolMode: k.DedicatedPoolMode,
				}
				if k.DedicatedAccountID != nil && *k.DedicatedAccountID > 0 {
					keysByAccount[*k.DedicatedAccountID] = append(keysByAccount[*k.DedicatedAccountID], summary)
				} else if k.DedicatedPoolMode && k.GroupID != nil && *k.GroupID > 0 {
					poolKeysByGroup[*k.GroupID] = append(poolKeysByGroup[*k.GroupID], summary)
				}
			}
		}
	}

	now := time.Now()
	multiplier := s.currentBalanceMultiplier(ctx)
	views := make([]UserDedicatedAccountView, 0, len(accounts))
	for i := range accounts {
		acc := &accounts[i]
		bound := append([]DedicatedBoundAPIKeySummary{}, keysByAccount[acc.ID]...)
		for _, gid := range acc.GroupIDs {
			bound = append(bound, poolKeysByGroup[gid]...)
		}
		dedup := deduplicateBoundKeys(bound)
		for j := range dedup {
			dedup[j].Name = acc.SanitizePublicText(dedup[j].Name)
		}
		views = append(views, s.buildUserAccountView(ctx, acc, dedup, multiplier, now))
	}
	return views, nil
}

// GetUserDedicatedAccountsUsage 批量查询用户名下独享账号的实时 5H/7D 剩余额度（严格校验归属并内置 60s 强刷冷却）。
func (s *DedicatedAccountService) GetUserDedicatedAccountsUsage(
	ctx context.Context,
	userID int64,
	accountIDs []int64,
	force bool,
) (map[int64]*DedicatedQuotaSnapshot, error) {
	result := make(map[int64]*DedicatedQuotaSnapshot)
	if s == nil || s.repo == nil || s.accountUsageService == nil || userID <= 0 {
		return result, nil
	}

	owned, err := s.repo.ListUserDedicatedAccounts(ctx, userID)
	if err != nil {
		return nil, err
	}
	ownedMap := make(map[int64]*Account, len(owned))
	for i := range owned {
		ownedMap[owned[i].ID] = &owned[i]
	}

	targetIDs := make([]int64, 0)
	if len(accountIDs) == 0 {
		for id := range ownedMap {
			targetIDs = append(targetIDs, id)
		}
	} else {
		seen := make(map[int64]struct{}, len(accountIDs))
		for _, id := range accountIDs {
			if id <= 0 {
				continue
			}
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			if _, ok := ownedMap[id]; !ok {
				return nil, ErrDedicatedAccountForbidden
			}
			targetIDs = append(targetIDs, id)
		}
	}

	for _, id := range targetIDs {
		allowForce := force && s.allowForceRefresh(id)
		usage, usageErr := s.accountUsageService.GetUsage(ctx, id, allowForce)
		if usageErr != nil {
			fallback := s.passiveQuotaSnapshot(ctx, ownedMap[id])
			if fallback == nil {
				fallback = &DedicatedQuotaSnapshot{}
			}
			fallback.StatusMessage = "暂无法刷新上游实时额度，当前展示最近快照"
			result[id] = fallback
			continue
		}
		result[id] = sanitizeUsageInfoToQuotaSnapshot(usage)
	}
	return result, nil
}

// GetUserAccountResetCredits 查询客户名下独享账号是否持有可用官方重置卡。
func (s *DedicatedAccountService) GetUserAccountResetCredits(
	ctx context.Context,
	userID int64,
	accountID int64,
	refresh bool,
) (*DedicatedResetCreditStatus, error) {
	acc, err := s.loadVerifiedOwnedAccount(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	return s.resolveResetCreditStatus(ctx, acc, refresh)
}

// RedeemUserAccountResetCredit 消耗客户名下独享账号持有的 1 张官方重置卡并立即恢复限流窗口。
func (s *DedicatedAccountService) RedeemUserAccountResetCredit(
	ctx context.Context,
	userID int64,
	accountID int64,
	idempotencyKey string,
) (*DedicatedResetCreditRedeemResult, error) {
	acc, err := s.loadVerifiedOwnedAccount(ctx, userID, accountID)
	if err != nil {
		return nil, err
	}
	if acc.IsShadow() {
		return nil, ErrSparkShadowResetNotSupported
	}

	// 1. 合成压测/演示账号支持本地无损模拟用卡验证全链路体验
	if acc.IsSyntheticUITest() {
		return s.redeemSyntheticResetCredit(ctx, acc)
	}

	// 2. Anthropic Claude OAuth 官方重置卡
	if acc.Platform == PlatformAnthropic && acc.Type == AccountTypeOAuth {
		if s.claudeResetCredits == nil {
			return nil, infraerrors.ServiceUnavailable("RESET_SERVICE_UNAVAILABLE", "重置卡服务暂不可用")
		}
		key := strings.TrimSpace(idempotencyKey)
		if key == "" {
			key = fmt.Sprintf("user-%d-acc-%d-%d", userID, accountID, time.Now().Unix()/10)
		}
		if _, err := s.claudeResetCredits.Redeem(ctx, accountID, key); err != nil {
			return nil, sanitizeDedicatedResetErrorForUser(acc, err)
		}
		if s.rateLimitService != nil {
			_ = s.rateLimitService.ClearRateLimit(ctx, accountID)
		}
		status, _ := s.resolveResetCreditStatus(ctx, acc, true)
		var quotaSnap *DedicatedQuotaSnapshot
		if s.accountUsageService != nil {
			if usage, uErr := s.accountUsageService.GetUsage(ctx, accountID, true); uErr == nil {
				quotaSnap = sanitizeUsageInfoToQuotaSnapshot(usage)
			}
		}
		return &DedicatedResetCreditRedeemResult{
			Success:       true,
			Message:       "重置卡已成功使用，5H 速率限制已重置",
			ResetCredits:  status,
			QuotaSnapshot: quotaSnap,
		}, nil
	}

	// 3. OpenAI Codex OAuth 官方重置卡
	if acc.Platform == PlatformOpenAI && acc.Type == AccountTypeOAuth {
		if s.openAIQuotaService == nil {
			return nil, infraerrors.ServiceUnavailable("RESET_SERVICE_UNAVAILABLE", "OpenAI 重置卡服务暂不可用")
		}
		if _, err := s.openAIQuotaService.ResetCredit(ctx, accountID); err != nil {
			return nil, sanitizeDedicatedResetErrorForUser(acc, err)
		}
		postCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 25*time.Second)
		defer cancel()
		var loadFn func(context.Context, int64) (*Account, error)
		if s.accountRepo != nil {
			loadFn = s.accountRepo.GetByID
		}
		_ = RunOpenAIQuotaResetPostProcess(postCtx, accountID, s.openAIQuotaService, s.rateLimitService, loadFn)

		refreshedAcc, _ := s.repo.GetByID(ctx, accountID)
		if refreshedAcc == nil {
			refreshedAcc = acc
		}
		status, _ := s.resolveResetCreditStatus(ctx, refreshedAcc, false)
		var quotaSnap *DedicatedQuotaSnapshot
		if s.accountUsageService != nil {
			if usage, uErr := s.accountUsageService.GetUsage(ctx, accountID, false); uErr == nil {
				quotaSnap = sanitizeUsageInfoToQuotaSnapshot(usage)
			}
		}
		return &DedicatedResetCreditRedeemResult{
			Success:       true,
			Message:       "重置卡已成功使用，账号额度窗口已恢复",
			ResetCredits:  status,
			QuotaSnapshot: quotaSnap,
		}, nil
	}

	return nil, infraerrors.BadRequest("RESET_CREDIT_UNSUPPORTED", "当前账号类型不支持使用重置卡")
}

// sanitizeDedicatedResetErrorForUser 确保官方重置卡调用失败时绝不向前端暴露上游原始报错中的邮箱、组织 ID 或令牌。
func sanitizeDedicatedResetErrorForUser(acc *Account, err error) error {
	if err == nil {
		return nil
	}
	var appErr *infraerrors.ApplicationError
	if errors.As(err, &appErr) && appErr != nil {
		safeMsg := RedactUpstreamSensitiveText(appErr.Message)
		if acc != nil {
			safeMsg = acc.SanitizePublicText(safeMsg)
		}
		return infraerrors.New(int(appErr.Code), appErr.Reason, safeMsg)
	}
	slog.Warn("dedicated reset credit upstream error shielded", "account_id", func() int64 {
		if acc != nil {
			return acc.ID
		}
		return 0
	}(), "error", err)
	return infraerrors.New(502, "RESET_CREDIT_UPSTREAM_FAILED", "官方重置卡上游服务调用失败，请稍后重试或联系管理员")
}

func (s *DedicatedAccountService) redeemSyntheticResetCredit(ctx context.Context, acc *Account) (*DedicatedResetCreditRedeemResult, error) {
	status := readCachedResetCreditsFromExtra(acc)
	if status == nil || status.AvailableCount <= 0 {
		return nil, infraerrors.BadRequest("NO_RESET_CREDITS", "当前账号没有可用的重置卡")
	}
	remaining := status.AvailableCount - 1
	now := time.Now().UTC()
	fiveHourReset := now.Add(5 * time.Hour).Format(time.RFC3339)

	remainingCredits := make([]OpenAIRateLimitResetCreditDetail, 0, remaining)
	expStr := now.Add(30 * 24 * time.Hour).Format(time.RFC3339)
	for i := 0; i < remaining; i++ {
		remainingCredits = append(remainingCredits, OpenAIRateLimitResetCreditDetail{ExpiresAt: expStr})
	}
	updates := map[string]any{
		openaiQuotaResetCreditsKey: OpenAIRateLimitResetCredits{
			AvailableCount: remaining,
			Credits:        remainingCredits,
		},
	}
	if acc.Platform == PlatformOpenAI {
		updates["codex_5h_used_percent"] = float64(0)
		updates["codex_5h_reset_after_seconds"] = 18000
		updates["codex_5h_reset_at"] = fiveHourReset
		updates["codex_usage_updated_at"] = now.Format(time.RFC3339)
	} else if acc.Platform == PlatformAnthropic {
		updates["passive_5h_utilization"] = float64(0)
		updates["passive_5h_resets_at"] = fiveHourReset
	}
	if s.accountRepo != nil {
		_ = s.accountRepo.UpdateExtra(ctx, acc.ID, updates)
		if acc.Platform == PlatformAnthropic {
			start := now
			end := now.Add(5 * time.Hour)
			_ = s.accountRepo.UpdateSessionWindow(ctx, acc.ID, &start, &end, "allowed")
		}
	}
	if s.rateLimitService != nil {
		_ = s.rateLimitService.ClearRateLimit(ctx, acc.ID)
	}

	refreshed, _ := s.repo.GetByID(ctx, acc.ID)
	if refreshed == nil {
		refreshed = acc
	}
	newStatus := readCachedResetCreditsFromExtra(refreshed)
	if newStatus == nil {
		newStatus = &DedicatedResetCreditStatus{Supported: true, Eligible: true, AvailableCount: remaining, Redeemable: remaining > 0}
	}
	return &DedicatedResetCreditRedeemResult{
		Success:       true,
		Message:       "重置卡已成功使用，5H 额度已恢复至 100%",
		ResetCredits:  newStatus,
		QuotaSnapshot: s.passiveQuotaSnapshot(ctx, refreshed),
	}, nil
}

func (s *DedicatedAccountService) resolveResetCreditStatus(ctx context.Context, acc *Account, refresh bool) (*DedicatedResetCreditStatus, error) {
	if acc == nil || acc.IsShadow() {
		return &DedicatedResetCreditStatus{Supported: false}, nil
	}
	if acc.IsSyntheticUITest() {
		if snap := readCachedResetCreditsFromExtra(acc); snap != nil {
			return snap, nil
		}
		return &DedicatedResetCreditStatus{Supported: true, Eligible: false, AvailableCount: 0, Redeemable: false}, nil
	}

	if acc.Platform == PlatformAnthropic && acc.Type == AccountTypeOAuth {
		if cached := readCachedResetCreditsFromExtra(acc); cached != nil && !refresh {
			return cached, nil
		}
		if !refresh || s.claudeResetCredits == nil || !s.allowForceRefresh(acc.ID+1_000_000_000) {
			return &DedicatedResetCreditStatus{Supported: true}, nil
		}
		credits, err := s.claudeResetCredits.Query(ctx, acc.ID)
		if err != nil || credits == nil {
			return &DedicatedResetCreditStatus{Supported: true}, nil
		}
		redeemable := false
		var nextExp *string
		for _, c := range credits.Credits {
			if c.Redeemable && c.ResetsLeft > 0 {
				redeemable = true
			}
			if nextExp == nil && c.ExpiresAt != nil {
				formatted := c.ExpiresAt.Format(time.RFC3339)
				nextExp = &formatted
			}
		}
		fetched := credits.FetchedAt
		return &DedicatedResetCreditStatus{
			Supported:      true,
			Eligible:       credits.Eligible,
			AvailableCount: credits.AvailableCount,
			Redeemable:     redeemable || credits.AvailableCount > 0,
			CooldownUntil:  credits.CooldownUntil,
			NextExpiresAt:  nextExp,
			FetchedAt:      &fetched,
		}, nil
	}

	if acc.Platform == PlatformOpenAI && acc.Type == AccountTypeOAuth {
		if refresh && s.openAIQuotaService != nil && s.allowForceRefresh(acc.ID+2_000_000_000) {
			if usage, err := s.openAIQuotaService.QueryUsage(ctx, acc.ID); err == nil && usage != nil {
				_ = s.openAIQuotaService.CacheCreditsSnapshot(ctx, acc.ID, usage)
				if usage.RateLimitResetCredits != nil {
					_ = s.openAIQuotaService.CacheResetCreditsSnapshot(ctx, acc.ID, usage.RateLimitResetCredits)
					count, nextExp := countUnexpiredOpenAIResetCredits(usage.RateLimitResetCredits, time.Now())
					now := time.Now()
					return &DedicatedResetCreditStatus{
						Supported:      true,
						Eligible:       count > 0,
						AvailableCount: count,
						Redeemable:     count > 0,
						NextExpiresAt:  nextExp,
						FetchedAt:      &now,
					}, nil
				}
			}
		}
		if snap := readCachedResetCreditsFromExtra(acc); snap != nil {
			return snap, nil
		}
		return &DedicatedResetCreditStatus{Supported: true, Eligible: false, AvailableCount: 0, Redeemable: false}, nil
	}

	return &DedicatedResetCreditStatus{Supported: false}, nil
}

func readCachedResetCreditsFromExtra(acc *Account) *DedicatedResetCreditStatus {
	if acc == nil || len(acc.Extra) == 0 {
		return nil
	}
	raw, ok := acc.Extra[openaiQuotaResetCreditsKey]
	if !ok || raw == nil {
		return nil
	}
	buf, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var parsed OpenAIRateLimitResetCredits
	if err := json.Unmarshal(buf, &parsed); err != nil {
		return nil
	}
	count, nextExp := countUnexpiredOpenAIResetCredits(&parsed, time.Now())
	return &DedicatedResetCreditStatus{
		Supported:      true,
		Eligible:       count > 0,
		AvailableCount: count,
		Redeemable:     count > 0,
		NextExpiresAt:  nextExp,
	}
}

func countUnexpiredOpenAIResetCredits(credits *OpenAIRateLimitResetCredits, now time.Time) (int, *string) {
	if credits == nil || credits.AvailableCount <= 0 {
		return 0, nil
	}
	if len(credits.Credits) == 0 {
		return credits.AvailableCount, nil
	}
	valid := 0
	var earliest *string
	for _, c := range credits.Credits {
		expStr := strings.TrimSpace(c.ExpiresAt)
		if expStr == "" {
			valid++
			continue
		}
		if expTime, err := time.Parse(time.RFC3339, expStr); err == nil {
			if expTime.After(now) {
				valid++
				if earliest == nil {
					cp := expStr
					earliest = &cp
				}
			}
		} else {
			valid++
		}
	}
	if valid > credits.AvailableCount {
		valid = credits.AvailableCount
	}
	return valid, earliest
}

func (s *DedicatedAccountService) loadVerifiedOwnedAccount(ctx context.Context, userID int64, accountID int64) (*Account, error) {
	if s == nil || s.repo == nil || userID <= 0 || accountID <= 0 {
		return nil, ErrDedicatedAccountNotFound
	}
	acc, err := s.repo.GetByID(ctx, accountID)
	if err != nil || acc == nil {
		return nil, ErrDedicatedAccountNotFound
	}
	if acc.OwnerUserID == nil || *acc.OwnerUserID != userID {
		return nil, ErrDedicatedAccountForbidden
	}
	return acc, nil
}

func (s *DedicatedAccountService) allowForceRefresh(key int64) bool {
	now := time.Now()
	if prev, ok := s.forceRefreshCooldown.Load(key); ok {
		if prevTime, ok := prev.(time.Time); ok && now.Sub(prevTime) < dedicatedUsageForceCooldown {
			return false
		}
	}
	s.forceRefreshCooldown.Store(key, now)
	return true
}

func (s *DedicatedAccountService) generateAPIKeyString() (string, error) {
	if s != nil && s.apiKeyService != nil {
		return s.apiKeyService.GenerateKey()
	}
	return "", fmt.Errorf("api key service unavailable")
}

func (s *DedicatedAccountService) passiveQuotaSnapshot(ctx context.Context, acc *Account) *DedicatedQuotaSnapshot {
	if acc == nil {
		return nil
	}
	if s != nil && s.accountUsageService != nil {
		if usage, err := s.accountUsageService.GetPassiveUsage(ctx, acc.ID); err == nil && usage != nil {
			if snap := sanitizeUsageInfoToQuotaSnapshot(usage); snap != nil && (snap.FiveHour != nil || snap.SevenDay != nil) {
				return snap
			}
		}
	}
	return extractExtraQuotaSnapshot(acc)
}

func extractExtraQuotaSnapshot(acc *Account) *DedicatedQuotaSnapshot {
	if acc == nil || len(acc.Extra) == 0 {
		return nil
	}
	snap := &DedicatedQuotaSnapshot{Source: "passive"}
	now := time.Now()
	if u5, ok := float64FromAny(acc.Extra["codex_5h_used_percent"]); ok {
		snap.FiveHour = buildWindowProgressFromUsed(u5, parseOptionalRFC3339(acc.Extra["codex_5h_reset_at"]), now)
	} else if u5, ok := float64FromAny(acc.Extra["passive_5h_utilization"]); ok {
		snap.FiveHour = buildWindowProgressFromUsed(u5, parseOptionalRFC3339(acc.Extra["passive_5h_resets_at"]), now)
	}
	if u7, ok := float64FromAny(acc.Extra["codex_7d_used_percent"]); ok {
		snap.SevenDay = buildWindowProgressFromUsed(u7, parseOptionalRFC3339(acc.Extra["codex_7d_reset_at"]), now)
	} else if u7, ok := float64FromAny(acc.Extra["passive_7d_utilization"]); ok {
		snap.SevenDay = buildWindowProgressFromUsed(u7, parseOptionalRFC3339(acc.Extra["passive_7d_resets_at"]), now)
	}
	if ts := parseOptionalRFC3339(acc.Extra["codex_usage_updated_at"]); ts != nil {
		snap.UpdatedAt = ts
	}
	if snap.FiveHour == nil && snap.SevenDay == nil {
		return nil
	}
	return snap
}

func buildWindowProgressFromUsed(used float64, resetsAt *time.Time, now time.Time) *DedicatedWindowProgress {
	clampedUsed := clampPercent(used)
	remSec := 0
	if resetsAt != nil && resetsAt.After(now) {
		remSec = int(resetsAt.Sub(now).Seconds())
	}
	return &DedicatedWindowProgress{
		UsedPercent:      clampedUsed,
		RemainingPercent: math.Round((100-clampedUsed)*10) / 10,
		ResetsAt:         resetsAt,
		RemainingSeconds: remSec,
	}
}

func sanitizeUsageInfoToQuotaSnapshot(info *UsageInfo) *DedicatedQuotaSnapshot {
	if info == nil {
		return nil
	}
	snap := &DedicatedQuotaSnapshot{
		Source:    info.Source,
		UpdatedAt: info.UpdatedAt,
	}
	if info.FiveHour != nil {
		snap.FiveHour = sanitizeUsageProgress(info.FiveHour)
	}
	if info.SevenDay != nil {
		snap.SevenDay = sanitizeUsageProgress(info.SevenDay)
	}
	if info.SevenDaySonnet != nil {
		snap.SevenDaySonnet = sanitizeUsageProgress(info.SevenDaySonnet)
	}
	if info.Error != "" {
		snap.StatusMessage = "额度状态存在延迟，系统将自动重试同步"
	}
	return snap
}

func sanitizeUsageProgress(p *UsageProgress) *DedicatedWindowProgress {
	if p == nil {
		return nil
	}
	used := clampPercent(p.Utilization)
	out := &DedicatedWindowProgress{
		UsedPercent:      used,
		RemainingPercent: math.Round((100-used)*10) / 10,
		ResetsAt:         p.ResetsAt,
		RemainingSeconds: p.RemainingSeconds,
	}
	if p.WindowStats != nil {
		out.Requests = p.WindowStats.Requests
		out.Tokens = p.WindowStats.Tokens
	}
	return out
}

func clampPercent(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return math.Round(v*10) / 10
}

func (s *DedicatedAccountService) buildUserAccountView(
	ctx context.Context,
	acc *Account,
	boundKeys []DedicatedBoundAPIKeySummary,
	multiplier float64,
	now time.Time,
) UserDedicatedAccountView {
	health, label := deriveUserFacingAccountHealth(acc, now)
	var groupID *int64
	if len(acc.GroupIDs) > 0 {
		gid := acc.GroupIDs[0]
		groupID = &gid
	} else if len(acc.Groups) > 0 && acc.Groups[0] != nil {
		gid := acc.Groups[0].ID
		groupID = &gid
	}
	resetStatus, _ := s.resolveResetCreditStatus(ctx, acc, false)
	if resetStatus == nil {
		resetStatus = &DedicatedResetCreditStatus{Supported: false}
	}
	if boundKeys == nil {
		boundKeys = []DedicatedBoundAPIKeySummary{}
	}
	effExpiresAt := effectiveDedicatedExpiresAt(acc)
	boughtOut := acc.IsBoughtOut()
	buyoutAvailable := acc.IsBuyoutEnabled()
	var buyoutCNY, buyoutBal float64
	if buyoutAvailable {
		buyoutCNY = decimal.NewFromFloat(acc.BuyoutPrice).Round(2).InexactFloat64()
		buyoutBal = calculateCreditedBalance(buyoutCNY, multiplier)
	}
	return UserDedicatedAccountView{
		ID:                        acc.ID,
		DisplayName:               acc.PublicDisplayName(),
		Platform:                  acc.Platform,
		Type:                      acc.Type,
		HealthStatus:              health,
		StatusLabel:               label,
		Concurrency:               acc.Concurrency,
		GroupID:                   groupID,
		GroupName:                 acc.SanitizePublicText(firstAccountGroupName(acc)),
		SoldAt:                    acc.SoldAt,
		ExpiresAt:                 effExpiresAt,
		RemainingDays:             computeRemainingDays(effExpiresAt, now),
		RateLimitResetAt:          acc.RateLimitResetAt,
		LastUsedAt:                acc.LastUsedAt,
		Quota:                     s.passiveQuotaSnapshot(ctx, acc),
		ResetCredits:              resetStatus,
		BoundAPIKeys:              boundKeys,
		BoughtOut:                 boughtOut,
		BoughtOutAt:               acc.BoughtOutAt,
		BuyoutAvailable:           buyoutAvailable,
		BuyoutPriceCNY:            buyoutCNY,
		BuyoutPriceBalance:        buyoutBal,
		BalanceCurrencyMultiplier: multiplier,
	}
}

func effectiveDedicatedExpiresAt(acc *Account) *time.Time {
	if acc == nil {
		return nil
	}
	if acc.ExpiresAt != nil {
		return acc.ExpiresAt
	}
	if len(acc.Credentials) > 0 {
		if t := parseOptionalRFC3339(acc.Credentials["subscription_expires_at"]); t != nil {
			return t
		}
	}
	return nil
}

func deriveUserFacingAccountHealth(acc *Account, now time.Time) (string, string) {
	if acc == nil {
		return "unknown", "未知"
	}
	if effExp := effectiveDedicatedExpiresAt(acc); effExp != nil && !effExp.After(now) {
		return "expired", "已到期"
	}
	if acc.Status == StatusError {
		return "error", "异常维护中"
	}
	if acc.Status != StatusActive || !acc.Schedulable {
		return "paused", "已暂停"
	}
	if acc.RateLimitResetAt != nil && acc.RateLimitResetAt.After(now) {
		return "rate_limited", "窗口限流冷却中"
	}
	if acc.OverloadUntil != nil && acc.OverloadUntil.After(now) {
		return "overloaded", "上游高峰拥塞中"
	}
	if acc.TempUnschedulableUntil != nil && acc.TempUnschedulableUntil.After(now) {
		return "cooldown", "短暂冷却恢复中"
	}
	return "healthy", "正常可用"
}

func computeRemainingDays(expiresAt *time.Time, now time.Time) *int {
	if expiresAt == nil {
		return nil
	}
	if !expiresAt.After(now) {
		zero := 0
		return &zero
	}
	days := int(math.Ceil(expiresAt.Sub(now).Hours() / 24.0))
	return &days
}

func firstAccountGroupName(acc *Account) string {
	if acc == nil {
		return ""
	}
	for _, g := range acc.Groups {
		if g != nil && strings.TrimSpace(g.Name) != "" {
			return strings.TrimSpace(g.Name)
		}
	}
	return ""
}

func derefOptionalString(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}

func derefRawString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func float64FromAny(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func parseOptionalRFC3339(v any) *time.Time {
	s, ok := v.(string)
	if !ok {
		return nil
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}
	return nil
}

func deduplicateBoundKeys(in []DedicatedBoundAPIKeySummary) []DedicatedBoundAPIKeySummary {
	if len(in) <= 1 {
		return in
	}
	seen := make(map[int64]struct{}, len(in))
	out := make([]DedicatedBoundAPIKeySummary, 0, len(in))
	for _, item := range in {
		if _, ok := seen[item.ID]; ok {
			continue
		}
		seen[item.ID] = struct{}{}
		out = append(out, item)
	}
	return out
}
