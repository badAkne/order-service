package rhealth

import (
	"github.com/gin-gonic/gin"

	rhandler "github.com/badAkne/order-service/internal/app/handler"
)

type handler struct{}

func NewHandler() rhandler.Health {
	return &handler{}
}

func (h *handler) LastCheck(c *gin.Context) {
	c.String(200, "ok")
}
