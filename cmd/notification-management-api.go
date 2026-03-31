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
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/adapters/persistence/postgresql"
	emailPersistence "insider-one/infrastructure/adapters/persistence/postgresql/notification/email"
	pushPersistence "insider-one/infrastructure/adapters/persistence/postgresql/notification/push"
	smsPersistence "insider-one/infrastructure/adapters/persistence/postgresql/notification/sms"
	"insider-one/infrastructure/config"
	"insider-one/infrastructure/middleware"
	"insider-one/projects/notification-management-api/api/healthcheck"
	emailController "insider-one/projects/notification-management-api/api/v1/email"
	pushController "insider-one/projects/notification-management-api/api/v1/push"
	smsController "insider-one/projects/notification-management-api/api/v1/sms"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

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

	cfg, err := config.Load(cmd.Use, env)
	if err != nil {
		return err
	}

	pool, err := postgresql.Connect(cfg.DB, ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	rabbitMqClient, err := rabbitmq2.New(cfg)
	if err != nil {
		return err
	}
	defer rabbitMqClient.Close()

	publisher := rabbitmq2.NewPublisher(rabbitMqClient)
	batchPublisherChannel, err := rabbitMqClient.Channel()
	if err != nil {
		return err
	}

	batchPublisher := rabbitmq2.NewBatchPublisher(batchPublisherChannel)

	emailRepository := emailPersistence.NewRepository(pool)

	sendEmailCommand := emailCommand.NewSendCommand(publisher)
	cancelEmailCommand := emailCommand.NewCancelCommand(emailRepository, *publisher)
	sendBatchEmailCommand := emailCommand.NewSendBatchCommand(batchPublisher)

	getAllEmailQuery := emailQuery.NewGetAllQuery(emailRepository)
	getStatusByIDEmailQuery := emailQuery.NewGetStatusByIDQuery(emailRepository)
	getStatusByBatchIDEmailQuery := emailQuery.NewGetStatusByBatchIDQuery(emailRepository)
	emailController := emailController.NewController(sendEmailCommand, sendBatchEmailCommand, cancelEmailCommand, getAllEmailQuery, getStatusByBatchIDEmailQuery, getStatusByIDEmailQuery)

	smsRepository := smsPersistence.NewRepository(pool)

	sendSmsCommand := smsCommand.NewSendCommand(publisher)
	cancelSmsCommand := smsCommand.NewCancelCommand(smsRepository, *publisher)
	sendBatchSmsCommand := smsCommand.NewSendBatchCommand(batchPublisher)

	getAllSmsQuery := smsQuery.NewGetAllQuery(smsRepository)
	getStatusByIDSmsQuery := smsQuery.NewGetStatusByIDQuery(smsRepository)
	getStatusByBatchIDSmsQuery := smsQuery.NewGetStatusByBatchIDQuery(smsRepository)
	smsController := smsController.NewController(sendSmsCommand, sendBatchSmsCommand, cancelSmsCommand, getAllSmsQuery, getStatusByBatchIDSmsQuery, getStatusByIDSmsQuery)

	pushRepository := pushPersistence.NewRepository(pool)

	sendPushCommand := pushCommand.NewSendCommand(publisher)
	cancelPushCommand := pushCommand.NewCancelCommand(pushRepository, *publisher)
	sendBatchPushCommand := pushCommand.NewSendBatchCommand(batchPublisher)

	getAllPushQuery := pushQuery.NewGetAllQuery(pushRepository)
	getStatusByIDPushQuery := pushQuery.NewGetStatusByIDQuery(pushRepository)
	getStatusByBatchIDPushQuery := pushQuery.NewGetStatusByBatchIDQuery(pushRepository)
	pushController := pushController.NewController(sendPushCommand, sendBatchPushCommand, cancelPushCommand, getAllPushQuery, getStatusByBatchIDPushQuery, getStatusByIDPushQuery)

	router := gin.Default()

	healthcheckController := healthcheck.NewController(pool, rabbitMqClient)
	// Public routes -- no auth required
	router.GET("/health", healthcheckController.HealthCheck)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

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

			sms := v1.Group("/sms")
			sms.POST("/", smsController.Send)
			sms.POST("/batch", smsController.SendBatch)
			sms.GET("/", smsController.List)
			sms.PUT("/cancel", smsController.Cancel)

			push := v1.Group("/push")
			push.POST("/", pushController.Send)
			push.POST("/batch", pushController.SendBatch)
			push.GET("/", pushController.List)
			push.PUT("/cancel", pushController.Cancel)
		}
	}

	fmt.Println(fmt.Sprintf("%s Starting.", cfg.App.Name))
	err = router.Run(fmt.Sprintf(":%v", cfg.App.Port))
	return err
}
