package postgres

import (
	"fmt"
	"time"

	"github.com/ConsenSys/orchestrate/pkg/tls"
	"github.com/ConsenSys/orchestrate/pkg/tls/certificate"
	"github.com/go-pg/pg/v9"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault(DBUserViperKey, dbUserDefault)
	_ = viper.BindEnv(DBUserViperKey, dbUserEnv)
	viper.SetDefault(DBPasswordViperKey, dbPasswordDefault)
	_ = viper.BindEnv(DBPasswordViperKey, dbPasswordEnv)
	viper.SetDefault(DBDatabaseViperKey, dbDatabaseDefault)
	_ = viper.BindEnv(DBDatabaseViperKey, dbDatabaseEnv)
	viper.SetDefault(DBHostViperKey, dbHostDefault)
	_ = viper.BindEnv(DBHostViperKey, dbHostEnv)
	viper.SetDefault(DBPortViperKey, dbPortDefault)
	_ = viper.BindEnv(DBPortViperKey, dbPortEnv)
	viper.SetDefault(DBPoolSizeViperKey, dbPoolSizeDefault)
	_ = viper.BindEnv(DBPoolSizeViperKey, dbPoolSizeEnv)
	viper.SetDefault(DBPoolTimeoutViperKey, dbPoolTimeoutDefault)
	_ = viper.BindEnv(DBPoolTimeoutViperKey, dbPoolTimeoutEnv)
	viper.SetDefault(DBTLSCertViperKey, dbTLSCertDefault)
	_ = viper.BindEnv(DBTLSCertViperKey, dbTLSCertEnv)
	viper.SetDefault(DBTLSKeyViperKey, dbTLSKeyDefault)
	_ = viper.BindEnv(DBTLSKeyViperKey, dbTLSKeyEnv)
	viper.SetDefault(DBTLSCAViperKey, dbTLSCADefault)
	_ = viper.BindEnv(DBTLSCAViperKey, dbTLSCAEnv)
	viper.SetDefault(DBTLSSSLModeViperKey, dbTLSSSLModeDefault)
	_ = viper.BindEnv(DBTLSSSLModeViperKey, dbTLSSSLModeEnv)
	viper.SetDefault(DBKeepAliveKey, dbKeepAliveDefault)
	_ = viper.BindEnv(DBKeepAliveKey, dbKeepAliveEnv)
}

// PGFlags register flags for Postgres database
func PGFlags(f *pflag.FlagSet) {
	DBUser(f)
	DBPassword(f)
	DBDatabase(f)
	DBHost(f)
	DBPort(f)
	DBPoolSize(f)
	DBPoolTimeout(f)
	DBKeepAliveInterval(f)
	DBTLSSSLMode(f)
	DBTLSCert(f)
	DBTLSKey(f)
	DBTLSCA(f)
}

const (
	dbUserFlag     = "db-user"
	DBUserViperKey = "db.user"
	dbUserDefault  = "postgres"
	dbUserEnv      = "DB_USER"
)

// DBUser register flag for db user
func DBUser(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Database User.
Environment variable: %q`, dbUserEnv)
	f.String(dbUserFlag, dbUserDefault, desc)
	_ = viper.BindPFlag(DBUserViperKey, f.Lookup(dbUserFlag))
}

const (
	dbPasswordFlag     = "db-password"
	DBPasswordViperKey = "db.password"
	dbPasswordDefault  = "postgres"
	dbPasswordEnv      = "DB_PASSWORD"
)

// DBPassword register flag for db password
func DBPassword(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Database User password
Environment variable: %q`, dbPasswordEnv)
	f.String(dbPasswordFlag, dbPasswordDefault, desc)
	_ = viper.BindPFlag(DBPasswordViperKey, f.Lookup(dbPasswordFlag))
}

const (
	dbDatabaseFlag     = "db-database"
	DBDatabaseViperKey = "db.database"
	dbDatabaseDefault  = "postgres"
	dbDatabaseEnv      = "DB_DATABASE"
)

