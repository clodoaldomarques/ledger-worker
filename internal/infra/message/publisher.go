package message

import (
	"context"

	"github.com/clodoaldomarques/core-sdk/pkg/aws/sqs"
	"github.com/clodoaldomarques/ledger-worker/config"
)

type Publisher struct {
	p *sqs.Publisher
}

func NewPublisher(ctx context.Context) *Publisher {
	return &Publisher{
		p: sqs.NewPublisher(ctx, config.New()),
	}
}

func (p *Publisher) Health() error {
	return p.p.Health()
}
