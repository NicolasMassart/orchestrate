package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/consensys/orchestrate/pkg/utils"
	kafka "github.com/consensys/orchestrate/src/infra/kafka/sarama"

	"github.com/consensys/orchestrate/cmd/flags"
	"github.com/consensys/orchestrate/pkg/backoff"
	orchestrateclient "github.com/consensys/orchestrate/pkg/sdk/client"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/viper"
)

type Config struct {
	TestData       *TestDataCfg
	KafkaCfg       *kafka.Config
	OrchestrateCfg *orchestrateclient.Config
	Artifacts      map[string]*Artifact
}

type Artifact struct {
	ABI              interface{}   `json:"abi" validate:"required"`                                                                   // ABI of the contract.
	Bytecode         hexutil.Bytes `json:"bytecode,omitempty" example:"0x6080604052348015600f57600080f" swaggertype:"string"`         // Bytecode of the contract.
	DeployedBytecode hexutil.Bytes `json:"deployedBytecode,omitempty" example:"0x6080604052348015600f57600080f" swaggertype:"string"` // Deployed bytecode of the contract.
}

func NewE2eConfig(vipr *viper.Viper) (*Config, error) {
	cfg := &Config{}
	rawTestData := viper.GetString(testDataViperKey)
	if rawTestData == testDataDefault {
		return nil, fmt.Errorf("missing test data %s", testDataEnv)
	}
	testData := &TestDataCfg{}
	err := json.Unmarshal([]byte(rawTestData), testData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test data. %v", err.Error())
	}
	if testData.ArtifactPath == "" {
		testData.ArtifactPath = "../artifacts"
	}
	if testData.Timeout == "" {
		testData.Timeout = "30s"
	}
	if testData.Topic == "" {
		testData.Topic = "test-topic-" + utils.RandString(5)
	}

	cfg.Artifacts, err = readArtifacts(testData.ArtifactPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read artifacts. %s", err.Error())
	}
	cfg.TestData = testData
	cfg.KafkaCfg = flags.NewKafkaConfig(vipr)
	cfg.OrchestrateCfg = orchestrateclient.NewConfigFromViper(vipr, backoff.ConstantBackOffWithMaxRetries(time.Second, 3))

	return cfg, nil
}

func readArtifacts(path string) (map[string]*Artifact, error) {
	artifacts := map[string]*Artifact{}
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		content, err := ioutil.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			return nil, err
		}

		artifact := &Artifact{}
		err = json.Unmarshal(content, artifact)
		if err != nil {
			return nil, err
		}

		contractName := strings.Replace(file.Name(), filepath.Ext(file.Name()), "", 1)
		artifacts[contractName] = artifact
	}

	return artifacts, nil
}
