package cmd

import (
	"context"
	"errors"
	"fmt"
	pushCommand "insider-one/application/command/notification/push"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"
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

	client, err := rabbitmq2.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	genericQueueName := fmt.Sprintf(rabbitmq2.Queue_PushCreated_Generic, priority)
	genericRoutingKey := fmt.Sprintf(rabbitmq2.RoutingKey_Generic, priority)

	err = rabbitmq2.DeclareTopology(ctx, client, rabbitmq2.TopologyOptions{
		ExchangeName: rabbitmq2.Exchange_PushCreated,
		QueueName:    genericQueueName,
		RoutingKey:   genericRoutingKey,
	})

	pool, err := postgresql.Connect(cfg.DB, ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	publisher := rabbitmq2.NewPublisher(client)
	pushRepository := pushPersistence.NewRepository(pool)
	pushCreateCommand := pushCommand.NewCreateCommand(pushRepository, *publisher)
	createpushHandler := handler.NewCreatePushHandler(pushCreateCommand)

	prometheusWrapper := prometheusWrapper.InitForConsumer()

	go func(port int) {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
		if err != nil {
			logging.Error(ctx, errors.New("error starting http server"))
		}
	}(cfg.App.Port)

	consumer := rabbitmq2.NewConsumer(client, genericQueueName, createpushHandler.HandleMessage, prometheusWrapper, pushCreatedConsumerOptions)
	if err = consumer.Start(ctx); err != nil {
		log.Fatal(err)
	}

	return err
}

func pushCreatedConsumerOptions(o *rabbitmq2.ConsumerOptions) {
	o.WorkerCount = 10
	o.PrefetchCount = 10
	o.MaxRetry = 5
}
