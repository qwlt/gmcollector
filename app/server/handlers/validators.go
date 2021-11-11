package handlers

import (
	"time"

	"github.com/google/uuid"
)

type MeasurementValidator struct {
	DeviceID  uuid.UUID `json:"id" validate:"required"`
	Value     float64   `json:"value" validate:"required"`
	Timestamp time.Time `json:"timestamp" validate:"required"`
}
