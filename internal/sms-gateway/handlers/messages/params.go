package messages

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/messages"
)

// thirdPartyPostQueryParams aliases smsgateway.SendOptions so that the query
// parameter shape stays in sync with client-go. Note that any field with a
// `query` struct tag added to SendOptions in a future client-go release will
// automatically widen the surface of this endpoint; additions to SendOptions
// must be reviewed deliberately here before merging.
type thirdPartyPostQueryParams smsgateway.SendOptions
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
		options.Limit = max(min(*p.Limit, maxLimit), 1)
	} else {
		options.Limit = 50
	}

	if p.Offset != nil {
		options.Offset = max(*p.Offset, 0)
	}

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
