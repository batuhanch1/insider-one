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
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var smsCreatedConsumerCmd = &cobra.Command{
	Use:   "sms-created-consumer",
	Short: "",
	Long:  ``,
	RunE:  smsCreatedConsumerCmdRun,
}

func init() {
	rootCmd.AddCommand(smsCreatedConsumerCmd)
}

func smsCreatedConsumerCmdRun(cmd *cobra.Command, args []string) (err error) {
	var environment = os.Getenv("APP_ENV")
	cfg, err := config.Load(cmd.Use, environment)
	if err != nil {
		return err
	}

	client, err := rabbitmq2.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	genericQueueName := fmt.Sprintf(rabbitmq2.Queue_SmsCreated_Generic, priority)
	genericRoutingKey := fmt.Sprintf(rabbitmq2.RoutingKey_Generic, priority)

	err = rabbitmq2.DeclareTopology(client, rabbitmq2.TopologyOptions{
		ExchangeName: rabbitmq2.Exchange_SmsCreated,
		QueueName:    genericQueueName,
		RoutingKey:   genericRoutingKey,
	})
	ctx := context.Background()

	pool, err := postgresql.Connect(cfg.DB, ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	publisher := rabbitmq2.NewPublisher(client)
	smsRepository := SmsPersistence.NewRepository(pool)
	smsCreateCommand := SmsCommand.NewCreateCommand(smsRepository, *publisher)
	createSmsHandler := handler.NewCreateSmsHandler(smsCreateCommand)

	prometheusWrapper := prometheusWrapper.InitForConsumer()

	go func(port int) {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
		if err != nil {
			logging.Error(ctx, errors.New("error starting http server"))
		}
	}(cfg.App.Port)

	consumer := rabbitmq2.NewConsumer(client, genericQueueName, createSmsHandler.HandleMessage, prometheusWrapper, smsCreatedConsumerOptions)
	if err = consumer.Start(ctx); err != nil {
		log.Fatal(err)
	}

	return err
}

func smsCreatedConsumerOptions(o *rabbitmq2.ConsumerOptions) {
	o.WorkerCount = 10
	o.PrefetchCount = 10
	o.MaxRetry = 5
}
