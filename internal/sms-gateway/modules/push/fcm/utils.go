package fcm

import (
	"encoding/json"
	"fmt"

	"github.com/android-sms-gateway/server/internal/sms-gateway/modules/push/client"
)

func eventToMap(event client.Event) (map[string]string, error) {
	json, err := json.Marshal(event.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	return map[string]string{
		"event": string(event.Type),
		"data":  string(json),
	}, nil
}
