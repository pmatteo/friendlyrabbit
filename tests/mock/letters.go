package mock

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
	fr "github.com/pmatteo/friendlyrabbit"
	amqp "github.com/rabbitmq/amqp091-go"
)

var mockRandomSource = rand.NewSource(time.Now().UnixNano())
var mockRandom = rand.New(mockRandomSource)

// CreateMockLetter creates a mock letter for publishing.
func CreateMockLetter(exchangeName string, routingKey string, body []byte) *fr.Letter {

	if body == nil { //   h   e   l   l   o       w   o   r   l   d
		body = []byte("\x68\x65\x6c\x6c\x6f\x20\x77\x6f\x72\x6c\x64")
	}

	e := fr.NewEnvelope(context.Background(), exchangeName, routingKey, nil)
	return fr.NewLetter(uuid.New(), e, body)
}

// CreateMockRandomLetter creates a mock letter for publishing with random sizes and random Ids.
func CreateMockRandomLetter(routingKey string) *fr.Letter {

	body := RandomBytes(mockRandom.Intn(randomMax-randomMin) + randomMin)

	env := fr.NewEnvelope(context.Background(), "", routingKey, amqp.Table{
		"x-testheader": "HelloWorldHeader",
	})

	return fr.NewLetter(uuid.New(), env, body)
}

type BasicStruct struct {
	N int
	S string
}

// CreateMockRandomWrappedBodyLetter creates a mock fr.Letter for publishing with random sizes and random Ids.
func CreateMockRandomWrappedBodyLetter(routingKey string) *fr.Letter {
	n := RandomNumber(1000)
	data := &BasicStruct{
		N: n,
		S: RepeatedRandomString(n, RandomNumber(10)),
	}

	e := fr.NewEnvelope(context.Background(), "", routingKey, nil)
	l, err := fr.NewLetterWithPayload(e, data, nil, nil, true, "")
	if err != nil {
		panic(err)
	}

	return l
}
