package ganache

import (
	"context"
	"fmt"
	"time"

	ethclient "github.com/consensys/orchestrate/src/infra/ethclient/rpc"
	log "github.com/sirupsen/logrus"

	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

const defaultGanacheImage = "trufflesuite/ganache-cli:v6.12.2"
const defaultHostPort = "8545"
const defaultHost = "localhost"
const defaultChainID = "666"

var accounts = map[string]string{
	"0x56202652fdffd802b7252a456dbd8f3ecc0352bbde76c23b40afe8aebd714e2e": "1000000000000000000000", // 0x7E654d251Da770A068413677967F6d3Ea2FeA9E4
	"0x5FBB50BFF6DFAD35C4A374C9237BA2F7EAED9C6868E0108CB259B62D68029B1A": "1000000000000000000000", // 0xdbb881a51CD4023E4400CEF3ef73046743f08da3
	"0x1476c66de79a57e8ab4cadceccbe858c99e5edf3bffea5404b15322b5421e18c": "1000000000000000000000", // 0x93f7274c9059e601be4512F656B57b830e019E41
}

type Ganache struct{}

type Config struct {
	Image   string
	Port    string
	Host    string
	ChainID string
}

func NewDefault() *Config {
	return &Config{
		Image:   defaultGanacheImage,
		Port:    defaultHostPort,
		Host:    defaultHost,
		ChainID: defaultChainID,
	}
}

func (cfg *Config) SetHostPort(port string) *Config {
	cfg.Port = port
	return cfg
}

func (cfg *Config) SetHost(host string) *Config {
	cfg.Host = host
	return cfg
}

func (cfg *Config) SetChainID(chainID string) *Config {
	cfg.ChainID = chainID
	return cfg
}

func (*Ganache) GenerateContainerConfig(_ context.Context, configuration interface{}) (*dockercontainer.Config, *dockercontainer.HostConfig, *network.NetworkingConfig, error) {
	cfg, ok := configuration.(*Config)
	if !ok {
		return nil, nil, nil, fmt.Errorf("invalid configuration type (expected %T but got %T)", cfg, configuration)
	}

	cmd := []string{"ganache-cli", "--blockTime", "1", "--chainId", cfg.ChainID}
	for privKey, balance := range accounts {
		cmd = append(cmd, fmt.Sprintf(`--account="%s,%s"`, privKey, balance))
	}

	containerCfg := &dockercontainer.Config{
		Image: cfg.Image,
		ExposedPorts: nat.PortSet{
			"8545/tcp": struct{}{},
		},
		// Cmd: []string{"ganache-cli", "--mnemonic", "surge arm pulse bus piano poet thrive erase angry dwarf cargo vanish", "--blockTime", "1"},
		Cmd: cmd,
	}

	hostConfig := &dockercontainer.HostConfig{}
	if cfg.Port != "" {
		hostConfig.PortBindings = nat.PortMap{
			"8545/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: cfg.Port}},
		}
	}

	return containerCfg, hostConfig, nil, nil
}

func (*Ganache) WaitForService(ctx context.Context, configuration interface{}, timeout time.Duration) error {
	cfg, ok := configuration.(*Config)
	if !ok {
		return fmt.Errorf("invalid configuration type (expected %T but got %T)", cfg, configuration)
	}

	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	retryT := time.NewTicker(2 * time.Second)
	defer retryT.Stop()

	ethclient.Init(ctx)
	client := ethclient.GlobalClient()

	var cerr error
waitForServiceLoop:
	for {
		select {
		case <-rctx.Done():
			cerr = rctx.Err()
			break waitForServiceLoop
		case <-retryT.C:
			_, err := client.Network(ctx, fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port))
			if err != nil {
				log.WithContext(rctx).WithError(err).Warnf("waiting for Ganache service to start")
			} else {
				log.WithContext(rctx).Info("ganache container service is ready")
				break waitForServiceLoop
			}
		}
	}

	return cerr
}
