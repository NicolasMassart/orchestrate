package listener

import (
	"github.com/consensys/orchestrate/src/infra/messenger"
)

var (
	EventLogsMessageType messenger.ConsumerRequestMessageType = "event-logs"
	JobUpdateMessageType messenger.ConsumerRequestMessageType = "job-update"
)
