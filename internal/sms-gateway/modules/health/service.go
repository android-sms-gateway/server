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
			case StatusPass:
			case StatusFail:
				level = max(level, levelFail)
			case StatusWarn:
				level = max(level, levelWarn)
			default:
				// Unknown status – log it and fail-safe by escalating to `levelFail`.
				s.logger.Warn("health check returned unknown status",
					zap.String("provider", p.Name()),
					zap.String("check", name),
					zap.String("status", string(detail.Status)),
				)
				level = max(level, levelFail)
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
