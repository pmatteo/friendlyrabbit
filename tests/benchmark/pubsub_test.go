package benchmark_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"

	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/pmatteo/friendlyrabbit/tests/mock"
)

func verifyAccuracyB(b *testing.B, conMap cmap.ConcurrentMap) int {
	// Breakpoint here and check the conMap for 100% accuracy.
	var percentage int

	notReceived := 0
	for item := range conMap.IterBuffered() {
		state := item.Val.(bool)
		assert.True(b, state, "LetterID: %s was not received.\r\n", item.Key)
		if !state {
			notReceived++
		}
	}

	b.Logf("messages not received: %d\r\n", notReceived)

	if conMap.Count() > 0 {
		percentage = (notReceived * 100) / conMap.Count()
	}

	return percentage
}

func BenchmarkPublishAndConsumeMany(b *testing.B) {
	purgeQueue(b)

	messageCount := 1000000
	if testing.Short() {
		messageCount = 10000
	}

	b.ReportAllocs()

	b.Logf("Benchmark Starts: %s\r\n", time.Now())

	consumerConfig, ok := seasoning.ConsumerConfigs["TurboCookedRabbitConsumer"]
	assert.True(b, ok)
	consumer := fr.NewConsumer(consumerConfig, connectionPool)
	consumer.StartConsuming(nil)

	publisher := fr.NewPublisher(seasoning, connectionPool)
	publisher.StartAutoPublishing()
	defer publisher.Shutdown(false)

	go func() {
		for i := 0; i < messageCount; i++ {
			letter := mock.CreateMockRandomLetter("TestBenchmarkQueue")
			letter.LetterID = uuid.New()

			publisher.QueueLetter(letter)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(1*time.Minute))
	messagesReceived := 0
	messagesPublished := 0
	messagesFailedToPublish := 0
	consumerErrors := 0
	channelPoolErrors := 0

ReceivePublishConfirmations:
	for {
		select {
		case <-ctx.Done():
			fmt.Print("\r\nContextTimeout\r\n")
			break ReceivePublishConfirmations
		case publish := <-publisher.PublishReceipts():
			if publish.Success {
				messagesPublished++
			} else {
				b.Logf("%s: Failed to published Failed Letter %s: %s\r\n", time.Now(), publish.LetterID.String(), publish.Error)
				messagesFailedToPublish++
			}
		case err := <-consumer.Errors():
			b.Logf("%s: Consumer - Error: %s\r\n", time.Now(), err)
			consumerErrors++
		case <-consumer.ReceivedMessages():
			//b.Logf("%s: MessageReceived\r\n", time.Now())
			messagesReceived++
		default:
			time.Sleep(time.Microsecond)
		}

		if messagesReceived+messagesFailedToPublish == messageCount {
			break ReceivePublishConfirmations
		}
	}

	assert.Equal(b, messageCount, messagesReceived+messagesFailedToPublish)
	b.Logf("Channel Pool Errors: %d\r\n", channelPoolErrors)
	b.Logf("Messages Published: %d\r\n", messagesPublished)
	b.Logf("Messages Failed to Publish: %d\r\n", messagesFailedToPublish)
	b.Logf("Consumer Errors: %d\r\n", consumerErrors)
	b.Logf("Consumer Messages Received: %d\r\n", messagesReceived)

	assert.True(b, consumer.Started())
	consumer.StopConsuming(true)

	// Wait processing to stop
	time.Sleep(50 * time.Millisecond)

	assert.False(b, consumer.Started())

	cancel()
}

func BenchmarkPublishConsumeAckForDuration(b *testing.B) {
	purgeQueue(b)

	timeDuration := time.Duration(5 * time.Minute)
	conTimeoutDuration := timeDuration + (5 * time.Second)
	if testing.Short() {
		timeDuration = time.Duration(1 * time.Minute)
		conTimeoutDuration = timeDuration + (5 * time.Second)
	}

	b.Logf("Benchmark Starts: %s\r\n", time.Now())
	b.Logf("Est. Benchmark End: %s\r\n", time.Now().Add(timeDuration))

	publishDone, consumerDone := make(chan bool, 1), make(chan bool, 1)

	conMap := cmap.New()

	b.ReportAllocs()

	publisher := fr.NewPublisher(seasoning, connectionPool)
	defer publisher.Shutdown(false)

	consumerConfig, ok := seasoning.ConsumerConfigs["TurboCookedRabbitConsumer-Ackable"]
	assert.True(b, ok)
	consumer := fr.NewConsumer(consumerConfig, connectionPool)
	consumer.StartConsuming(nil)

	go publishWithConfirmation(b, conMap, publishDone, timeDuration, publisher)
	go consumeLoop(b, conMap, consumerDone, conTimeoutDuration, publisher, consumer)

	<-publishDone
	<-consumerDone

	assert.True(b, consumer.Started())
	consumer.StopConsuming(true)

	// Wait processing to stop
	time.Sleep(50 * time.Millisecond)

	assert.False(b, consumer.Started())

	b.Logf("Percentage of messages not received: %d\r\n", verifyAccuracyB(b, conMap))
}
