package steps

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/consensys/orchestrate/src/infra/push_notification/client"

	"github.com/consensys/orchestrate/src/api/service/formatters"
	"github.com/consensys/orchestrate/src/entities"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/go-kit/kit/transport/http/jsonrpc"

	clientutils "github.com/consensys/orchestrate/pkg/toolkit/app/http/client-utils"

	"github.com/cenkalti/backoff/v4"
	utils4 "github.com/consensys/orchestrate/pkg/utils"
	api "github.com/consensys/orchestrate/src/api/service/types"

	"github.com/consensys/orchestrate/pkg/errors"
	"github.com/consensys/orchestrate/pkg/ethereum/account"
	utils2 "github.com/consensys/orchestrate/src/infra/ethclient/utils"
	"github.com/consensys/orchestrate/tests/service/e2e/cucumber/alias"
	"github.com/consensys/orchestrate/tests/service/e2e/utils"
	"github.com/cucumber/godog"
	gherkin "github.com/cucumber/messages-go/v10"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gofrs/uuid"
	log "github.com/sirupsen/logrus"
)

const aliasHeaderValue = "alias"

var aliasRegex = regexp.MustCompile("{{([^}]*)}}")
var AddressPtrType = reflect.TypeOf(new(common.Address))

func (sc *ScenarioContext) appendTxResponse(txResponse *client.TxResponse) {
	sc.mux.Lock()
	defer sc.mux.Unlock()
	sc.txResponses = append(sc.txResponses, txResponse)
}

func (sc *ScenarioContext) txResponseShouldBeInTxDecodedTopic(msgID string) error {
	ctx := context.Background()

	cerr := make(chan error, 1)

	go func() {
		txResponse, err := sc.consumerTracker.WaitForTxResponseInTopic(ctx, msgID, "", sc.waitForEnvelope)
		sc.appendTxResponse(txResponse)
		cerr <- err
	}()

	go func() {
		txResponse, err := sc.consumerTracker.WaitForTxResponseInTopic(ctx, msgID, "", sc.waitForEnvelope)
		if err == nil {
			cerr <- fmt.Errorf("envelope found in tx-recover %s", txResponse.Error)
		}
	}()

	return <-cerr
}
func (sc *ScenarioContext) txResponseShouldBeInTxRecoverTopic(msgID string) error {
	ctx := context.Background()
	_, err := sc.consumerTracker.WaitForTxResponseInTopic(ctx, msgID, "", sc.waitForEnvelope)
	return err
}

func (sc *ScenarioContext) envelopesShouldHaveTheFollowingValues(table *gherkin.PickleStepArgument_PickleTable) error {
	header := table.Rows[0]
	rows := table.Rows[1:]
	if len(rows) != len(sc.txResponses) {
		return fmt.Errorf("expected as much rows as envelopes tracked")
	}

	for idx, row := range rows {
		val := reflect.ValueOf(sc.txResponses[idx]).Elem()
		txResponse, err := json.Marshal(sc.txResponses[idx])
		log.WithError(err).Debugf("Marshaled envelope: %s", utils4.ShortString(fmt.Sprint(txResponse), 30))
		for c, col := range row.Cells {
			fieldName := header.Cells[c].Value
			field, err := utils.GetField(fieldName, val)
			if err != nil {
				return err
			}

			if err := utils.CmpField(field, col.Value); err != nil {
				return fmt.Errorf("(%d/%d) %v %v", idx+1, len(rows), fieldName, err)
			}
		}
	}

	return nil
}

func (sc *ScenarioContext) iRegisterTheFollowingEnvelopeFields(table *gherkin.PickleStepArgument_PickleTable) (err error) {
	txResponses := make(map[string]*client.TxResponse)
	for _, txResponse := range sc.txResponses {
		txResponses[txResponse.Job.ScheduleUUID] = txResponse
		txResponses[txResponse.Job.Labels["id"]] = txResponse
	}

	header := table.Rows[0]
	rows := table.Rows[1:]
	for idx, h := range []string{"id", "alias", "path"} {
		if header.Cells[idx].Value != h {
			return fmt.Errorf("invalid first column table header: expected '%s', found '%s'", h, header.Cells[idx].Value)
		}
	}

	for i, row := range rows {
		msgID := row.Cells[0].Value
		if aliasRegex.MatchString(msgID) {
			msgID = aliasRegex.FindString(msgID)
		}

		evlp, ok := txResponses[msgID]
		if !ok {
			return fmt.Errorf("envelope %s is not found: %q", msgID, row)
		}

		a := row.Cells[1].Value
		bodyPath := table.Rows[i+1].Cells[2].Value
		val, err := utils.GetField(bodyPath, reflect.ValueOf(evlp))
		if err != nil {
			return err
		}

		switch val.Type() {
		case AddressPtrType:
			sc.aliases.Set(val.Interface().(*common.Address).Hex(), sc.Pickle.Id, a)
		default:
			sc.aliases.Set(val, sc.Pickle.Id, a)
		}
	}

	return nil
}

