package rprocessor

import (
	"net/http"

	rhandler "github.com/badAkne/order-service/internal/app/handler"
	"github.com/gin-gonic/gin"
)

func v1GenericRegOrder(r *gin.RouterGroup, h rhandler.Order) {
	regRoute(r, http.MethodPost, "/order", h.Create, "api.v1.create_order")
	regRoute(r, http.MethodGet, "/order/{order_guid}", h.Get, "api.v1.get_order")
	regRoute(r, http.MethodDelete, "/order/{order_guid}", h.Delete, "api.v1.delete_order")
	regRoute(r, http.MethodPatch, "/order/{order_guid}", h.Update, "api.v1.update_order")
}
