package executor

import (
	"context"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/android-sms-gateway/server/internal/worker/locker"
	"go.uber.org/zap"
)

type Service struct {
	tasks  []PeriodicTask
	locker locker.Locker

	stopChan chan struct{}
	wg       sync.WaitGroup

	metrics *metrics
	logger  *zap.Logger
}

func NewService(tasks []PeriodicTask, locker locker.Locker, metrics *metrics, logger *zap.Logger) *Service {
	return &Service{
		tasks:  tasks,
		locker: locker,

		stopChan: make(chan struct{}),
		wg:       sync.WaitGroup{},

		metrics: metrics,
		logger:  logger,
	}
}

func (s *Service) Start() error {
	ctx, cancel := context.WithCancel(context.Background())

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		<-s.stopChan
		cancel()
	}()

	for index, task := range s.tasks {
		if task.Interval() <= 0 {
			s.logger.Info("skipping task", zap.String("name", task.Name()), zap.Duration("interval", task.Interval()))
			continue
		}

		s.wg.Add(1)
		go func(index int, task PeriodicTask) {
			defer s.wg.Done()
			s.logger.Info("starting task", zap.Int("index", index), zap.String("name", task.Name()), zap.Duration("interval", task.Interval()))
			s.runTask(ctx, task)
			s.logger.Info("task stopped", zap.Int("index", index), zap.String("name", task.Name()))
		}(index, task)
	}

	return nil
}

func (s *Service) runTask(ctx context.Context, task PeriodicTask) {
	initialDelay := time.Duration(math.Floor(rand.Float64()*task.Interval().Seconds())) * time.Second

	s.logger.Info("initial delay", zap.String("name", task.Name()), zap.Duration("delay", initialDelay))

	select {
	case <-ctx.Done():
		s.logger.Info("stopping task", zap.String("name", task.Name()))
		return
	case <-time.After(initialDelay):
	}

	ticker := time.NewTicker(task.Interval())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("stopping task", zap.String("name", task.Name()))
			return
		case <-ticker.C:
			if err := s.locker.AcquireLock(ctx, task.Name()); err != nil {
				s.logger.Error("can't acquire lock", zap.String("name", task.Name()), zap.Error(err))
				continue
			}

			s.execute(ctx, task)

			if err := s.locker.ReleaseLock(ctx, task.Name()); err != nil {
				s.logger.Error("can't release lock", zap.String("name", task.Name()), zap.Error(err))
			}
		}
	}
}

func (s *Service) execute(ctx context.Context, task PeriodicTask) {
	logger := s.logger.With(zap.String("name", task.Name()))

	s.metrics.IncActiveTasks()

	logger.Info("running task")

	start := time.Now()
	if err := task.Run(ctx); err != nil {
		s.metrics.ObserveTaskResult(task.Name(), metricsTaskResultError, time.Since(start))
		logger.Error("task failed", zap.Duration("duration", time.Since(start)), zap.Error(err))
	} else {
		s.metrics.ObserveTaskResult(task.Name(), metricsTaskResultSuccess, time.Since(start))
		logger.Info("task succeeded", zap.Duration("duration", time.Since(start)))
	}

	s.metrics.DecActiveTasks()
}

func (s *Service) Stop() error {
	close(s.stopChan)
	s.wg.Wait()

	return nil
}
