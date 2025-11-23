package config

import (
	"encoding"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

type Duration time.Duration

// Duration returns the underlying time.Duration value.
func (d *Duration) Duration() time.Duration {
	if d == nil {
		return 0
	}
	return time.Duration(*d)
}

// String returns the string representation of the duration.
func (d *Duration) String() string {
	if d == nil {
		return ""
	}
	return time.Duration(*d).String()
}

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

	return d.UnmarshalText([]byte(s))
}

var _ yaml.Unmarshaler = (*Duration)(nil)
var _ encoding.TextUnmarshaler = (*Duration)(nil)
