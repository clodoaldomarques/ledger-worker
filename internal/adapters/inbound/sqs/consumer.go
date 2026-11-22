package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	coresqs "github.com/clodoaldomarques/core-sdk/pkg/aws/sqs"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/tracer"
	"github.com/clodoaldomarques/core-sdk/pkg/zap/logger"
	domain "github.com/clodoaldomarques/ledger-worker/internal/core/ledger/domain"
	"github.com/clodoaldomarques/ledger-worker/internal/core/ledger/ports"
	"github.com/shopspring/decimal"
)

// Client é a dependência de transporte injetada pela raiz de composição.
type Client interface {
	Start() error
	Health() error
}

// Closer representa o adaptador de saída cujo ciclo de vida é encerrado pela raiz.
type Closer interface {
	Close()
}

var _ ports.HealthChecker = (*Consumer)(nil)

type Consumer struct {
	client  Client
	closer  Closer
	useCase ports.ProcessEventUseCase
}

func NewConsumer(client Client, useCase ports.ProcessEventUseCase, closer Closer) *Consumer {
	return &Consumer{
		client:  client,
		closer:  closer,
		useCase: useCase,
	}
}

func (con *Consumer) Start() error {
	return con.client.Start()
}

func (con *Consumer) Health() error {
	return con.client.Health()
}

func (con *Consumer) Close() {
	if con.closer != nil {
		con.closer.Close()
	}
}

// Handle é o handler entregue ao cliente SQS na raiz de composição.
func (con *Consumer) Handle(ctx context.Context, msg *coresqs.Message) error {
	span, ctx := tracer.NewSpanFromContext(ctx, "Message::Handler", map[string]any{
		"MessageID": msg.MessageID,
		"Body":      msg.Body,
	})
	defer span.End()

	e, err := buildLedgerEvent(msg)
	if err != nil {
		span.AddEvent(err.Error(), map[string]any{
			"MessageID": msg.MessageID,
			"Body":      msg.Body,
		})
		return err
	}

	if err := con.useCase.CreateEvent(ctx, e); err != nil {
		span.AddEvent(err.Error(), map[string]any{
			"Cid":            e.Cid,
			"OrgID":          e.OrgID,
			"ProgramID":      e.ProgramID,
			"AccountID":      e.AccountID,
			"ProcessingCode": e.ProcessingCode,
		})
		span.SetError(err)
		logger.Error(ctx, err.Error(), logger.Fields{
			"Cid":            e.Cid,
			"OrgID":          e.OrgID,
			"ProgramID":      e.ProgramID,
			"AccountID":      e.AccountID,
			"ProcessingCode": e.ProcessingCode,
		})
		return err
	}

	return nil
}

func buildLedgerEvent(msg *coresqs.Message) (domain.Event, error) {
	var t TransactionCreated
	if err := json.Unmarshal([]byte(msg.Body), &t); err != nil {
		return domain.Event{}, fmt.Errorf("%w: invalid JSON: %v", ErrInvalidMessage, err)
	}

	if t.ProcessingCode == nil || strings.TrimSpace(*t.ProcessingCode) == "" {
		return domain.Event{}, fmt.Errorf("%w: processing_code is required", ErrInvalidMessage)
	}

	amounts, err := buildAmounts(t.Amount)
	if err != nil {
		return domain.Event{}, err
	}
	fees, err := buildFees(t.Tax)
	if err != nil {
		return domain.Event{}, err
	}

	return domain.Event{
		Producer:       "regular",
		OrgID:          t.OrgID,
		AccountID:      t.AccountID,
		ProgramID:      t.Program.ID,
		Cid:            t.CorrelationID,
		ProcessingCode: *t.ProcessingCode,
		Amounts:        amounts,
		Fees:           fees,
	}, nil
}

func buildAmounts(amount []Amount) (map[string]decimal.Decimal, error) {
	values := make(map[string]decimal.Decimal, len(amount))
	for i, a := range amount {
		if a.Description == nil || strings.TrimSpace(*a.Description) == "" {
			return nil, fmt.Errorf("%w: amount[%d].description is required", ErrInvalidMessage, i)
		}
		v, err := parseDecimal(a.Value, fmt.Sprintf("amount[%d].value", i))
		if err != nil {
			return nil, err
		}
		values[*a.Description] = v
	}
	return values, nil
}

func buildFees(tax []Tax) (map[string]decimal.Decimal, error) {
	values := make(map[string]decimal.Decimal, len(tax))
	for i, t := range tax {
		v, err := parseDecimal(t.Value, fmt.Sprintf("tax[%d].value", i))
		if err != nil {
			return nil, err
		}
		values[t.Type] = v
	}
	return values, nil
}

func parseDecimal(v json.Number, field string) (decimal.Decimal, error) {
	if v == "" {
		return decimal.Decimal{}, fmt.Errorf("%w: %s is required", ErrInvalidMessage, field)
	}
	d, err := decimal.NewFromString(v.String())
	if err != nil {
		return decimal.Decimal{}, fmt.Errorf("%w: invalid %s: %v", ErrInvalidMessage, field, err)
	}
	return d, nil
}
