package flags

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Shopify/sarama"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	// Kafka Consumer
	viper.SetDefault(kafkaConsumerMaxWaitTimeViperKey, kafkaConsumerMaxWaitTimeDefault)
	_ = viper.BindEnv(kafkaConsumerMaxWaitTimeViperKey, kafkaConsumerMaxWaitTimeEnv)
	viper.SetDefault(kafkaConsumerMaxProcessingTimeViperKey, kafkaConsumerMaxProcessingTimeDefault)
	_ = viper.BindEnv(kafkaConsumerMaxProcessingTimeViperKey, kafkaConsumerMaxProcessingTimeEnv)
	viper.SetDefault(kafkaConsumerGroupSessionTimeoutViperKey, kafkaConsumerGroupSessionTimeoutDefault)
	_ = viper.BindEnv(kafkaConsumerGroupSessionTimeoutViperKey, kafkaConsumerGroupSessionTimeoutEnv)
	viper.SetDefault(kafkaConsumerGroupHeartbeatIntervalViperKey, kafkaConsumerGroupHeartbeatIntervalDefault)
	_ = viper.BindEnv(kafkaConsumerGroupHeartbeatIntervalViperKey, kafkaConsumerGroupHeartbeatIntervalEnv)
	viper.SetDefault(kafkaConsumerGroupRebalanceTimeoutViperKey, kafkaConsumerGroupRebalanceTimeoutDefault)
	_ = viper.BindEnv(kafkaConsumerGroupRebalanceTimeoutViperKey, kafkaConsumerGroupRebalanceTimeoutEnv)
	viper.SetDefault(ConsumerGroupNameViperKey, consumerGroupNameDefault)
	_ = viper.BindEnv(ConsumerGroupNameViperKey, consumerGroupNameEnv)
	viper.SetDefault(kafkaNConsumerViperKey, kafkaNConsumerDefault)
	_ = viper.BindEnv(kafkaNConsumerViperKey, kafkaNConsumerEnv)
}

var rebalanceStrategy = map[string]sarama.BalanceStrategy{
	"Range":      sarama.BalanceStrategyRange,
	"RoundRobin": sarama.BalanceStrategyRoundRobin,
	"Sticky":     sarama.BalanceStrategySticky,
}

func KafkaConsumerFlags(f *pflag.FlagSet) {
	consumerGroupName(f)
	kafkaConsumerMaxWaitTime(f)
	kafkaConsumerMaxProcessingTime(f)
	kafkaConsumerGroupSessionTimeout(f)
	kafkaConsumerGroupHeartbeatInterval(f)
	kafkaConsumerGroupRebalanceTimeout(f)
	kafkaConsumerGroupRebalanceStrategy(f)
	kafkaNumberConsumers(f)
}

// Kafka Consumer group environment variables
const (
	consumerGroupNameFlag     = "consumer-group-name"
	ConsumerGroupNameViperKey = "kafka.consumer.group.name"
	consumerGroupNameEnv      = "KAFKA_CONSUMER_GROUP_NAME"
	consumerGroupNameDefault  = "group-sender"
)

