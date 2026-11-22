package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	coresqs "github.com/clodoaldomarques/core-sdk/pkg/aws/sqs"
	"github.com/clodoaldomarques/core-sdk/pkg/otel"
	"github.com/clodoaldomarques/core-sdk/pkg/zap/logger"
	"github.com/clodoaldomarques/ledger-worker/config"
	"github.com/clodoaldomarques/ledger-worker/internal/adapters/inbound/http/server"
	sqsadapter "github.com/clodoaldomarques/ledger-worker/internal/adapters/inbound/sqs"
	ledgerapi "github.com/clodoaldomarques/ledger-worker/internal/adapters/outbound/ledger-api"
	application "github.com/clodoaldomarques/ledger-worker/internal/core/ledger/application"
)

func main() {
	ctx := context.Background()
	otel.Start(ctx)
	c := config.New()

	// Composition root: cada adaptador é criado uma única vez e injetado nas portas.
	publisher := ledgerapi.New(ctx, c.LedgerEventsApiUrl)
	useCase := application.New(publisher)

	// O handler captura con, preenchido logo abaixo; nenhuma mensagem é
	// processada antes de con.Start().
	var con *sqsadapter.Consumer
	client := coresqs.NewConsumer(ctx, c, func(ctx context.Context, msg *coresqs.Message) error {
		return con.Handle(ctx, msg)
	})
	con = sqsadapter.NewConsumer(client, useCase, publisher)
	s := server.New(con)

	go func() {
		err := s.Start(c.AppPort)
		if err != http.ErrServerClosed {
			logger.Fatal(ctx, err.Error(), logger.Fields{})
		}
	}()

	go func() {
		if err := con.Start(); err != nil {
			logger.Fatal(ctx, err.Error(), logger.Fields{})
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Recebido sinal de desligamento, iniciando graceful shutdown...")

	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		fmt.Printf("Erro no shutdown do servidor HTTP: %v\n", err)
	} else {
		fmt.Println("Servidor HTTP finalizado com sucesso")
	}

	con.Close()

	if err := otel.Shutdown(ctx); err != nil {
		fmt.Printf("erro no shutdown do opentelemetry: %v\n", err)
	}

	fmt.Println("Graceful shutdown concluído")
}
