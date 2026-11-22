package sqs

import "errors"

// ErrInvalidMessage identifica mensagens SQS que não podem ser convertidas em ledger.Event.
var ErrInvalidMessage = errors.New("invalid message")
