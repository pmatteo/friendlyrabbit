package integration_tests

import (
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/pmatteo/friendlyrabbit/tests/mock"
	"github.com/stretchr/testify/assert"
)

// TestConsumingAfterPublish is a combination test of Consuming and Publishing
func TestConsumingAfterPublish(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	connectionPool, err := fr.NewConnectionPool(Seasoning.PoolConfig)
	assert.NoError(t, err)
	defer connectionPool.Shutdown()

	consumerConfig, ok := Seasoning.ConsumerConfigs["TurboCookedRabbitConsumer-Ackable"]
	assert.True(t, ok)

	timeoutAfter := time.After(time.Minute * 2)
	consumer := fr.NewConsumer(consumerConfig, connectionPool)
	assert.NotNil(t, consumer)

	done1 := make(chan struct{}, 1)
	done2 := make(chan struct{}, 1)
	consumer.StartConsuming(nil)

	publisher := fr.NewPublisher(Seasoning, connectionPool)
	letter := mock.CreateMockRandomLetter("TestIntegrationQueue")
	count := 1000

	go monitorPublish(t, timeoutAfter, publisher, count, done1)
	go monitorConsumer(t, timeoutAfter, consumer, count, done2)

	for i := 0; i < count; i++ {
		_ = publisher.Publish(letter, true)
	}

	<-done1
	<-done2

	err = consumer.StopConsuming(false, false)
	assert.NoError(t, err)
}

// TestConsumingAftersPublishLarge is a combination test of Consuming and Publishing
func TestConsumingAftersPublishLarge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestConsumingAftersPublishLarge test in short mode.")
	}

	defer leaktest.CheckTimeout(t, 1*time.Minute)() // Fail on leaked goroutines.

	connectionPool, err := fr.NewConnectionPool(Seasoning.PoolConfig)
	assert.NoError(t, err)
	defer connectionPool.Shutdown()

	consumerConfig, ok := Seasoning.ConsumerConfigs["TurboCookedRabbitConsumer-Ackable"]
	assert.True(t, ok)

	timeoutAfter := time.After(time.Minute * 5)
	consumer := fr.NewConsumer(consumerConfig, connectionPool)
	assert.NotNil(t, consumer)

	done1 := make(chan struct{}, 1)
	done2 := make(chan struct{}, 1)
	consumer.StartConsuming(nil)

	publisher := fr.NewPublisher(Seasoning, connectionPool)
	letter := mock.CreateMockRandomLetter("TestIntegrationQueue")
	count := 100000

	go monitorPublish(t, timeoutAfter, publisher, count, done1)
	go monitorConsumer(t, timeoutAfter, consumer, count, done2)

	for i := 0; i < count; i++ {
		_ = publisher.Publish(letter, false)
	}

	<-done1
	<-done2

	err = consumer.StopConsuming(false, false)
	assert.NoError(t, err)
}

// TestConsumingAfterPublishConfirmationLarge is a combination test of Consuming and Publishing with confirmation.
func TestConsumingAfterPublishConfirmationLarge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestConsumingAfterPublishConfirmationLarge test in short mode.")
	}

	defer leaktest.Check(t)() // Fail on leaked goroutines.

	connectionPool, err := fr.NewConnectionPool(Seasoning.PoolConfig)
	assert.NoError(t, err)
	defer connectionPool.Shutdown()

	consumerConfig, ok := Seasoning.ConsumerConfigs["TurboCookedRabbitConsumer-Ackable"]
	assert.True(t, ok)

	timeoutAfter := time.After(time.Minute * 2)
	consumer := fr.NewConsumer(consumerConfig, connectionPool)
	assert.NotNil(t, consumer)

	done1 := make(chan struct{}, 1)
	done2 := make(chan struct{}, 1)
	consumer.StartConsuming(nil)

	publisher := fr.NewPublisher(Seasoning, connectionPool)
	letter := mock.CreateMockRandomLetter("TestIntegrationQueue")
	count := 10000

	go monitorPublish(t, timeoutAfter, publisher, count, done1)
	go monitorConsumer(t, timeoutAfter, consumer, count, done2)

	for i := 0; i < count; i++ {
		_ = publisher.PublishWithConfirmation(letter, 500*time.Millisecond)
	}

	<-done1
	<-done2

	err = consumer.StopConsuming(false, false)
	assert.NoError(t, err)
}

// TestPublishConfirmation is a combination test of Consuming and Publishing with confirmation.
func TestPublishConfirmation(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	timeoutAfter := time.After(time.Minute * 2)
	done1 := make(chan struct{}, 1)

	connectionPool, err := fr.NewConnectionPool(Seasoning.PoolConfig)
	assert.NoError(t, err)
	defer connectionPool.Shutdown()

	publisher := fr.NewPublisher(Seasoning, connectionPool)
	letter := mock.CreateMockRandomLetter("TestIntegrationQueue")
	count := 100

	go monitorPublish(t, timeoutAfter, publisher, count, done1)

	for i := 0; i < count; i++ {
		e := publisher.PublishWithConfirmation(letter, time.Millisecond*500)
		assert.NoError(t, e)
	}

	<-done1
}

func monitorConsumer(t *testing.T, timeoutAfter <-chan time.Time, con *fr.Consumer, count int, done chan struct{}) {
	receivedMessageCount := 0
WaitForConsumer:
	for {
		select {
		case <-timeoutAfter:
			assert.Fail(t, "Publish Timeout")
			done <- struct{}{}
			return

		case message := <-con.ReceivedMessages():
			_ = message.Acknowledge()
			receivedMessageCount++
			if receivedMessageCount == count {
				break WaitForConsumer
			}

		default:
			time.Sleep(time.Millisecond)
		}
	}

	assert.Equal(t, count, receivedMessageCount, "Received Message Count: %d  Expected Count: %d", receivedMessageCount, count)
	done <- struct{}{}
}

func monitorPublish(t *testing.T, timeoutAfter <-chan time.Time, pub *fr.Publisher, count int, done chan struct{}) {
	publishSuccessCount := 0
	publishFailureCount := 0

WaitForReceiptsLoop:
	for {
		select {
		case <-timeoutAfter:
			return
		case receipt := <-pub.PublishReceipts():
			if receipt.Success {
				publishSuccessCount++
				if count == publishSuccessCount+publishFailureCount {
					break WaitForReceiptsLoop
				}
			} else {
				publishFailureCount++
			}

		default:
			time.Sleep(time.Millisecond * 1)
		}
	}

	actualCount := publishSuccessCount + publishFailureCount
	assert.Equal(t, count, actualCount, "Publish Success Count: %d  Expected Count: %d", publishSuccessCount, count)
	done <- struct{}{}
}
