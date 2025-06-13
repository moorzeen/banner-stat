package model

import (
	"encoding/json"
	"strconv"
	"time"
)

const timeFormat = "2006-01-02T15:04:05"

type Banner struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ClickStats struct {
	Timestamp TimeNoTZ `json:"ts"`
	Value     int      `json:"v"`
}

type StatsRequest struct {
	From TimeNoTZ `json:"from"`
	To   TimeNoTZ `json:"to"`
}

type TimeNoTZ time.Time

func (t *TimeNoTZ) asTime() time.Time {
	return time.Time(*t)
}

func (t *TimeNoTZ) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.asTime().Format(timeFormat))
}

func (t *TimeNoTZ) UnmarshalJSON(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}

	tt, err := time.Parse(timeFormat, s)
	if err != nil {
		return err
	}

	*t = TimeNoTZ(tt)
	return nil
}

type StatsResponse struct {
	Stats []ClickStats `json:"stats"`
}
