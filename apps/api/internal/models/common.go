package models

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type JSONB json.RawMessage

func NewJSONB(v any) (JSONB, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	if !json.Valid(data) {
		return nil, errors.New("models: invalid JSON")
	}

	return JSONB(data), nil
}

func (j JSONB) Value() (driver.Value, error) {
	if j.IsNull() {
		return nil, nil
	}

	if !json.Valid(j) {
		return nil, errors.New("models: invalid JSONB")
	}

	return []byte(j), nil
}

func (j *JSONB) Scan(src any) error {
	if j == nil {
		return errors.New("models: JSONB: Scan on nil pointer")
	}

	if src == nil {
		*j = nil
		return nil
	}

	switch value := src.(type) {
	case []byte:
		if !json.Valid(value) {
			return errors.New("models: invalid JSONB")
		}

		*j = append((*j)[:0], value...)

	case string:
		if !json.Valid([]byte(value)) {
			return errors.New("models: invalid JSONB")
		}

		*j = append((*j)[:0], value...)

	default:
		return fmt.Errorf("models: cannot scan type %T into JSONB", src)
	}

	return nil
}

func (j JSONB) MarshalJSON() ([]byte, error) {
	if j.IsNull() {
		return []byte("null"), nil
	}

	if !json.Valid(j) {
		return nil, errors.New("models: invalid JSONB")
	}

	return []byte(j), nil
}

func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("models: JSONB: UnmarshalJSON on nil pointer")
	}

	if string(data) == "null" {
		*j = nil
		return nil
	}

	if !json.Valid(data) {
		return errors.New("models: invalid JSON")
	}

	*j = append((*j)[:0], data...)

	return nil
}

func (j JSONB) Unmarshal(v any) error {
	if j.IsNull() {
		return nil
	}

	return json.Unmarshal(j, v)
}

func (j JSONB) IsNull() bool {
	return len(j) == 0 || string(j) == "null"
}

func (j JSONB) IsObject() bool {
	trimmed := bytes.TrimSpace(j)
	return len(trimmed) > 0 && trimmed[0] == '{'
}

func (j JSONB) IsArray() bool {
	trimmed := bytes.TrimSpace(j)
	return len(trimmed) > 0 && trimmed[0] == '['
}

type StringArray = []string

type DateOnly time.Time

func NewDateOnly(t time.Time) DateOnly {
	year, month, day := t.UTC().Date()

	return DateOnly(
		time.Date(
			year,
			month,
			day,
			0,
			0,
			0,
			0,
			time.UTC,
		),
	)
}

func (d DateOnly) Time() time.Time {
	return time.Time(d)
}

func (d DateOnly) String() string {
	return time.Time(d).Format("2006-01-02")
}

func (d DateOnly) IsZero() bool {
	return time.Time(d).IsZero()
}

func (d DateOnly) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}

	return d.String(), nil
}

func (d *DateOnly) Scan(src any) error {
	if d == nil {
		return errors.New("models: DateOnly: Scan on nil pointer")
	}

	if src == nil {
		*d = DateOnly{}
		return nil
	}

	switch value := src.(type) {
	case time.Time:
		*d = NewDateOnly(value)

	case string:
		t, err := time.Parse("2006-01-02", value)
		if err != nil {
			return fmt.Errorf("models: invalid DateOnly %q: %w", value, err)
		}

		*d = NewDateOnly(t)

	case []byte:
		t, err := time.Parse("2006-01-02", string(value))
		if err != nil {
			return fmt.Errorf("models: invalid DateOnly %q: %w", value, err)
		}

		*d = NewDateOnly(t)

	default:
		return fmt.Errorf("models: cannot scan type %T into DateOnly", src)
	}

	return nil
}

func (d DateOnly) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(d.String())
}

func (d *DateOnly) UnmarshalJSON(data []byte) error {
	if d == nil {
		return errors.New("models: DateOnly: UnmarshalJSON on nil pointer")
	}

	if string(data) == "null" {
		*d = DateOnly{}
		return nil
	}

	var value string

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if value == "" {
		*d = DateOnly{}
		return nil
	}

	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return fmt.Errorf("models: invalid DateOnly %q: %w", value, err)
	}

	*d = NewDateOnly(t)

	return nil
}

type TimeOfDay time.Time

func NewTimeOfDay(t time.Time) TimeOfDay {
	hour, min, sec := t.Clock()

	return TimeOfDay(time.Date(0, 1, 1, hour, min, sec, 0, time.UTC))
}

func (t TimeOfDay) Time() time.Time {
	return time.Time(t)
}

func (t TimeOfDay) String() string {
	return time.Time(t).Format("15:04:05")
}

func (t TimeOfDay) IsZero() bool {
	return time.Time(t).IsZero()
}

func (t TimeOfDay) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}

	return t.String(), nil
}

func parseTimeOfDay(value string) (TimeOfDay, error) {
	layouts := []string{"15:04:05.999999", "15:04:05", "15:04"}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return NewTimeOfDay(parsed), nil
		}
	}

	return TimeOfDay{}, fmt.Errorf("models: invalid TimeOfDay %q", value)
}

func (t *TimeOfDay) Scan(src any) error {
	if t == nil {
		return errors.New("models: TimeOfDay: Scan on nil pointer")
	}

	if src == nil {
		*t = TimeOfDay{}
		return nil
	}

	switch value := src.(type) {
	case time.Time:
		*t = NewTimeOfDay(value)

	case string:
		parsed, err := parseTimeOfDay(value)
		if err != nil {
			return err
		}

		*t = parsed

	case []byte:
		parsed, err := parseTimeOfDay(string(value))
		if err != nil {
			return err
		}

		*t = parsed

	default:
		return fmt.Errorf("models: cannot scan type %T into TimeOfDay", src)
	}

	return nil
}

func (t TimeOfDay) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(t.String())
}

func (t *TimeOfDay) UnmarshalJSON(data []byte) error {
	if t == nil {
		return errors.New("models: TimeOfDay: UnmarshalJSON on nil pointer")
	}

	if string(data) == "null" {
		*t = TimeOfDay{}
		return nil
	}

	var value string

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	if value == "" {
		*t = TimeOfDay{}
		return nil
	}

	parsed, err := parseTimeOfDay(value)
	if err != nil {
		return err
	}

	*t = parsed

	return nil
}
