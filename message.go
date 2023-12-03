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

type LetterOpts struct {
	Ctx context.Context

	Exchange      string
	RoutingKey    string
	ContentType   string
	CorrelationID string
	MessageType   string
	Mandatory     bool
	Immediate     bool
	Headers       map[string]interface{}
	DeliveryMode  uint8
	Priority      uint8
	RetryCount    uint16

	EConf *EncryptionConfig
	CConf *CompressionConfig
}

type LetterOptsFun func(*LetterOpts)

func defaultOpts(e, rk string) *LetterOpts {
	return &LetterOpts{
		Ctx:           context.Background(),
		Exchange:      e,
		RoutingKey:    rk,
		ContentType:   "application/json",
		CorrelationID: "",
		MessageType:   "",
		Mandatory:     false,
		Immediate:     false,
		Headers:       map[string]interface{}{},
		DeliveryMode:  amqp.Persistent,
		Priority:      0,
		RetryCount:    3,
		EConf:         nil,
		CConf:         nil,
	}
}

func WithContext(ctx context.Context) LetterOptsFun {
	return func(o *LetterOpts) {
		o.Ctx = ctx
	}
}

func WithCorrelationID(correlationID string) LetterOptsFun {
	return func(o *LetterOpts) {
		o.CorrelationID = correlationID
	}
}

func WithHeaders(h map[string]interface{}) LetterOptsFun {
	return func(o *LetterOpts) {
		for k, v := range h {
			o.Headers[k] = v
		}
	}
}

func WithEncryptionConfig(eConf *EncryptionConfig) LetterOptsFun {
	return func(o *LetterOpts) {
		o.EConf = eConf
	}
}

func WithCompressionConfig(cConf *CompressionConfig) LetterOptsFun {
	return func(o *LetterOpts) {
		o.CConf = cConf
	}

// Letter contains the message body and address of where things are going.
type Letter struct {
	LetterID uuid.UUID
	Body     []byte
	opts     *LetterOpts
}

func NewLetter(
	exchange string,
	routingKey string,
	data interface{},
	optsFuns ...LetterOptsFun,
) (*Letter, error) {
	opts := defaultOpts(exchange, routingKey)
	for _, f := range optsFuns {
		f(opts)
	}

	var body []byte
	var err error

	body, err = ToPayload(data, opts.CConf, opts.EConf)

	if err != nil {
		return nil, err
	}

	return &Letter{LetterID: uuid.New(), Body: body, opts: opts}, nil
}

func (l *Letter) Done() <-chan struct{} {
	return l.opts.Ctx.Done()
}

func (l *Letter) Options() LetterOpts {
	return *l.opts
}
