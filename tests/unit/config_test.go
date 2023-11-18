package unit_tests

import (
	"testing"

	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {
	fileNamePath := "../data/testseasoning.json"

	assert.FileExists(t, fileNamePath)

	config, err := fr.ConvertJSONFileToConfig(fileNamePath)

	assert.Nil(t, err)
	assert.NotEqual(t, "", config.PoolConfig.URI, "RabbitMQ URI should not be blank.")
}

func TestReadTopologyConfig(t *testing.T) {
	fileNamePath := "../data/testtopology.json"

	assert.FileExists(t, fileNamePath)

	config, err := fr.ConvertJSONFileToTopologyConfig(fileNamePath)

	assert.Nil(t, err)
	assert.NotEqual(t, 0, len(config.Exchanges))
	assert.NotEqual(t, 0, len(config.Queues))
	assert.NotEqual(t, 0, len(config.QueueBindings))
	assert.NotEqual(t, 0, len(config.ExchangeBindings))
}
