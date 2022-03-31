package flags

import (
	"fmt"

	"github.com/consensys/orchestrate/src/infra/kafka/sarama"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	// Kafka general parameters
	viper.SetDefault(KafkaURLViperKey, kafkaURLDefault)
	_ = viper.BindEnv(KafkaURLViperKey, KafkaURLEnv)

	// Kafka SASL
	viper.SetDefault(KafkaSASLEnabledViperKey, kafkaSASLEnabledDefault)
	_ = viper.BindEnv(KafkaSASLEnabledViperKey, kafkaSASLEnabledEnv)
	viper.SetDefault(KafkaSASLMechanismViperKey, kafkaSASLMechanismDefault)
	_ = viper.BindEnv(KafkaSASLMechanismViperKey, kafkaSASLMechanismEnv)
	viper.SetDefault(KafkaSASLHandshakeViperKey, kafkaSASLHandshakeDefault)
	_ = viper.BindEnv(KafkaSASLHandshakeViperKey, kafkaSASLHandshakeEnv)
	viper.SetDefault(KafkaSASLUserViperKey, kafkaSASLUserDefault)
	_ = viper.BindEnv(KafkaSASLUserViperKey, kafkaSASLUserEnv)
	viper.SetDefault(KafkaSASLPasswordViperKey, kafkaSASLPasswordDefault)
	_ = viper.BindEnv(KafkaSASLPasswordViperKey, kafkaSASLPasswordEnv)
	viper.SetDefault(KafkaSASLSCRAMAuthzIDViperKey, kafkaSASLSCRAMAuthzIDDefault)
	_ = viper.BindEnv(KafkaSASLSCRAMAuthzIDViperKey, kafkaSASLSCRAMAuthzIDEnv)

	// Kafka TLS
	viper.SetDefault(KafkaTLSEnableViperKey, kafkaTLSEnableDefault)
	_ = viper.BindEnv(KafkaTLSEnableViperKey, kafkaTLSEnableEnv)
	viper.SetDefault(KafkaTLSInsecureSkipVerifyViperKey, kafkaTLSInsecureSkipVerifyDefault)
	_ = viper.BindEnv(KafkaTLSInsecureSkipVerifyViperKey, kafkaTLSInsecureSkipVerifyEnv)
	viper.SetDefault(KafkaTLSClientCertFilePathViperKey, kafkaTLSClientCertFilePathDefault)
	_ = viper.BindEnv(KafkaTLSClientCertFilePathViperKey, kafkaTLSClientCertFilePathEnv)
	viper.SetDefault(KafkaTLSClientKeyFilePathViperKey, kafkaTLSClientKeyFilePathDefault)
	_ = viper.BindEnv(KafkaTLSClientKeyFilePathViperKey, kafkaTLSClientKeyFilePathEnv)
	viper.SetDefault(KafkaTLSCACertFilePathViperKey, kafkaTLSCACertFilePathDefault)
	_ = viper.BindEnv(KafkaTLSCACertFilePathViperKey, kafkaTLSCACertFilePathEnv)

	// Kafka
	viper.SetDefault(KafkaVersionViperKey, kafkaVersionDefault)
	_ = viper.BindEnv(KafkaVersionViperKey, kafkaVersionEnv)
}

func KafkaFlags(f *pflag.FlagSet) {
	kafkaSASLFlags(f)
	kafkaTLSFlags(f)
	kafkaVersion(f)
	kafkaURL(f)
}

var (
	kafkaURLFlag     = "kafka-url"
	KafkaURLViperKey = "kafka.url"
	kafkaURLDefault  = []string{"localhost:9092"}
	KafkaURLEnv      = "KAFKA_URL"
)

// KafkaURL register flag for Kafka server
func kafkaURL(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`URL (addresses) of Kafka server(s) to connect to.
Environment variable: %q`, KafkaURLEnv)
	f.StringSlice(kafkaURLFlag, kafkaURLDefault, desc)
	_ = viper.BindPFlag(KafkaURLViperKey, f.Lookup(kafkaURLFlag))
}

// kafkaSASLFlags register flags for SASL authentication
func kafkaSASLFlags(f *pflag.FlagSet) {
	kafkaSASLEnable(f)
	kafkaSASLMechanism(f)
	kafkaSASLHandshake(f)
	kafkaSASLUser(f)
	kafkaSASLPassword(f)
	kafkaSASLSCRAMAuthzID(f)
}

