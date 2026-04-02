package cmd

import (
	"context"
	"errors"
	"fmt"
	smsCommand "insider-one/application/command/notification/sms"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/adapters/messaging/rabbitmq/handler"
	"insider-one/infrastructure/adapters/persistence/postgresql"
	smsPersistence "insider-one/infrastructure/adapters/persistence/postgresql/notification/sms"
	"insider-one/infrastructure/config"
	"insider-one/infrastructure/logging"
	prometheusWrapper "insider-one/infrastructure/prometheus"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var cancelSmsConsumerCmd = &cobra.Command{
	Use:   "cancel-sms-consumer",
	Short: "",
	Long:  ``,
	RunE:  cancelSmsConsumerCmdRun,
}

func init() {
	rootCmd.AddCommand(cancelSmsConsumerCmd)
}

func cancelSmsConsumerCmdRun(cmd *cobra.Command, args []string) (err error) {
	ctx := context.Background()
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

	err = rabbitmq.DeclareTopology(ctx, rabbitMqClient, rabbitmq.TopologyOptions{
		ExchangeName: rabbitmq.Exchange_CancelSms,
		QueueName:    rabbitmq.Queue_CancelSms,
		RoutingKey:   rabbitmq.RoutingKey_Asterisk,
	})

	pool, err := postgresql.Connect(cfg.DB, ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	smsCommandRepository := smsPersistence.NewCommandRepository(pool)
	smsCancelCommand := smsCommand.NewCancelCommand(smsCommandRepository)
	cancelSmsHandler := handler.NewCancelSmsHandler(smsCancelCommand)

	prometheusWrapper := prometheusWrapper.InitForConsumer()

	go func(port int) {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
		if err != nil {
			logging.Error(ctx, errors.New("error starting http server"))
		}
	}(cfg.App.Port)

	consumer := rabbitmq.NewConsumer(rabbitMqClient, rabbitmq.Queue_CancelSms, cancelSmsHandler.HandleMessage, prometheusWrapper, cancelSmsConsumerOptions)
	if err = consumer.Start(ctx); err != nil {
		log.Fatal(err)
	}

	return err
}

func cancelSmsConsumerOptions(o *rabbitmq.ConsumerOptions) {
	o.WorkerCount = 10
	o.PrefetchCount = 10
	o.MaxRetry = 5
}
