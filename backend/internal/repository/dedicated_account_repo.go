package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	dbaccount "github.com/Wei-Shaw/sub2api/ent/account"
	dbapikey "github.com/Wei-Shaw/sub2api/ent/apikey"
	dbgroup "github.com/Wei-Shaw/sub2api/ent/group"
	dbuser "github.com/Wei-Shaw/sub2api/ent/user"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/shopspring/decimal"
)

type dedicatedAccountRepository struct {
	client *dbent.Client
	sql    *sql.DB
}

func NewDedicatedAccountRepository(client *dbent.Client, sqlDB *sql.DB) service.DedicatedAccountRepository {
	return &dedicatedAccountRepository{
		client: client,
		sql:    sqlDB,
	}
}

func (r *dedicatedAccountRepository) ReleaseExpiredLocks(ctx context.Context, now time.Time) (int, error) {
	if r == nil || r.sql == nil {
		return 0, nil
	}
	rows, err := r.sql.QueryContext(ctx, `
		UPDATE accounts
		SET sale_status = $1,
		    locked_by_order_id = NULL,
		    locked_until = NULL,
		    updated_at = NOW()
		WHERE deleted_at IS NULL
		  AND sale_status = $2
		  AND owner_user_id IS NULL
		  AND locked_until IS NOT NULL
		  AND locked_until <= $3
		RETURNING id
	`, service.AccountSaleStatusListed, service.AccountSaleStatusLocked, now)
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()

	var unlockedIDs []int64
	for rows.Next() {
		var id int64
		if scanErr := rows.Scan(&id); scanErr == nil {
			unlockedIDs = append(unlockedIDs, id)
		}
	}
	if err := rows.Err(); err != nil {
		return len(unlockedIDs), err
	}
	for _, parentID := range unlockedIDs {
		_, _ = r.sql.ExecContext(ctx, `
			UPDATE accounts
			SET sale_status = $1, updated_at = NOW()
			WHERE parent_account_id = $2 AND deleted_at IS NULL
		`, service.AccountSaleStatusListed, parentID)
	}
	if len(unlockedIDs) > 0 {
		payload := map[string]any{"account_ids": unlockedIDs}
		if outboxErr := enqueueSchedulerOutbox(ctx, r.sql, service.SchedulerOutboxEventAccountBulkChanged, nil, nil, payload); outboxErr != nil {
			logger.LegacyPrintf("repository.dedicated_account", "[SchedulerOutbox] release expired locks outbox failed: %v", outboxErr)
		}
	}
	return len(unlockedIDs), nil
}

func (r *dedicatedAccountRepository) ListMarketAccounts(ctx context.Context, platform string) ([]service.Account, error) {
	_, _ = r.ReleaseExpiredLocks(ctx, time.Now())

	q := r.client.Account.Query().
		Where(
			dbaccount.DeletedAtIsNil(),
			dbaccount.ParentAccountIDIsNil(),
			dbaccount.OwnerUserIDIsNil(),
			dbaccount.StatusEQ(service.StatusActive),
			dbaccount.SaleStatusIn(service.AccountSaleStatusListed, service.AccountSaleStatusLocked),
		).
		WithGroups(func(gq *dbent.GroupQuery) {
			gq.Where(dbgroup.DeletedAtIsNil())
		}).
		Order(dbent.Desc(dbaccount.FieldListedAt), dbent.Desc(dbaccount.FieldID))

	if p := strings.TrimSpace(platform); p != "" {
		q = q.Where(dbaccount.PlatformEQ(p))
	}

	entities, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]service.Account, 0, len(entities))
	for _, entity := range entities {
		if mapped := dedicatedAccountEntityToService(entity); mapped != nil {
			out = append(out, *mapped)
		}
	}
	return out, nil
}

