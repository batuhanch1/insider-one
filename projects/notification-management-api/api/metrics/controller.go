package healthcheck

import (
	"github.com/gin-gonic/gin"
)

type Controller interface {
	Metrics(g *gin.Context)
}

type controller struct {
}

func NewController() Controller {
	return &controller{}
}

func (c *controller) Metrics(g *gin.Context) {
	return
}
