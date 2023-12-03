package unit_tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	fr "github.com/pmatteo/friendlyrabbit"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/stretchr/testify/assert"
)

func TestCreateTopologyFromTopologyConfig(t *testing.T) {
	fileNamePath := "../data/testtopology.json"
	assert.FileExists(t, fileNamePath)

	topologyConfig, err := fr.ConvertJSONFileToTopologyConfig(fileNamePath)
	assert.NoError(t, err)

	connectionPool, err := fr.NewConnectionPool(Seasoning.PoolConfig, nil, nil)
	assert.NoError(t, err)
	defer connectionPool.Shutdown()

	topologer := fr.NewTopologer(connectionPool)

	err = topologer.BuildTopology(topologyConfig, true)
	assert.NoError(t, err)
}

func TestCreateMultipleTopologyFromTopologyConfig(t *testing.T) {
	topologer := fr.NewTopologer(RabbitService.ConnectionPool)

	topologyConfigs := make([]string, 0)
	configRoot := "./"
	err := filepath.Walk(configRoot, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, "topology") {
			topologyConfigs = append(topologyConfigs, path)
		}
		return nil
	})
	assert.NoError(t, err)

	for _, filePath := range topologyConfigs {
		topologyConfig, err := fr.ConvertJSONFileToTopologyConfig(filePath)
		if err != nil {
			assert.NoError(t, err)
		} else {
			err = topologer.BuildTopology(topologyConfig, false)
			assert.NoError(t, err)
		}
	}
}

func TestUnbindQueue(t *testing.T) {
	connectionPool, err := fr.NewConnectionPool(Seasoning.PoolConfig, nil, nil)
	assert.NoError(t, err)
	defer connectionPool.Shutdown()

	topologer := fr.NewTopologer(connectionPool)

	err = topologer.UnbindQueue("QueueAttachedToExch01", "RoutingKey1", "MyTestExchange.Child01", nil)
	assert.NoError(t, err)
}

func TestCreateQuorumQueue(t *testing.T) {
	connectionPool, err := fr.NewConnectionPool(Seasoning.PoolConfig, nil, nil)
	assert.NoError(t, err)
	defer connectionPool.Shutdown()

	topologer := fr.NewTopologer(connectionPool)

	table := amqp.Table{"x-queue-type": fr.QueueTypeQuorum}
	err = topologer.CreateQueue("TcrTestQuorumQueue", false, true, false, false, false, table)
	assert.NoError(t, err)

	_, err = topologer.QueueDelete("TcrTestQuorumQueue", false, false, false)
	assert.NoError(t, err)
}
