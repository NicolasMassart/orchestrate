package steps

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/cucumber/godog"
	gherkin "github.com/cucumber/messages-go/v10"
	"github.com/gofrs/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	nonceUtils "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/handlers/nonce/utils"
	authutils "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/auth/utils"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/encoding/json"
	encoding "gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/encoding/sarama"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/errors"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/ethereum/account"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/keystore/session"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/pkg/types/tx"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/services/chain-registry/store/models"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/tests/service/cucumber/alias"
	"gitlab.com/ConsenSys/client/fr/core-stack/orchestrate.git/tests/service/cucumber/utils"
)

const aliasHeaderValue = "alias"

var aliasRegex = regexp.MustCompile("{{([^}]*)}}")

func (sc *ScenarioContext) sendEnvelope(topic string, e *tx.Envelope) error {
	// Prepare message to be sent
	msg := &sarama.ProducerMessage{
		Topic: viper.GetString(fmt.Sprintf("topic.%v", topic)),
		Key:   sarama.StringEncoder(e.PartitionKey()),
	}

	err := encoding.Marshal(e.TxEnvelopeAsRequest(), msg)
	if err != nil {
		return err
	}

	// Send message
	_, _, err = sc.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"id":            e.GetID(),
		"scenario.id":   sc.Pickle.Id,
		"scenario.name": sc.Pickle.Name,
	}).Debugf("scenario: envelope sent")

	return nil
}

func (sc *ScenarioContext) iSendEnvelopesToTopic(topic string, table *gherkin.PickleStepArgument_PickleTable) error {
	// Parse table
	if err := sc.replaceAliases(table); err != nil {
		return err
	}

	envelopes, err := utils.ParseEnvelope(table)
	if err != nil {
		return errors.DataError("could not parse tx request - got %v", err)
	}

	// Set trackers for each envelope
	sc.setTrackers(sc.newTrackers(envelopes))

	// Send envelopes
	for _, t := range sc.trackers {
		err := sc.sendEnvelope(topic, t.Current)
		if err != nil {
			return errors.InternalError("could not send tx request - got %v", err)
		}
	}

	return nil
}

func (sc *ScenarioContext) registerEnvelopeTracker(value string) error {
	envelopeID, ok := sc.aliases.Get(sc.Pickle.Id, value)
	if !ok {
		envelopeID, ok = sc.aliases.Get("global", value)
		if !ok {
			envelopeID = value
		}
	}

	evlp := tx.NewEnvelope()
	_ = evlp.SetID(envelopeID.(string)).
		SetContextLabelsValue("debug", "true").
		SetContextLabelsValue("scenario.id", sc.Pickle.Id)

	sc.setTrackers(append(sc.trackers, sc.newTracker(evlp)))

	return nil
}

func (sc *ScenarioContext) envelopeShouldBeInTopic(topic string) error {
	for i, t := range sc.trackers {
		err := t.Load(topic, viper.GetDuration(CucumberTimeoutViperKey))
		if err != nil {
			e := t.Load("tx.recover", time.Millisecond)
			if e != nil {
				return fmt.Errorf("%v: envelope n°%v neither in topic %q nor in %q", sc.Pickle.Id, i, topic, "tx.recover")
			}
			return fmt.Errorf("%v: envelope n°%v not in topic %q but found in %q - envelope.Errors %q", sc.Pickle.Id, i, topic, "tx.recover", t.Current.Error())
		}
	}
	return nil
}

func (sc *ScenarioContext) envelopesShouldHaveTheFollowingValues(table *gherkin.PickleStepArgument_PickleTable) error {
	header := table.Rows[0]
	rows := table.Rows[1:]
	if len(rows) != len(sc.trackers) {
		return fmt.Errorf("expected as much rows as envelopes tracked")
	}

	for r, row := range rows {
		val := reflect.ValueOf(sc.trackers[r].Current).Elem()
		sEvlp, err := json.Marshal(*sc.trackers[r].Current)
		log.WithError(err).Debugf("Marshaled envelope: %s", sEvlp)
		for c, col := range row.Cells {
			fieldName := header.Cells[c].Value
			field, err := utils.GetField(fieldName, val)
			if err != nil {
				return err
			}

			if err := utils.CmpField(field, col.Value); err != nil {
				return fmt.Errorf("(%d/%d) %v %v", r+1, len(rows), fieldName, err)
			}
		}
	}

	return nil
}

