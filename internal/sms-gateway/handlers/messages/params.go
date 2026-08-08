package messages

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
	"github.com/samber/lo"
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
	const defaultLimit = 50

	var options messages.SelectOptions
	options.WithRecipients = true
	options.WithStates = true

	if p.IncludeContent != nil {
		options.WithContent = *p.IncludeContent
	}

	options.Limit = lo.FromPtrOr(p.Limit, int(defaultLimit))

	if p.Sort != nil {
		switch *p.Sort {
		case smsgateway.CreatedAtAscending:
			options.SortField = messages.SortFieldCreatedAtAsc
		case smsgateway.CreatedAtDescending:
			options.SortField = messages.SortFieldCreatedAtDesc
		}
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
