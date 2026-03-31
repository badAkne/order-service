package rhealth

import (
	rhandler "github.com/badAkne/order-service/internal/app/handler"
	"github.com/gin-gonic/gin"
)

type handler struct{}

func NewHandler() rhandler.Health {
	return &handler{}
}

func (h *handler) LastCheck(c *gin.Context) {
	c.String(200, "ok")
}
