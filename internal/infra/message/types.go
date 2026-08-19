package message

type TransactionCreated struct {
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
	AuthorizationTracking *string         `json:"authorization_tracking_id,omitempty"`
	CustomerID            *int64          `json:"customer_id,omitempty"`
	Details               *string         `json:"details,omitempty"`
	DueDate               *string         `json:"due_date,omitempty"`
	Installment           *int64          `json:"installment,omitempty"`
	InterestRate          *float64        `json:"interest_rate,omitempty"`
	NumberOfInstallments  *int64          `json:"number_of_installments,omitempty"`
	Origin                *string         `json:"origin,omitempty"`
	PaymentDate           *string         `json:"payment_date,omitempty"`
	PaymentDatetime       *string         `json:"payment_datetime,omitempty"`
	ProcessingCode        *string         `json:"processing_code,omitempty"`
	ProcessingDescription *string         `json:"processing_description,omitempty"`
	Rates                 []Rate          `json:"rates,omitempty"`
	RefToCardholderRate   *float64        `json:"reference_to_cardholder_exchange_rate,omitempty"`
	SoftDescriptor        *string         `json:"soft_descriptor,omitempty"`
	StatementID           *int64          `json:"statement_id,omitempty"`
	Tax                   []Tax           `json:"tax,omitempty"`
	TransactionGroup      *string         `json:"transaction_group,omitempty"`
	UserCategory          *string         `json:"user_category,omitempty"`
	CreatedAt             *string         `json:"created_at,omitempty"`
}

// Amount representa os detalhes do valor da transação.
type Amount struct {
	Currency    *string `json:"currency,omitempty"`
	Value       float64 `json:"value"`
	Description *string `json:"description,omitempty"`
}

// Authorization representa o registro de autorização da transação.
type Authorization struct {
	Type                      *string  `json:"type,omitempty"`
	ID                        *int64   `json:"id,omitempty"`
	TID                       *string  `json:"tid,omitempty"`
	CardID                    *string  `json:"card_id,omitempty"`
	CardHash                  *string  `json:"card_hash,omitempty"`
	AuthorizationCode         *string  `json:"authorization_code,omitempty"`
	RetrievalReferenceNumber  *string  `json:"retrieval_reference_number,omitempty"`
	PrincipalAmount           *float64 `json:"principal_amount,omitempty"`
	CorrelatedAuthorizationID *int64   `json:"correlated_authorization_id,omitempty"`
	Currency                  *string  `json:"currency,omitempty"`
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
