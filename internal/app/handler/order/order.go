package rorder

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/badAkne/order-service/internal/app/entity"
	rhandler "github.com/badAkne/order-service/internal/app/handler"
	rservice "github.com/badAkne/order-service/internal/app/service"
	"github.com/badAkne/order-service/internal/pkg/http/httph"
)

type handlerOrder struct {
	serviceOrder rservice.Order
}

func NewHandler(serviceOrder rservice.Order) rhandler.Order {
	return &handlerOrder{
		serviceOrder: serviceOrder,
	}
}

func (h *handlerOrder) Create(c *gin.Context) {
	var req entity.RequestOrderCreate

	if err := c.ShouldBindJSON(&req); err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}

	response, err := h.serviceOrder.Create(c.Request.Context(), req)
	if err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}

	httph.SendJSON(c.Writer, http.StatusCreated, response)
}

func (h *handlerOrder) Get(c *gin.Context) {
	guid, err := uuid.Parse(c.Param("order_guid"))
	if err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}

	res, err := h.serviceOrder.Get(c.Request.Context(), guid)
	if err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}

	httph.SendJSON(c.Writer, http.StatusOK, res)
}

func (h *handlerOrder) Update(c *gin.Context) {
	guid, err := uuid.Parse(c.Param("order_guid"))
	if err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}

	var req entity.RequestUpdateOrder
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		httph.ErrorApply(c.Request, err)
	}

	res, err := h.serviceOrder.Update(c.Request.Context(), guid, req.Status)
	if err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}

	httph.SendJSON(c.Writer, http.StatusOK, res)
}

func (h *handlerOrder) Delete(c *gin.Context) {
	guid, err := uuid.Parse(c.Param("order_guid"))
	if err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}

	err = h.serviceOrder.Delete(c.Request.Context(), guid)
	if err != nil {
		httph.ErrorApply(c.Request, err)
		return
	}

	httph.SendEmpty(c.Writer, http.StatusNoContent)
}
