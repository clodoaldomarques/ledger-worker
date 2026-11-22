package ledger

import "github.com/shopspring/decimal"

type Event struct {
	Cid            string
	OrgID          string
	ProgramID      int64
	AccountID      int64
	ProcessingCode string
	Producer       string
	Amounts        map[string]decimal.Decimal
	Fees           map[string]decimal.Decimal
}