// Kafka SASL Enable environment variables
const (
	kafkaSASLEnabledFlag     = "kafka-sasl-enabled"
	KafkaSASLEnabledViperKey = "kafka.sasl.enabled"
	kafkaSASLEnabledEnv      = "KAFKA_SASL_ENABLED"
	kafkaSASLEnabledDefault  = false
)

// kafkaSASLEnable register flag
func kafkaSASLEnable(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Whether or not to use SASL authentication when connecting to the broker
Environment variable: %q`, kafkaSASLEnabledEnv)
	f.Bool(kafkaSASLEnabledFlag, kafkaSASLEnabledDefault, desc)
	_ = viper.BindPFlag(KafkaSASLEnabledViperKey, f.Lookup(kafkaSASLEnabledFlag))
}

// Kafka SASL mechanism environment variables
const (
	kafkaSASLMechanismFlag     = "kafka-sasl-mechanism"
	KafkaSASLMechanismViperKey = "kafka.sasl.mechanism"
	kafkaSASLMechanismEnv      = "KAFKA_SASL_MECHANISM"
	kafkaSASLMechanismDefault  = ""
)

// kafkaSASLMechanism register flag
func kafkaSASLMechanism(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`SASLMechanism is the name of the enabled SASL mechanism. Possible values: OAUTHBEARER, PLAIN (defaults to PLAIN).
Environment variable: %q`, kafkaSASLMechanismEnv)
	f.String(kafkaSASLMechanismFlag, kafkaSASLMechanismDefault, desc)
	_ = viper.BindPFlag(KafkaSASLMechanismViperKey, f.Lookup(kafkaSASLMechanismFlag))
}

// Kafka SASL Handshake environment variables
const (
	kafkaSASLHandshakeFlag     = "kafka-sasl-handshake"
	KafkaSASLHandshakeViperKey = "kafka.sasl.handshake"
	kafkaSASLHandshakeEnv      = "KAFKA_SASL_HANDSHAKE"
	kafkaSASLHandshakeDefault  = true
)

// kafkaSASLHandshake register flag
func kafkaSASLHandshake(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Whether or not to send the Kafka SASL handshake first if enabled (defaults to true). You should only set this to false if you're using a non-Kafka SASL proxy.
Environment variable: %q`, kafkaSASLHandshakeEnv)
	f.Bool(kafkaSASLHandshakeFlag, kafkaSASLHandshakeDefault, desc)
	_ = viper.BindPFlag(KafkaSASLHandshakeViperKey, f.Lookup(kafkaSASLHandshakeFlag))
}

// Kafka SASL User environment variables
const (
	kafkaSASLUserFlag     = "kafka-sasl-user"
	KafkaSASLUserViperKey = "kafka.sasl.user"
	kafkaSASLUserEnv      = "KAFKA_SASL_USER"
	kafkaSASLUserDefault  = ""
)

// kafkaSASLUser register flag
func kafkaSASLUser(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Username for SASL/PLAIN or SASL/SCRAM auth.
Environment variable: %q`, kafkaSASLUserEnv)
	f.String(kafkaSASLUserFlag, kafkaSASLUserDefault, desc)
	_ = viper.BindPFlag(KafkaSASLUserViperKey, f.Lookup(kafkaSASLUserFlag))
}

// Kafka SASL Password environment variables
const (
	kafkaSASLPasswordFlag     = "kafka-sasl-password"
	KafkaSASLPasswordViperKey = "kafka.sasl.password"
	kafkaSASLPasswordEnv      = "KAFKA_SASL_PASSWORD"
	kafkaSASLPasswordDefault  = ""
)

// kafkaSASLPassword register flag
func kafkaSASLPassword(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Password for SASL/PLAIN or SASL/SCRAM auth.
Environment variable: %q`, kafkaSASLPasswordEnv)
	f.String(kafkaSASLPasswordFlag, kafkaSASLPasswordDefault, desc)
	_ = viper.BindPFlag(KafkaSASLPasswordViperKey, f.Lookup(kafkaSASLPasswordFlag))
}

// Kafka SASL SCRAMAuthzID environment variables
const (
	kafkaSASLSCRAMAuthzIDFlag     = "kafka-sasl-scramauthzid"
	KafkaSASLSCRAMAuthzIDViperKey = "kafka.sasl.scramauthzid"
	kafkaSASLSCRAMAuthzIDEnv      = "KAFKA_SASL_SCRAMAUTHZID"
	kafkaSASLSCRAMAuthzIDDefault  = ""
)