func (r *dedicatedAccountRepository) GetByID(ctx context.Context, id int64) (*service.Account, error) {
	entity, err := r.client.Account.Query().
		Where(dbaccount.IDEQ(id), dbaccount.DeletedAtIsNil()).
		WithGroups(func(gq *dbent.GroupQuery) {
			gq.Where(dbgroup.DeletedAtIsNil())
		}).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrDedicatedAccountNotFound
		}
		return nil, err
	}
	return dedicatedAccountEntityToService(entity), nil
}

func (r *dedicatedAccountRepository) ListUserDedicatedAccounts(ctx context.Context, userID int64) ([]service.Account, error) {
	entities, err := r.client.Account.Query().
		Where(
			dbaccount.DeletedAtIsNil(),
			dbaccount.ParentAccountIDIsNil(),
			dbaccount.OwnerUserIDEQ(userID),
		).
		WithGroups(func(gq *dbent.GroupQuery) {
			gq.Where(dbgroup.DeletedAtIsNil())
		}).
		Order(dbent.Desc(dbaccount.FieldSoldAt), dbent.Desc(dbaccount.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]service.Account, 0, len(entities))
	for _, entity := range entities {
		if mapped := dedicatedAccountEntityToService(entity); mapped != nil {
			out = append(out, *mapped)
		}
	}
	return out, nil
}

func (r *dedicatedAccountRepository) EnsureAccountHasGroup(ctx context.Context, accountID int64, platform string) (int64, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	groupID, err := r.ensureAccountHasGroupInTx(ctx, tx, accountID, platform)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	_ = enqueueSchedulerOutbox(ctx, r.sql, service.SchedulerOutboxEventAccountChanged, &accountID, &groupID, nil)
	return groupID, nil
}

func (r *dedicatedAccountRepository) ensureAccountHasGroupInTx(ctx context.Context, tx *dbent.Tx, accountID int64, platform string) (int64, error) {
	acc, err := tx.Account.Query().
		Where(dbaccount.IDEQ(accountID), dbaccount.DeletedAtIsNil()).
		WithGroups(func(gq *dbent.GroupQuery) {
			gq.Where(dbgroup.DeletedAtIsNil())
		}).
		Only(ctx)
	if err != nil {
		return 0, err
	}
	for _, g := range acc.Edges.Groups {
		if g != nil && g.Status == service.StatusActive {
			return g.ID, nil
		}
	}
	if len(acc.Edges.Groups) > 0 && acc.Edges.Groups[0] != nil {
		return acc.Edges.Groups[0].ID, nil
	}

	normPlatform := strings.TrimSpace(platform)
	if normPlatform == "" {
		normPlatform = acc.Platform
	}
	groupName := dedicatedDefaultGroupName(normPlatform)
	groupEntity, gErr := tx.Group.Query().
		Where(
			dbgroup.NameEQ(groupName),
			dbgroup.PlatformEQ(normPlatform),
			dbgroup.DeletedAtIsNil(),
		).
		First(ctx)
	if gErr != nil {
		if !dbent.IsNotFound(gErr) {
			return 0, gErr
		}
		desc := "系统自动创建的独享账号专属隔离分组"
		groupEntity, gErr = tx.Group.Create().
			SetName(groupName).
			SetDescription(desc).
			SetPlatform(normPlatform).
			SetRateMultiplier(1.0).
			SetIsExclusive(true).
			SetStatus(service.StatusActive).
			SetSubscriptionType(service.SubscriptionTypeStandard).
			Save(ctx)
		if gErr != nil {
			return 0, gErr
		}
	}

	_, execErr := tx.Client().ExecContext(ctx, `
		INSERT INTO account_groups (account_id, group_id, priority, created_at)
		VALUES ($1, $2, 1, NOW())
		ON CONFLICT (account_id, group_id) DO NOTHING
	`, accountID, groupEntity.ID)
	if execErr != nil {
		return 0, execErr
	}

	// 同步把该母账号的影子子账号也挂到同一独享组，保障 Spark 等同源影子正常调度
	_, _ = tx.Client().ExecContext(ctx, `
		INSERT INTO account_groups (account_id, group_id, priority, created_at)
		SELECT id, $2, 1, NOW()
		FROM accounts
		WHERE parent_account_id = $1 AND deleted_at IS NULL
		ON CONFLICT (account_id, group_id) DO NOTHING
	`, accountID, groupEntity.ID)

	return groupEntity.ID, nil
}