// consumerGroupName register flag for a kafka consumer group name
func consumerGroupName(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Kafka consumer group name
Environment variable: %q`, consumerGroupNameEnv)
	f.String(consumerGroupNameFlag, consumerGroupNameDefault, desc)
	_ = viper.BindPFlag(ConsumerGroupNameViperKey, f.Lookup(consumerGroupNameFlag))
}

// Kafka Consumer MaxWaitTime wait time environment variables
const (
	kafkaConsumerMaxWaitTimeViperFlag = "kafka-consumer-max-wait-time"
	kafkaConsumerMaxWaitTimeViperKey  = "kafka.consumer.max.wait.time"
	kafkaConsumerMaxWaitTimeEnv       = "KAFKA_CONSUMER_MAX_WAIT_TIME"
	kafkaConsumerMaxWaitTimeDefault   = time.Millisecond * 250
)

// kafkaConsumerMaxWaitTime configuration
func kafkaConsumerMaxWaitTime(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`The maximum amount of time the broker will wait for Consumer.Fetch.Min bytes to become available before it returns fewer than that anyways.
Environment variable: %q in ms`, kafkaConsumerMaxWaitTimeEnv)
	f.Duration(kafkaConsumerMaxWaitTimeViperFlag, kafkaConsumerMaxWaitTimeDefault, desc)
	_ = viper.BindPFlag(kafkaConsumerMaxWaitTimeViperKey, f.Lookup(kafkaConsumerMaxWaitTimeViperFlag))
}

// kafkaConsumerMaxProcessingTime environment variables
const (
	kafkaConsumerMaxProcessingTimeFlag     = "kafka-consumer-max-processing-time"
	kafkaConsumerMaxProcessingTimeViperKey = "kafka.consumer.maxprocessingtime"
	kafkaConsumerMaxProcessingTimeEnv      = "KAFKA_CONSUMER_MAXPROCESSINGTIME"
	kafkaConsumerMaxProcessingTimeDefault  = time.Millisecond * 100
)

// kafkaConsumerMaxProcessingTime configuration
func kafkaConsumerMaxProcessingTime(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`The maximum amount of time the consumer expects a message takes to process for the user. If writing to the Messages channel takes longer than this, that partition will stop fetching more messages until it can proceed again.
Environment variable: %q`, kafkaConsumerMaxProcessingTimeEnv)
	f.Duration(kafkaConsumerMaxProcessingTimeFlag, kafkaConsumerMaxProcessingTimeDefault, desc)
	_ = viper.BindPFlag(kafkaConsumerMaxProcessingTimeViperKey, f.Lookup(kafkaConsumerMaxProcessingTimeFlag))
}

// kafkaConsumerGroupSessionTimeout environment variables
const (
	kafkaConsumerGroupSessionTimeoutFlag     = "kafka-consumer-group-session-timeout"
	kafkaConsumerGroupSessionTimeoutViperKey = "kafka.consumer.group.session.timeout"
	kafkaConsumerGroupSessionTimeoutEnv      = "KAFKA_CONSUMER_GROUP_SESSION_TIMEOUT"
	kafkaConsumerGroupSessionTimeoutDefault  = time.Second * 10
)

// kafkaConsumerGroupSessionTimeout configuration
func kafkaConsumerGroupSessionTimeout(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`The timeout used to detect consumer failures when using Kafka's group management facility. The consumer sends periodic heartbeats to indicate its liveness to the broker. If no heartbeats are received by the broker before the expiration of this session timeout, then the broker will remove this consumer from the group and initiate a rebalance.
Environment variable: %q`, kafkaConsumerGroupSessionTimeoutEnv)
	f.Duration(kafkaConsumerGroupSessionTimeoutFlag, kafkaConsumerGroupSessionTimeoutDefault, desc)
	_ = viper.BindPFlag(kafkaConsumerGroupSessionTimeoutViperKey, f.Lookup(kafkaConsumerGroupSessionTimeoutFlag))
}

// kafkaConsumerGroupHeartbeatInterval environment variables
const (
	kafkaConsumerGroupHeartbeatIntervalFlag     = "kafka-consumer-group-heartbeat-interval"
	kafkaConsumerGroupHeartbeatIntervalViperKey = "kafka.consumer.group.heartbeat.interval"
	kafkaConsumerGroupHeartbeatIntervalEnv      = "KAFKA_CONSUMER_GROUP_HEARTBEAT_INTERVAL"
	kafkaConsumerGroupHeartbeatIntervalDefault  = time.Second * 3
)

// kafkaConsumerGroupHeartbeatInterval configuration
func kafkaConsumerGroupHeartbeatInterval(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`The expected time between heartbeats to the consumer coordinator when using Kafka's group management facilities. Heartbeats are used to ensure that the consumer's session stays active and to facilitate rebalancing when new consumers join or leave the group.
Environment variable: %q in ms`, kafkaConsumerGroupHeartbeatIntervalEnv)
	f.Duration(kafkaConsumerGroupHeartbeatIntervalFlag, kafkaConsumerGroupHeartbeatIntervalDefault, desc)
	_ = viper.BindPFlag(kafkaConsumerGroupHeartbeatIntervalViperKey, f.Lookup(kafkaConsumerGroupHeartbeatIntervalFlag))
}

// kafkaConsumerGroupRebalanceTimeout environment variables
const (
	kafkaConsumerGroupRebalanceTimeoutFlag     = "kafka-consumer-group-rebalance-timeout"
	kafkaConsumerGroupRebalanceTimeoutViperKey = "kafka.consumer.group.rebalance.timeout"
	kafkaConsumerGroupRebalanceTimeoutEnv      = "KAFKA_CONSUMER_GROUP_REBALANCE_TIMEOUT"
	kafkaConsumerGroupRebalanceTimeoutDefault  = time.Second * 60
)

// kafkaConsumerGroupRebalanceTimeout configuration
func kafkaConsumerGroupRebalanceTimeout(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`The maximum allowed time for each worker to join the group once a rebalance has begun. This is basically a limit on the amount of time needed for all tasks to flush any pending data and commit offsets. If the timeout is exceeded, then the worker will be removed from the group, which will cause offset commit failures
Environment variable: %q`, kafkaConsumerGroupRebalanceTimeoutEnv)
	f.Duration(kafkaConsumerGroupRebalanceTimeoutFlag, kafkaConsumerGroupRebalanceTimeoutDefault, desc)
	_ = viper.BindPFlag(kafkaConsumerGroupRebalanceTimeoutViperKey, f.Lookup(kafkaConsumerGroupRebalanceTimeoutFlag))
}

// kafkaConsumerGroupRebalanceStrategy environment variables
const (
	kafkaConsumerGroupRebalanceStrategyFlag     = "kafka-consumer-group-rebalance-strategy"
	kafkaConsumerGroupRebalanceStrategyViperKey = "kafka.consumer.group.rebalance.strategy"
	kafkaConsumerGroupRebalanceStrategyEnv      = "KAFKA_CONSUMER_GROUP_REBALANCE_STRATEGY"
	kafkaConsumerGroupRebalanceStrategyDefault  = "Range"
)

// kafkaConsumerGroupRebalanceStrategy configuration
func kafkaConsumerGroupRebalanceStrategy(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Strategy for allocating topic partitions to members (one of %q).
Environment variable: %q`, reflect.ValueOf(rebalanceStrategy).MapKeys(), kafkaConsumerGroupRebalanceStrategyEnv)
	f.String(kafkaConsumerGroupRebalanceStrategyFlag, kafkaConsumerGroupRebalanceStrategyDefault, desc)
	_ = viper.BindPFlag(kafkaConsumerGroupRebalanceStrategyViperKey, f.Lookup(kafkaConsumerGroupRebalanceStrategyFlag))
}

const (
	kafkaNConsumersFlag    = "kafka-num-consumers"
	kafkaNConsumerViperKey = "kafka.num-consumers"
	kafkaNConsumerDefault  = uint8(1)
	kafkaNConsumerEnv      = "KAFKA_NUM_CONSUMERS"
)

func kafkaNumberConsumers(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Number of parallel kafka consumers to initialize.
Environment variable: %q`, kafkaNConsumerEnv)
	f.Uint8(kafkaNConsumersFlag, kafkaNConsumerDefault, desc)
	_ = viper.BindPFlag(kafkaNConsumerViperKey, f.Lookup(kafkaNConsumersFlag))
}
