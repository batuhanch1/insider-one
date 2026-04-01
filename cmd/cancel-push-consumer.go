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

var cancelPushConsumerCmd = &cobra.Command{
	Use:   "cancel-push-consumer",
	Short: "",
	Long:  ``,
	RunE:  cancelPushConsumerCmdRun,
}

func init() {
	rootCmd.AddCommand(cancelPushConsumerCmd)
}

func cancelPushConsumerCmdRun(cmd *cobra.Command, args []string) (err error) {
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

	err = rabbitmq2.DeclareTopology(ctx, client, rabbitmq2.TopologyOptions{
		ExchangeName: rabbitmq2.Exchange_CancelPush,
		QueueName:    rabbitmq2.Queue_CancelPush,
		RoutingKey:   rabbitmq2.RoutingKey_Asterisk,
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

	consumer := rabbitmq2.NewConsumer(client, rabbitmq2.Queue_CancelPush, createpushHandler.HandleMessage, prometheusWrapper, cancelPushConsumerOptions)
	if err = consumer.Start(ctx); err != nil {
		log.Fatal(err)
	}

	return err
}

func cancelPushConsumerOptions(o *rabbitmq2.ConsumerOptions) {
	o.WorkerCount = 10
	o.PrefetchCount = 10
	o.MaxRetry = 5
}
