package message

import (
	"context"
	"encoding/json"

	"github.com/clodoaldomarques/core-sdk/pkg/logger"
	"github.com/clodoaldomarques/core-sdk/pkg/sqs"
	"github.com/clodoaldomarques/ledger-worker/internal/domain/ledger"
	"github.com/clodoaldomarques/ledger-worker/internal/infra/ledger/events"
	"github.com/shopspring/decimal"
)

func Handler(ctx context.Context, msg *sqs.Message) error {
	api := events.New(ctx)
	srv := ledger.New(api)

	e, err := buildLedgerEvent(msg)
	if err != nil {
		return err
	}

	err = srv.CreateEvent(ctx, e)
	if err != nil {
		logger.Error(ctx, err.Error(), logger.Fields{})
		return err
	}

	return nil
}

func buildLedgerEvent(msg *sqs.Message) (ledger.Event, error) {
	var t TransactionCreated
	err := json.Unmarshal([]byte(msg.Body), &t)
	if err != nil {
		return ledger.Event{}, err
	}
	return ledger.Event{
		OrgID:          "TN-Test",
		AccountID:      t.AccountID,
		ProgramID:      t.Program.ID,
		Cid:            t.CorrelationID,
		ProcessingCode: *t.ProcessingCode,
		Amounts:        buildAmounts(t.Amount),
		Fees:           buildFees(t.Tax),
	}, nil
}

func buildAmounts(amount []Amount) map[string]decimal.Decimal {
	values := make(map[string]decimal.Decimal, len(amount))
	for _, a := range amount {
		values[*a.Description] = decimal.NewFromFloat(a.Value)
	}
	return values
}

func buildFees(tax []Tax) map[string]decimal.Decimal {
	values := make(map[string]decimal.Decimal, len(tax))
	for _, t := range tax {
		values[t.Type] = decimal.NewFromFloat(t.Value)
	}
	return values
}
