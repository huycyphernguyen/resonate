package system

import (
	"log/slog"
	"time"

	"github.com/resonatehq/resonate/internal/aio"
	"github.com/resonatehq/resonate/internal/api"
	"github.com/resonatehq/resonate/internal/metrics"

	"github.com/resonatehq/resonate/internal/kernel/scheduler"
	"github.com/resonatehq/resonate/internal/kernel/types"
	"github.com/resonatehq/resonate/internal/util"
)

type System struct {
	cfg       *Config
	api       api.API
	aio       aio.AIO
	metrics   *metrics.Metrics
	scheduler *scheduler.Scheduler
	onRequest map[types.APIKind]func(int64, *types.Request, func(*types.Response, error)) *scheduler.Coroutine
	onTick    map[int][]func(int64, *Config) *scheduler.Coroutine
	ticks     int64
}

type Config struct {
	PromiseCacheSize      int
	TimeoutCacheSize      int
	NotificationCacheSize int
	SubmissionBatchSize   int
	CompletionBatchSize   int
}

func New(cfg *Config, api api.API, aio aio.AIO, metrics *metrics.Metrics) *System {
	return &System{
		cfg:       cfg,
		api:       api,
		aio:       aio,
		metrics:   metrics,
		scheduler: scheduler.NewScheduler(aio, metrics),
		onRequest: map[types.APIKind]func(int64, *types.Request, func(*types.Response, error)) *scheduler.Coroutine{},
		onTick:    map[int][]func(int64, *Config) *scheduler.Coroutine{},
	}
}

func (s *System) Loop() error {
	for {
		t := time.Now().Unix()
		s.Tick(t, time.After(10*time.Millisecond))

		if s.api.Done() && s.scheduler.Done() {
			return nil
		}
	}
}

func (s *System) Tick(t int64, timeoutCh <-chan time.Time) {
	defer s.housekeeping(t)

	if !s.api.Done() {
		// add request coroutines
		for _, sqe := range s.api.Dequeue(s.cfg.SubmissionBatchSize, timeoutCh) {
			if coroutine, ok := s.onRequest[sqe.Submission.Kind]; ok {
				slog.Debug("api:dequeue", "sqe", sqe.Submission)
				s.scheduler.Add(coroutine(t, sqe.Submission, sqe.Callback))
			} else {
				panic("invalid api request")
			}
		}

		// add tick coroutines
		for _, coroutines := range util.OrderedRangeKV(s.onTick) {
			if s.ticks%int64(coroutines.Key) == 0 {
				for _, coroutine := range coroutines.Value {
					s.scheduler.Add(coroutine(t, s.cfg))
				}
			}
		}
	}

	// tick scheduler
	s.scheduler.Tick(t, s.cfg.CompletionBatchSize)
}

func (s *System) AddOnRequest(kind types.APIKind, constructor func(int64, *types.Request, func(*types.Response, error)) *scheduler.Coroutine) {
	s.onRequest[kind] = constructor
}

func (s *System) AddOnTick(n int, constructor func(int64, *Config) *scheduler.Coroutine) {
	util.Assert(n > 0, "n must be greater than 0")
	s.onTick[n] = append(s.onTick[n], constructor)
}

func (s *System) housekeeping(int64) {
	s.ticks++
}
