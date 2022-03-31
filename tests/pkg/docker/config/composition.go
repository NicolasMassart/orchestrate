package config

import (
	"github.com/consensys/orchestrate/pkg/utils"
	"github.com/consensys/orchestrate/tests/pkg/docker/container/ganache"
	"github.com/consensys/orchestrate/tests/pkg/docker/container/hashicorp"
	"github.com/consensys/orchestrate/tests/pkg/docker/container/kafka"
	"github.com/consensys/orchestrate/tests/pkg/docker/container/postgres"
	quorumkeymanager "github.com/consensys/orchestrate/tests/pkg/docker/container/quorum-key-manager"
	"github.com/consensys/orchestrate/tests/pkg/docker/container/zookeeper"
)

type Composition struct {
	Containers map[string]*Container
}

type Container struct {
	Postgres                *postgres.Config
	Zookeeper               *zookeeper.Config
	Kafka                   *kafka.Config
	HashicorpVault          *hashicorp.Config
	Ganache                 *ganache.Config
	QuorumKeyManager        *quorumkeymanager.Config
	QuorumKeyManagerMigrate *quorumkeymanager.ConfigMigrate
}

func (c *Container) Field() (interface{}, error) {
	return utils.ExtractField(c)
}
