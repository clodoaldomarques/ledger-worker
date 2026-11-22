package message

import (
	"testing"

	coresqs "github.com/clodoaldomarques/core-sdk/pkg/aws/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildLedgerEvent(t *testing.T) {
	body := `{
		"org_id": "TN-1",
		"account_id": 100,
		"program": {"id": 200},
		"correlation_id": "cid-1",
		"processing_code": "101",
		"amount": [{"description": "principal", "value": 12345678901234567890.123456789}],
		"tax": [{"type": "IOF", "value": 6.25}]
	}`

	e, err := buildLedgerEvent(&coresqs.Message{MessageID: "msg-1", Body: body})
	require.NoError(t, err)

	assert.Equal(t, "cid-1", e.Cid)
	assert.Equal(t, "TN-1", e.OrgID)
	assert.Equal(t, int64(100), e.AccountID)
	assert.Equal(t, int64(200), e.ProgramID)
	assert.Equal(t, "101", e.ProcessingCode)
	assert.Equal(t, "regular", e.Producer)
	assert.Equal(t, "12345678901234567890.123456789", e.Amounts["principal"].String())
	assert.Equal(t, "6.25", e.Fees["IOF"].String())
}

func TestBuildLedgerEventInvalid(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "invalid json", body: `{"processing_code":`},
		{name: "missing processing code", body: `{"correlation_id": "cid-1"}`},
		{name: "empty processing code", body: `{"processing_code": "  "}`},
		{
			name: "missing amount description",
			body: `{"processing_code": "101", "amount": [{"value": 1}]}`,
		},
		{
			name: "empty amount description",
			body: `{"processing_code": "101", "amount": [{"description": "", "value": 1}]}`,
		},
		{
			name: "missing amount value",
			body: `{"processing_code": "101", "amount": [{"description": "principal"}]}`,
		},
		{
			name: "invalid amount value",
			body: `{"processing_code": "101", "amount": [{"description": "principal", "value": "abc"}]}`,
		},
		{
			name: "invalid tax value",
			body: `{"processing_code": "101", "tax": [{"type": "IOF", "value": "abc"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildLedgerEvent(&coresqs.Message{MessageID: "msg-1", Body: tt.body})
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidMessage)
		})
	}
}