func (sc *ScenarioContext) iRegisterTheFollowingEnvelopeFields(table *gherkin.PickleStepArgument_PickleTable) (err error) {

	evlps := make(map[string]*tx.Envelope)
	for _, tracker := range sc.trackers {
		evlp := tracker.Current
		evlps[evlp.GetID()] = evlp
		evlps[evlp.GetContextLabelsValue("id")] = evlp
	}

	header := table.Rows[0]
	rows := table.Rows[1:]
	for idx, h := range []string{"id", "alias", "path"} {
		if header.Cells[idx].Value != h {
			return fmt.Errorf("invalid first column table header: expected '%s', found '%s'", h, header.Cells[idx].Value)
		}
	}

	for i, row := range rows {
		evlpID := row.Cells[0].Value
		if aliasRegex.MatchString(evlpID) {
			evlpID = aliasRegex.FindString(evlpID)
		}

		evlp, ok := evlps[evlpID]
		if !ok {
			return fmt.Errorf("envelope %s is not found: %q", evlpID, row)
		}

		a := row.Cells[1].Value
		bodyPath := table.Rows[i+1].Cells[2].Value
		val, err := utils.GetField(bodyPath, reflect.ValueOf(evlp))
		if err != nil {
			return err
		}

		sc.aliases.Set(val, sc.Pickle.Id, a)
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

		for i, cell := range row.Cells {
			switch v := headers.Cells[i].Value; {
			case v == aliasHeaderValue:
				a = cell.Value
			case v == "tenantID":
				tenantID = cell.Value
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
		token, err := sc.jwtGenerator.GenerateAccessTokenWithTenantID(tenantID, 24*time.Hour)
		if err != nil {
			return err
		}
		tenantMap["token"] = token
		tenantMap["tenantID"] = tenantID
		sc.aliases.Set(tenantMap, sc.Pickle.Id, a)
	}

	return nil
}

func (sc *ScenarioContext) iHaveCreatedTheFollowingAccounts(table *gherkin.PickleStepArgument_PickleTable) error {
	aliasTable := utils.ExtractColumns(table, []string{aliasHeaderValue})
	if aliasTable == nil {
		return errors.DataError("alias column is mandatory")
	}

	envelopes, err := utils.ParseEnvelope(table)
	if err != nil {
		return err
	}

	var childEnvelopes []*tx.Envelope
	for _, e := range envelopes {
		if childTxID := e.GetContextLabelsValue("faucetChildTxID"); childTxID != "" {
			childEvlp := tx.NewEnvelope().SetID(childTxID).SetContextLabelsValue("id", childTxID)
			childEnvelopes = append(childEnvelopes, childEvlp)
		}
	}

	trackers := sc.newTrackers(envelopes)
	sc.setTrackers(sc.newTrackers(childEnvelopes))

	// Send envelopes
	for _, t := range trackers {
		if sc.sendEnvelope("account.generator", t.Current) != nil {
			return errors.InternalError("could not send tx request - got %v", err)
		}
	}

	// TODO: Should we able to delete keys after the scenario complete?

	// Catch envelope after it has been decoded
	for i, t := range trackers {
		accAlias := aliasTable.Rows[i+1].Cells[0].Value
		if t.Load("account.generated", 30*time.Second) != nil {
			return errors.DataError("could not generate account for %s - got %v", accAlias, err)
		} else if t.Current.GetFrom() == nil {
			return errors.DataError("no address found")
		}
		sc.aliases.Set(t.Current.GetFrom().Hex(), sc.Pickle.Id, accAlias)
	}

	// Check that accounts has been funded
	err = sc.envelopeShouldBeInTopic("tx.decoded")
	if err != nil {
		return err
	}

	for i, t := range sc.trackers {
		if t.Current.Receipt.Status != 1 {
			a, _ := sc.aliases.Get(sc.Pickle.Id, aliasTable.Rows[i+1].Cells[0].Value)
			return errors.EthereumError("Account '%s' has been created but not funded", a)
		}
	}

	// If accounts are well funded reset tracker
	sc.trackers = nil
	return nil
}

func (sc *ScenarioContext) iRegisterTheFollowingChains(table *gherkin.PickleStepArgument_PickleTable) error {
	utilsCols := utils.ExtractColumns(table, []string{aliasHeaderValue, "Headers.Authorization"})
	if utilsCols == nil {
		return errors.DataError("One of the following columns is missing %q", utilsCols)
	}
	interfaceSlices, err := utils.ParseTable(models.Chain{}, table)
	if err != nil {
		return err
	}

	onTearDown := func(uuid, token string) func() {
		return func() {
			_ = sc.ChainRegistry.DeleteChainByUUID(
				authutils.WithAuthorization(context.Background(), token),
				uuid)
		}
	}

	for i, chain := range interfaceSlices {
		token := utilsCols.Rows[i+1].Cells[1].Value

		res, err := sc.ChainRegistry.RegisterChain(authutils.WithAuthorization(context.Background(), token), chain.(*models.Chain))
		if err != nil {
			return err
		}
		sc.TearDownFunc = append(sc.TearDownFunc, onTearDown(res.UUID, token))

		// set aliases
		sc.aliases.Set(res, sc.Pickle.Id, utilsCols.Rows[i+1].Cells[0].Value)
	}

	return nil
}

func (sc *ScenarioContext) iRegisterTheFollowingFaucets(table *gherkin.PickleStepArgument_PickleTable) error {
	tokenTable := utils.ExtractColumns(table, []string{"Headers.Authorization"})
	interfaceSlices, err := utils.ParseTable(models.Faucet{}, table)

	if err != nil {
		return err
	}

	f := func(uuid, token string) func() {
		return func() {
			_ = sc.ChainRegistry.DeleteFaucetByUUID(
				authutils.WithAuthorization(context.Background(), token),
				uuid)
		}
	}

	for i, faucet := range interfaceSlices {
		token := tokenTable.Rows[i+1].Cells[0].Value

		res, err := sc.ChainRegistry.RegisterFaucet(authutils.WithAuthorization(context.Background(), token), faucet.(*models.Faucet))
		if err != nil {
			return err
		}
		sc.TearDownFunc = append(sc.TearDownFunc, f(res.UUID, token))
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
				w := account.NewAccount()
				_ = w.Generate()
				s = strings.Replace(s, matchedAlias[0], w.Address().Hex(), 1)
			case "int":
				s = strings.Replace(s, matchedAlias[0], fmt.Sprintf("%d", rand.Int()), 1)
			}
			continue
		}

		if !strings.HasPrefix(matchedAlias[1], "global.") {
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
			strb, _ := json.Marshal(v)
			str = string(strb)
		default:
			str = fmt.Sprintf("%v", v)
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

func (sc *ScenarioContext) iTrackTheFollowingEnvelopes(table *gherkin.PickleStepArgument_PickleTable) error {
	if len(table.Rows[0].Cells) != 1 {
		return errors.DataError("invalid table")
	}

	var envelopes []*tx.Envelope
	for _, r := range table.Rows[1:] {
		if r.Cells[0].Value != "" {
			envelopes = append(envelopes, tx.NewEnvelope().SetID(r.Cells[0].Value))
		}
	}
	sc.setTrackers(sc.newTrackers(envelopes))

	return nil
}

func (sc *ScenarioContext) iSignTheFollowingTransactions(table *gherkin.PickleStepArgument_PickleTable) error {
	helpersColumns := []string{aliasHeaderValue, "privateKey", "Headers.Authorization"}
	helpersTable := utils.ExtractColumns(table, helpersColumns)
	if helpersTable == nil {
		return errors.DataError("One of the following columns is missing %q", helpersColumns)
	}

	envelopes, err := utils.ParseEnvelope(table)
	if err != nil {
		return err
	}

	// Sign tx for each envelopes
	for i, e := range envelopes {
		err := sc.craftAndSignEnvelope(
			authutils.WithAuthorization(context.Background(), helpersTable.Rows[i+1].Cells[2].Value),
			e,
			helpersTable.Rows[i+1].Cells[1].Value,
		)
		if err != nil {
			return nil
		}
		sc.aliases.Set(e, sc.Pickle.Id, helpersTable.Rows[i+1].Cells[0].Value)
	}

	return nil
}

func (sc *ScenarioContext) craftAndSignEnvelope(ctx context.Context, e *tx.Envelope, privKey string) error {
	a := account.NewAccount()
	err := a.FromPrivateKey(privKey)
	if err != nil {
		return nil
	}

	chainRegistry, ok := sc.aliases.Get(alias.GlobalAka, "chain-registry")
	if !ok {
		return errors.DataError("Could not find the chain registry endpoint")
	}
	endpoint := fmt.Sprintf("%s/%s", chainRegistry.(string), e.GetChainUUID())
	if e.GetChainID() == nil && e.GetChainUUID() != "" {
		chainID, errNetwork := sc.ec.Network(ctx, endpoint)
		if errNetwork != nil {
			return errNetwork
		}
		_ = e.SetChainID(chainID)
	}

	s := session.NewSigningSession()
	err = s.SetChain(e.GetChainID())
	if err != nil {
		return err
	}
	err = s.SetAccount(a)
	if err != nil {
		return err
	}
	_ = e.SetFrom(a.Address())

	if e.GetNonce() == nil {
		nonceKey := e.PartitionKey()
		lastAttributed, ok, errNonce := sc.nonceManager.GetLastAttributed(nonceKey)
		if errNonce != nil {
			return errNonce
		}
		var n uint64
		if !ok {
			pendingNonce, errNonce := nonceUtils.GetNonce(ctx, sc.ec, e, endpoint)
			if errNonce != nil {
				return errNonce
			}
			n = pendingNonce
		} else {
			n = lastAttributed + 1
		}
		_ = e.SetNonce(n)
		err = sc.nonceManager.SetLastAttributed(nonceKey, n)
		if err != nil {
			return err
		}
	}

	if e.GetGasPrice() == nil {
		gasPrice, errGasPrice := sc.ec.SuggestGasPrice(ctx, endpoint)
		if errGasPrice != nil {
			return errGasPrice
		}
		_ = e.SetGasPrice(gasPrice)
	}

	t, err := e.GetTransaction()
	if err != nil {
		return err
	}

	// TODO able to sign private tx
	raw, hash, err := s.ExecuteForTx(t)
	if err != nil {
		return err
	}

	_ = e.SetRaw(raw).SetTxHash(*hash)

	return nil
}

func initEnvelopeSteps(s *godog.ScenarioContext, sc *ScenarioContext) {
	s.Step(`^I register the following chains$`, sc.preProcessTableStep(sc.iRegisterTheFollowingChains))
	s.Step(`^I register the following faucets$`, sc.preProcessTableStep(sc.iRegisterTheFollowingFaucets))
	s.Step(`^I have the following tenants$`, sc.preProcessTableStep(sc.iHaveTheFollowingTenant))
	s.Step(`^I register the following alias$`, sc.preProcessTableStep(sc.iRegisterTheFollowingAliasAs))
	s.Step(`^I have created the following accounts$`, sc.preProcessTableStep(sc.iHaveCreatedTheFollowingAccounts))
	s.Step(`^I track the following envelopes$`, sc.preProcessTableStep(sc.iTrackTheFollowingEnvelopes))
	s.Step(`^I send envelopes to topic "([^"]*)"$`, sc.iSendEnvelopesToTopic)
	s.Step(`^Register new envelope tracker "([^"]*)"$`, sc.registerEnvelopeTracker)
	s.Step(`^Envelopes should be in topic "([^"]*)"$`, sc.envelopeShouldBeInTopic)
	s.Step(`^Envelopes should have the following fields$`, sc.preProcessTableStep(sc.envelopesShouldHaveTheFollowingValues))
	s.Step(`^I register the following envelope fields$`, sc.preProcessTableStep(sc.iRegisterTheFollowingEnvelopeFields))
	s.Step(`^I sign the following transactions$`, sc.preProcessTableStep(sc.iSignTheFollowingTransactions))
}
