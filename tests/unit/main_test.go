package unit_tests

import (
	"log"
	"os"
	"testing"

	fr "github.com/pmatteo/friendlyrabbit"
)

var Seasoning *fr.RabbitSeasoning
var RabbitService *fr.RabbitService
var DefaultPublisher *fr.Publisher

func setup() error {
	var err error
	Seasoning = &fr.RabbitSeasoning{
		EncryptionConfig:  &fr.EncryptionConfig{},
		CompressionConfig: &fr.CompressionConfig{},
		PoolConfig: &fr.PoolConfig{
			URI:                  "amqp://guest:guest@localhost:5672/",
			ApplicationName:      "TurboCookedRabbit",
			SleepOnErrorInterval: 100,
			MaxCacheChannelCount: 50,
			MaxConnectionCount:   1,
			Heartbeat:            6,
			ConnectionTimeout:    10,
			TLSConfig: &fr.TLSConfig{
				EnableTLS: false,
			},
		},
		ConsumerConfigs: map[string]*fr.ConsumerConfig{
			"TurboCookedRabbitConsumer": {
				QueueName:            "TestUnitQueue",
				ConsumerName:         "TurboCookedRabbitConsumer",
				AutoAck:              true,
				Exclusive:            false,
				NoWait:               false,
				QosCountOverride:     100,
				SleepOnErrorInterval: 0,
			},
			"TurboCookedRabbitConsumer-Ackable": {
				QueueName:            "TestUnitQueue",
				ConsumerName:         "TurboCookedRabbitConsumer-Ackable",
				AutoAck:              false,
				Exclusive:            false,
				NoWait:               false,
				QosCountOverride:     100,
				SleepOnErrorInterval: 0,
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

	RabbitService, err = fr.NewRabbitService(Seasoning, "", "", nil, nil)
	if err != nil {
		return err
	}

	err = RabbitService.Topologer.CreateQueue("TestUnitQueue", false, true, false, false, false, nil)
	if err != nil {
		return err
	}

	DefaultPublisher = RabbitService.Publisher

	return nil
}

func teardown() {
	_, err := RabbitService.Topologer.QueueDelete("TestUnitQueue", false, false, false)
	if err != nil {
		panic(err)
	}

	RabbitService.Shutdown(true)
}

func TestMain(m *testing.M) {
	defer func() {
		if e := recover(); e != nil {
			log.Fatalln("there was a panic", e)
		}
	}()

	if err := setup(); err != nil {
		log.Fatalln(err)
	}

	code := m.Run()

	teardown()

	os.Exit(code)
}
