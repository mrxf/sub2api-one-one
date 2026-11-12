package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	AccountSaleStatusNone   = "none"
	AccountSaleStatusListed = "listed"
	AccountSaleStatusLocked = "locked"
	AccountSaleStatusSold   = "sold"
)

var (
	ErrDedicatedAccountNotFound    = infraerrors.NotFound("DEDICATED_ACCOUNT_NOT_FOUND", "独享账号不存在")
	ErrDedicatedAccountForbidden   = infraerrors.Forbidden("DEDICATED_ACCOUNT_FORBIDDEN", "无权绑定或访问该独享账号")
	ErrDedicatedAccountUnavailable = infraerrors.Conflict("DEDICATED_ACCOUNT_UNAVAILABLE", "该独享账号已被锁定或售出，请刷新商城")
	ErrDedicatedSaleConfigInvalid  = infraerrors.BadRequest("DEDICATED_SALE_CONFIG_INVALID", "独享账号上架参数不合法")
)

// NormalizeAccountSaleStatus 归一化销售状态，空字符串回退为 "none"。
func NormalizeAccountSaleStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "", AccountSaleStatusNone:
		return AccountSaleStatusNone
	case AccountSaleStatusListed:
		return AccountSaleStatusListed
	case AccountSaleStatusLocked:
		return AccountSaleStatusLocked
	case AccountSaleStatusSold:
		return AccountSaleStatusSold
	default:
		return ""
	}
}

// ValidateAccountSaleConfig 校验上架价格与递减规则，确保无负数/NaN/零元白嫖漏洞。
func ValidateAccountSaleConfig(saleStatus string, ownerUserID *int64, salePrice, dailyPriceDecay, minSalePrice float64) error {
	norm := NormalizeAccountSaleStatus(saleStatus)
	if norm == "" {
		return fmt.Errorf("%w: invalid sale_status %q", ErrDedicatedSaleConfigInvalid, saleStatus)
	}
	for _, v := range []float64{salePrice, dailyPriceDecay, minSalePrice} {
		if math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
			return fmt.Errorf("%w: price values must be finite and non-negative", ErrDedicatedSaleConfigInvalid)
		}
	}
	if norm == AccountSaleStatusListed || norm == AccountSaleStatusLocked {
		if ownerUserID != nil && *ownerUserID > 0 {
			return fmt.Errorf("%w: 已有归属用户的独享账号不可同时处于待售货架状态", ErrDedicatedSaleConfigInvalid)
		}
		if salePrice < 0.01 {
			return fmt.Errorf("%w: 上架价格必须不少于 0.01 元", ErrDedicatedSaleConfigInvalid)
		}
		if minSalePrice > salePrice {
			return fmt.Errorf("%w: 保底价格不能高于初始上架价格", ErrDedicatedSaleConfigInvalid)
		}
	}
	return nil
}

// IsOnSaleShelf 报告账号是否处于商城货架锁定/待售状态（任何网关流量均禁止调度）。
func (a *Account) IsOnSaleShelf() bool {
	if a == nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(a.SaleStatus)) {
	case AccountSaleStatusListed, AccountSaleStatusLocked:
		return true
	default:
		return false
	}
}

// IsOwnedDedicated 报告账号是否已归属于具体独享用户。
func (a *Account) IsOwnedDedicated() bool {
	return a != nil && a.OwnerUserID != nil && *a.OwnerUserID > 0
}

// PublicDisplayName 返回面向普通用户与商城的安全展示名称（隐藏管理员内部账号名）。
func (a *Account) PublicDisplayName() string {
	if a == nil {
		return ""
	}
	if a.DisplayAlias != nil {
		if alias := strings.TrimSpace(*a.DisplayAlias); alias != "" {
			return alias
		}
	}
	platformLabel := strings.ToUpper(strings.TrimSpace(a.Platform))
	if platformLabel == "ANTHROPIC" {
		platformLabel = "Claude"
	} else if platformLabel == "OPENAI" {
		platformLabel = "OpenAI"
	} else if platformLabel == "" {
		platformLabel = "AI"
	}
	return fmt.Sprintf("%s 独享 #%d", platformLabel, a.ID)
}

