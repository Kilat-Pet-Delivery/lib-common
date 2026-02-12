package kafka

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CloudEvent is a lightweight envelope for domain events following the CloudEvents spec.
type CloudEvent struct {
	ID              string          `json:"id"`
	Source          string          `json:"source"`
	Type            string          `json:"type"`
	Time            time.Time       `json:"time"`
	DataContentType string          `json:"datacontenttype"`
	Data            json.RawMessage `json:"data"`
}

// NewCloudEvent creates a new CloudEvent.
func NewCloudEvent(source, eventType string, data interface{}) (CloudEvent, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return CloudEvent{}, fmt.Errorf("failed to marshal event data: %w", err)
	}

	return CloudEvent{
		ID:              uuid.New().String(),
		Source:          source,
		Type:            eventType,
		Time:            time.Now().UTC(),
		DataContentType: "application/json",
		Data:            dataBytes,
	}, nil
}

// ParseCloudEvent deserializes bytes into a CloudEvent.
func ParseCloudEvent(data []byte) (CloudEvent, error) {
	var event CloudEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return CloudEvent{}, fmt.Errorf("failed to parse cloud event: %w", err)
	}
	return event, nil
}

// ParseData deserializes the event data into the given target.
func (e CloudEvent) ParseData(target interface{}) error {
	return json.Unmarshal(e.Data, target)
}