// DBDatabase register flag for db database name
func DBDatabase(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Target Database name
Environment variable: %q`, dbDatabaseEnv)
	f.String(dbDatabaseFlag, dbDatabaseDefault, desc)
	_ = viper.BindPFlag(DBDatabaseViperKey, f.Lookup(dbDatabaseFlag))
}

const (
	dbHostFlag     = "db-host"
	DBHostViperKey = "db.host"
	dbHostDefault  = "127.0.0.1"
	dbHostEnv      = "DB_HOST"
)

// DBHost register flag for database host
func DBHost(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Database host
Environment variable: %q`, dbHostEnv)
	f.String(dbHostFlag, dbHostDefault, desc)
	_ = viper.BindPFlag(DBHostViperKey, f.Lookup(dbHostFlag))
}

const (
	dbPortFlag     = "db-port"
	DBPortViperKey = "db.port"
	dbPortDefault  = 5432
	dbPortEnv      = "DB_PORT"
)

// DBPort register flag for database port
func DBPort(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Database port
Environment variable: %q`, dbPortEnv)
	f.Int(dbPortFlag, dbPortDefault, desc)
	_ = viper.BindPFlag(DBPortViperKey, f.Lookup(dbPortFlag))
}

const (
	dbPoolSizeFlag     = "db-poolsize"
	DBPoolSizeViperKey = "db.poolsize"
	dbPoolSizeDefault  = 0
	dbPoolSizeEnv      = "DB_POOLSIZE"
)

// DBPoolSize register flag for database pool size
func DBPoolSize(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Maximum number of connections on database
Environment variable: %q`, dbPoolSizeEnv)
	f.Int(dbPoolSizeFlag, dbPoolSizeDefault, desc)
	_ = viper.BindPFlag(DBPoolSizeViperKey, f.Lookup(dbPoolSizeFlag))
}

const (
	dbPoolTimeoutFlag     = "db-pool-timeout"
	DBPoolTimeoutViperKey = "db.pool-timeout"
	dbPoolTimeoutDefault  = time.Second * 30
	dbPoolTimeoutEnv      = "DB_POOL_TIMEOUT"
)

func DBPoolTimeout(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Time for which client waits for free connection if all connections are busy
Environment variable: %q`, dbPoolTimeoutEnv)
	f.Duration(dbPoolTimeoutFlag, dbPoolTimeoutDefault, desc)
	_ = viper.BindPFlag(DBPoolTimeoutViperKey, f.Lookup(dbPoolTimeoutFlag))
}

const (
	dbKeepAliveFlag    = "db-keepalive"
	DBKeepAliveKey     = "db.keepalive"
	dbKeepAliveDefault = time.Minute
	dbKeepAliveEnv     = "DB_KEEPALIVE"
)

// DBPoolSize register flag for database pool size
func DBKeepAliveInterval(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Controls the number of seconds after which a TCP keepalive message should be sent 
Environment variable: %q`, dbKeepAliveEnv)
	f.Duration(dbKeepAliveFlag, dbKeepAliveDefault, desc)
	_ = viper.BindPFlag(DBKeepAliveKey, f.Lookup(dbKeepAliveFlag))
}

const (
	requireSSLMode    = "require"
	disableSSLMode    = "disable"
	verifyCASSLMode   = "verify-ca"
	verifyFullSSLMode = "verify-full"
)

var availableSSLModes = []string{
	requireSSLMode,
	disableSSLMode,
	verifyCASSLMode,
	verifyFullSSLMode,
}

const (
	dbTLSSSLModeFlag     = "db-sslmode"
	DBTLSSSLModeViperKey = "db.tls.sslmode"
	dbTLSSSLModeDefault  = disableSSLMode
	dbTLSSSLModeEnv      = "DB_TLS_SSLMODE"
)

// DBTLSSSLMode register flag for TLS SSL mode used to connect to the database
func DBTLSSSLMode(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`TLS/SSL mode to connect to database (one of %q)
Environment variable: %q`, dbTLSSSLModeEnv, availableSSLModes)
	f.String(dbTLSSSLModeFlag, dbTLSSSLModeDefault, desc)
	_ = viper.BindPFlag(DBTLSSSLModeViperKey, f.Lookup(dbTLSSSLModeFlag))
}

const (
	dbTLSCertFlag     = "db-tls-cert"
	DBTLSCertViperKey = "db.tls.cert"
	dbTLSCertDefault  = ""
	dbTLSCertEnv      = "DB_TLS_CERT"
)

