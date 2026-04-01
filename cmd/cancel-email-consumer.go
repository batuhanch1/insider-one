package cmd

import (
	"context"
	"errors"
	"fmt"
	emailCommand "insider-one/application/command/notification/email"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/adapters/messaging/rabbitmq/handler"
	"insider-one/infrastructure/adapters/persistence/postgresql"
	emailPersistence "insider-one/infrastructure/adapters/persistence/postgresql/notification/email"
	"insider-one/infrastructure/config"
	"insider-one/infrastructure/logging"
	prometheusWrapper "insider-one/infrastructure/prometheus"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var cancelEmailConsumerCmd = &cobra.Command{
	Use:   "cancel-email-consumer",
	Short: "",
	Long:  ``,
	RunE:  cancelEmailConsumerCmdRun,
}

func init() {
	rootCmd.AddCommand(cancelEmailConsumerCmd)
}

func cancelEmailConsumerCmdRun(cmd *cobra.Command, args []string) (err error) {
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

	err = rabbitmq.DeclareTopology(ctx, rabbitMqClient, rabbitmq.TopologyOptions{
		ExchangeName: rabbitmq.Exchange_CancelEmail,
		QueueName:    rabbitmq.Queue_CancelEmail,
		RoutingKey:   rabbitmq.RoutingKey_Asterisk,
	})

	pool, err := postgresql.Connect(cfg.DB, ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	publisher := rabbitmq.NewPublisher(rabbitMqClient)
	emailRepository := emailPersistence.NewRepository(pool)
	emailCreateCommand := emailCommand.NewCreateCommand(emailRepository, publisher)
	createEmailHandler := handler.NewCreateEmailHandler(emailCreateCommand)
	prometheusWrapper := prometheusWrapper.InitForConsumer()

	go func(port int) {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
		if err != nil {
			logging.Error(ctx, errors.New("error starting http server"))
		}
	}(cfg.App.Port)

	consumer := rabbitmq.NewConsumer(rabbitMqClient, rabbitmq.Queue_CancelEmail, createEmailHandler.HandleMessage, prometheusWrapper, cancelEmailConsumerOptions)
	if err = consumer.Start(ctx); err != nil {
		log.Fatal(err)
	}

	return err
}

func cancelEmailConsumerOptions(o *rabbitmq.ConsumerOptions) {
	o.WorkerCount = 10
	o.PrefetchCount = 10
	o.MaxRetry = 5
}
