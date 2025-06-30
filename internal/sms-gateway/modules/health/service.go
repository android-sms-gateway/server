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

func (s *Service) HealthCheck(ctx context.Context) (Check, error) {
	check := Check{
		Status: StatusPass,
		Checks: map[string]CheckDetail{},
	}

	level := levelPass
	for _, p := range s.healthProviders {
		healthChecks, err := p.HealthCheck(ctx)
		if err != nil {
			s.logger.Error("Error getting health check", zap.String("provider", p.Name()), zap.Error(err))
		}
		if len(healthChecks) == 0 {
			continue
		}

		for name, detail := range healthChecks {
			check.Checks[p.Name()+":"+name] = detail

			switch detail.Status {
			case StatusFail:
				level = max(level, levelFail)
			case StatusWarn:
				level = max(level, levelWarn)
			}
		}
	}

	check.Status = statusLevels[level]

	return check, nil
}

func AsHealthProvider(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(HealthProvider)),
		fx.ResultTags(`group:"health-providers"`),
	)
}
