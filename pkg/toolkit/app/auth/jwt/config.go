package jwt

import (
	"fmt"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault(CertificateViperKey, certificateDefault)
	_ = viper.BindEnv(CertificateViperKey, certificateEnv)
	viper.SetDefault(ClaimsNamespaceViperKey, claimsNamespaceDefault)
	_ = viper.BindEnv(ClaimsNamespaceViperKey, claimsNamespaceEnv)
}

func Flags(f *pflag.FlagSet) {
	certificateFlags(f)
	claimsNamespace(f)
}

// Provision trusted certificate of the authentication service (base64 encoded)
const (
	certificateFlag     = "auth-jwt-certificate"
	CertificateViperKey = "auth.jwt.certificate"
	certificateDefault  = ""
	certificateEnv      = "AUTH_JWT_CERTIFICATE"
)

// certificateFlag register flag for Authentication service certificate
func certificateFlags(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`certificate of the authentication service encoded in base64.
Environment variable: %q`, certificateEnv)
	f.String(certificateFlag, certificateDefault, desc)
	_ = viper.BindPFlag(CertificateViperKey, f.Lookup(certificateFlag))
}

// Provision tenant namespace to retrieve the tenant id in the OpenId or Access Token (JWT)
const (
	claimsNamespaceFlag     = "auth-jwt-claims-namespace"
	ClaimsNamespaceViperKey = "auth.jwt.claims.namespace"
	claimsNamespaceDefault  = "orchestrate.info"
	claimsNamespaceEnv      = "AUTH_JWT_CLAIMS_NAMESPACE"
)

// ClaimsNamespace register flag for tenant namespace
func claimsNamespace(f *pflag.FlagSet) {
	desc := fmt.Sprintf(`Tenant Namespace to retrieve the tenant id in the Access Token (JWT).
Environment variable: %q`, claimsNamespaceEnv)
	f.String(claimsNamespaceFlag, claimsNamespaceDefault, desc)
	_ = viper.BindPFlag(ClaimsNamespaceViperKey, f.Lookup(claimsNamespaceFlag))
}

type Config struct {
	Certificate          []byte
	ClaimsNamespace      string
	SkipClaimsValidation bool
	ValidMethods         []string
}

func NewConfig(vipr *viper.Viper) *Config {
	return &Config{
		ClaimsNamespace: vipr.GetString(ClaimsNamespaceViperKey),
		Certificate:     []byte(vipr.GetString(CertificateViperKey)),
	}
}
