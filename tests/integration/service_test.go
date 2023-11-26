package integration_tests

import (
	"testing"

	"github.com/fortytw2/leaktest"
	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/pmatteo/friendlyrabbit/tests/mock"
	"github.com/stretchr/testify/assert"
)

// TestOutageAndQueueLetterAccuracy is similar to the above. It is designed to be slow test connection recovery by severing all connections
// and then verify connections and channels properly (and evenly over connections) restore.
func TestOutageAndQueueLetterAccuracy(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	count := 10000
	if testing.Short() {
		count = 500
	}

	letter := mock.CreateMockRandomLetter("TestIntegrationQueue")

	// RabbitService is configured (out of the box) to use PublishWithConfirmations on AutoPublish.
	// All publish receipt events outside of the standard publish confirmation for streadway/amqp are also
	// listened to. During a failed publish receipt event, the default (and overridable behavior) is to automatically
	// requeue into the AutoPublisher for retry.
	//
	// This is fairly close to the most durable asynchronous form of Publishing you probably could create.
	// The only thing that would probably improve this, is directly call PublishWithConfirmations more
	// directly (which is an option).
	//
	// RabbitService has been hardened with a solid shutdown mechanism so please hook into this
	// in your application.
	//
	// This still leaves a few open ended vulnerabilities.
	// ApplicationSide:
	//    1.) Uncontrolled shutdown event or panic.
	//    2.) RabbitService publish receipt indicates a retry scenario, but we are mid-shutdown.
	// ServerSide:
	//    1.) A storage failure (without backup).
	//    2.) Split Brain event in HA mode.

	for i := 0; i < count; i++ {
		err := RabbitService.QueueLetter(letter)
		assert.NoError(t, err)
	}
}

// TestPublishWithHeaderAndVerify verifies headers are publishing.
func TestPublishWithHeaderAndVerify(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.
	letter := mock.CreateMockRandomLetter("TestIntegrationQueue")

	err := RabbitService.PublishLetter(letter)
	assert.NoError(t, err)

	consumer, err := RabbitService.GetConsumer("TurboCookedRabbitConsumer-Ackable")
	assert.NoError(t, err)

	delivery, err := consumer.Get("TestIntegrationQueue")
	assert.NoError(t, err)
	assert.NotNil(t, delivery)

	testHeader, ok := delivery.Headers["x-testheader"]
	assert.True(t, ok)
	assert.NotNil(t, testHeader)
	assert.Equal(t, "HelloWorldHeader", testHeader.(string))
}

// TestPublishWithHeaderAndConsumerReceivedHeader verifies headers are being consumed into ReceivedData.
func TestPublishWithHeaderAndConsumerReceivedHeader(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	service, err := fr.NewRabbitService(Seasoning, "", "", nil, nil)
	defer service.Shutdown(true)
	assert.NoError(t, err)

	letter := mock.CreateMockRandomLetter("TestIntegrationQueue")

	err = service.PublishLetter(letter)
	assert.NoError(t, err)

	consumer, err := service.GetConsumer("TurboCookedRabbitConsumer-Ackable")
	consumer.StartConsuming(nil)
	assert.NoError(t, err)

	data := <-consumer.ReceivedMessages()

	testHeader, ok := data.Delivery.Headers["x-testheader"]
	assert.True(t, ok)

	t.Logf("Header Received: %s\r\n", testHeader.(string))
}
