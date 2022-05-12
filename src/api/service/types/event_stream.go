package types

import (
	"time"

	"github.com/consensys/orchestrate/src/entities"
)

type CreateEventStreamRequest struct {
	Channel string            `json:"channel" validate:"required,isChannel" example:"webhook"`
	Name    string            `json:"name" validate:"required" example:"my-webhook-stream"`
	Chain   string            `json:"chain,omitempty" validate:"omitempty" example:"mainnet"`
	Webhook *WebhookRequest   `json:"webhook,omitempty" validate:"omitempty"`
	Kafka   *KafkaRequest     `json:"kafka,omitempty" validate:"omitempty"`
	Labels  map[string]string `json:"labels,omitempty" validate:"omitempty"`
}

type WebhookRequest struct {
	URL     string            `json:"url" validate:"required,url" example:"https://my-event-steam-endpoint.com"`
	Headers map[string]string `json:"headers,omitempty" validate:"omitempty"`
}

type KafkaRequest struct {
	Topic string `json:"topic" validate:"required" example:"my-notification-topic"`
}

func (r *CreateEventStreamRequest) ToEntity() *entities.EventStream {
	es := &entities.EventStream{
		Name:    r.Name,
		Channel: entities.EventStreamChannel(r.Channel),
		Status:  entities.EventStreamStatusLive,
		Labels:  r.Labels,
	}

	switch es.Channel {
	case entities.EventStreamChannelWebhook:
		es.Webhook = &entities.EventStreamWebhookSpec{
			URL:     r.Webhook.URL,
			Headers: r.Webhook.Headers,
		}
	case entities.EventStreamChannelKafka:
		es.Kafka = &entities.EventStreamKafkaSpec{
			Topic: r.Kafka.Topic,
		}
	}

	return es
}

func (r *CreateEventStreamRequest) Validate() bool {
	switch r.Channel {
	case string(entities.EventStreamChannelWebhook):
		return r.Webhook != nil
	case string(entities.EventStreamChannelKafka):
		return r.Kafka != nil
	}

	return false
}

type UpdateEventStreamRequest struct {
	Name    string            `json:"name,omitempty" validate:"omitempty" example:"my-kafka-stream"`
	URL     string            `json:"url,omitempty" validate:"omitempty,url" example:"https://my-event-steam-endpoint.com"`
	Headers map[string]string `json:"headers,omitempty" validate:"omitempty"`
	Topic   string            `json:"topic,omitempty" validate:"omitempty" example:"my-notification-topic"`
	Status  string            `json:"status,omitempty" validate:"omitempty,isEventStreamStatus" example:"PAUSED"`
	Labels  map[string]string `json:"labels,omitempty" validate:"omitempty"`
}

func (r *UpdateEventStreamRequest) ToEntity(uuid string) *entities.EventStream {
	es := &entities.EventStream{
		UUID:   uuid,
		Name:   r.Name,
		Status: entities.EventStreamStatus(r.Status),
		Labels: r.Labels,
	}

	if r.URL != "" || r.Headers != nil {
		es.Channel = entities.EventStreamChannelWebhook
		es.Webhook = &entities.EventStreamWebhookSpec{
			URL:     r.URL,
			Headers: r.Headers,
		}
	}

	if r.Topic != "" {
		es.Channel = entities.EventStreamChannelKafka
		es.Kafka = &entities.EventStreamKafkaSpec{
			Topic: r.Topic,
		}
	}

	return es
}

type EventStreamResponse struct {
	UUID      string            `json:"uuid"  example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
	Name      string            `json:"name"  example:"my-stream"`
	ChainUUID string            `json:"chainUUID,omitempty"  example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
	TenantID  string            `json:"tenantID"  example:"foo"`
	OwnerID   string            `json:"ownerID,omitempty"  example:"foo"`
	Specs     interface{}       `json:"specs"`
	Channel   string            `json:"channel"  example:"webhook"`
	Status    string            `json:"status"  example:"LIVE"`
	Labels    map[string]string `json:"labels,omitempty"`
	CreatedAt time.Time         `json:"createdAt"  example:"2020-07-09T12:35:42.115395Z"`
	UpdatedAt time.Time         `json:"updatedAt"  example:"2020-07-09T12:35:42.115395Z"`
}

func NewEventStreamResponse(e *entities.EventStream) *EventStreamResponse {
	resp := &EventStreamResponse{
		UUID:      e.UUID,
		Name:      e.Name,
		ChainUUID: e.ChainUUID,
		TenantID:  e.TenantID,
		OwnerID:   e.OwnerID,
		Channel:   string(e.Channel),
		Status:    string(e.Status),
		Labels:    e.Labels,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}

	switch e.Channel {
	case entities.EventStreamChannelKafka:
		resp.Specs = e.Kafka
	case entities.EventStreamChannelWebhook:
		resp.Specs = e.Webhook
	}

	return resp
}

func NewEventStreamResponses(eventStreams []*entities.EventStream) []*EventStreamResponse {
	response := []*EventStreamResponse{}
	for _, e := range eventStreams {
		response = append(response, NewEventStreamResponse(e))
	}

	return response
}

type SuspendEventStreamRequestMessage struct {
	UUID string `json:"uuid,omitempty" validate:"required" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
}
