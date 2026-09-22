package ledger

import (
	"context"
	"errors"
	"testing"

	"github.com/clodoaldomarques/ledger-worker/internal/domain/ledger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestService_CreateEvent(t *testing.T) {
	tests := []struct {
		name  string
		setup func(ctrl *gomock.Controller) *Service
		args  func() ledger.Event
		want  func(t *testing.T, e error)
	}{
		{
			name: "when create a event with success",
			setup: func(ctrl *gomock.Controller) *Service {
				a := NewMockEventsProvider(ctrl)
				a.EXPECT().CreateEvent(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				return New(a)
			},
			args: func() ledger.Event {
				return fakeEvent()
			},
			want: func(t *testing.T, e error) {
				assert.Nil(t, e)
			},
		},
		{
			name: "when create a event with error",
			setup: func(ctrl *gomock.Controller) *Service {
				a := NewMockEventsProvider(ctrl)
				a.EXPECT().CreateEvent(gomock.Any(), gomock.Any()).Return(errors.New("wrong event")).Times(1)
				return New(a)
			},
			args: func() ledger.Event {
				return fakeEvent()
			},
			want: func(t *testing.T, e error) {
				assert.Equal(t, "wrong event", e.Error())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			s := tt.setup(ctrl)
			evt := tt.args()
			err := s.CreateEvent(context.Background(), evt)
			tt.want(t, err)

		})
	}
}

func fakeEvent() ledger.Event {
	return ledger.Event{
		Cid:            "cid-123-456-789",
		OrgID:          "TN-abc-def-ghi-jkl",
		ProgramID:      123,
		AccountID:      456,
		ProcessingCode: "101",
		Producer:       "regular",
		Amounts: map[string]decimal.Decimal{
			"Principal":  decimal.NewFromFloat(123.00),
			"Secundario": decimal.NewFromFloat(456.00),
		},
		Fees: map[string]decimal.Decimal{
			"taxa":    decimal.NewFromFloat(6.00),
			"imposto": decimal.NewFromFloat(5.00),
			"contrib": decimal.NewFromFloat(4.00),
		},
	}
}
