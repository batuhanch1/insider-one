package cmd

import (
	"context"
	"errors"
	"fmt"
	SmsCommand "insider-one/application/command/notification/sms"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/adapters/messaging/rabbitmq/handler"
	"insider-one/infrastructure/adapters/persistence/postgresql"
	SmsPersistence "insider-one/infrastructure/adapters/persistence/postgresql/notification/sms"
	"insider-one/infrastructure/config"
	"insider-one/infrastructure/logging"
	prometheusWrapper "insider-one/infrastructure/prometheus"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var createSmsConsumerCmd = &cobra.Command{
	Use:   "create-sms-consumer",
	Short: "",
	Long:  ``,
	RunE:  createSmsConsumerCmdRun,
}

func init() {
	rootCmd.AddCommand(createSmsConsumerCmd)
}

func createSmsConsumerCmdRun(cmd *cobra.Command, args []string) (err error) {
	ctx := context.Background()

	cfg, err := config.Load(ctx, cmd.Use, env)
	if err != nil {
		return err
	}

	rabbitMqClient, err := rabbitmq2.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitMqClient.Close()

	genericQueueName := fmt.Sprintf(rabbitmq2.Queue_CreateSms_Generic, priority)
	genericRoutingKey := fmt.Sprintf(rabbitmq2.RoutingKey_Generic, priority)

	err = rabbitmq2.DeclareTopology(ctx, rabbitMqClient, rabbitmq2.TopologyOptions{
		ExchangeName: rabbitmq2.Exchange_CreateSms,
		QueueName:    genericQueueName,
		RoutingKey:   genericRoutingKey,
	})

	pool, err := postgresql.Connect(cfg.DB, ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	publisher := rabbitmq2.NewPublisher(rabbitMqClient)
	smsRepository := SmsPersistence.NewRepository(pool)
	smsCreateCommand := SmsCommand.NewCreateCommand(smsRepository, publisher)
	createSmsHandler := handler.NewCreateSmsHandler(smsCreateCommand)

	prometheusWrapper := prometheusWrapper.InitForConsumer()

	go func(port int) {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
		if err != nil {
			logging.Error(ctx, errors.New("error starting http server"))
		}
	}(cfg.App.Port)

	consumer := rabbitmq2.NewConsumer(rabbitMqClient, genericQueueName, createSmsHandler.HandleMessage, prometheusWrapper, createSmsConsumerOptions)
	if err = consumer.Start(ctx); err != nil {
		log.Fatal(err)
	}

	return err
}

func createSmsConsumerOptions(o *rabbitmq2.ConsumerOptions) {
	o.WorkerCount = 10
	o.PrefetchCount = 10
	o.MaxRetry = 5
}
