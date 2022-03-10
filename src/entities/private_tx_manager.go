package entities

import "time"

type PrivateTxType string
type PrivateTxManagerType string

var (
	PrivateTxTypeRestricted PrivateTxType = "restricted"

	GoQuorumChainType PrivateTxManagerType = "GoQuorum"
	EEAChainType      PrivateTxManagerType = "EEA"
)

const (
	// Minimum gas is calculated by the size of the enclaveKey
	TesseraGasLimit = 60000
)

func (ptx *PrivateTxManagerType) String() string {
	return string(*ptx)
}

func (ptt *PrivateTxType) String() string {
	return string(*ptt)
}

type PrivateTxManager struct {
	UUID      string               // UUID of the private transaction manager.
	ChainUUID string               // UUID of the registered chain.
	URL       string               // Transaction manager endpoint.
	Type      PrivateTxManagerType // Currently supports `Tessera` and `EEA`.
	CreatedAt time.Time            // Date and time that the private transaction manager was registered with the chain.
}
