package config

import (
	"encoding"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type Duration time.Duration

func (d *Duration) UnmarshalText(text []byte) error {
	t, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("can't parse duration: %w", err)
	}
	*d = Duration(t)
	return nil
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return fmt.Errorf("can't unmarshal duration: %w", err)
	}

	t, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("can't parse duration: %w", err)
	}
	*d = Duration(t)
	return nil
}

var _ yaml.Unmarshaler = (*Duration)(nil)
var _ encoding.TextUnmarshaler = (*Duration)(nil)
