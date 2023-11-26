package integration_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/pmatteo/friendlyrabbit/tests/mock"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/stretchr/testify/assert"
)

// TestBasicPublishToNonExistentExchange tests what happen when a publish to exchange
// that doesn't exist also doesn't error. This is a demonstration of server
// side Dead Lettering.
func TestBasicPublishToNonExistentExchange(t *testing.T) {
	defer leaktest.Check(t)()

	letter := mock.CreateMockLetter("DoesNotExist", "TestIntegrationQueue", nil)
	amqpConn, err := amqp.Dial(Seasoning.PoolConfig.URI)
	assert.NoError(t, err)

	amqpChan, err := amqpConn.Channel()
	assert.NoError(t, err)

	err = amqpChan.PublishWithContext(context.Background(),
		letter.Envelope.Exchange,
		letter.Envelope.RoutingKey,
		letter.Envelope.Mandatory,
		letter.Envelope.Immediate,
		amqp.Publishing{
			ContentType: letter.Envelope.ContentType,
			Body:        letter.Body,
			MessageId:   letter.LetterID.String(),
			Timestamp:   time.Now().UTC(),
			AppId:       "TCR-Test",
		})

	assert.NoError(t, err)

	amqpChan.Close()
	amqpConn.Close()
}

func TestPublishAndWaitForReceipt(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	publisher := fr.NewPublisher(Seasoning, RabbitService.ConnectionPool)
	assert.NotNil(t, publisher)

	letter := mock.CreateMockRandomLetter("TestIntegrationQueue")
	_ = publisher.Publish(letter, true)

WaitLoop:
	for {
		select {
		case receipt := <-publisher.PublishReceipts():
			assert.Equal(t, receipt.Success, true)
			assert.NoError(t, receipt.Error)
			assert.Nil(t, receipt.FailedLetter)
			break WaitLoop
		default:
			time.Sleep(time.Millisecond * 1)
		}
	}
}

func TestPublishWithConfirmation(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	publisher := fr.NewPublisher(Seasoning, RabbitService.ConnectionPool)

	letter := mock.CreateMockRandomLetter("TestIntegrationQueue")
	e := publisher.PublishWithConfirmation(letter, time.Millisecond*500)
	assert.NoError(t, e)

WaitLoop:
	for {
		select {
		case receipt := <-publisher.PublishReceipts():
			assert.Equal(t, receipt.Success, true)
			assert.NoError(t, receipt.Error)
			assert.Nil(t, receipt.FailedLetter)
			break WaitLoop
		case <-time.After(time.Second * 1):
			t.Log("PublishWithConfirmation timed out")
			t.Fail()
			break WaitLoop
		}
	}
}

func TestPublishAccuracy(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	count, successCount := 10000, 0
	if testing.Short() {
		count = 1000
	}

	t1 := time.Now()
	fmt.Printf("TestPublishAccuracy Starts: %s\r\n", t1)
	defer func() {
		t2 := time.Now()
		diff := t2.Sub(t1)
		fmt.Printf("TestPublishAccuracy End: %s\r\n", t2)
		fmt.Printf("Messages: %f msg/s\r\n", float64(count)/diff.Seconds())
	}()

	publisher := fr.NewPublisher(Seasoning, RabbitService.ConnectionPool)
	letter := mock.CreateMockRandomLetter("TestIntegrationQueue")
	letter.Envelope.DeliveryMode = amqp.Transient

	for i := 0; i < count; i++ {
		_ = publisher.Publish(letter, true)
	}

	for ; successCount < count; successCount++ {
		receipt := <-publisher.PublishReceipts()
		if !receipt.Success {
			t.Log(receipt.Error)
			t.Fail()
		}
	}

	assert.Equal(t, count, successCount)
}

func TestPublishWithConfirmationAccuracy(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	count, successCount := 10000, 0
	if testing.Short() {
		count = 100
	}

	t.Log(count)

	t1 := time.Now()
	fmt.Printf("TestPublishWithConfirmationAccuracy Starts: %s\r\n", t1)
	defer func() {
		t2 := time.Now()
		diff := t2.Sub(t1)
		fmt.Printf("TestPublishWithConfirmationAccuracy End: %s\r\n", t2)
		fmt.Printf("Messages: %f msg/s\r\n", float64(count)/diff.Seconds())
	}()

	connectionPool, err := fr.NewConnectionPool(Seasoning.PoolConfig)
	assert.NoError(t, err)

	publisher := fr.NewPublisher(Seasoning, connectionPool)
	defer publisher.Shutdown(true)

	letter := mock.CreateMockRandomLetter("TestIntegrationQueue")
	for i := 0; i < count; i++ {
		e := publisher.PublishWithConfirmation(letter, time.Millisecond*500)
		assert.NoError(t, e)
	}

	for ; successCount < count; successCount++ {
		receipt := <-publisher.PublishReceipts()
		assert.True(t, receipt.Success)
		assert.NoError(t, receipt.Error)
	}

	assert.Equal(t, count, successCount)
}
