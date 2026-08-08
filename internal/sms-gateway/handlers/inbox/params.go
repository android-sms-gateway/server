package inbox

import (
	"github.com/android-sms-gateway/client-go/smsgateway"
	"github.com/android-sms-gateway/server/internal/sms-gateway/inbox"
	"github.com/samber/lo"
)

type thirdPartyListParams smsgateway.ListInboxOptions

func (p thirdPartyListParams) toFilter(userID string) inbox.ListFilter {
	return inbox.ListFilter{
		UserID:    userID,
		DeviceID:  lo.FromPtr(p.DeviceID),
		Type:      inbox.MessageType(lo.FromPtr(p.Type)),
		StartDate: lo.FromPtr(p.From),
		EndDate:   lo.FromPtr(p.To),
	}
}

func (p thirdPartyListParams) toOptions() inbox.ListOptions {
	const defaultLimit = 50

	return inbox.ListOptions{
		Limit:  lo.FromPtrOr(p.Limit, defaultLimit),
		Offset: lo.FromPtr(p.Offset),
	}
}
