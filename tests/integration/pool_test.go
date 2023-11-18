package integration_tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/stretchr/testify/assert"
)

func TestConnectionPoolGetConnectionAndReturnLoop(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	count := 100000
	if testing.Short() {
		count = 10000
	}

	for i := 0; i < count; i++ {
		connHost, err := RabbitService.ConnectionPool.GetConnection()
		assert.NoError(t, err)

		RabbitService.ConnectionPool.ReturnConnection(connHost, false)
	}
}

func TestConnectionPoolGetChannelFromPoolAndReturnLoop(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	count := 100000
	if testing.Short() {
		count = 10000
	}

	for i := 0; i < count; i++ {
		chanHost := RabbitService.ConnectionPool.GetChannelFromPool()

		RabbitService.ConnectionPool.ReturnChannel(chanHost, false)
	}
}

// TestConnectionGetConnectionAndReturnSlowLoop is designed to be slow test connection recovery by severing all connections
// and then verify connections properly restore.
func TestConnectionPoolGetConnectionAndReturnSlowLoop(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	count := 10000
	if testing.Short() {
		count = 1000
	}

	wg := &sync.WaitGroup{}
	semaphore := make(chan bool, 100)
	for i := 0; i < count; i++ {

		wg.Add(1)
		semaphore <- true
		go func() {
			defer wg.Done()

			connHost, err := RabbitService.ConnectionPool.GetConnection()
			assert.NoError(t, err)

			time.Sleep(time.Millisecond * 10)

			RabbitService.ConnectionPool.ReturnConnection(connHost, false)

			<-semaphore
		}()
	}

	wg.Wait()
}

// TestConnectionGetConnectionAndReturnSlowLoop is designed to be slow test connection recovery by severing all connections
// and then verify connections and channels properly (and evenly over connections) restore and continues publishing.
// All publish errors are not recovered but used to trigger channel recovery, total publish count should be lower
// than expected 10,000 during outages.
func TestConnectionGetChannelAndReturnSlowLoop(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	count := 10000
	if testing.Short() {
		count = 1000
	}

	body := []byte("\x68\x65\x6c\x6c\x6f\x20\x77\x6f\x72\x6c\x64")

	wg := &sync.WaitGroup{}
	semaphore := make(chan bool, 200) // at most 300 requests a time
	for i := 0; i < count; i++ {      // total request to try

		wg.Add(1)
		semaphore <- true
		go func() {
			defer wg.Done()

			chanHost := RabbitService.ConnectionPool.GetChannelFromPool()

			time.Sleep(time.Millisecond * 100) // artificially create channel poool contention by long exposure

			err := chanHost.Channel.PublishWithContext(context.Background(), "", "TestIntegrationQueue", false, false, amqp.Publishing{
				ContentType:  "plaintext/text",
				Body:         body,
				DeliveryMode: 2,
			})

			RabbitService.ConnectionPool.ReturnChannel(chanHost, err != nil)

			<-semaphore
		}()
	}

	wg.Wait() // wait for the final batch of requests to finish
}
