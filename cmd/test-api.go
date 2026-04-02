package cmd

import (
	"context"
	"fmt"
	"insider-one/infrastructure/adapters/client"
	"insider-one/infrastructure/config"
	"insider-one/infrastructure/middleware"
	prometheusWrapper "insider-one/infrastructure/prometheus"
	_ "insider-one/projects/notification-management-api"
	test_api "insider-one/projects/test-api"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

// @title           Notification Management API
// @version         1.0
// @description     REST API for managing email, SMS, and push notifications
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey ApiKeyAuth
// @in              header
// @name            Authorization

var testApiCmd = &cobra.Command{
	Use:   "test-api",
	Short: "",
	Long:  ``,
	RunE:  testApiCmdRun,
}

func init() {
	rootCmd.AddCommand(testApiCmd)
}

func testApiCmdRun(cmd *cobra.Command, args []string) (err error) {
	ctx := context.Background()

	cfg, err := config.Load(ctx, cmd.Use, env)
	if err != nil {
		return err
	}

	httpClient := client.NewClient()
	emailController := test_api.NewEmailController(httpClient)
	smsController := test_api.NewSmsController(httpClient)
	pushController := test_api.NewPushController(httpClient)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	prometheusWrapper := prometheusWrapper.InitForAPI()
	router.Use(middleware.PromMiddleware(prometheusWrapper))
	router.Use(middleware.CorrelationID())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.StartTime())

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := router.Group("/test")
	api.Use()
	{
		v1 := api.Group("/api")
		{
			email := v1.Group("/email")
			email.POST("/", emailController.Send)
			email.POST("/batch", emailController.SendBatch)
			email.PUT("/cancel", emailController.Cancel)

			sms := v1.Group("/sms")
			sms.POST("/", smsController.Send)
			sms.POST("/batch", smsController.SendBatch)
			sms.PUT("/cancel", smsController.Cancel)

			push := v1.Group("/push")
			push.POST("/", pushController.Send)
			push.POST("/batch", pushController.SendBatch)
			push.PUT("/cancel", pushController.Cancel)
		}
	}

	fmt.Println(fmt.Sprintf("%s Starting.", cfg.App.Name))
	err = router.Run(fmt.Sprintf(":%v", cfg.App.Port))
	return err
}
