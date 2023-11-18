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

	publisher := fr.NewPublisherFromConfig(Seasoning, RabbitService.ConnectionPool)
	assert.NotNil(t, publisher)

	publisher.Shutdown(false)
}

func TestPublisherPublishWithError(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	letter := mock.CreateMockRandomLetter("TestUnitQueue")
	e := DefaultPublisher.PublishWithError(letter, false)

	assert.NoError(t, e)
}

func TestPublisherPublishWithConfirmationContextWithError(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	letter := mock.CreateMockRandomLetter("TestUnitQueue")
	ctx, cancelFun := context.WithCancel(context.Background())

	// call function istantly so when publisher try to publish it match the ctx.Done() condition
	cancelFun()

	e := DefaultPublisher.PublishWithConfirmationContextError(ctx, letter)
	assert.Error(t, e)
}

func TestPublisherPublishWithConfirmationContextWithoutError(t *testing.T) {
	defer leaktest.Check(t)() // Fail on leaked goroutines.

	letter := mock.CreateMockRandomLetter("TestUnitQueue")
	ctx, cancelFun := context.WithCancel(context.Background())

	// call function later when publisher already published
	defer cancelFun()

	e := DefaultPublisher.PublishWithConfirmationContextError(ctx, letter)
	assert.NoError(t, e)
}
