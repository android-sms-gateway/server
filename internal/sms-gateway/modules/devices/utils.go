package devices

import "errors"

var rules = map[string]any{
	"encryption": map[string]any{
		"passphrase": "",
	},
	"gateway": map[string]any{
		"cloud_url":     "",
		"private_token": "",
	},
	"messages": map[string]any{
		"send_interval_min":  "",
		"send_interval_max":  "",
		"limit_period":       "",
		"limit_value":        "",
		"sim_selection_mode": "",
		"log_lifetime_days":  "",
	},
	"localserver": map[string]any{
		"PORT": "",
	},
	"ping": map[string]any{
		"interval_seconds": "",
	},
	"logs": map[string]any{
		"lifetime_days": "",
	},
	"webhooks": map[string]any{
		"internet_required": "",
		"retry_count":       "",
		"signing_key":       "",
	},
}

var rulesPublic = map[string]any{
	"encryption": map[string]any{},
	"gateway": map[string]any{
		"cloud_url": "",
	},
	"messages": map[string]any{
		"send_interval_min":  "",
		"send_interval_max":  "",
		"limit_period":       "",
		"limit_value":        "",
		"sim_selection_mode": "",
		"log_lifetime_days":  "",
	},
	"localserver": map[string]any{
		"PORT": "",
	},
	"ping": map[string]any{
		"interval_seconds": "",
	},
	"logs": map[string]any{
		"lifetime_days": "",
	},
	"webhooks": map[string]any{
		"internet_required": "",
		"retry_count":       "",
	},
}

func filterMap(m map[string]any, r map[string]any) (map[string]any, error) {
	var err error

	result := make(map[string]any)
	for field, rule := range r {
		if ruleObj, ok := rule.(map[string]any); ok {
			if dataObj, ok := m[field].(map[string]any); ok {
				result[field], err = filterMap(dataObj, ruleObj)
				if err != nil {
					return nil, err
				}
			} else if m[field] == nil {
				continue
			} else {
				return nil, errors.New("The field: '" + field + "' is not a map to dive")
			}
		} else if _, ok := rule.(string); ok {
			if _, ok := m[field]; !ok {
				continue
			}
			result[field] = m[field]
		}
	}

	return result, nil
}

func appendMap(m1, m2 map[string]any, rules map[string]any) map[string]any {

	for field, rule := range rules {
		if ruleObj, ok := rule.(map[string]any); ok {
			if dataObj, ok := m2[field].(map[string]any); ok {
				m1[field] = appendMap(m1[field].(map[string]any), dataObj, ruleObj)
			} else if m2[field] == nil {
				continue
			}
		} else if _, ok := rule.(string); ok {
			if _, ok := m2[field]; !ok {
				continue
			}
			m1[field] = m2[field]
		}
	}

	return m1
}
