package echosight

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Source: https://stackoverflow.com/a/54571600

type Duration time.Duration

func (d Duration) String() string {
	duration := time.Duration(d)
	return fmt.Sprintf("%.0fs", duration.Seconds())
}

func (d *Duration) setDurationFromValue(value interface{}) error {
	switch v := value.(type) {
	case float64:
		*d = Duration(time.Duration(v))
		return nil
	case int64:
		*d = Duration(time.Duration(v))
		return nil
	case string:
		tmp, err := time.ParseDuration(v)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

// JSON
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	return d.setDurationFromValue(v)
}

// TOML
func (d Duration) MarshalTOML() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d *Duration) UnmarshalTOML(data interface{}) error {
	return d.setDurationFromValue(data)
}

// YAML
func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

func (d *Duration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v interface{}
	if err := unmarshal(&v); err != nil {
		return err
	}
	return d.setDurationFromValue(v)
}
