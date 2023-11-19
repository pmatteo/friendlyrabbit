package benchmark_tests

import (
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	cmap "github.com/orcaman/concurrent-map"
	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/pmatteo/friendlyrabbit/tests/mock"
	"github.com/stretchr/testify/assert"
)

var seasoning *fr.RabbitSeasoning
var connectionPool *fr.ConnectionPool

func setup() error {
	var err error
	seasoning = &fr.RabbitSeasoning{
		EncryptionConfig:  &fr.EncryptionConfig{},
		CompressionConfig: &fr.CompressionConfig{},
		PoolConfig: &fr.PoolConfig{
			URI:                  "amqp://guest:guest@localhost:5672/",
			ApplicationName:      "TurboCookedRabbit",
			SleepOnErrorInterval: 100,
			MaxCacheChannelCount: 50,
			MaxConnectionCount:   3,
			Heartbeat:            6,
			ConnectionTimeout:    10,
			TLSConfig: &fr.TLSConfig{
				EnableTLS: false,
			},
		},
		ConsumerConfigs: map[string]*fr.ConsumerConfig{
			"TurboCookedRabbitConsumer": {
				Enabled:              true,
				QueueName:            "TestBenchmarkQueue",
				ConsumerName:         "TurboCookedRabbitConsumer",
				AutoAck:              true,
				Exclusive:            false,
				NoWait:               false,
				QosCountOverride:     100,
				SleepOnErrorInterval: 0,
				SleepOnIdleInterval:  0,
			},
			"TurboCookedRabbitConsumer-Ackable": {
				Enabled:              true,
				QueueName:            "TestBenchmarkQueue",
				ConsumerName:         "TurboCookedRabbitConsumer-Ackable",
				AutoAck:              false,
				Exclusive:            false,
				NoWait:               false,
				QosCountOverride:     100,
				SleepOnErrorInterval: 0,
				SleepOnIdleInterval:  0,
			},
		},
		PublisherConfig: &fr.PublisherConfig{
			AutoAck:                false,
			SleepOnIdleInterval:    0,
			SleepOnErrorInterval:   0,
			PublishTimeOutInterval: 100,
			MaxRetryCount:          3,
		},
	}

	connectionPool, err = fr.NewConnectionPool(seasoning.PoolConfig)
	if err != nil {
		return err
	}

	err = fr.NewTopologer(connectionPool).CreateQueue("TestBenchmarkQueue", false, true, false, false, false, nil)
	if err != nil {
		return err
	}

	return nil
}

func purgeQueue(b *testing.B) {
	_, err := fr.NewTopologer(connectionPool).PurgeQueue("TestBenchmarkQueue", false)
	assert.NoError(b, err)
}

func teardown() {
	_, err := fr.NewTopologer(connectionPool).QueueDelete("TestBenchmarkQueue", false, false, false)
	if err != nil {
		panic(err)
	}

	connectionPool.Shutdown()
}

func TestMain(m *testing.M) {
	defer func() {
		if e := recover(); e != nil {
			teardown()
			log.Fatalln("there was a panic", e)
		}
	}()

	if err := setup(); err != nil {
		log.Fatalln(err)
	}

	os.Exit(m.Run())
}

func publishWithConfirmation(
	b *testing.B,
	conMap cmap.ConcurrentMap,
	done chan bool,
	timeoutDuration time.Duration,
	publisher *fr.Publisher,
) {
	messagesPublished := 0
	messagesFailedToPublish := 0
	publisherErrors := 0

	timeout := time.After(timeoutDuration)

PublishLoop:
	for {
		select {
		case <-timeout:
			break PublishLoop
		default:
			newLetter := mock.CreateMockRandomWrappedBodyLetter("TestBenchmarkQueue")
			conMap.Set(newLetter.LetterID.String(), false)
			_ = publisher.PublishWithConfirmation(newLetter, time.Second)

			notice := <-publisher.PublishReceipts()
			if notice.Success {
				messagesPublished++
				notice = nil
			} else {
				messagesFailedToPublish++
				notice = nil
			}
		}
	}

	b.Logf("Publisher Errors: %d\r\n", publisherErrors)
	b.Logf("Messages Published: %d\r\n", messagesPublished)
	b.Logf("Messages Failed to Publish: %d\r\n", messagesFailedToPublish)

	done <- true
}

func consumeLoop(
	b *testing.B,
	conMap cmap.ConcurrentMap,
	done chan bool,
	timeoutDuration time.Duration,
	publisher *fr.Publisher,
	consumer *fr.Consumer,
) {
	messagesReceived := 0
	messagesAcked := 0
	messagesFailedToAck := 0
	consumerErrors := 0

	timeout := time.After(timeoutDuration)

	ticker := time.NewTicker(time.Millisecond)

ConsumeLoop:
	for {
		select {
		case <-timeout:
			break ConsumeLoop

		case err := <-consumer.Errors():
			b.Logf("%s: Consumer Error - %s\r\n", time.Now(), err)
			consumerErrors++

		case message := <-consumer.ReceivedMessages():
			body, err := readWrappedBodyFromJSONBytes(message.Delivery.Body)
			assert.NoError(b, err, "message was not deserializeable")

			// Accuracy check
			tmp, ok := conMap.Get(body.LetterID.String())
			assert.True(b, ok, "letter (%s) received that wasn't published!", body.LetterID.String())

			state := tmp.(bool)
			assert.False(b, state, "duplicate letter (%s) received!", body.LetterID.String())

			messagesReceived++
			conMap.Set(body.LetterID.String(), true)

			err = message.Acknowledge()
			if err != nil {
				messagesFailedToAck++
			} else {
				messagesAcked++
			}
		}
	}

	ticker.Stop()

	b.Logf("Consumer Errors: %d\r\n", consumerErrors)
	b.Logf("Messages Acked: %d\r\n", messagesAcked)
	b.Logf("Messages Failed to Ack: %d\r\n", messagesFailedToAck)
	b.Logf("Messages Received: %d\r\n", messagesReceived)

	done <- true
}

// ReadWrappedBodyFromJSONBytes simply read the bytes as a Letter.
func readWrappedBodyFromJSONBytes(data []byte) (*fr.WrappedBody, error) {
	body := &fr.WrappedBody{}
	err := json.Unmarshal(data, body)

	return body, err
}
