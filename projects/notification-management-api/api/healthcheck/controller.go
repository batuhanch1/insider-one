package healthcheck

import (
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Controller interface {
	HealthCheck(g *gin.Context)
}

type controller struct {
	pool           *pgxpool.Pool
	rabbitMqClient *rabbitmq.Client
}

func NewController(pool *pgxpool.Pool, rabbitMqClient *rabbitmq.Client) Controller {
	return &controller{pool, rabbitMqClient}
}

func (c *controller) HealthCheck(g *gin.Context) {
	ctx := g.Copy()

	var dbStatus bool
	if err := c.pool.Ping(ctx); err == nil {
		dbStatus = true
	}

	rabbitStatus := c.rabbitMqClient.IsAlive()

	g.JSON(http.StatusOK, gin.H{
		"status":   "ok",
		"db":       dbStatus,
		"rabbitMQ": rabbitStatus,
	})
}
