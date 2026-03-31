package cmd

import (
	"context"
	"errors"
	"fmt"
	emailCommand "insider-one/application/command/notification/email"
	rabbitmq2 "insider-one/infrastructure/adapters/messaging/rabbitmq"
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
	cfg, err := config.Load(cmd.Use, env)
	if err != nil {
		return err
	}

	client, err := rabbitmq2.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	err = rabbitmq2.DeclareTopology(client, rabbitmq2.TopologyOptions{
		ExchangeName: rabbitmq2.Exchange_CancelEmail,
		QueueName:    rabbitmq2.Queue_CancelEmail,
		RoutingKey:   rabbitmq2.RoutingKey_Asterisk,
	})
	ctx := context.Background()

	pool, err := postgresql.Connect(cfg.DB, ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	publisher := rabbitmq2.NewPublisher(client)
	emailRepository := emailPersistence.NewRepository(pool)
	emailCreateCommand := emailCommand.NewCreateCommand(emailRepository, *publisher)
	createEmailHandler := handler.NewCreateEmailHandler(emailCreateCommand)
	prometheusWrapper := prometheusWrapper.InitForConsumer()

	go func(port int) {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
		if err != nil {
			logging.Error(ctx, errors.New("error starting http server"))
		}
	}(cfg.App.Port)

	consumer := rabbitmq2.NewConsumer(client, rabbitmq2.Queue_CancelEmail, createEmailHandler.HandleMessage, prometheusWrapper, cancelEmailConsumerOptions)
	if err = consumer.Start(ctx); err != nil {
		log.Fatal(err)
	}

	return err
}

func cancelEmailConsumerOptions(o *rabbitmq2.ConsumerOptions) {
	o.WorkerCount = 10
	o.PrefetchCount = 10
	o.MaxRetry = 5
}
