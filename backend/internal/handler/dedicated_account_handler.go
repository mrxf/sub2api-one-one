package handler

import (
	"context"
	"io"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// DedicatedAccountHandler 处理用户端独享账号商城、钱包买断、只读 5H/7D 看板与官方重置卡接口。
type DedicatedAccountHandler struct {
	svc *service.DedicatedAccountService
}

// NewDedicatedAccountHandler 创建用户端独享账号 Handler。
func NewDedicatedAccountHandler(svc *service.DedicatedAccountService) *DedicatedAccountHandler {
	return &DedicatedAccountHandler{svc: svc}
}

func requireAuthenticatedUserID(c *gin.Context) (int64, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "User not authenticated")
		return 0, false
	}
	return subject.UserID, true
}

// ListMarket 获取商城当前上架中的独享买断账号列表。
// GET /api/v1/dedicated-accounts/market
func (h *DedicatedAccountHandler) ListMarket(c *gin.Context) {
	if _, ok := requireAuthenticatedUserID(c); !ok {
		return
	}
	items, err := h.svc.ListMarketAccounts(c.Request.Context(), c.Query("platform"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
}

// BuyWithBalance 使用平台账户钱包余额原子买断上架中的独享账号，并自动创建默认 1 对 1 专属 API Key。
// POST /api/v1/dedicated-accounts/market/:id/buy-with-balance
func (h *DedicatedAccountHandler) BuyWithBalance(c *gin.Context) {
	userID, ok := requireAuthenticatedUserID(c)
	if !ok {
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "Invalid dedicated account ID")
		return
	}

	if strings.TrimSpace(c.GetHeader("Idempotency-Key")) != "" {
		payload := struct {
			AccountID int64 `json:"account_id"`
		}{AccountID: accountID}
		executeUserIdempotentJSON(
			c,
			"user.dedicated_accounts.buy_with_balance",
			payload,
			service.DefaultWriteIdempotencyTTL(),
			func(ctx context.Context) (any, error) {
				return h.svc.PurchaseWithBalance(ctx, userID, accountID)
			},
		)
		return
	}

	res, err := h.svc.PurchaseWithBalance(c.Request.Context(), userID, accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, res)
}

// BuyoutWithBalance 使用平台钱包余额原子买断独享账号（支持商城直接买断或已持有独享账号升级买断），
// 成功后交付纯文本账号信息并永久标记免封号售后。
// POST /api/v1/dedicated-accounts/:id/buyout-with-balance
func (h *DedicatedAccountHandler) BuyoutWithBalance(c *gin.Context) {
	userID, ok := requireAuthenticatedUserID(c)
	if !ok {
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "Invalid dedicated account ID")
		return
	}

	if strings.TrimSpace(c.GetHeader("Idempotency-Key")) != "" {
		payload := struct {
			AccountID int64 `json:"account_id"`
		}{AccountID: accountID}
		executeUserIdempotentJSON(
			c,
			"user.dedicated_accounts.buyout_with_balance",
			payload,
			service.DefaultWriteIdempotencyTTL(),
			func(ctx context.Context) (any, error) {
				return h.svc.BuyoutWithBalance(ctx, userID, accountID)
			},
		)
		return
	}

	res, err := h.svc.BuyoutWithBalance(c.Request.Context(), userID, accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, res)
}

// GetBuyoutInfo 仅在确认当前用户为账号归属人且已完成买断后，返回买断交付信息（纯文本）。
// GET /api/v1/dedicated-accounts/:id/buyout-info
func (h *DedicatedAccountHandler) GetBuyoutInfo(c *gin.Context) {
	userID, ok := requireAuthenticatedUserID(c)
	if !ok {
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "Invalid dedicated account ID")
		return
	}
	info, err := h.svc.GetUserBoughtOutAccountInfo(c.Request.Context(), userID, accountID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, info)
}

// ListMyAccounts 返回用户名下全部已购买/已分配独享账号的只读安全视图（严格隐藏邮箱、凭据、代理、内部备注等敏感字段）。
// GET /api/v1/dedicated-accounts
func (h *DedicatedAccountHandler) ListMyAccounts(c *gin.Context) {
	userID, ok := requireAuthenticatedUserID(c)
	if !ok {
		return
	}
	items, err := h.svc.ListUserDedicatedAccounts(c.Request.Context(), userID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"items": items})
}

type dedicatedAccountsUsageRequest struct {
	AccountIDs   []int64 `json:"account_ids"`
	ForceRefresh bool    `json:"force_refresh"`
}

// GetMyAccountsUsage 批量查询用户名下独享账号的实时 5H / 7D 剩余额度与重置倒计时。
// POST /api/v1/dedicated-accounts/usage
func (h *DedicatedAccountHandler) GetMyAccountsUsage(c *gin.Context) {
	userID, ok := requireAuthenticatedUserID(c)
	if !ok {
		return
	}
	var req dedicatedAccountsUsageRequest
	if err := c.ShouldBindJSON(&req); err != nil && err != io.EOF {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	quotas, err := h.svc.GetUserDedicatedAccountsUsage(c.Request.Context(), userID, req.AccountIDs, req.ForceRefresh)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"quotas": quotas})
}

// GetResetCredits 查询指定自有独享账号是否持有可用的官方重置卡。
// GET /api/v1/dedicated-accounts/:id/reset-credits
func (h *DedicatedAccountHandler) GetResetCredits(c *gin.Context) {
	userID, ok := requireAuthenticatedUserID(c)
	if !ok {
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "Invalid dedicated account ID")
		return
	}
	refresh := c.Query("refresh") == "true" || c.Query("refresh") == "1"
	status, err := h.svc.GetUserAccountResetCredits(c.Request.Context(), userID, accountID, refresh)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, status)
}

// RedeemResetCredit 消耗指定自有独享账号持有的 1 张官方重置卡并立即恢复限流窗口。
// POST /api/v1/dedicated-accounts/:id/reset-credits/redeem
func (h *DedicatedAccountHandler) RedeemResetCredit(c *gin.Context) {
	userID, ok := requireAuthenticatedUserID(c)
	if !ok {
		return
	}
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || accountID <= 0 {
		response.BadRequest(c, "Invalid dedicated account ID")
		return
	}
	res, err := h.svc.RedeemUserAccountResetCredit(
		c.Request.Context(),
		userID,
		accountID,
		c.GetHeader("Idempotency-Key"),
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, res)
}
