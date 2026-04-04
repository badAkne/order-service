package transaction

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	rcpostgres "github.com/badAkne/order-service/internal/app/repository/conn/postgres"
)

type Repository struct {
	db *gorm.DB
}

func NewTxRepository(db *gorm.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) OpenTx(ctx context.Context, fn func(c context.Context) error) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		c := rcpostgres.CtxWithTx(ctx, tx)
		return fn(c)
	})
	if err != nil {
		return fmt.Errorf("unable to procces transaction: %w", err)
	}

	return nil
}
