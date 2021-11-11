package models

import (
	"time"

	"github.com/google/uuid"
)

type Model interface {
	Flatten() []interface{}
}

// Measurement - basic model to recieve and process data from user devices
type Measurement struct {
	DeviceID  uuid.UUID
	Value     float64
	Timestamp time.Time
	Metadata  map[string]interface{}
}

func (m Measurement) Flatten() []interface{} {
	fields := make([]interface{}, 0, 8)
	fields = append(fields, m.DeviceID, m.Timestamp, m.Value)
	return fields
}