// DBTLSCert register flag for TLS certificate used to connect to the database
func DBTLSCert(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`TLS Certificate to connect to database
Environment variable: %q`, dbTLSCertEnv)
	f.String(dbTLSCertFlag, dbTLSCertDefault, desc)
	_ = viper.BindPFlag(DBTLSCertViperKey, f.Lookup(dbTLSCertFlag))
}

const (
	dbTLSKeyFlag     = "db-tls-key"
	DBTLSKeyViperKey = "db.tls.key"
	dbTLSKeyDefault  = ""
	dbTLSKeyEnv      = "DB_TLS_KEY"
)

// DBTLSKey register flag for database TLS private key
func DBTLSKey(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`TLS Private Key to connect to database
Environment variable: %q`, dbTLSKeyEnv)
	f.String(dbTLSKeyFlag, dbTLSKeyDefault, desc)
	_ = viper.BindPFlag(DBTLSKeyViperKey, f.Lookup(dbTLSKeyFlag))
}

const (
	dbTLSCAFlag     = "db-tls-ca"
	DBTLSCAViperKey = "db.tls.ca"
	dbTLSCADefault  = ""
	dbTLSCAEnv      = "DB_TLS_CA"
)

// DBTLSCert register flag for database pool size
func DBTLSCA(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Trusted Certificate Authority
Environment variable: %q`, dbTLSCAEnv)
	f.String(dbTLSCAFlag, dbTLSCADefault, desc)
	_ = viper.BindPFlag(DBTLSCAViperKey, f.Lookup(dbTLSCAFlag))
}

type Config struct {
	Host              string
	Port              string
	User              string
	Password          string
	Database          string
	PoolSize          int
	PoolTimeout       time.Duration
	DialTimeout       time.Duration
	KeepAliveInterval time.Duration
	TLS               *tls.Option
	ApplicationName   string
	SSLMode           string
}

func (cfg *Config) PGOptions() (*pg.Options, error) {
	opt := &pg.Options{
		Addr:            fmt.Sprintf("%v:%v", cfg.Host, cfg.Port),
		User:            cfg.User,
		Password:        cfg.Password,
		Database:        cfg.Database,
		PoolSize:        cfg.PoolSize,
		ApplicationName: cfg.ApplicationName,
		PoolTimeout:     cfg.PoolTimeout,
	}

	dialer, err := NewTLSDialer(cfg)
	if err != nil {
		return nil, err
	}

	if dialer != nil {
		opt.Dialer = dialer.DialContext
	} else {
		opt.Dialer = Dialer(cfg).DialContext
	}

	return opt, nil
}

func NewConfig(vipr *viper.Viper) *Config {
	cfg := &Config{
		Host:              vipr.GetString(DBHostViperKey),
		Port:              vipr.GetString(DBPortViperKey),
		User:              vipr.GetString(DBUserViperKey),
		Password:          vipr.GetString(DBPasswordViperKey),
		Database:          vipr.GetString(DBDatabaseViperKey),
		PoolSize:          vipr.GetInt(DBPoolSizeViperKey),
		PoolTimeout:       vipr.GetDuration(DBPoolTimeoutViperKey),
		SSLMode:           vipr.GetString(DBTLSSSLModeViperKey),
		KeepAliveInterval: vipr.GetDuration(DBKeepAliveKey),
		DialTimeout:       time.Second * 10, // Using double of default PG value
		TLS:               &tls.Option{},
	}

	if vipr.GetString(DBTLSCertViperKey) != "" {
		cfg.TLS = &tls.Option{
			Certificates: []*certificate.KeyPair{
				{
					Cert: []byte(vipr.GetString(DBTLSCertViperKey)),
					Key:  []byte(vipr.GetString(DBTLSKeyViperKey)),
				},
			},
		}
	}

	if vipr.GetString(DBTLSCAViperKey) != "" {
		cfg.TLS.CAs = [][]byte{
			[]byte(vipr.GetString(DBTLSCAViperKey)),
		}
	}

	return cfg
}

func Copy(opts *pg.Options) *pg.Options {
	if opts == nil {
		return nil
	}
	o := (*opts)
	return &o
}
