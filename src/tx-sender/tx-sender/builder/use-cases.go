package builder

import (
	"github.com/consensys/orchestrate/pkg/sdk"
	"github.com/consensys/orchestrate/src/infra/ethclient"
	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/nonce"
	usecases "github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases"
	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases/crafter"
	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases/sender"
	"github.com/consensys/orchestrate/src/tx-sender/tx-sender/use-cases/signer"
	keymanager "github.com/consensys/quorum-key-manager/pkg/client"
)

type useCases struct {
	sendETHTx             usecases.SendETHTxUseCase
	sendETHRawTx          usecases.SendETHRawTxUseCase
	sendEEAPrivateTx      usecases.SendEEAPrivateTxUseCase
	sendGoQuorumPrivateTx usecases.SendGoQuorumPrivateTxUseCase
	sendGoQuorumMarkingTx usecases.SendGoQuorumMarkingTxUseCase
}

func NewUseCases(messengerAPI sdk.MessengerAPI,
	keyManagerClient keymanager.KeyManagerClient,
	ec ethclient.MultiClient,
	nonceManager nonce.Manager,
	chainRegistryURL string,
) usecases.UseCases {
	signETHTransactionUC := signer.NewSignETHTransactionUseCase(keyManagerClient)
	signEEATransactionUC := signer.NewSignEEAPrivateTransactionUseCase(keyManagerClient)
	signQuorumTransactionUC := signer.NewSignGoQuorumPrivateTransactionUseCase(keyManagerClient)

	crafterUC := crafter.NewCraftTransactionUseCase(ec, chainRegistryURL, nonceManager)

	return &useCases{
		sendETHTx:             sender.NewSendEthTxUseCase(signETHTransactionUC, crafterUC, ec, messengerAPI, chainRegistryURL, nonceManager),
		sendETHRawTx:          sender.NewSendETHRawTxUseCase(ec, messengerAPI, chainRegistryURL),
		sendEEAPrivateTx:      sender.NewSendEEAPrivateTxUseCase(signEEATransactionUC, crafterUC, ec, messengerAPI, chainRegistryURL, nonceManager),
		sendGoQuorumPrivateTx: sender.NewSendGoQuorumPrivateTxUseCase(ec, crafterUC, messengerAPI, chainRegistryURL),
		sendGoQuorumMarkingTx: sender.NewSendGoQuorumMarkingTxUseCase(signQuorumTransactionUC, crafterUC, ec, messengerAPI, chainRegistryURL, nonceManager),
	}
}

func (u *useCases) SendETHTx() usecases.SendETHTxUseCase {
	return u.sendETHTx
}

func (u *useCases) SendETHRawTx() usecases.SendETHRawTxUseCase {
	return u.sendETHRawTx
}

func (u *useCases) SendEEAPrivateTx() usecases.SendEEAPrivateTxUseCase {
	return u.sendEEAPrivateTx
}

func (u *useCases) SendGoQuorumPrivateTx() usecases.SendGoQuorumPrivateTxUseCase {
	return u.sendGoQuorumPrivateTx
}

func (u *useCases) SendGoQuorumMarkingTx() usecases.SendGoQuorumMarkingTxUseCase {
	return u.sendGoQuorumMarkingTx
}