func (sc *ScenarioContext) tearDown(s *gherkin.Pickle, err error) {
	var wg sync.WaitGroup
	wg.Add(len(sc.TearDownFunc))

	for _, f := range sc.TearDownFunc {
		f := f
		go func() {
			defer wg.Done()
			f()
		}()
	}
	wg.Wait()
}

func (sc *ScenarioContext) iHaveTheFollowingTenant(table *gherkin.PickleStepArgument_PickleTable) error {
	headers := table.Rows[0]
	for _, row := range table.Rows[1:] {
		tenantMap := make(map[string]interface{})
		var a string
		var tenantID string
		var username string

		for i, cell := range row.Cells {
			switch v := headers.Cells[i].Value; {
			case v == aliasHeaderValue:
				a = cell.Value
			case v == "tenantID":
				tenantID = cell.Value
			case v == "username":
				username = cell.Value
			default:
				tenantMap[v] = cell.Value
			}
		}
		if a == "" {
			return errors.DataError("need an alias")
		}
		if tenantID == "" {
			tenantID = uuid.Must(uuid.NewV4()).String()
		}

		tenantMap["tenantID"] = tenantID
		if username != "" {
			tenantMap["username"] = username
		}
		sc.aliases.Set(tenantMap, sc.Pickle.Id, a)
	}

	return nil
}

func (sc *ScenarioContext) iHaveTheFollowingAccount(table *gherkin.PickleStepArgument_PickleTable) error {
	headers := table.Rows[0]
	for _, row := range table.Rows[1:] {
		accountMap := make(map[string]interface{})
		var aliass string

		for i, cell := range row.Cells {
			switch v := headers.Cells[i].Value; {
			case v == aliasHeaderValue:
				aliass = cell.Value
			default:
				accountMap[v] = cell.Value
			}
		}

		if aliass == "" {
			return errors.DataError("need an alias")
		}

		w, _ := account.NewAccount()
		accountMap["address"] = w.Address.String()
		accountMap["private_key"] = hexutil.Encode(w.Priv())
		sc.aliases.Set(accountMap, sc.Pickle.Id, aliass)
	}

	return nil
}

func (sc *ScenarioContext) iHaveCreatedTheFollowingAccounts(table *gherkin.PickleStepArgument_PickleTable) error {
	tenantCol := utils.ExtractColumns(table, []string{"Tenant"})
	apiKeyCol := utils.ExtractColumns(table, []string{"API-KEY"})
	accIDCol := utils.ExtractColumns(table, []string{"ID"})
	accChainCol := utils.ExtractColumns(table, []string{"ChainName"})
	aliasCol := utils.ExtractColumns(table, []string{aliasHeaderValue})
	if aliasCol == nil {
		return errors.DataError("alias column is mandatory")
	}

	for idx := range apiKeyCol.Rows[1:] {
		req := &api.CreateAccountRequest{}

		if accIDCol != nil {
			req.Alias = accIDCol.Rows[idx+1].Cells[0].Value
		}

		if accChainCol != nil {
			req.Chain = accChainCol.Rows[idx+1].Cells[0].Value
		}

		tenant := ""
		if tenantCol != nil {
			tenant = tenantCol.Rows[idx+1].Cells[0].Value
		}
		headers := utils.GetHeaders(apiKeyCol.Rows[idx+1].Cells[0].Value, tenant, "")
		accRes, err := sc.client.CreateAccount(context.WithValue(context.Background(), clientutils.RequestHeaderKey, headers), req)

		if err != nil {
			return err
		}

		sc.aliases.Set(accRes.Address, sc.Pickle.Id, aliasCol.Rows[idx+1].Cells[0].Value)
	}

	return nil
}

