package unit_tests

import (
	"testing"

	"github.com/fortytw2/leaktest"
	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/stretchr/testify/assert"
)

func TestConnectionPoolCreateWithWrongURI(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	old := Seasoning.PoolConfig.URI
	Seasoning.PoolConfig.URI = "amqp://wrong:wrong@localhost:1111/"

	cp, err := fr.NewConnectionPool(Seasoning.PoolConfig)
	assert.Nil(t, cp)
	assert.Error(t, err)

	Seasoning.PoolConfig.URI = old
}

func TestConnectionPoolCreateWithZeroConnections(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	old := Seasoning.PoolConfig.MaxConnectionCount
	Seasoning.PoolConfig.MaxConnectionCount = 0

	cp, err := fr.NewConnectionPool(Seasoning.PoolConfig)
	assert.Nil(t, cp)
	assert.Error(t, err)

	Seasoning.PoolConfig.MaxConnectionCount = old
}

func TestConnectionPoolCreateWithErrorHandler(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	seasoning, err := fr.ConvertJSONFileToConfig("../data/badtest.json")
	if err != nil {
		return
	}

	cp, err := fr.NewConnectionPoolWithErrorHandler(seasoning.PoolConfig, func(err error) {})
	assert.Nil(t, cp)
	assert.Error(t, err)
}

func TestConnectionPoolGetConnection(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	assert.Equal(t, int64(1), RabbitService.ConnectionPool.ActiveConnections())

	conHost, err := RabbitService.ConnectionPool.GetConnection()
	assert.NotNil(t, conHost)
	assert.NoError(t, err)

	RabbitService.ConnectionPool.ReturnConnection(conHost, false)
	assert.Equal(t, int64(1), RabbitService.ConnectionPool.ActiveConnections())
}

func TestConnectionPoolCreateAndGetAckableChannel(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	assert.Equal(t, int64(1), RabbitService.ConnectionPool.ActiveConnections())

	conHost, err := RabbitService.ConnectionPool.GetConnection()
	assert.NotNil(t, conHost)
	assert.NoError(t, err)

	chanHost := RabbitService.ConnectionPool.GetChannelFromPool()
	assert.NotNil(t, chanHost)

	RabbitService.ConnectionPool.ReturnConnection(conHost, false)
	assert.Equal(t, int64(1), RabbitService.ConnectionPool.ActiveConnections())
}

func TestConnectionPoolGetChannel(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	chanHost := RabbitService.ConnectionPool.GetChannelFromPool()
	assert.NotNil(t, chanHost)
	chanHost.Close()
}