// kafkaSASLSCRAMAuthzID register flag
func kafkaSASLSCRAMAuthzID(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Authz id used for SASL/SCRAM authentication
Environment variable: %q`, kafkaSASLSCRAMAuthzIDEnv)
	f.String(kafkaSASLSCRAMAuthzIDFlag, kafkaSASLSCRAMAuthzIDDefault, desc)
	_ = viper.BindPFlag(KafkaSASLSCRAMAuthzIDViperKey, f.Lookup(kafkaSASLSCRAMAuthzIDFlag))
}

// kafkaTLSFlags register flags for SASL and SSL
func kafkaTLSFlags(f *pflag.FlagSet) {
	kafkaTLSEnable(f)
	kafkaTLSInsecureSkipVerify(f)
	kafkaTLSClientCertFilePath(f)
	kafkaTLSClientKeyFilePath(f)
	kafkaTLSCaCertFilePath(f)
}

// Kafka TLS Enable environment variables
const (
	kafkaTLSEnableFlag     = "kafka-tls-enabled"
	KafkaTLSEnableViperKey = "kafka.tls.enabled"
	kafkaTLSEnableEnv      = "KAFKA_TLS_ENABLED"
	kafkaTLSEnableDefault  = false
)

// kafkaTLSEnable register flag
func kafkaTLSEnable(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Whether or not to use TLS when connecting to the broker (defaults to false).
Environment variable: %q`, kafkaTLSEnableEnv)
	f.Bool(kafkaTLSEnableFlag, kafkaTLSEnableDefault, desc)
	_ = viper.BindPFlag(KafkaTLSEnableViperKey, f.Lookup(kafkaTLSEnableFlag))
}

// Kafka TLS InsecureSkipVerify environment variables
const (
	kafkaTLSInsecureSkipVerifyFlag     = "kafka-tls-insecure-skip-verify"
	KafkaTLSInsecureSkipVerifyViperKey = "kafka.tls.insecure.skip.verify"
	kafkaTLSInsecureSkipVerifyEnv      = "KAFKA_TLS_INSECURE_SKIP_VERIFY"
	kafkaTLSInsecureSkipVerifyDefault  = false
)

// kafkaTLSInsecureSkipVerify register flag
func kafkaTLSInsecureSkipVerify(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Controls whether a client verifies the server's certificate chain and host name. If InsecureSkipVerify is true, TLS accepts any certificate presented by the server and any host name in that certificate. In this mode, TLS is susceptible to man-in-the-middle attacks. This should be used only for testing.
Environment variable: %q`, kafkaTLSInsecureSkipVerifyEnv)
	f.Bool(kafkaTLSInsecureSkipVerifyFlag, kafkaTLSInsecureSkipVerifyDefault, desc)
	_ = viper.BindPFlag(KafkaTLSInsecureSkipVerifyViperKey, f.Lookup(kafkaTLSInsecureSkipVerifyFlag))
}

// Kafka TLS ClientCertFilePath environment variables
const (
	kafkaTLSClientCertFilePathFlag     = "kafka-tls-client-cert-file"
	KafkaTLSClientCertFilePathViperKey = "kafka.tls.client.cert.file"
	kafkaTLSClientCertFilePathEnv      = "KAFKA_TLS_CLIENT_CERT_FILE"
	kafkaTLSClientCertFilePathDefault  = ""
)

// kafkaTLSClientCertFilePath register flag
func kafkaTLSClientCertFilePath(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Client Cert File Path.
Environment variable: %q`, kafkaTLSClientCertFilePathEnv)
	f.String(kafkaTLSClientCertFilePathFlag, kafkaTLSClientCertFilePathDefault, desc)
	_ = viper.BindPFlag(KafkaTLSClientCertFilePathViperKey, f.Lookup(kafkaTLSClientCertFilePathFlag))
}

// Kafka TLS ClientKeyFilePath environment variables
const (
	kafkaTLSClientKeyFilePathFlag     = "kafka-tls-client-key-file"
	KafkaTLSClientKeyFilePathViperKey = "kafka.tls.client.key.file"
	kafkaTLSClientKeyFilePathEnv      = "KAFKA_TLS_CLIENT_KEY_FILE"
	kafkaTLSClientKeyFilePathDefault  = ""
)

