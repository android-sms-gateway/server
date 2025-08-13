package messages

import (
	"fmt"
	"time"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
)

type thirdPartyPostQueryParams struct {
	SkipPhoneValidation bool `query:"skipPhoneValidation"`
	DeviceActiveWithin  uint `query:"deviceActiveWithin"`
}

type thirdPartyGetQueryParams struct {
	StartDate string `query:"from" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	EndDate   string `query:"to" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	State     string `query:"state" validate:"omitempty,oneof=Pending Processed Sent Delivered Failed"`
	DeviceID  string `query:"deviceId" validate:"omitempty,len=21"`
	Limit     int    `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset    int    `query:"offset" validate:"omitempty,min=0"`
}

func (p *thirdPartyGetQueryParams) Validate() error {
	if p.StartDate != "" && p.EndDate != "" && p.StartDate > p.EndDate {
		return fmt.Errorf("`from` date must be before `to` date")
	}

	return nil
}

func (p *thirdPartyGetQueryParams) ToFilter() messages.MessagesSelectFilter {
	filter := messages.MessagesSelectFilter{}

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

func (p *thirdPartyGetQueryParams) ToOptions() messages.MessagesSelectOptions {
	options := messages.MessagesSelectOptions{
		WithRecipients: true,
		WithStates:     true,
	}

	if p.Limit > 0 {
		options.Limit = min(p.Limit, 100)
	} else {
		options.Limit = 50
	}

	if p.Offset > 0 {
		options.Offset = p.Offset
	}

	return options
}

type mobileGetQueryParams struct {
	Order messages.MessagesOrder `query:"order" validate:"omitempty,oneof=lifo fifo"`
}

func (p *mobileGetQueryParams) OrderOrDefault() messages.MessagesOrder {
	if p.Order != "" {
		return p.Order
	}
	return messages.MessagesOrderLIFO

}
