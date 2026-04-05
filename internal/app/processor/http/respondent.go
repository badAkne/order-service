package rprocessor

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/badAkne/order-service/internal/app/entity"
	"github.com/badAkne/order-service/internal/pkg/http/binding"
	"github.com/badAkne/order-service/internal/pkg/http/respondent"
)

func makeErrorMiddleware() gin.HandlerFunc {
	const (
		_40001 = "Bad request"
		_40002 = "Заказ с таким номером уже существует"
		_40003 = "Товар не найден"
		_40004 = "Некорректный статус заказа"
		_40401 = "Not found"
		_50001 = "Internal server error"
	)

	const (
		_40401D = "Entity was deleted or never exist"
		_50001D = "Try again later"
	)

	makeFallbackExtractor := func(status, errorCode int, message, detail string) respondent.ManifestExtractor {
		genericManifest := respondent.Manifest{
			Status:      status,
			ErrorCode:   errorCode,
			Error:       message,
			ErrorDetail: detail,
		}

		return func(_ error) *respondent.Manifest { return &genericManifest }
	}

	expander := respondent.NewSimpleExpander().
		ExtractFor(
			binding.ErrValidationFailed,
			binding.NewRespondentManifestExtractor(http.StatusBadRequest, 40001, _40001)).
		WithoutDetail(entity.ErrOrderDuplicate, http.StatusBadRequest, 40002, _40002).
		WithoutDetail(entity.ErrProductNotFound, http.StatusBadRequest, 40003, _40003).
		WithoutDetail(entity.ErrInvalidOrderStatus, http.StatusBadRequest, 40004, _40004).
		WithDetail(entity.ErrNotFound, http.StatusNotFound, 40401, _40401, _40401D).
		FallbackExtractor(makeFallbackExtractor(http.StatusInternalServerError, 50001, _50001, _50001D))

	applicator := respondent.NewSimpleApplicator()

	return respondent.NewMiddleware(expander, applicator)
}
