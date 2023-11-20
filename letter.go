package friendlyrabbit

import (
	"context"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Letter contains the message body and address of where things are going.
type Letter struct {
	LetterID   uuid.UUID
	RetryCount uint32
	Body       []byte
	Envelope   *Envelope
}

// Envelope contains all the address details of where a letter is going.
type Envelope struct {
	Ctx           context.Context
	Exchange      string
	RoutingKey    string
	ContentType   string
	CorrelationID string
	Type          string
	Mandatory     bool
	Immediate     bool
	Headers       amqp.Table
	DeliveryMode  uint8
	Priority      uint8
}

// WrappedBody is to go inside a Letter struct with indications of the body of data being modified (ex., compressed).
type WrappedBody struct {
	LetterID       uuid.UUID   `json:"LetterID"`
	Body           *ModdedBody `json:"Body"`
	LetterMetadata string      `json:"LetterMetadata"`
}

// ModdedBody is a payload with modifications and indicators of what was modified.
type ModdedBody struct {
	Encrypted   bool   `json:"Encrypted"`
	EType       string `json:"EncryptionType,omitempty"`
	Compressed  bool   `json:"Compressed"`
	CType       string `json:"CompressionType,omitempty"`
	UTCDateTime string `json:"UTCDateTime"`
	Data        []byte `json:"Data"`
}

func NewEnvelope(
	ctx context.Context,
	exchangeName string,
	routingKey string,
	headers amqp.Table,
) *Envelope {
	return &Envelope{
		Ctx:          context.Background(),
		Exchange:     exchangeName,
		RoutingKey:   routingKey,
		ContentType:  "application/json",
		Mandatory:    false,
		Immediate:    false,
		DeliveryMode: amqp.Persistent,
		Headers:      headers,
	}
}

func NewLetterWithPayload(
	envelope *Envelope,
	data interface{},
	encryption *EncryptionConfig,
	compression *CompressionConfig,
	wrapPayload bool,
	metadata string,
) (*Letter, error) {
	var letterID = uuid.New()

	if wrapPayload {
		body, err := CreateWrappedPayload(data, letterID, metadata, compression, encryption)
		if err != nil {
			return nil, err
		}
		return NewLetter(letterID, envelope, body), nil
	}

	body, err := CreatePayload(data, compression, encryption)
	if err != nil {
		return nil, err
	}
	return NewLetter(letterID, envelope, body), nil
}

// CreateLetter creates a simple letter for publishing.
func NewLetter(
	letterID uuid.UUID,
	envelope *Envelope,
	body []byte,
) *Letter {
	return &Letter{
		LetterID:   letterID,
		RetryCount: uint32(3),
		Body:       body,
		Envelope:   envelope,
	}
}
