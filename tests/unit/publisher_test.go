package unit_tests

import (
	"context"
	"testing"

	"github.com/fortytw2/leaktest"
	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/pmatteo/friendlyrabbit/tests/mock"
	"github.com/stretchr/testify/assert"
)

func TestPublisherCreate(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	publisher := fr.NewPublisher(Seasoning, RabbitService.ConnectionPool)
	assert.NotNil(t, publisher)

	publisher.Shutdown(false)
}

func TestPublisherPublishWithError(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	letter := mock.CreateMockRandomLetter("TestUnitQueue")
	e := DefaultPublisher.Publish(letter, false)

	assert.NoError(t, e)
}

func TestPublisherPublishWithConfirmationContextWithError(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	letter := mock.CreateMockRandomLetter("TestUnitQueue")
	ctx, cancelFun := context.WithCancel(context.Background())
	letter.Envelope.Ctx = ctx

	// call function istantly so when publisher try to publish it match the ctx.Done() condition
	cancelFun()

	e := DefaultPublisher.PublishWithConfirmation(letter, DefaultPublisher.PublishTimeout)
	assert.Error(t, e)
}

func TestPublisherPublishWithConfirmationContextWithoutError(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	letter := mock.CreateMockRandomLetter("TestUnitQueue")
	ctx, cancelFun := context.WithCancel(context.Background())
	letter.Envelope.Ctx = ctx

	// call function later when publisher already published
	defer cancelFun()

	e := DefaultPublisher.PublishWithConfirmation(letter, DefaultPublisher.PublishTimeout)
	assert.NoError(t, e)
}
