package cmd

import (
	"context"
	"errors"
	"fmt"
	pushCommand "insider-one/application/command/notification/push"
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

var createPushConsumerCmd = &cobra.Command{
	Use:   "create-push-consumer",
	Short: "",
	Long:  ``,
	RunE:  createPushConsumerCmdRun,
	PreRun: func(cmd *cobra.Command, args []string) {
		priority, _ := cmd.Flags().GetString("priority")

		if !rabbitmq.IsPriorityRoutingKeyValid(priority) {
			panic("invalid priority")
		}
	},
}

func init() {
	createPushConsumerCmd.Flags().String("priority", "", "priority (HIGH|MEDIUM|LOW)")
	rootCmd.AddCommand(createPushConsumerCmd)
}

func createPushConsumerCmdRun(cmd *cobra.Command, args []string) (err error) {
	ctx := context.Background()
	priority, _ := cmd.Flags().GetString("priority")
	env, _ := cmd.Flags().GetString("env")
	cfg, err := config.Load(ctx, cmd.Use, env)
	if err != nil {
		return err
	}

	rabbitMqClient, err := rabbitmq.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitMqClient.Close()

	genericQueueName := fmt.Sprintf(rabbitmq.Queue_CreatePush_Generic, priority)
	genericRoutingKey := fmt.Sprintf(rabbitmq.RoutingKey_Generic, priority)

	err = rabbitmq.DeclareTopology(ctx, rabbitMqClient, rabbitmq.TopologyOptions{
		ExchangeName: rabbitmq.Exchange_CreatePush,
		ExchangeType: "direct",
		QueueName:    genericQueueName,
		RoutingKey:   genericRoutingKey,
	})

	pool, err := postgresql.Connect(cfg.DB, ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	publisher := rabbitmq.NewPublisher(rabbitMqClient)
	pushRepository := pushPersistence.NewCommandRepository(pool)
	pushCreateCommand := pushCommand.NewCreateCommand(pushRepository, publisher)
	createpushHandler := handler.NewCreatePushHandler(pushCreateCommand)

	prometheusWrapper := prometheusWrapper.InitForConsumer()

	go func(port int) {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
		if err != nil {
			logging.Error(ctx, errors.New("error starting http server"))
		}
	}(cfg.App.Port)

	consumer := rabbitmq.NewConsumer(rabbitMqClient, genericQueueName, createpushHandler.HandleMessage, prometheusWrapper, createPushConsumerOptions)
	if err = consumer.Start(ctx); err != nil {
		log.Fatal(err)
	}

	return err
}

func createPushConsumerOptions(o *rabbitmq.ConsumerOptions) {
	o.WorkerCount = 10
	o.PrefetchCount = 10
	o.MaxRetry = 5
}
