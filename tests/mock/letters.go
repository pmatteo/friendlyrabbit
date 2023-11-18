package mock

import (
	"context"
	"encoding/json"
	"math/rand"
	"time"

	"github.com/google/uuid"
	fr "github.com/pmatteo/friendlyrabbit"
	utils_json "github.com/pmatteo/friendlyrabbit/internal/utils/json"
	amqp "github.com/rabbitmq/amqp091-go"
)

var mockRandomSource = rand.NewSource(time.Now().UnixNano())
var mockRandom = rand.New(mockRandomSource)

// CreateMockLetter creates a mock letter for publishing.
func CreateMockLetter(exchangeName string, routingKey string, body []byte) *fr.Letter {

	if body == nil { //   h   e   l   l   o       w   o   r   l   d
		body = []byte("\x68\x65\x6c\x6c\x6f\x20\x77\x6f\x72\x6c\x64")
	}

	envelope := &fr.Envelope{
		Ctx:          context.Background(),
		Exchange:     exchangeName,
		RoutingKey:   routingKey,
		ContentType:  "application/json",
		DeliveryMode: 2,
	}

	return &fr.Letter{
		LetterID:   uuid.New(),
		RetryCount: uint32(3),
		Body:       body,
		Envelope:   envelope,
	}
}

// CreateMockRandomLetter creates a mock letter for publishing with random sizes and random Ids.
func CreateMockRandomLetter(routingKey string) *fr.Letter {

	body := RandomBytes(mockRandom.Intn(randomMax-randomMin) + randomMin)

	envelope := &fr.Envelope{
		Ctx:          context.Background(),
		Exchange:     "",
		RoutingKey:   routingKey,
		ContentType:  "application/json",
		DeliveryMode: 2,
		Headers:      make(amqp.Table),
	}

	envelope.Headers["x-fr-testheader"] = "HelloWorldHeader"

	return &fr.Letter{
		LetterID:   uuid.New(),
		RetryCount: uint32(0),
		Body:       body,
		Envelope:   envelope,
	}
}

// CreateMockRandomWrappedBodyLetter creates a mock fr.Letter for publishing with random sizes and random Ids.
func CreateMockRandomWrappedBodyLetter(routingKey string) *fr.Letter {

	body := RandomBytes(mockRandom.Intn(randomMax-randomMin) + randomMin)

	envelope := &fr.Envelope{
		Ctx:          context.Background(),
		Exchange:     "",
		RoutingKey:   routingKey,
		ContentType:  "application/json",
		DeliveryMode: 2,
	}

	wrappedBody := &fr.WrappedBody{
		LetterID: uuid.New(),
		Body: &fr.ModdedBody{
			Encrypted:   false,
			Compressed:  false,
			UTCDateTime: utils_json.UtcTimestamp(),
			Data:        body,
		},
	}

	data, _ := json.Marshal(wrappedBody)

	letter := &fr.Letter{
		LetterID:   wrappedBody.LetterID,
		RetryCount: uint32(0),
		Body:       data,
		Envelope:   envelope,
	}

	return letter
}
