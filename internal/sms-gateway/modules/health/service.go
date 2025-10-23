package health

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ServiceParams struct {
	fx.In

	HealthProviders []HealthProvider `group:"health-providers"`

	Logger *zap.Logger
}

type Service struct {
	healthProviders []HealthProvider

	logger *zap.Logger
}

func NewService(params ServiceParams) *Service {
	return &Service{
		healthProviders: params.HealthProviders,

		logger: params.Logger,
	}
}

func (s *Service) checkProvider(ctx context.Context, probe func(context.Context, HealthProvider) (Checks, error)) CheckResult {
	check := CheckResult{
		Checks: map[string]CheckDetail{},
	}

	for _, p := range s.healthProviders {
		select {
		case <-ctx.Done():
			return check
		default:
		}

		healthChecks, err := probe(ctx, p)
		if err != nil {
			s.logger.Error("Failed check", zap.String("provider", p.Name()), zap.Error(err))
			check.Checks[p.Name()] = CheckDetail{
				Description:   "Failed check",
				ObservedUnit:  "",
				ObservedValue: 0,
				Status:        StatusFail,
			}
			continue
		}

		if len(healthChecks) == 0 {
			continue
		}

		for name, detail := range healthChecks {
			check.Checks[p.Name()+":"+name] = detail
		}
	}

	return check
}

func (s *Service) CheckReadiness(ctx context.Context) CheckResult {
	return s.checkProvider(ctx, func(ctx context.Context, p HealthProvider) (Checks, error) {
		return p.ReadyProbe(ctx)
	})
}

func (s *Service) CheckLiveness(ctx context.Context) CheckResult {
	return s.checkProvider(ctx, func(ctx context.Context, p HealthProvider) (Checks, error) {
		return p.LiveProbe(ctx)
	})
}

func (s *Service) CheckStartup(ctx context.Context) CheckResult {
	return s.checkProvider(ctx, func(ctx context.Context, p HealthProvider) (Checks, error) {
		return p.StartedProbe(ctx)
	})
}
