package fcm

import (
	"encoding/json"
	"fmt"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/types"
)

func eventToMap(event types.Event) (map[string]string, error) {
	json, err := json.Marshal(event.Data)
	if err != nil {
		return nil, fmt.Errorf("can't marshal event data: %w", err)
	}

	return map[string]string{
		"event": string(event.Type),
		"data":  string(json),
	}, nil
}
