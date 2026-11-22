package message

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/clodoaldomarques/core-sdk/pkg/aws/sqs"
	"github.com/clodoaldomarques/core-sdk/pkg/otel/tracer"
	"github.com/clodoaldomarques/core-sdk/pkg/zap/logger"
	"github.com/clodoaldomarques/ledger-worker/config"
	service "github.com/clodoaldomarques/ledger-worker/internal/application/ledger"
	"github.com/clodoaldomarques/ledger-worker/internal/domain/ledger"
	"github.com/clodoaldomarques/ledger-worker/internal/infra/ledger/events"
	"github.com/shopspring/decimal"
)

type Consumer struct {
	c   *sqs.Consumer
	api *events.LedgerEventsApi
	srv *service.Service
}

// NewConsumer cria o cliente da API uma única vez, de modo que o circuit breaker
// compartilhe estado entre mensagens e possa ser fechado no graceful shutdown.
func NewConsumer(ctx context.Context) *Consumer {
	api := events.New(ctx)
	con := &Consumer{
		api: api,
		srv: service.New(api),
	}
	con.c = sqs.NewConsumer(ctx, config.New(), con.handle)
	return con
}

func (con *Consumer) Start() error {
	return con.c.Start()
}

func (con *Consumer) Health() error {
	return con.c.Health()
}

func (con *Consumer) Close() {
	con.api.Close()
}

func (con *Consumer) handle(ctx context.Context, msg *sqs.Message) error {
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

	if err := con.srv.CreateEvent(ctx, e); err != nil {
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

func buildLedgerEvent(msg *sqs.Message) (ledger.Event, error) {
	var t TransactionCreated
	if err := json.Unmarshal([]byte(msg.Body), &t); err != nil {
		return ledger.Event{}, fmt.Errorf("%w: invalid JSON: %v", ErrInvalidMessage, err)
	}

	if t.ProcessingCode == nil || strings.TrimSpace(*t.ProcessingCode) == "" {
		return ledger.Event{}, fmt.Errorf("%w: processing_code is required", ErrInvalidMessage)
	}

	amounts, err := buildAmounts(t.Amount)
	if err != nil {
		return ledger.Event{}, err
	}
	fees, err := buildFees(t.Tax)
	if err != nil {
		return ledger.Event{}, err
	}

	return ledger.Event{
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
