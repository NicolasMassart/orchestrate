package config

import "github.com/ethereum/go-ethereum/common/hexutil"

type TestData struct {
	Nodes Nodes `json:"nodes"`
	OIDC  *OIDC `json:"oidc"`
}

type Nodes struct {
	Besu   []ChainData `json:"besu,omitempty"`
	Quorum []ChainData `json:"quorum,omitempty"`
	Geth   []ChainData `json:"geth,omitempty"`
}

type OIDC struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	TokenURL     string `json:"tokenURL"`
}

type ChainData struct {
	URLs                []string        `json:"URLs,omitempty"`
	PrivateAddress      []string        `json:"privateAddress,omitempty"`
	FundedPublicKeys    []string        `json:"fundedPublicKeys,omitempty"`
	FundedPrivateKeys   []hexutil.Bytes `json:"fundedPrivateKeys,omitempty"`
	PrivateTxManagerURL string          `json:"privateTxManagerURL,omitempty"`
}
