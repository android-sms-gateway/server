package messages

import (
	"errors"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
)

type thirdPartyPostQueryParams struct {
	SkipPhoneValidation bool `query:"skipPhoneValidation"`
	DeviceActiveWithin  int  `query:"deviceActiveWithin"  validate:"omitempty,min=1"`
}

type thirdPartyGetQueryParams struct {
	StartDate string `query:"from"     validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	EndDate   string `query:"to"       validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	State     string `query:"state"    validate:"omitempty,oneof=Pending Processed Sent Delivered Failed"`
	DeviceID  string `query:"deviceId" validate:"omitempty,len=21"`
	Limit     int    `query:"limit"    validate:"omitempty,min=1,max=100"`
	Offset    int    `query:"offset"   validate:"omitempty,min=0"`
}

func (p *thirdPartyGetQueryParams) Validate() error {
	if p.StartDate != "" && p.EndDate != "" && p.StartDate > p.EndDate {
		return errors.New("`from` date must be before `to` date") //nolint:err113 // won't be used directly
	}

	return nil
}

func (p *thirdPartyGetQueryParams) ToFilter() messages.SelectFilter {
	var filter messages.SelectFilter

	if p.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, p.StartDate); err == nil {
			filter.StartDate = t
		}
	}

	if p.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, p.EndDate); err == nil {
			filter.EndDate = t
		}
	}

	if p.State != "" {
		filter.State = messages.ProcessingState(p.State)
	}

	if p.DeviceID != "" {
		filter.DeviceID = p.DeviceID
	}

	return filter
}

func (p *thirdPartyGetQueryParams) ToOptions() messages.SelectOptions {
	const maxLimit = 100

	var options messages.SelectOptions
	options.WithRecipients = true
	options.WithStates = true

	if p.Limit > 0 {
		options.Limit = min(p.Limit, maxLimit)
	} else {
		options.Limit = 50
	}

	if p.Offset > 0 {
		options.Offset = p.Offset
	}

	return options
}

type mobileGetQueryParams struct {
	Order messages.Order `query:"order" validate:"omitempty,oneof=lifo fifo"`
}

func (p *mobileGetQueryParams) OrderOrDefault() messages.Order {
	if p.Order != "" {
		return p.Order
	}
	return messages.MessagesOrderLIFO
}
