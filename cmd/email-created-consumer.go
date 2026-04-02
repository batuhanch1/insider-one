package cmd

import (
	"context"
	"errors"
	"fmt"
	emailCommand "insider-one/application/command/notification/email"
	"insider-one/infrastructure/adapters/client"
	emailProvider "insider-one/infrastructure/adapters/client/email-provider"
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

var emailCreatedConsumerCmd = &cobra.Command{
	Use:   "email-created-consumer",
	Short: "",
	Long:  ``,
	RunE:  emailCreatedConsumerCmdRun,
	PreRun: func(cmd *cobra.Command, args []string) {
		priority, _ := cmd.Flags().GetString("priority")

		if !rabbitmq.IsPriorityRoutingKeyValid(priority) {
			panic("invalid priority")
		}
	},
}

func init() {
	emailCreatedConsumerCmd.Flags().String("priority", "", "priority (HIGH|MEDIUM|LOW)")
	rootCmd.AddCommand(emailCreatedConsumerCmd)
}

func emailCreatedConsumerCmdRun(cmd *cobra.Command, args []string) (err error) {
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

	genericQueueName := fmt.Sprintf(rabbitmq.Queue_EmailCreated_Generic, priority)
	genericRoutingKey := fmt.Sprintf(rabbitmq.RoutingKey_Generic, priority)

	err = rabbitmq.DeclareTopology(ctx, rabbitMqClient, rabbitmq.TopologyOptions{
		ExchangeName: rabbitmq.Exchange_EmailCreated,
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
	httpClient := client.NewClient()
	emailRepository := emailPersistence.NewCommandRepository(pool)
	emailProvider := emailProvider.NewEmailProvider(httpClient, cfg.EmailProvider)
	emailDeliverCommand := emailCommand.NewDeliverCommand(emailRepository, emailProvider, publisher)
	emailCreatedHandler := handler.NewEmailCreatedHandler(emailDeliverCommand)

	prometheusWrapper := prometheusWrapper.InitForConsumer()

	go func(port int) {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
		if err != nil {
			logging.Error(ctx, errors.New("error starting http server"))
		}
	}(cfg.App.Port)

	consumer := rabbitmq.NewConsumer(rabbitMqClient, genericQueueName, emailCreatedHandler.HandleMessage, prometheusWrapper, emailCreatedConsumerOptions)
	if err = consumer.Start(ctx); err != nil {
		log.Fatal(err)
	}

	return err
}

func emailCreatedConsumerOptions(o *rabbitmq.ConsumerOptions) {
	o.WorkerCount = 10
	o.PrefetchCount = 10
	o.MaxRetry = 5
}
