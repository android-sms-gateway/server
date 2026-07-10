package messages

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
)

type thirdPartyPostQueryParams struct {
	SkipPhoneValidation bool `query:"skipPhoneValidation"`
	DeviceActiveWithin  int  `query:"deviceActiveWithin"  validate:"omitempty,min=1"`
}

type thirdPartyGetQueryParams smsgateway.ListMessagesOptions

func (p *thirdPartyGetQueryParams) ToFilter() messages.SelectFilter {
	var filter messages.SelectFilter

	if p.From != nil {
		filter.StartDate = *p.From
	}

	if p.To != nil {
		filter.EndDate = *p.To
	}

	if p.State != nil {
		filter.State = append(filter.State, messages.ProcessingState(*p.State))
	}

	if p.DeviceID != nil {
		filter.DeviceID = *p.DeviceID
	}

	return filter
}

func (p *thirdPartyGetQueryParams) ToOptions() messages.SelectOptions {
	const maxLimit = 100

	var options messages.SelectOptions
	options.WithRecipients = true
	options.WithStates = true

	if p.IncludeContent != nil {
		options.WithContent = *p.IncludeContent
	}

	if p.Limit != nil {
		options.Limit = min(*p.Limit, maxLimit)
	} else {
		options.Limit = 50
	}

	if p.Offset != nil {
		options.Offset = max(*p.Offset, 0)
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