func (sc *ScenarioContext) iRegisterTheFollowingChains(table *gherkin.PickleStepArgument_PickleTable) error {
	aliasCol := utils.ExtractColumns(table, []string{"alias"})
	tenantCol := utils.ExtractColumns(table, []string{"Tenant"})
	apiKeyCol := utils.ExtractColumns(table, []string{"API-KEY"})

	interfaceSlices, err := utils.ParseTable(api.RegisterChainRequest{}, table)
	if err != nil {
		return err
	}

	onTearDown := func(uuid string, headers map[string]string) func() {
		return func() {
			time.Sleep(time.Second * 3) // Wait few seconds to allow ongoing work to complete
			_ = sc.client.DeleteChain(context.WithValue(context.Background(), clientutils.RequestHeaderKey, headers), uuid)
		}
	}

	ctx := context.Background()
	for i, chain := range interfaceSlices {
		apiKey := apiKeyCol.Rows[i+1].Cells[0].Value

		tenant := ""
		if tenantCol != nil {
			tenant = tenantCol.Rows[i+1].Cells[0].Value
		}

		headers := utils.GetHeaders(apiKey, tenant, "")
		res, err := sc.client.RegisterChain(context.WithValue(context.Background(), clientutils.RequestHeaderKey, headers), chain.(*api.RegisterChainRequest))
		if err != nil {
			return err
		}
		sc.TearDownFunc = append(sc.TearDownFunc, onTearDown(res.UUID, headers))

		apiURL, _ := sc.aliases.Get(alias.GlobalAka, "api")
		proxyURL := utils4.GetProxyURL(apiURL.(string), res.UUID)
		err = backoff.RetryNotify(
			func() error {
				_, err2 := sc.ec.Network(ctx, proxyURL)
				return err2
			},
			backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 5),
			func(err error, duration time.Duration) {
				log.WithFields(log.Fields{
					"chain_uuid": res.UUID,
				}).WithError(err).Debug("scenario: chain proxy is still not ready")
			},
		)

		if err != nil {
			return err
		}

		// set aliases
		sc.aliases.Set(res, sc.Pickle.Id, aliasCol.Rows[i+1].Cells[0].Value)
	}

	return nil
}

func (sc *ScenarioContext) iRegisterTheFollowingFaucets(table *gherkin.PickleStepArgument_PickleTable) error {
	tenantCol := utils.ExtractColumns(table, []string{"Tenant"})
	apiKeyCol := utils.ExtractColumns(table, []string{"API-KEY"})

	interfaceSlices, err := utils.ParseTable(api.RegisterFaucetRequest{}, table)
	if err != nil {
		return err
	}

	f := func(ctx context.Context, uuid string) func() {
		return func() { _ = sc.client.DeleteFaucet(ctx, uuid) }
	}

	for i, faucet := range interfaceSlices {
		apiKey := apiKeyCol.Rows[i+1].Cells[0].Value

		tenant := ""
		if tenantCol != nil {
			tenant = tenantCol.Rows[i+1].Cells[0].Value
		}

		ctx := context.WithValue(context.Background(), clientutils.RequestHeaderKey, utils.GetHeaders(apiKey, tenant, ""))
		res, err := sc.client.RegisterFaucet(ctx, faucet.(*api.RegisterFaucetRequest))
		if err != nil {
			return err
		}
		sc.TearDownFunc = append(sc.TearDownFunc, f(ctx, res.UUID))
	}

	return nil
}

func (sc *ScenarioContext) replace(s string) (string, error) {
	for _, matchedAlias := range aliasRegex.FindAllStringSubmatch(s, -1) {
		aka := []string{matchedAlias[1]}
		if strings.HasPrefix(matchedAlias[1], "random.") {
			random := strings.Split(matchedAlias[1], ".")
			switch random[1] {
			case "uuid":
				s = strings.Replace(s, matchedAlias[0], uuid.Must(uuid.NewV4()).String(), 1)
			case "account":
				w, _ := account.NewAccount()
				s = strings.Replace(s, matchedAlias[0], w.Address.Hex(), 1)
			case "private_key":
				w, _ := account.NewAccount()
				s = strings.Replace(s, matchedAlias[0], hexutil.Encode(w.Priv())[2:], 1)
			case "int":
				s = strings.Replace(s, matchedAlias[0], fmt.Sprintf("%d", rand.Int()), 1)
			}
			continue
		}

		if !strings.HasPrefix(matchedAlias[1], "global.") && !strings.HasPrefix(matchedAlias[1], "chain.") {
			aka = append([]string{sc.Pickle.Id}, aka...)
		}
		v, ok := sc.aliases.Get(aka...)
		if !ok {
			return "", fmt.Errorf("could not replace alias '%s'", matchedAlias[1])
		}

		val := reflect.ValueOf(v)

		var str string
		switch val.Kind() {
		case reflect.Array, reflect.Slice:
			switch reflect.TypeOf(v).String() {
			case reflect.TypeOf(new(hexutil.Bytes)).String():
				str = v.(*hexutil.Bytes).String()
			case reflect.TypeOf(hexutil.Bytes{}).String():
				str = v.(hexutil.Bytes).String()
			default:
				strb, _ := json.Marshal(v)
				str = string(strb)
			}
		default:
			switch reflect.TypeOf(v).String() {
			case reflect.TypeOf(new(hexutil.Big)).String():
				str = v.(*hexutil.Big).String()
			default:
				str = fmt.Sprintf("%v", v)
			}
		}

		s = strings.Replace(s, matchedAlias[0], str, 1)
	}

	return s, nil
}

