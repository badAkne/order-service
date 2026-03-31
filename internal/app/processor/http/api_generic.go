package rprocessor

import (
	"net/http"

	rhandler "github.com/badAkne/order-service/internal/app/handler"
	"github.com/gin-gonic/gin"
)

func GenericRegHealthCheck(r *gin.Engine, h rhandler.Health) {
	regRoute(r, http.MethodGet, "/health", h.LastCheck, "api.generic.health_check")
}
