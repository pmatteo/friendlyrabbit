package unit_tests

import (
	"fmt"
	"testing"

	"github.com/fortytw2/leaktest"
	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/stretchr/testify/assert"
)

func TestCreateConsumer(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	consumerConfig, err := RabbitService.GetConsumerConfig("TurboCookedRabbitConsumer")
	assert.NoError(t, err)

	ackableConsumerConfig, err := RabbitService.GetConsumerConfig("TurboCookedRabbitConsumer-Ackable")
	assert.NoError(t, err)

	consumer1 := fr.NewConsumerFromConfig(ackableConsumerConfig, RabbitService.ConnectionPool)
	assert.NotNil(t, consumer1)

	consumer2 := fr.NewConsumerFromConfig(consumerConfig, RabbitService.ConnectionPool)
	assert.NotNil(t, consumer2)
}

func TestConsumerStartStop(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	consumerConfig, err := RabbitService.GetConsumerConfig("TurboCookedRabbitConsumer")
	assert.NoError(t, err)

	consumer := fr.NewConsumerFromConfig(consumerConfig, RabbitService.ConnectionPool)
	assert.NotNil(t, consumer)

	consumer.StartConsuming()
	err = consumer.StopConsuming(false, false)
	assert.NoError(t, err)
}

func TestConsumerStartWithActionStop(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	consumerConfig, err := RabbitService.GetConsumerConfig("TurboCookedRabbitConsumer")
	assert.NoError(t, err)

	consumer := fr.NewConsumerFromConfig(consumerConfig, RabbitService.ConnectionPool)
	assert.NotNil(t, consumer)

	consumer.StartConsumingWithAction(func(msg *fr.ReceivedMessage) {
		if err := msg.Acknowledge(); err != nil {
			fmt.Printf("Error acking message: %v\r\n", msg.Delivery.Body)
		}
	})
	err = consumer.StopConsuming(false, false)
	assert.NoError(t, err)
}

func TestConsumerGet(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	consumerConfig, err := RabbitService.GetConsumerConfig("TurboCookedRabbitConsumer")
	assert.NoError(t, err)

	_, e := RabbitService.Topologer.PurgeQueue("TestUnitQueue", false)
	assert.NoError(t, e)

	consumer := fr.NewConsumerFromConfig(consumerConfig, RabbitService.ConnectionPool)
	assert.NotNil(t, consumer)

	delivery, err := consumer.Get("TestUnitQueue")
	assert.Nil(t, delivery) // empty queue should be nil
	assert.NoError(t, err)
}