func dedicatedDefaultGroupName(platform string) string {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case service.PlatformAnthropic:
		return "Claude 独享专区"
	case service.PlatformOpenAI:
		return "OpenAI 独享专区"
	case service.PlatformGemini:
		return "Gemini 独享专区"
	default:
		return fmt.Sprintf("%s 独享专区", strings.ToUpper(strings.TrimSpace(platform)))
	}
}

func (r *dedicatedAccountRepository) LockAccountForOrderInTx(
	ctx context.Context,
	tx *dbent.Tx,
	accountID int64,
	orderID int64,
	lockUntil time.Time,
) (*service.Account, error) {
	now := time.Now()
	acc, err := tx.Account.Query().
		Where(dbaccount.IDEQ(accountID), dbaccount.DeletedAtIsNil()).
		ForUpdate().
		WithGroups(func(gq *dbent.GroupQuery) {
			gq.Where(dbgroup.DeletedAtIsNil())
		}).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrDedicatedAccountNotFound
		}
		return nil, err
	}
	if acc.ParentAccountID != nil || acc.Status != service.StatusActive || acc.OwnerUserID != nil {
		return nil, service.ErrDedicatedAccountUnavailable
	}
	saleStatus := service.NormalizeAccountSaleStatus(acc.SaleStatus)
	lockExpired := saleStatus == service.AccountSaleStatusLocked && acc.LockedUntil != nil && !acc.LockedUntil.After(now)
	if saleStatus != service.AccountSaleStatusListed && !lockExpired {
		return nil, service.ErrDedicatedAccountUnavailable
	}

	updated, err := tx.Account.UpdateOneID(acc.ID).
		SetSaleStatus(service.AccountSaleStatusLocked).
		SetLockedByOrderID(orderID).
		SetLockedUntil(lockUntil).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	_, _ = tx.Client().ExecContext(ctx, `
		UPDATE accounts
		SET sale_status = $1, updated_at = NOW()
		WHERE parent_account_id = $2 AND deleted_at IS NULL
	`, service.AccountSaleStatusLocked, acc.ID)
	_ = enqueueSchedulerOutbox(ctx, tx.Client(), service.SchedulerOutboxEventAccountChanged, &acc.ID, nil, nil)

	updated.Edges.Groups = acc.Edges.Groups
	return dedicatedAccountEntityToService(updated), nil
}

func (r *dedicatedAccountRepository) UnlockAccountForOrder(ctx context.Context, accountID int64, orderID int64) error {
	if r == nil || r.sql == nil {
		return nil
	}
	res, err := r.sql.ExecContext(ctx, `
		UPDATE accounts
		SET sale_status = $1,
		    locked_by_order_id = NULL,
		    locked_until = NULL,
		    updated_at = NOW()
		WHERE id = $2
		  AND deleted_at IS NULL
		  AND owner_user_id IS NULL
		  AND sale_status = $3
		  AND locked_by_order_id = $4
	`, service.AccountSaleStatusListed, accountID, service.AccountSaleStatusLocked, orderID)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected > 0 {
		_, _ = r.sql.ExecContext(ctx, `
			UPDATE accounts
			SET sale_status = $1, updated_at = NOW()
			WHERE parent_account_id = $2 AND deleted_at IS NULL
		`, service.AccountSaleStatusListed, accountID)
		_ = enqueueSchedulerOutbox(ctx, r.sql, service.SchedulerOutboxEventAccountChanged, &accountID, nil, nil)
	}
	return nil
}

func (r *dedicatedAccountRepository) FulfillOrderInTx(
	ctx context.Context,
	tx *dbent.Tx,
	accountID int64,
	buyerUserID int64,
	orderID int64,
	keyGen func() (string, error),
) (*service.Account, *service.APIKey, error) {
	return r.assignDedicatedAccountInTx(ctx, tx, accountID, buyerUserID, orderID, keyGen)
}

