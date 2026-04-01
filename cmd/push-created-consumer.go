package cmd

import (
	"context"
	"errors"
	"fmt"
	pushCommand "insider-one/application/command/notification/push"
	"insider-one/infrastructure/adapters/client"
	pushProvider "insider-one/infrastructure/adapters/client/push-provider"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/adapters/messaging/rabbitmq/handler"
	"insider-one/infrastructure/adapters/persistence/postgresql"
	pushPersistence "insider-one/infrastructure/adapters/persistence/postgresql/notification/push"
	"insider-one/infrastructure/config"
	"insider-one/infrastructure/logging"
	prometheusWrapper "insider-one/infrastructure/prometheus"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var pushCreatedConsumerCmd = &cobra.Command{
	Use:   "push-created-consumer",
	Short: "",
	Long:  ``,
	RunE:  pushCreatedConsumerCmdRun,
}

func init() {
	rootCmd.AddCommand(pushCreatedConsumerCmd)
}

func pushCreatedConsumerCmdRun(cmd *cobra.Command, args []string) (err error) {
	ctx := context.Background()
	cfg, err := config.Load(ctx, cmd.Use, env)
	if err != nil {
		return err
	}

	rabbitMqClient, err := rabbitmq.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitMqClient.Close()

	genericQueueName := fmt.Sprintf(rabbitmq.Queue_PushCreated_Generic, priority)
	genericRoutingKey := fmt.Sprintf(rabbitmq.RoutingKey_Generic, priority)

	err = rabbitmq.DeclareTopology(ctx, rabbitMqClient, rabbitmq.TopologyOptions{
		ExchangeName: rabbitmq.Exchange_PushCreated,
		QueueName:    genericQueueName,
		RoutingKey:   genericRoutingKey,
	})

	pool, err := postgresql.Connect(cfg.DB, ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	publisher := rabbitmq.NewPublisher(rabbitMqClient)
	httpClient := client.NewClient()
	pushRepository := pushPersistence.NewRepository(pool)
	pushProvider := pushProvider.NewPushProvider(httpClient, cfg.PushProvider)
	pushDeliverCommand := pushCommand.NewDeliverCommand(pushRepository, pushProvider, publisher)
	pushCreatedHandler := handler.NewPushCreatedHandler(pushDeliverCommand)

	prometheusWrapper := prometheusWrapper.InitForConsumer()

	go func(port int) {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
		if err != nil {
			logging.Error(ctx, errors.New("error starting http server"))
		}
	}(cfg.App.Port)

	consumer := rabbitmq.NewConsumer(rabbitMqClient, genericQueueName, pushCreatedHandler.HandleMessage, prometheusWrapper, pushCreatedConsumerOptions)
	if err = consumer.Start(ctx); err != nil {
		log.Fatal(err)
	}

	return err
}

func pushCreatedConsumerOptions(o *rabbitmq.ConsumerOptions) {
	o.WorkerCount = 10
	o.PrefetchCount = 10
	o.MaxRetry = 5
}
