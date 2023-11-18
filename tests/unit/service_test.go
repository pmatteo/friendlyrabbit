package unit_tests

import (
	"context"
	"testing"
	"time"

	"github.com/fortytw2/leaktest"
	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/pmatteo/friendlyrabbit/tests/mock"
	"github.com/stretchr/testify/assert"
)

func TestCreateRabbitService(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	service, err := fr.NewRabbitService(Seasoning, "", "", nil, nil)
	defer service.Shutdown(true)
	assert.NoError(t, err)

	assert.NotNil(t, service)
}

func TestCreateRabbitServiceWithEncryption(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	Seasoning.EncryptionConfig.Enabled = true
	service, err := fr.NewRabbitService(Seasoning, "PasswordyPassword", "SaltySalt", nil, nil)
	defer service.Shutdown(true)

	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestRabbitServicePublish(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	data := mock.RandomBytes(1000)
	err := RabbitService.Publish(context.Background(), data, "", "TestUnitQueue", "", false, nil)
	assert.NoError(t, err)
}

func TestRabbitServicePublishLetter(t *testing.T) {
	defer leaktest.CheckTimeout(t, 90*time.Second)() // Fail on leaked goroutines.

	letter := mock.CreateMockRandomLetter("TestUnitQueue")
	err := RabbitService.PublishLetter(letter)
	assert.NoError(t, err)
}

func TestRabbitServicePublishAndConsumeLetter(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.
	letter := mock.CreateMockRandomLetter("TestUnitQueue")
	err := RabbitService.PublishLetter(letter)
	assert.NoError(t, err)
}

func TestRabbitServicePublishLetterToNonExistentQueueForRetryTesting(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	Seasoning.PublisherConfig.PublishTimeOutInterval = 0 // triggering instant timeouts for retry test

	letter := mock.CreateMockRandomLetter("QueueDoesNotExist")
	err := RabbitService.QueueLetter(letter)
	assert.NoError(t, err)

	<-time.After(time.Duration(2 * time.Second))
}