func (r *dedicatedAccountRepository) PurchaseWithBalance(
	ctx context.Context,
	buyerUserID int64,
	accountID int64,
	balanceMultiplier float64,
	keyGen func() (string, error),
) (*service.DedicatedBalancePurchaseOutcome, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now()
	accEntity, err := tx.Account.Query().
		Where(dbaccount.IDEQ(accountID), dbaccount.DeletedAtIsNil()).
		ForUpdate().
		WithGroups(func(gq *dbent.GroupQuery) {
			gq.Where(dbgroup.DeletedAtIsNil())
		}).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, service.ErrDedicatedAccountNotFound
		}
		return nil, err
	}

	accSvc := dedicatedAccountEntityToService(accEntity)
	if accSvc == nil || accSvc.ParentAccountID != nil || accSvc.Status != service.StatusActive || accSvc.OwnerUserID != nil {
		return nil, service.ErrDedicatedAccountUnavailable
	}
	saleStatus := service.NormalizeAccountSaleStatus(accSvc.SaleStatus)
	lockExpired := saleStatus == service.AccountSaleStatusLocked && accSvc.LockedUntil != nil && !accSvc.LockedUntil.After(now)
	if saleStatus != service.AccountSaleStatusListed && !lockExpired {
		return nil, service.ErrDedicatedAccountUnavailable
	}

	priceCNY := accSvc.CurrentSalePrice(now)
	if priceCNY < 0.01 {
		return nil, service.ErrDedicatedSaleConfigInvalid
	}
	if balanceMultiplier <= 0 {
		balanceMultiplier = 1.0
	}
	costBalance := decimal.NewFromFloat(priceCNY).
		Mul(decimal.NewFromFloat(balanceMultiplier)).
		Round(2).
		InexactFloat64()
	if costBalance < 0.01 {
		costBalance = 0.01
	}

	userEntity, err := tx.User.Query().
		Where(dbuser.IDEQ(buyerUserID), dbuser.DeletedAtIsNil()).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, err
	}
	if userEntity.Status != service.StatusActive {
		return nil, infraerrors.Forbidden("USER_INACTIVE", "用户账号已被禁用")
	}
	if decimal.NewFromFloat(userEntity.Balance).LessThan(decimal.NewFromFloat(costBalance)) {
		return nil, infraerrors.BadRequest(
			"INSUFFICIENT_BALANCE",
			fmt.Sprintf("账户可用余额不足（需 %.2f，当前 %.2f）", costBalance, userEntity.Balance),
		)
	}

	newBalance := decimal.NewFromFloat(userEntity.Balance).
		Sub(decimal.NewFromFloat(costBalance)).
		Round(8).
		InexactFloat64()
	if newBalance < 0 {
		newBalance = 0
	}
	if _, err := tx.User.UpdateOneID(userEntity.ID).SetBalance(newBalance).Save(ctx); err != nil {
		return nil, err
	}

	assignedAcc, createdKey, err := r.assignLockedAccountEntityInTx(ctx, tx, accEntity, buyerUserID, now, keyGen)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &service.DedicatedBalancePurchaseOutcome{
		Account:          assignedAcc,
		APIKey:           createdKey,
		PriceCNY:         priceCNY,
		BalanceDeducted:  costBalance,
		RemainingBalance: newBalance,
	}, nil
}

