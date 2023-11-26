package integration_tests

import (
	"log"
	"os"
	"testing"

	fr "github.com/pmatteo/friendlyrabbit"
)

var Seasoning *fr.RabbitSeasoning
var RabbitService *fr.RabbitService

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
				QueueName:            "TestIntegrationQueue",
				ConsumerName:         "TurboCookedRabbitConsumer",
				AutoAck:              true,
				Exclusive:            false,
				NoWait:               false,
				QosCountOverride:     100,
				SleepOnErrorInterval: 0,
			},
			"TurboCookedRabbitConsumer-Ackable": {
				QueueName:            "TestIntegrationQueue",
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

	err = RabbitService.Topologer.CreateQueue("TestIntegrationQueue", false, true, false, false, false, nil)
	if err != nil {
		return err
	}

	return nil
}

func teardown() {
	_, err := RabbitService.Topologer.QueueDelete("TestIntegrationQueue", false, false, false)
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
