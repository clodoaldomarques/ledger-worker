package message

type TransactionCreated struct {
	OrgID                 string          `json:"org_id"`
	AccountID             int64           `json:"account_id"`
	AccountingDate        string          `json:"accounting_date"`
	Amount                []Amount        `json:"amount"`
	Authorization         Authorization   `json:"authorization"`
	CorrelationID         string          `json:"correlation_id"`
	EventDate             string          `json:"event_date"`
	EventDatetime         string          `json:"event_datetime"`
	ID                    int64           `json:"id"`
	Program               Program         `json:"program"`
	TransactionType       TransactionType `json:"transaction_type"`
	AuthorizationTracking *string         `json:"authorization_tracking_id"`
	CustomerID            *int64          `json:"customer_id"`
	Details               *string         `json:"details"`
	DueDate               *string         `json:"due_date"`
	Installment           *int64          `json:"installment"`
	InterestRate          *float64        `json:"interest_rate"`
	NumberOfInstallments  *int64          `json:"number_of_installments"`
	Origin                *string         `json:"origin"`
	PaymentDate           *string         `json:"payment_date"`
	PaymentDatetime       *string         `json:"payment_datetime"`
	ProcessingCode        *string         `json:"processing_code"`
	ProcessingDescription *string         `json:"processing_description"`
	Rates                 []Rate          `json:"rates"`
	RefToCardholderRate   *float64        `json:"reference_to_cardholder_exchange_rate"`
	SoftDescriptor        *string         `json:"soft_descriptor"`
	StatementID           *int64          `json:"statement_id"`
	Tax                   []Tax           `json:"tax"`
	TransactionGroup      *string         `json:"transaction_group"`
	UserCategory          *string         `json:"user_category"`
	CreatedAt             *string         `json:"created_at"`
}

// Amount representa os detalhes do valor da transação.
type Amount struct {
	Currency    *string `json:"currency"`
	Value       float64 `json:"value"`
	Description *string `json:"description"`
}

// Authorization representa o registro de autorização da transação.
type Authorization struct {
	Type                      *string  `json:"type"`
	ID                        *int64   `json:"id"`
	TID                       *string  `json:"tid"`
	CardID                    *string  `json:"card_id"`
	CardHash                  *string  `json:"card_hash"`
	AuthorizationCode         *string  `json:"authorization_code"`
	RetrievalReferenceNumber  *string  `json:"retrieval_reference_number"`
	PrincipalAmount           *float64 `json:"principal_amount"`
	CorrelatedAuthorizationID *int64   `json:"correlated_authorization_id"`
	Currency                  *string  `json:"currency"`
}

// Program representa o programa da transação.
type Program struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// TransactionType representa o tipo da transação.
type TransactionType struct {
	ID            int64  `json:"id"`
	Description   string `json:"description"`
	IsCredit      bool   `json:"is_credit"`
	StatementPost bool   `json:"statement_post"`
}

// Rate representa as taxas aplicadas ao valor da transação.
type Rate struct {
	Type  string  `json:"type"` // "settlement_conversion_rate" ou "reference_to_cardholder_exchange_rate"
	Value float64 `json:"value"`
}

// Tax representa os impostos aplicados à transação.
type Tax struct {
	Type  string  `json:"type"` // "IOF", "DAILY_IOF" ou "INTEREST"
	Value float64 `json:"value"`
}