func (r *dedicatedAccountRepository) assignDedicatedAccountInTx(
	ctx context.Context,
	tx *dbent.Tx,
	accountID int64,
	buyerUserID int64,
	orderID int64,
	keyGen func() (string, error),
) (*service.Account, *service.APIKey, error) {
	now := time.Now()
	accEntity, err := tx.Account.Query().
		Where(dbaccount.IDEQ(accountID), dbaccount.DeletedAtIsNil()).
		ForUpdate().
		WithGroups(func(gq *dbent.GroupQuery) {
			gq.Where(dbgroup.DeletedAtIsNil())
		}).
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil, service.ErrDedicatedAccountNotFound
		}
		return nil, nil, err
	}

	// 幂等保护：同一笔订单重复回调时直接返回既有归属与专属 Key
	if accEntity.OwnerUserID != nil && *accEntity.OwnerUserID == buyerUserID &&
		service.NormalizeAccountSaleStatus(accEntity.SaleStatus) == service.AccountSaleStatusSold {
		existingKey, _ := tx.APIKey.Query().
			Where(
				dbapikey.UserIDEQ(buyerUserID),
				dbapikey.DedicatedAccountIDEQ(accountID),
				dbapikey.DeletedAtIsNil(),
			).
			First(ctx)
		var keySvc *service.APIKey
		if existingKey != nil {
			keySvc = apiKeyEntityToService(existingKey)
		}
		return dedicatedAccountEntityToService(accEntity), keySvc, nil
	}

	if accEntity.ParentAccountID != nil || accEntity.OwnerUserID != nil {
		return nil, nil, service.ErrDedicatedAccountUnavailable
	}
	saleStatus := service.NormalizeAccountSaleStatus(accEntity.SaleStatus)
	switch saleStatus {
	case service.AccountSaleStatusLocked:
		if orderID > 0 && accEntity.LockedByOrderID != nil && *accEntity.LockedByOrderID != orderID {
			if accEntity.LockedUntil == nil || accEntity.LockedUntil.After(now) {
				return nil, nil, service.ErrDedicatedAccountUnavailable
			}
		}
	case service.AccountSaleStatusListed:
		// 允许订单超时刚解锁、在 5 分钟宽限期内晚到支付回调恢复履约（前提是未被他人锁单/买走）
	default:
		return nil, nil, service.ErrDedicatedAccountUnavailable
	}

	return r.assignLockedAccountEntityInTx(ctx, tx, accEntity, buyerUserID, now, keyGen)
}

func (r *dedicatedAccountRepository) assignLockedAccountEntityInTx(
	ctx context.Context,
	tx *dbent.Tx,
	accEntity *dbent.Account,
	buyerUserID int64,
	now time.Time,
	keyGen func() (string, error),
) (*service.Account, *service.APIKey, error) {
	groupID, err := r.ensureAccountHasGroupInTx(ctx, tx, accEntity.ID, accEntity.Platform)
	if err != nil {
		return nil, nil, err
	}

	updatedAcc, err := tx.Account.UpdateOneID(accEntity.ID).
		SetOwnerUserID(buyerUserID).
		SetSaleStatus(service.AccountSaleStatusSold).
		SetSoldAt(now).
		SetSchedulable(true).
		ClearLockedByOrderID().
		ClearLockedUntil().
		Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	// 同步该独享母账号派生的 Spark 影子账号归属与售出状态
	_, _ = tx.Client().ExecContext(ctx, `
		UPDATE accounts
		SET owner_user_id = $1,
		    sale_status = $2,
		    sold_at = $3,
		    schedulable = TRUE,
		    locked_by_order_id = NULL,
		    locked_until = NULL,
		    updated_at = NOW()
		WHERE parent_account_id = $4 AND deleted_at IS NULL
	`, buyerUserID, service.AccountSaleStatusSold, now, accEntity.ID)

	reloadedAcc, reloadErr := tx.Account.Query().
		Where(dbaccount.IDEQ(updatedAcc.ID)).
		WithGroups(func(gq *dbent.GroupQuery) {
			gq.Where(dbgroup.DeletedAtIsNil())
		}).
		Only(ctx)
	if reloadErr == nil && reloadedAcc != nil {
		updatedAcc = reloadedAcc
	}
	accSvc := dedicatedAccountEntityToService(updatedAcc)

	// 自动生成一把默认 1 对 1 绑定的专属 API Key，买完即刻复制可用
	var keyStr string
	for attempt := 0; attempt < 5; attempt++ {
		candidate, genErr := keyGen()
		if genErr != nil {
			return nil, nil, genErr
		}
		exists, existErr := tx.APIKey.Query().Where(dbapikey.KeyEQ(candidate)).Exist(ctx)
		if existErr != nil {
			return nil, nil, existErr
		}
		if !exists {
			keyStr = candidate
			break
		}
	}
	if keyStr == "" {
		return nil, nil, fmt.Errorf("failed to allocate unique dedicated api key")
	}

	keyName := truncateRunes(fmt.Sprintf("独享-%s", accSvc.PublicDisplayName()), 80)
	createdKeyEntity, err := tx.APIKey.Create().
		SetUserID(buyerUserID).
		SetKey(keyStr).
		SetName(keyName).
		SetGroupID(groupID).
		SetDedicatedAccountID(accEntity.ID).
		SetDedicatedPoolMode(false).
		SetStatus(service.StatusActive).
		Save(ctx)
	if err != nil {
		return nil, nil, err
	}

	_ = enqueueSchedulerOutbox(ctx, tx.Client(), service.SchedulerOutboxEventAccountChanged, &accEntity.ID, &groupID, nil)
	return accSvc, apiKeyEntityToService(createdKeyEntity), nil
}