// ElapsedListedDays 返回自 listed_at 起算已满多少个 24 小时自然天（向下取整，>=0）。
func (a *Account) ElapsedListedDays(now time.Time) int {
	if a == nil || a.ListedAt == nil {
		return 0
	}
	if !now.After(*a.ListedAt) {
		return 0
	}
	days := int(now.Sub(*a.ListedAt) / (24 * time.Hour))
	if days < 0 {
		return 0
	}
	return days
}

// EffectiveMinSalePrice 返回有效保底价格（当初始价 > 0 且未显式配置保底价时，默认兜底 0.01 元防止跌成 0 元漏洞）。
func (a *Account) EffectiveMinSalePrice() float64 {
	if a == nil || a.SalePrice <= 0 {
		return 0
	}
	if a.MinSalePrice >= 0.01 {
		if a.MinSalePrice > a.SalePrice {
			return a.SalePrice
		}
		return a.MinSalePrice
	}
	if a.SalePrice < 0.01 {
		return a.SalePrice
	}
	return 0.01
}

// CurrentSalePrice 使用高精度十进制计算当前时刻的实际买断价格（CNY 元，保留 2 位小数）：
// CurrentPrice = max(EffectiveMinSalePrice, SalePrice - floor((now - ListedAt)/24h) * DailyPriceDecay)
func (a *Account) CurrentSalePrice(now time.Time) float64 {
	if a == nil || a.SalePrice <= 0 {
		return 0
	}
	base := decimal.NewFromFloat(a.SalePrice)
	floorPrice := decimal.NewFromFloat(a.EffectiveMinSalePrice())
	days := a.ElapsedListedDays(now)
	if days > 0 && a.DailyPriceDecay > 0 {
		decayTotal := decimal.NewFromFloat(a.DailyPriceDecay).Mul(decimal.NewFromInt(int64(days)))
		base = base.Sub(decayTotal)
	}
	if base.LessThan(floorPrice) {
		base = floorPrice
	}
	if base.LessThan(decimal.Zero) {
		base = decimal.Zero
	}
	return base.Round(2).InexactFloat64()
}

// NextPriceDecayAt 返回下一次触发日递减降价的时间点；若未配置日递减或已触底则返回 nil。
func (a *Account) NextPriceDecayAt(now time.Time) *time.Time {
	if a == nil || a.ListedAt == nil || a.DailyPriceDecay <= 0 || a.SalePrice <= 0 {
		return nil
	}
	current := a.CurrentSalePrice(now)
	if current <= a.EffectiveMinSalePrice()+1e-9 {
		return nil
	}
	days := a.ElapsedListedDays(now)
	next := a.ListedAt.Add(time.Duration(days+1) * 24 * time.Hour)
	return &next
}

// DedicatedRoutingInfo 描述当前网关请求的独享路由上下文。
type DedicatedRoutingInfo struct {
	UserID             int64
	DedicatedAccountID *int64
	DedicatedPoolMode  bool
}

// IsDedicatedMode 报告当前请求上下文是否处于独享路由模式（禁止回退到公共池）。
func (d DedicatedRoutingInfo) IsDedicatedMode() bool {
	return (d.DedicatedAccountID != nil && *d.DedicatedAccountID > 0) || d.DedicatedPoolMode
}

// WithDedicatedRoutingContext 将独享路由上下文写入 context.Context（供测试或内部调用复用）。
func WithDedicatedRoutingContext(ctx context.Context, userID int64, dedicatedAccountID *int64, dedicatedPoolMode bool) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if userID > 0 {
		ctx = context.WithValue(ctx, ctxkey.UserID, userID)
	}
	if dedicatedAccountID != nil && *dedicatedAccountID > 0 {
		idCopy := *dedicatedAccountID
		ctx = context.WithValue(ctx, ctxkey.DedicatedAccountID, idCopy)
	}
	if dedicatedPoolMode {
		ctx = context.WithValue(ctx, ctxkey.DedicatedPoolMode, true)
	}
	return ctx
}

