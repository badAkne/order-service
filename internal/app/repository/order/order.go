package porder

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/badAkne/order-service/internal/app/entity"
	"github.com/badAkne/order-service/internal/app/repository"
	rcpostgres "github.com/badAkne/order-service/internal/app/repository/conn/postgres"
	"github.com/badAkne/order-service/internal/app/repository/transaction"
	"github.com/badAkne/order-service/internal/app/util"
)

type repoPG struct {
	db *gorm.DB
	*transaction.Repository
}

func NewRepo(_ context.Context, client *rcpostgres.Client) (repository.Order, error) {
	txRepo := transaction.NewTxRepository(client.DB())

	return &repoPG{
		db:         client.DB(),
		Repository: txRepo,
	}, nil
}

func (r *repoPG) Create(ctx context.Context, order entity.Order) (entity.Order, error) {
	db := rcpostgres.GetTxFromCtx(ctx, r.db)

	err := db.WithContext(ctx).Model(&order).Create(&order).Error
	if err != nil {
		return order, util.ReplaceErr2(err, gorm.ErrForeignKeyViolated, entity.ErrNotFound, gorm.ErrDuplicatedKey, entity.ErrOrderDuplicate)
	}

	return order, nil
}

func (r *repoPG) Get(ctx context.Context, guid uuid.UUID) (entity.Order, error) {
	db := rcpostgres.GetTxFromCtx(ctx, r.db)

	var order entity.Order
	err := db.WithContext(ctx).Preload("Items").Where("guid = ?", guid).Find(&order).Error
	if err != nil {
		return entity.Order{}, util.ReplaceError(err, gorm.ErrRecordNotFound, entity.ErrNotFound)
	}

	return order, nil
}

func (r *repoPG) Update(ctx context.Context, guid uuid.UUID, status string) (entity.Order, error) {
	db := rcpostgres.GetTxFromCtx(ctx, r.db)

	var order entity.Order
	res := db.WithContext(ctx).Model(&order).Preload("Items").Where("guid = ?", guid).Update("status", status).Scan(&order)

	if res.Error != nil {
		return entity.Order{}, res.Error
	}

	if res.RowsAffected == 0 {
		return entity.Order{}, entity.ErrNotFound
	}

	return order, nil
}

func (r *repoPG) Delete(ctx context.Context, guid uuid.UUID) error {
	db := rcpostgres.GetTxFromCtx(ctx, r.db)

	res := db.WithContext(ctx).Delete(&entity.Order{}, guid)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return entity.ErrNotFound
	}

	return nil
}
