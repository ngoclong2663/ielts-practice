package internal

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"go.uber.org/fx"

	usecases "ielts-practice/app/modules/vocab/internal/use-cases"
	"ielts-practice/configs"
)

const (
	idlePoll      = time.Second
	janitorPeriod = time.Minute
	stuckAfter    = 3 * time.Minute
)

// Worker is the in-process goroutine pool that generates cards (Epic 3 §5.2). Items are claimed with
// FOR UPDATE SKIP LOCKED, so it can later move to its own binary without schema changes.
type Worker struct {
	uc      *usecases.VocabUseCases
	workers int
	logger  zerolog.Logger
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewWorker(uc *usecases.VocabUseCases, cfg *configs.Config, logger zerolog.Logger) *Worker {
	return &Worker{uc: uc, workers: cfg.Vocab.Workers, logger: logger}
}

func RegisterWorker(lc fx.Lifecycle, w *Worker) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error { w.Start(); return nil },
		OnStop:  func(ctx context.Context) error { return w.Stop(ctx) },
	})
}

func (w *Worker) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel
	w.uc.RequeueStuck(ctx, stuckAfter)
	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go w.loop(ctx)
	}
	w.wg.Add(1)
	go w.janitor(ctx)
	w.logger.Info().Int("workers", w.workers).Msg("Card generation worker started")
}

func (w *Worker) Stop(ctx context.Context) error {
	if w.cancel != nil {
		w.cancel()
	}
	done := make(chan struct{})
	go func() { w.wg.Wait(); close(done) }()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (w *Worker) loop(ctx context.Context) {
	defer w.wg.Done()
	for ctx.Err() == nil {
		items, err := w.uc.ClaimBatch(ctx)
		if err != nil && ctx.Err() == nil {
			w.logger.Error().Err(err).Msg("Failed to claim generation items")
		}
		if len(items) == 0 {
			select {
			case <-ctx.Done():
				return
			case <-w.uc.Wake():
			case <-time.After(idlePoll):
			}
			continue
		}
		// Finish the batch even during shutdown; abandoned items would wait for the janitor.
		w.uc.ProcessBatch(context.WithoutCancel(ctx), items)
	}
}

func (w *Worker) janitor(ctx context.Context) {
	defer w.wg.Done()
	t := time.NewTicker(janitorPeriod)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.uc.RequeueStuck(ctx, stuckAfter)
		}
	}
}