// DedicatedRoutingInfoFromContext 从请求上下文提取独享路由信息。
func DedicatedRoutingInfoFromContext(ctx context.Context) DedicatedRoutingInfo {
	var info DedicatedRoutingInfo
	if ctx == nil {
		return info
	}
	if uid, ok := ctx.Value(ctxkey.UserID).(int64); ok && uid > 0 {
		info.UserID = uid
	}
	switch v := ctx.Value(ctxkey.DedicatedAccountID).(type) {
	case int64:
		if v > 0 {
			idCopy := v
			info.DedicatedAccountID = &idCopy
		}
	case *int64:
		if v != nil && *v > 0 {
			idCopy := *v
			info.DedicatedAccountID = &idCopy
		}
	}
	if poolMode, ok := ctx.Value(ctxkey.DedicatedPoolMode).(bool); ok {
		info.DedicatedPoolMode = poolMode
	}
	return info
}

// IsAccountAllowedForRequestContext 实现独享号池与公共号池的严格双向硬隔离：
// 1. 货架待售/锁单账号（listed/locked）对任何请求均不可用；
// 2. 独享 API Key 请求：
//    - 仅允许路由到 owner_user_id == reqUserID 的自有独享账号（永不溢出公共池、永不串号至他人）；
//    - 若未开启 dedicated_pool_mode 且绑定了 dedicated_account_id，则严格 1 对 1 锁定该账号（及其同源 Spark 影子）；
//    - 若开启了 dedicated_pool_mode，则可在该用户名下同分组/平台的多个独享账号间轮询/故障转移；
// 3. 普通公共池 API Key（或后台非独享测试）请求：
//    - 严格只允许 owner_user_id == nil 的公共账号，杜绝任何请求蹭用客户买断的独享账号。
func IsAccountAllowedForRequestContext(ctx context.Context, account *Account) bool {
	if account == nil || account.IsOnSaleShelf() {
		return false
	}
	routing := DedicatedRoutingInfoFromContext(ctx)
	if routing.IsDedicatedMode() {
		if routing.UserID <= 0 || account.OwnerUserID == nil || *account.OwnerUserID != routing.UserID {
			return false
		}
		if routing.DedicatedAccountID != nil && !routing.DedicatedPoolMode {
			targetID := *routing.DedicatedAccountID
			if account.ID == targetID {
				return true
			}
			// 允许该独享母账号派生的同源影子账号（例如 OpenAI Codex Spark 影子）透传调度
			if account.ParentAccountID != nil && *account.ParentAccountID == targetID {
				return true
			}
			return false
		}
		return true
	}

	// 公共池模式：严禁命中任何已归属给用户的独享账号
	return account.OwnerUserID == nil
}

// FilterAccountsForRequestContext 过滤账号切片（值类型），仅保留当前请求上下文允许调度的账号。
func FilterAccountsForRequestContext(ctx context.Context, accounts []Account) []Account {
	if len(accounts) == 0 {
		return accounts
	}
	filtered := make([]Account, 0, len(accounts))
	for i := range accounts {
		if IsAccountAllowedForRequestContext(ctx, &accounts[i]) {
			filtered = append(filtered, accounts[i])
		}
	}
	return filtered
}

// FilterAccountPointersForRequestContext 过滤账号指针切片，仅保留当前请求上下文允许调度的账号。
func FilterAccountPointersForRequestContext(ctx context.Context, accounts []*Account) []*Account {
	if len(accounts) == 0 {
		return accounts
	}
	filtered := make([]*Account, 0, len(accounts))
	for _, acc := range accounts {
		if IsAccountAllowedForRequestContext(ctx, acc) {
			filtered = append(filtered, acc)
		}
	}
	return filtered
}
