package config

import (
	"sync"

	"github.com/clodoaldomarques/core-sdk/pkg/env"
)

type Config struct {
	AppPort            int
	AwsAddress         string
	AwsRegion          string
	AwsAccessKeyID     string
	AwsSecretAccessKey string
	LedgerEventsApiUrl string
	BalanceQueueUrl    string
}

type Option func(*Config)

var (
	singleton sync.Once
	instance  *Config
)

func New(options ...Option) *Config {
	singleton.Do(func() {
		instance = &Config{
			AppPort:            env.GetInt("APP_PORT", 5002),
			AwsAddress:         env.GetString("AWS_ADDRESS", ""),
			AwsRegion:          env.GetString("AWS_REGION", ""),
			AwsAccessKeyID:     env.GetString("AWS_ACCESS_KEY_ID", ""),
			AwsSecretAccessKey: env.GetString("AWS_SECRET_ACCESS_KEY", ""),
			BalanceQueueUrl:    env.GetString("BALANCE_QUEUE", ""),
			LedgerEventsApiUrl: env.GetString("LEDGER_EVENTS_API_URL", ""),
		}
	})

	for _, optFunc := range options {
		optFunc(instance)
	}

	return instance
}

func WithAppPort(appPort int) Option {
	return func(c *Config) {
		c.AppPort = appPort
	}
}

func WithAwsAddress(awsAddress string) Option {
	return func(c *Config) {
		c.AwsAddress = awsAddress
	}
}
func WithAwsRegion(awsRegion string) Option {
	return func(c *Config) {
		c.AwsRegion = awsRegion
	}
}

func WithLedgerEventsApiUrl(ledgerEventsApiUrl string) Option {
	return func(c *Config) {
		c.LedgerEventsApiUrl = ledgerEventsApiUrl
	}
}

func (c Config) Region() string {
	return c.AwsRegion
}

func (c Config) Address() string {
	return c.AwsAddress
}
func (c Config) AccessKeyID() string {
	return c.AwsAccessKeyID
}
func (c Config) SecretAccessKey() string {
	return c.AwsSecretAccessKey
}
func (c Config) BalanceQueue() string {
	return c.BalanceQueueUrl
}

func (c Config) QueueURL() string {
	return c.BalanceQueueUrl
}

func (c Config) DeadLetterQueueURL() string {
	return c.BalanceQueueUrl
}

func (c Config) MaxReceiveCount() int {
	return int(10)
}
