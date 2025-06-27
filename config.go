package main

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	TableName string         `env:"TABLE_NAME" envDefault:"user_profiles"`
	Port      string         `env:"PORT" envDefault:":8080"`
	AWS       AWSConfig      `envPrefix:"AWS_"`
	DynamoDB  DynamoDBConfig `envPrefix:"DYNAMO_"`
}

type AWSConfig struct {
	Region    string `env:"REGION" envDefault:"us-east-1"`
	AccessKey string `env:"ACCESS_KEY_ID"`
	SecretKey string `env:"SECRET_ACCESS_KEY"`
}

type DynamoDBConfig struct {
	Endpoint string `env:"ENDPOINT" envDefault:"http://dynamodb:8000"`
}

func LoadConfig() (*Config, error) {
	var cfg Config
	return &cfg, env.Parse(&cfg)
}
