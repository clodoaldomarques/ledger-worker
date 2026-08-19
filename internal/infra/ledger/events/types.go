package events

import (
	"github.com/clodoaldomarques/ledger-worker/internal/domain/ledger"
	"github.com/shopspring/decimal"
)

type EventRequest struct {
	ProcessingCode string                     `json:"processing_code"`
	ProgramID      int64                      `json:"program_id"`
	AccountID      int64                      `json:"account_id"`
	Producer       string                     `json:"producer"`
	Amounts        map[string]decimal.Decimal `json:"amounts"`
	Fees           map[string]decimal.Decimal `json:"fees"`
}

func NewEventRequest(e ledger.Event) EventRequest {
	return EventRequest{
		ProcessingCode: e.ProcessingCode,
		ProgramID:      e.ProgramID,
		AccountID:      e.AccountID,
		Producer:       e.Producer,
		Amounts:        e.Amounts,
		Fees:           e.Fees,
	}
}
