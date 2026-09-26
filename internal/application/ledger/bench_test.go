package ledger

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"
)

// =============================================================================
// Helpers
// =============================================================================

// newBenchService monta o Service com um mock de EventsProvider que sempre
// responde com sucesso.
func newBenchService(b *testing.B) *Service {
	ctrl := gomock.NewController(b)

	a := NewMockEventsProvider(ctrl)
	a.EXPECT().
		CreateEvent(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()

	return New(a)
}

// newBenchServiceErr monta o Service com mock que sempre retorna erro.
func newBenchServiceErr(b *testing.B) *Service {
	ctrl := gomock.NewController(b)

	a := NewMockEventsProvider(ctrl)
	a.EXPECT().
		CreateEvent(gomock.Any(), gomock.Any()).
		Return(errors.New("wrong event")).
		AnyTimes()

	return New(a)
}

// =============================================================================
// Benchmarks
// =============================================================================

// caminho feliz: só delega para o provider
func BenchmarkCreateEvent_Success(b *testing.B) {
	svc := newBenchService(b)

	ctx := context.Background()
	evt := fakeEvent()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := svc.CreateEvent(ctx, evt); err != nil {
			b.Fatal(err)
		}
	}
}

// mesmo caminho, mas com Event montado dentro do loop
// (útil para medir o custo de construir o Event, se for relevante)
func BenchmarkCreateEvent_WithBuildEvent(b *testing.B) {
	svc := newBenchService(b)

	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		evt := fakeEvent()
		if err := svc.CreateEvent(ctx, evt); err != nil {
			b.Fatal(err)
		}
	}
}

// caminho de erro: mede o custo de propagar erro
func BenchmarkCreateEvent_Error(b *testing.B) {
	svc := newBenchServiceErr(b)

	ctx := context.Background()
	evt := fakeEvent()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = svc.CreateEvent(ctx, evt)
	}
}

// paralelo: aqui provavelmente vai regredir, como no ledger-config
func BenchmarkCreateEvent_Parallel(b *testing.B) {
	svc := newBenchService(b)

	ctx := context.Background()
	evt := fakeEvent()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := svc.CreateEvent(ctx, evt); err != nil {
				b.Fatal(err)
			}
		}
	})
}
