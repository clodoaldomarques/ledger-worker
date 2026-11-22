package sqs

import (
	"context"
	"errors"
	"testing"

	coresqs "github.com/clodoaldomarques/core-sdk/pkg/aws/sqs"
	domain "github.com/clodoaldomarques/ledger-worker/internal/core/ledger/domain"
	"github.com/clodoaldomarques/ledger-worker/internal/core/ledger/ports"
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

type fakeProcessEventUseCase struct {
	calls int
	event domain.Event
	err   error
}

var _ ports.ProcessEventUseCase = (*fakeProcessEventUseCase)(nil)

func (f *fakeProcessEventUseCase) CreateEvent(_ context.Context, e domain.Event) error {
	f.calls++
	f.event = e
	return f.err
}

func TestConsumerHandleUsesProcessEventUseCase(t *testing.T) {
	body := `{
		"org_id": "TN-1",
		"account_id": 100,
		"program": {"id": 200},
		"correlation_id": "cid-1",
		"processing_code": "101",
		"amount": [{"description": "principal", "value": 10.5}],
		"tax": [{"type": "IOF", "value": 1.25}]
	}`

	useCase := &fakeProcessEventUseCase{}
	con := NewConsumer(nil, useCase, nil)

	err := con.Handle(context.Background(), &coresqs.Message{MessageID: "msg-1", Body: body})
	require.NoError(t, err)
	require.Equal(t, 1, useCase.calls)
	assert.Equal(t, "cid-1", useCase.event.Cid)
	assert.Equal(t, "101", useCase.event.ProcessingCode)
	assert.Equal(t, "10.5", useCase.event.Amounts["principal"].String())
	assert.Equal(t, "1.25", useCase.event.Fees["IOF"].String())
}

func TestConsumerHandleDoesNotCallUseCaseForInvalidMessage(t *testing.T) {
	useCase := &fakeProcessEventUseCase{}
	con := NewConsumer(nil, useCase, nil)

	err := con.Handle(context.Background(), &coresqs.Message{MessageID: "msg-1", Body: `{"processing_code":`})
	require.ErrorIs(t, err, ErrInvalidMessage)
	assert.Zero(t, useCase.calls)
}

func TestConsumerHandlePropagatesUseCaseError(t *testing.T) {
	useCase := &fakeProcessEventUseCase{err: errors.New("publish failed")}
	con := NewConsumer(nil, useCase, nil)

	err := con.Handle(context.Background(), &coresqs.Message{MessageID: "msg-1", Body: `{"processing_code": "101"}`})
	require.EqualError(t, err, "publish failed")
	assert.Equal(t, 1, useCase.calls)
}

type fakeCloser struct {
	closed bool
}

func (f *fakeCloser) Close() { f.closed = true }

func TestConsumerCloseClosesInjectedAdapter(t *testing.T) {
	closer := &fakeCloser{}
	con := NewConsumer(nil, &fakeProcessEventUseCase{}, closer)

	con.Close()
	assert.True(t, closer.closed)
}
