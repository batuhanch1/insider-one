package cmd

import (
	"context"
	"fmt"
	smsCommand "insider-one/application/command/notification/sms"
	"insider-one/infrastructure/adapters/messaging/rabbitmq"
	"insider-one/infrastructure/adapters/persistence/postgresql"
	"insider-one/infrastructure/adapters/persistence/postgresql/notification/sms"
	"insider-one/infrastructure/config"
	"log"
	"time"

	redislock "github.com/go-co-op/gocron-redis-lock/v2"
	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
)

var smsScheduledJobCmd = &cobra.Command{
	Use:   "sms-scheduled-job",
	Short: "",
	Long:  ``,
	RunE:  smsScheduledJobCmdRun,
}

func init() {
	rootCmd.AddCommand(smsScheduledJobCmd)
}

func smsScheduledJobCmdRun(cmd *cobra.Command, args []string) (err error) {
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

	pool, err := postgresql.Connect(cfg.DB, ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	smsRepository := sms.NewQueryRepository(pool)
	publisher := rabbitmq.NewPublisher(rabbitMqClient)
	scheduleCommand := smsCommand.NewScheduleCommand(smsRepository, publisher)

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Username: cfg.Redis.User,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	locker, err := redislock.NewRedisLocker(rdb)
	if err != nil {
		panic(err)
	}

	s, err := gocron.NewScheduler(
		gocron.WithDistributedLocker(locker),
		gocron.WithLocation(time.UTC),
	)
	if err != nil {
		panic(err)
	}

	_, err = s.NewJob(
		gocron.DurationJob(5*time.Minute),
		gocron.NewTask(scheduleCommand.Execute, ctx),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		panic(err)
	}

	s.Start()

	<-ctx.Done()
	return
}
