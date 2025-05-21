package settings

import "fmt"

var rules = map[string]any{
	"encryption": map[string]any{
		"passphrase": "",
	},
	"messages": map[string]any{
		"send_interval_min":  "",
		"send_interval_max":  "",
		"limit_period":       "",
		"limit_value":        "",
		"sim_selection_mode": "",
		"log_lifetime_days":  "",
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
	"messages": map[string]any{
		"send_interval_min":  "",
		"send_interval_max":  "",
		"limit_period":       "",
		"limit_value":        "",
		"sim_selection_mode": "",
		"log_lifetime_days":  "",
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
				return nil, fmt.Errorf("the field: '%s' is not a map to dive", field)
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

func appendMap(m1, m2 map[string]any, rules map[string]any) (map[string]any, error) {
	var err error

	for field, rule := range rules {
		if ruleObj, ok := rule.(map[string]any); ok {
			if dataObj, ok := m2[field].(map[string]any); ok {
				if m1Field, ok := m1[field].(map[string]any); ok {
					m1[field], err = appendMap(m1Field, dataObj, ruleObj)
					if err != nil {
						return nil, err
					}
				} else {
					// Initialize if not present or not a map
					newMap := make(map[string]any)
					m1[field], err = appendMap(newMap, dataObj, ruleObj)
					if err != nil {
						return nil, err
					}
				}
			} else if m2[field] == nil {
				continue
			} else {
				return nil, fmt.Errorf("expected field '%s' to be a map, but got %T", field, m2[field])
			}
		} else if _, ok := rule.(string); ok {
			if _, ok := m2[field]; !ok {
				continue
			}
			m1[field] = m2[field]
		}
	}

	return m1, nil
}
