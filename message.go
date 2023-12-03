package friendlyrabbit

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	utils_json "github.com/pmatteo/friendlyrabbit/internal/utils/json"
	amqp "github.com/rabbitmq/amqp091-go"
)

// PublishReceipt is a way to monitor publishing success and to initiate a retry when using async publishing.
type PublishReceipt struct {
	LetterID     uuid.UUID
	FailedLetter *Letter
	Success      bool
	Error        error
}

// ToString allows you to quickly log the PublishReceipt struct as a string.
func (not *PublishReceipt) ToString() string {
	if not.Success {
		return fmt.Sprintf("[LetterID: %s] - Publish successful.\r\n", not.LetterID.String())
	}

	return fmt.Sprintf("[LetterID: %s] - Publish failed.\r\nError: %s\r\n", not.LetterID.String(), not.Error.Error())
}

// ReceivedMessage allow for you to acknowledge, after processing the received payload, by its RabbitMQ tag and Channel pointer.
type ReceivedMessage struct {
	IsAckable     bool
	Body          []byte
	MessageID     string // LetterID
	ApplicationID string
	PublishDate   string
	Delivery      amqp.Delivery // Access everything.
}

// NewReceivedMessage creates a new ReceivedMessage.
func NewReceivedMessage(
	isAckable bool,
	delivery amqp.Delivery) *ReceivedMessage {

	return &ReceivedMessage{
		IsAckable:     isAckable,
		Body:          delivery.Body,
		MessageID:     delivery.MessageId,
		ApplicationID: delivery.AppId,
		PublishDate:   utils_json.UtcTimestampFromTime(delivery.Timestamp),
		Delivery:      delivery,
	}
}

// Acknowledge allows for you to acknowledge message on the original channel it was received.
// Will fail if channel is closed and this is by design per RabbitMQ server.
// Can't ack from a different channel.
func (msg *ReceivedMessage) Acknowledge() error {
	if !msg.IsAckable {
		return errors.New("can't acknowledge, not an ackable message")
	}

	if msg.Delivery.Acknowledger == nil {
		return errors.New("can't acknowledge, internal channel is nil")
	}

	return msg.Delivery.Acknowledger.Ack(msg.Delivery.DeliveryTag, false)
}

// Nack allows for you to negative acknowledge message on the original channel it was received.
// Will fail if channel is closed and this is by design per RabbitMQ server.
func (msg *ReceivedMessage) Nack(requeue bool) error {
	if !msg.IsAckable {
		return errors.New("can't nack, not an ackable message")
	}

	if msg.Delivery.Acknowledger == nil {
		return errors.New("can't nack, internal channel is nil")
	}

	return msg.Delivery.Acknowledger.Nack(msg.Delivery.DeliveryTag, false, requeue)
}

// Reject allows for you to reject on the original channel it was received.
// Will fail if channel is closed and this is by design per RabbitMQ server.
func (msg *ReceivedMessage) Reject(requeue bool) error {
	if !msg.IsAckable {
		return errors.New("can't reject, not an ackable message")
	}

	if msg.Delivery.Acknowledger == nil {
		return errors.New("can't reject, internal channel is nil")
	}

	return msg.Delivery.Acknowledger.Reject(msg.Delivery.DeliveryTag, requeue)
}

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
		body, err := ToWrappedPayload(data, letterID, metadata, compression, encryption)
		if err != nil {
			return nil, err
		}
		return NewLetter(letterID, envelope, body), nil
	}

	body, err := ToPayload(data, compression, encryption)
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
