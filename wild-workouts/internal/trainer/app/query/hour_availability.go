package query

import (
	"context"
	"time"
	
	"github.com/sirupsen/logrus"
	
	"github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/common/decorator"
	"github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/internal/trainer/domain/hour"
)

type (
	Date struct {
		Date         time.Time
		HasFreeHours bool
		Hours        []Hour
	}
	
	Hour struct {
		Available            bool
		HasTrainingScheduled bool
		Hour                 time.Time
	}
	
	HourAvailability struct {
		Hour time.Time
	}
)

type HourAvailabilityHandler decorator.QueryHandler[HourAvailability, bool]

type hourAvailabilityHandler struct {
	hourRepo hour.Repository
}

func NewHourAvailabilityHandler(
	hourRepo hour.Repository,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) HourAvailabilityHandler {
	if hourRepo == nil {
		panic("nil hourRepo")
	}
	
	return decorator.ApplyQueryDecorators[HourAvailability, bool](
		hourAvailabilityHandler{hourRepo: hourRepo},
		logger,
		metricsClient,
	)
}

func (h hourAvailabilityHandler) Handle(ctx context.Context, query HourAvailability) (bool, error) {
	hour, err := h.hourRepo.GetHour(ctx, query.Hour)
	if err != nil {
		return false, err
	}
	
	return hour.IsAvailable(), nil
}
