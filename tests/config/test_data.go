package config

import "github.com/ethereum/go-ethereum/common/hexutil"

type TestDataCfg struct {
	Nodes        NodesCfg  `json:"nodes"`
	OIDC         *OIDCCfg  `json:"oidc"`
	UserInfo     *UserInfo `json:"userInfo"`
	QKMStoreID   string    `json:"qkmStoreID"`
	ArtifactPath string    `json:"artifacts"`
	Timeout      string    `json:"timeout"`
	KafkaTopic   string    `json:"topic"`
}

type UserInfo struct {
	TenantID string `json:"tenantId"`
	Username string `json:"username"`
}

type NodesCfg struct {
	Besu     []ChainDataCfg `json:"besu,omitempty"`
	GoQuorum []ChainDataCfg `json:"quorum,omitempty"`
	Geth     []ChainDataCfg `json:"geth,omitempty"`
}

type OIDCCfg struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	TokenURL     string `json:"tokenURL"`
	Audience     string `json:"audience"`
}

type ChainDataCfg struct {
	URLs                []string        `json:"URLs,omitempty"`
	PrivateAddress      []string        `json:"privateAddress,omitempty"`
	FundedPublicKeys    []string        `json:"fundedPublicKeys,omitempty"`
	FundedPrivateKeys   []hexutil.Bytes `json:"fundedPrivateKeys,omitempty"`
	PrivateTxManagerURL string          `json:"privateTxManagerURL,omitempty"`
}