// kafkaTLSClientKeyFilePath register flag
func kafkaTLSClientKeyFilePath(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Client key file Path.
Environment variable: %q`, kafkaTLSClientKeyFilePathEnv)
	f.String(kafkaTLSClientKeyFilePathFlag, kafkaTLSClientKeyFilePathDefault, desc)
	_ = viper.BindPFlag(KafkaTLSClientKeyFilePathViperKey, f.Lookup(kafkaTLSClientKeyFilePathFlag))
}

// Kafka TLS CACertFilePath environment variables
const (
	kafkaTLSCACertFilePathFlag     = "kafka-tls-ca-cert-file"
	KafkaTLSCACertFilePathViperKey = "kafka.tls.ca.cert.file"
	kafkaTLSCACertFilePathEnv      = "KAFKA_TLS_CA_CERT_FILE"
	kafkaTLSCACertFilePathDefault  = ""
)

// kafkaTLSCaCertFilePath register flag
func kafkaTLSCaCertFilePath(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`CA cert file Path.
Environment variable: %q`, kafkaTLSCACertFilePathEnv)
	f.String(kafkaTLSCACertFilePathFlag, kafkaTLSCACertFilePathDefault, desc)
	_ = viper.BindPFlag(KafkaTLSCACertFilePathViperKey, f.Lookup(kafkaTLSCACertFilePathFlag))
}

// kafkaVersion environment variables
const (
	kafkaVersionFlag     = "kafka-version"
	KafkaVersionViperKey = "kafka.version"
	kafkaVersionEnv      = "KAFKA_VERSION"
	kafkaVersionDefault  = "1.0.0"
)

// kafkaVersion configuration
func kafkaVersion(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`The version of Kafka that Sarama will assume it is running against. Defaults to the oldest supported stable version. Since Kafka provides backwards-compatibility, setting it to a version older than you have will not break anything, although it may prevent you from using the latest features. Setting it to a version greater than you are actually running may lead to random breakage.
Environment variable: %q`, kafkaVersionEnv)
	f.String(kafkaVersionFlag, kafkaVersionDefault, desc)
	_ = viper.BindPFlag(KafkaVersionViperKey, f.Lookup(kafkaVersionFlag))
}

func NewKafkaConfig(vipr *viper.Viper) *sarama.Config {
	return &sarama.Config{
		Version:           vipr.GetString(KafkaVersionViperKey),
		ClientID:          "sarama-orchestrate",
		SASLEnable:        vipr.GetBool(KafkaSASLEnabledViperKey),
		SASLMechanism:     vipr.GetString(KafkaSASLMechanismViperKey),
		SASLHandshake:     vipr.GetBool(KafkaSASLHandshakeViperKey),
		SASLUser:          vipr.GetString(KafkaSASLUserViperKey),
		SASLPassword:      vipr.GetString(KafkaSASLPasswordViperKey),
		SASLSCRAMAuthzID:  vipr.GetString(KafkaSASLSCRAMAuthzIDViperKey),
		TSLEnable:         vipr.GetBool(KafkaTLSEnableViperKey),
		TLSCert:           vipr.GetString(KafkaTLSClientCertFilePathViperKey),
		TLSKey:            vipr.GetString(KafkaTLSClientKeyFilePathViperKey),
		TLSCA:             vipr.GetString(KafkaTLSCACertFilePathViperKey),
		TLSSkipVerify:     vipr.GetBool(KafkaTLSInsecureSkipVerifyViperKey),
		URLs:              viper.GetStringSlice(KafkaURLViperKey),
		GroupName:         vipr.GetString(ConsumerGroupNameViperKey),
		ReturnErrors:      true,
		ReturnSuccesses:   true,
		AutoCommit:        false,
		MaxWaitTime:       vipr.GetDuration(kafkaConsumerMaxWaitTimeViperKey),
		MaxProcessingTime: vipr.GetDuration(kafkaConsumerMaxProcessingTimeViperKey),
		SessionTimeout:    vipr.GetDuration(kafkaConsumerGroupSessionTimeoutViperKey),
		HeartbeatInterval: vipr.GetDuration(kafkaConsumerGroupHeartbeatIntervalViperKey),
		RebalanceTimeout:  vipr.GetDuration(kafkaConsumerGroupRebalanceTimeoutViperKey),
		RebalanceStrategy: vipr.GetString(kafkaConsumerGroupRebalanceStrategyViperKey),
	}
}
