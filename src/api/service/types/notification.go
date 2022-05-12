package types

type AckNotificationRequestMessage struct {
	UUID string `json:"uuid,omitempty" validate:"required" example:"b4374e6f-b28a-4bad-b4fe-bda36eaf849c"`
}
