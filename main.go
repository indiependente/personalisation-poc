package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"personalisation-poc/repository/ddb"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/guregu/dynamo/v2"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(log)

	conf, err := LoadConfig()
	if err != nil {
		log.Error("unable to load config", "error", err)
		os.Exit(1)
	}

	if err := run(log, conf); err != nil {
		log.Error("error running", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger, conf *Config) error {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %w", err)
	}

	db := dynamo.New(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(conf.DynamoDB.Endpoint)
		o.Region = conf.AWS.Region
		o.Credentials = credentials.NewStaticCredentialsProvider(conf.AWS.AccessKey, conf.AWS.SecretKey, "")
	})
	table := db.Table(conf.TableName)

	repo := ddb.NewDB(table)

	server := newServer(repo, log)

	go func() {
		log.Info("starting server", "port", conf.Port)
		if err := http.ListenAndServe(conf.Port, server.router); err != nil {
			if err == http.ErrServerClosed {
				log.Info("server closed")
			} else {
				log.Error("error starting server", "error", err)
				os.Exit(1)
			}
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down")
	os.Exit(0)

	return nil
}
