package cmd

import (
	"context"
	"fmt"
	emailCommand "insider-one/application/command/notification/email"
	pushCommand "insider-one/application/command/notification/push"
	smsCommand "insider-one/application/command/notification/sms"
	emailQuery "insider-one/application/query/notification/email"
	pushQuery "insider-one/application/query/notification/push"
	smsQuery "insider-one/application/query/notification/sms"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/adapters/persistence/postgresql"
	emailPersistence "insider-one/infrastructure/adapters/persistence/postgresql/notification/email"
	pushPersistence "insider-one/infrastructure/adapters/persistence/postgresql/notification/push"
	smsPersistence "insider-one/infrastructure/adapters/persistence/postgresql/notification/sms"
	"insider-one/infrastructure/config"
	"insider-one/infrastructure/middleware"
	prometheusWrapper "insider-one/infrastructure/prometheus"
	_ "insider-one/projects/notification-management-api"
	notification_management_api "insider-one/projects/notification-management-api"
	"insider-one/projects/notification-management-api/api/healthcheck"
	emailController "insider-one/projects/notification-management-api/api/v1/email"
	pushController "insider-one/projects/notification-management-api/api/v1/push"
	smsController "insider-one/projects/notification-management-api/api/v1/sms"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Notification Management API
// @version         1.0
// @description     REST API for managing email, SMS, and push notifications
// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey ApiKeyAuth
// @in              header
// @name            Authorization

var notificationManagementApiCmd = &cobra.Command{
	Use:   "notification-management-api",
	Short: "",
	Long:  ``,
	RunE:  notificationManagementApiCmdRun,
}

func init() {
	rootCmd.AddCommand(notificationManagementApiCmd)
}

func notificationManagementApiCmdRun(cmd *cobra.Command, args []string) (err error) {
	ctx := context.Background()

	env, _ := cmd.Flags().GetString("env")
	cfg, err := config.Load(ctx, cmd.Use, env)
	if err != nil {
		return err
	}

	pool, err := postgresql.Connect(cfg.DB, ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	err = postgresql.Migrate(ctx, pool)
	if err != nil {
		return err
	}

	rabbitMqClient, err := rabbitmq.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer rabbitMqClient.Close()

	publisher := rabbitmq.NewPublisher(rabbitMqClient)
	batchPublisher := rabbitmq.NewBatchPublisher(rabbitMqClient)

	emailQueryRepository := emailPersistence.NewQueryRepository(pool)

	sendEmailCommand := emailCommand.NewSendCommand(publisher)
	emailEnqueueCancelCommand := emailCommand.NewEnqueueCancelCommand(emailQueryRepository, publisher)
	sendBatchEmailCommand := emailCommand.NewSendBatchCommand(batchPublisher)

	getAllEmailQuery := emailQuery.NewGetAllQuery(emailQueryRepository)
	getStatusByIDEmailQuery := emailQuery.NewGetStatusByIDQuery(emailQueryRepository)
	getStatusByBatchIDEmailQuery := emailQuery.NewGetStatusByBatchIDQuery(emailQueryRepository)
	emailController := emailController.NewController(sendEmailCommand, sendBatchEmailCommand, emailEnqueueCancelCommand, getAllEmailQuery, getStatusByBatchIDEmailQuery, getStatusByIDEmailQuery)

	smsQueryRepository := smsPersistence.NewQueryRepository(pool)

	sendSmsCommand := smsCommand.NewSendCommand(publisher)
	smsEnqueueCancelCommand := smsCommand.NewEnqueueCancelCommand(smsQueryRepository, publisher)
	sendBatchSmsCommand := smsCommand.NewSendBatchCommand(batchPublisher)

	getAllSmsQuery := smsQuery.NewGetAllQuery(smsQueryRepository)
	getStatusByIDSmsQuery := smsQuery.NewGetStatusByIDQuery(smsQueryRepository)
	getStatusByBatchIDSmsQuery := smsQuery.NewGetStatusByBatchIDQuery(smsQueryRepository)
	smsController := smsController.NewController(sendSmsCommand, sendBatchSmsCommand, smsEnqueueCancelCommand, getAllSmsQuery, getStatusByBatchIDSmsQuery, getStatusByIDSmsQuery)

	pushQueryRepository := pushPersistence.NewQueryRepository(pool)

	sendPushCommand := pushCommand.NewSendCommand(publisher)
	pushEnqueueCancelCommand := pushCommand.NewEnqueueCancelCommand(pushQueryRepository, publisher)
	sendBatchPushCommand := pushCommand.NewSendBatchCommand(batchPublisher)

	getAllPushQuery := pushQuery.NewGetAllQuery(pushQueryRepository)
	getStatusByIDPushQuery := pushQuery.NewGetStatusByIDQuery(pushQueryRepository)
	getStatusByBatchIDPushQuery := pushQuery.NewGetStatusByBatchIDQuery(pushQueryRepository)
	pushController := pushController.NewController(sendPushCommand, sendBatchPushCommand, pushEnqueueCancelCommand, getAllPushQuery, getStatusByBatchIDPushQuery, getStatusByIDPushQuery)

	gin.SetMode(gin.ReleaseMode)
	notification_management_api.RegisterValidators()
	router := gin.Default()

	prometheusWrapper := prometheusWrapper.InitForAPI()
	router.Use(middleware.PromMiddleware(prometheusWrapper))
	router.Use(middleware.CorrelationID())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.StartTime())

	healthcheckController := healthcheck.NewController(pool, rabbitMqClient)
	// Public routes -- no auth required
	router.GET("/health", healthcheckController.HealthCheck)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Private routes -- Needs Auth
	privateAPI := router.Group("/api")
	privateAPI.Use(middleware.AuthRequired())
	{
		v1 := privateAPI.Group("/v1")
		{
			email := v1.Group("/email")
			email.POST("/", emailController.Send)
			email.POST("/batch", emailController.SendBatch)
			email.GET("/", emailController.List)
			email.PUT("/cancel", emailController.Cancel)
			email.GET("/status", emailController.GetStatusByID)
			email.POST("/status/batch", emailController.GetStatusByIDs)

			sms := v1.Group("/sms")
			sms.POST("/", smsController.Send)
			sms.POST("/batch", smsController.SendBatch)
			sms.GET("/", smsController.List)
			sms.PUT("/cancel", smsController.Cancel)
			sms.GET("/status", smsController.GetStatusByID)
			sms.POST("/status/batch", smsController.GetStatusByIDs)

			push := v1.Group("/push")
			push.POST("/", pushController.Send)
			push.POST("/batch", pushController.SendBatch)
			push.GET("/", pushController.List)
			push.PUT("/cancel", pushController.Cancel)
			push.GET("/status", pushController.GetStatusByID)
			push.POST("/status/batch", pushController.GetStatusByIDs)
		}
	}

	fmt.Println(fmt.Sprintf("%s Starting.", cfg.App.Name))
	err = router.Run(fmt.Sprintf(":%v", cfg.App.Port))
	return err
}
