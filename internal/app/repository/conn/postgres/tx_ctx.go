package rcpostgres

import (
	"context"

	"gorm.io/gorm"
)

type _ctxKeyTx struct{}

func GetTxFromCtx(ctx context.Context, db *gorm.DB) *gorm.DB {
	tx, ok := ctx.Value(_ctxKeyTx{}).(*gorm.DB)
	if ok {
		return tx
	}

	return db
}

func CtxWithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, _ctxKeyTx{}, tx)
}