func (r *dedicatedAccountRepository) RevokeOwnershipOnRefund(
	ctx context.Context,
	accountID int64,
	expectedOwnerUserID int64,
) ([]string, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	accEntity, err := tx.Account.Query().
		Where(dbaccount.IDEQ(accountID), dbaccount.DeletedAtIsNil()).
		ForUpdate().
		Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	if accEntity.OwnerUserID == nil || *accEntity.OwnerUserID != expectedOwnerUserID {
		return nil, nil
	}

	if _, err := tx.Account.UpdateOneID(accountID).
		ClearOwnerUserID().
		SetSaleStatus(service.AccountSaleStatusNone).
		SetSchedulable(false).
		ClearSoldAt().
		ClearLockedByOrderID().
		ClearLockedUntil().
		Save(ctx); err != nil {
		return nil, err
	}

	_, _ = tx.Client().ExecContext(ctx, `
		UPDATE accounts
		SET owner_user_id = NULL,
		    sale_status = $1,
		    schedulable = FALSE,
		    sold_at = NULL,
		    updated_at = NOW()
		WHERE parent_account_id = $2 AND deleted_at IS NULL
	`, service.AccountSaleStatusNone, accountID)

	pinnedKeys, err := tx.APIKey.Query().
		Where(
			dbapikey.UserIDEQ(expectedOwnerUserID),
			dbapikey.DedicatedAccountIDEQ(accountID),
			dbapikey.DeletedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}
	disabledKeyStrings := make([]string, 0, len(pinnedKeys))
	for _, k := range pinnedKeys {
		disabledKeyStrings = append(disabledKeyStrings, k.Key)
		_, _ = tx.APIKey.UpdateOneID(k.ID).
			SetStatus(service.StatusAPIKeyInactive).
			ClearDedicatedAccountID().
			SetDedicatedPoolMode(false).
			Save(ctx)
	}

	_ = enqueueSchedulerOutbox(ctx, tx.Client(), service.SchedulerOutboxEventAccountChanged, &accountID, nil, nil)

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return disabledKeyStrings, nil
}

func truncateRunes(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxRunes])
}

func dedicatedAccountEntityToService(m *dbent.Account) *service.Account {
	out := accountEntityToService(m)
	if out == nil || m == nil || len(m.Edges.Groups) == 0 {
		return out
	}
	groups := make([]*service.Group, 0, len(m.Edges.Groups))
	groupIDs := make([]int64, 0, len(m.Edges.Groups))
	for _, g := range m.Edges.Groups {
		if g == nil {
			continue
		}
		if mapped := groupEntityToService(g); mapped != nil {
			groups = append(groups, mapped)
		}
		groupIDs = append(groupIDs, g.ID)
	}
	out.Groups = groups
	out.GroupIDs = groupIDs
	return out
}