func (sc *ScenarioContext) replaceAliases(table *gherkin.PickleStepArgument_PickleTable) error {
	for _, row := range table.Rows {
		for _, r := range row.Cells {
			s, err := sc.replace(r.Value)
			if err != nil {
				return err
			}
			r.Value = s
		}
	}
	return nil
}

func (sc *ScenarioContext) iRegisterTheFollowingAliasAs(table *gherkin.PickleStepArgument_PickleTable) error {
	aliasTable := utils.ExtractColumns(table, []string{aliasHeaderValue})
	if aliasTable == nil {
		return errors.DataError("alias column is mandatory")
	}

	for i, row := range aliasTable.Rows[1:] {
		a := row.Cells[0].Value
		value := table.Rows[i+1].Cells[0].Value
		ok := sc.aliases.Set(value, sc.Pickle.Id, a)
		if !ok {
			return errors.DataError("could not register alias")
		}
	}
	return nil
}

func (sc *ScenarioContext) iSignTheFollowingTransactions(table *gherkin.PickleStepArgument_PickleTable) error {
	tenantCol := utils.ExtractColumns(table, []string{"Tenant"})
	apiKeyCol := utils.ExtractColumns(table, []string{"API-KEY"})
	chainUUUIDCol := utils.ExtractColumns(table, []string{"ChainUUID"})

	helpersColumns := []string{aliasHeaderValue, "privateKey"}
	helpersTable := utils.ExtractColumns(table, helpersColumns)
	if helpersTable == nil {
		return errors.DataError("One of the following columns is missing %q", helpersColumns)
	}

	txReqs, err := utils.ParseTransactions(table)
	if err != nil {
		return err
	}

	// Sign tx for each envelope
	ctx := utils2.RetryConnectionError(context.Background(), true)
	for i, txReq := range txReqs {
		apiKey := apiKeyCol.Rows[i+1].Cells[0].Value

		chainUUID := ""
		if chainUUUIDCol != nil {
			chainUUID = chainUUUIDCol.Rows[i+1].Cells[0].Value
		}

		tenant := ""
		if tenantCol != nil {
			tenant = tenantCol.Rows[i+1].Cells[0].Value
		}

		headers := utils.GetHeaders(apiKey, tenant, "")
		txRaw, txHash, err := sc.signTransaction(
			context.WithValue(ctx, clientutils.RequestHeaderKey, headers),
			formatters.FormatETHTransactionRequest(txReq),
			chainUUID,
			helpersTable.Rows[i+1].Cells[1].Value,
		)
		if err != nil {
			return err
		}

		sc.aliases.Set(struct {
			Raw    string
			TxHash string
		}{
			Raw:    string(txRaw),
			TxHash: txHash,
		}, sc.Pickle.Id, helpersTable.Rows[i+1].Cells[0].Value)
	}

	return nil
}

func (sc *ScenarioContext) iHaveTheFollowingJWTTokens(table *gherkin.PickleStepArgument_PickleTable) error {
	headers := table.Rows[0]
	for _, row := range table.Rows[1:] {
		tenantMap := make(map[string]interface{})
		var a string
		var audience string

		for i, cell := range row.Cells {
			switch v := headers.Cells[i].Value; {
			case v == aliasHeaderValue:
				a = cell.Value
			case v == "audience":
				audience = cell.Value
			default:
				tenantMap[v] = cell.Value
			}
		}
		if a == "" {
			return errors.DataError("need an alias")
		}
		if audience == "" {
			return errors.DataError("need an audience")
		}

		jwtToken, err := sc.getJWT(audience)
		if err != nil {
			return err
		}

		tenantMap["token"] = fmt.Sprintf("Bearer %s", jwtToken)
		sc.aliases.Set(tenantMap, sc.Pickle.Id, a)
	}

	return nil
}

type accessTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func (sc *ScenarioContext) getJWT(audience string) (string, error) {
	clientID, ok := sc.aliases.Get(alias.GlobalAka, "oidc.clientID")
	if !ok {
		return "", errors.DataError("Could not find oidc client ID")
	}

	clientSecret, ok := sc.aliases.Get(alias.GlobalAka, "oidc.clientSecret")
	if !ok {
		return "", errors.DataError("Could not find oidc client secret")
	}

	idpURL, ok := sc.aliases.Get(alias.GlobalAka, "oidc.tokenURL")
	if !ok {
		return "", errors.DataError("Could not find token url")
	}

	body := new(bytes.Buffer)
	_ = json.NewEncoder(body).Encode(map[string]interface{}{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"audience":      audience,
		"grant_type":    "client_credentials",
	})

	resp, err := http.DefaultClient.Post(idpURL.(string), jsonrpc.ContentType, body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	acessToken := &accessTokenResponse{}
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(acessToken)
		if err != nil {
			return "", err
		}

		return acessToken.AccessToken, nil
	}

	// Read body
	respMsg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return "", fmt.Errorf(string(respMsg))
}

// nolint
func (sc *ScenarioContext) signTransaction(ctx context.Context, tx *entities.ETHTransaction, chainUUID, privKey string) ([]byte, string, error) {
	chain, err := sc.client.GetChain(ctx, chainUUID)
	if err != nil {
		log.WithError(err).WithField("private_key", privKey).Error("failed to create account using private key")
		return nil, "", err
	}

	chainID, _ := new(big.Int).SetString(chain.ChainID, 10)
	signer := types.NewEIP155Signer(chainID)
	acc, err := crypto.HexToECDSA(privKey)
	if err != nil {
		log.WithError(err).WithField("private_key", privKey).Error("failed to create account using private key")
		return nil, "", err
	}

	if tx.GasPrice == nil {
		gasPrice, errGasPrice := sc.ec.SuggestGasPrice(ctx, chain.URLs[0])
		if errGasPrice != nil {
			log.WithError(errGasPrice).Error("failed to suggest gas price")
			return nil, "", errGasPrice
		}
		tx.GasPrice = (*hexutil.Big)(gasPrice)
	}

	transaction := tx.ToETHTransaction(chainID)
	signature, err := signTransaction(transaction, acc, signer)
	if err != nil {
		log.WithError(err).Error("failed to sign transaction")
		return nil, "", err
	}

	signedTx, err := transaction.WithSignature(signer, signature)
	if err != nil {
		log.WithError(err).Error("failed to set signature in transaction")
		return nil, "", err
	}

	signedRaw, err := rlp.EncodeToBytes(signedTx)
	if err != nil {
		log.WithError(err).Error("failed to RLP encode signed transaction")
		return nil, "", err
	}

	return signedRaw, signedTx.Hash().String(), nil
}

func signTransaction(transaction *types.Transaction, privKey *ecdsa.PrivateKey, signer types.Signer) ([]byte, error) {
	h := signer.Hash(transaction)
	decodedSignature, err := crypto.Sign(h[:], privKey)
	if err != nil {
		return nil, errors.CryptoOperationError(err.Error())
	}

	return decodedSignature, nil
}

func initEnvelopeSteps(s *godog.ScenarioContext, sc *ScenarioContext) {
	s.Step(`^I register the following chains$`, sc.preProcessTableStep(sc.iRegisterTheFollowingChains))
	s.Step(`^I register the following faucets$`, sc.preProcessTableStep(sc.iRegisterTheFollowingFaucets))
	s.Step(`^I have the following tenants$`, sc.preProcessTableStep(sc.iHaveTheFollowingTenant))
	s.Step(`^I have the following account`, sc.preProcessTableStep(sc.iHaveTheFollowingAccount))
	s.Step(`^I register the following alias$`, sc.preProcessTableStep(sc.iRegisterTheFollowingAliasAs))
	s.Step(`^I have created the following accounts$`, sc.preProcessTableStep(sc.iHaveCreatedTheFollowingAccounts))
	s.Step(`^TxResponse was found in tx-decoded topic "([^"]*)"$`, sc.txResponseShouldBeInTxDecodedTopic)
	s.Step(`^TxResponse was found in tx-recover topic "([^"]*)"$`, sc.txResponseShouldBeInTxRecoverTopic)
	s.Step(`^Envelopes should have the following fields$`, sc.preProcessTableStep(sc.envelopesShouldHaveTheFollowingValues))
	s.Step(`^I register the following envelope fields$`, sc.preProcessTableStep(sc.iRegisterTheFollowingEnvelopeFields))
	s.Step(`^I sign the following transactions$`, sc.preProcessTableStep(sc.iSignTheFollowingTransactions))
	s.Step(`^I have the following jwt tokens$`, sc.preProcessTableStep(sc.iHaveTheFollowingJWTTokens))
}
