package friendlyrabbit

import (
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	CONFIRMATION_TIMEOUT = "publish confirmation timeout after %d millisecond - recommend retry/requeue (LetterID %s)"
	CONFIRMATION_CANCEL  = "publish confirmation not received before context expired - recommend retry/requeue (LetterID %s)"
)

// Publisher contains everything you need to publish a message.
type Publisher struct {
	Config           *RabbitSeasoning
	ConnectionPool   *ConnectionPool
	letters          chan *Letter
	autoStop         chan bool
	publishReceipts  chan *PublishReceipt
	autoStarted      bool
	autoPublishGroup *sync.WaitGroup
	sleepOnIdle      time.Duration
	sleepOnError     time.Duration
	PublishTimeout   time.Duration
	pubLock          *sync.Mutex
	pubRWLock        *sync.RWMutex
}

// NewPublisherFromConfig creates and configures a new Publisher.
func NewPublisherFromConfig(conf *RabbitSeasoning, cp *ConnectionPool) *Publisher {

	if conf.PublisherConfig.MaxRetryCount == 0 {
		conf.PublisherConfig.MaxRetryCount = 5
	}

	return &Publisher{
		Config:           conf,
		ConnectionPool:   cp,
		letters:          make(chan *Letter, 1000),
		autoStop:         make(chan bool, 1),
		autoPublishGroup: &sync.WaitGroup{},
		publishReceipts:  make(chan *PublishReceipt, 1000),
		sleepOnIdle:      time.Duration(conf.PublisherConfig.SleepOnIdleInterval) * time.Millisecond,
		sleepOnError:     time.Duration(conf.PublisherConfig.SleepOnErrorInterval) * time.Millisecond,
		PublishTimeout:   time.Duration(conf.PublisherConfig.PublishTimeOutInterval) * time.Millisecond,
		pubLock:          &sync.Mutex{},
		pubRWLock:        &sync.RWMutex{},
		autoStarted:      false,
	}
}

// Publish sends a single message to the address on the letter using a cached ChannelHost.
// Subscribe to PublishReceipts to see success and errors.
//
// For proper resilience (at least once delivery guarantee over shaky network) use PublishWithConfirmation
func (pub *Publisher) Publish(letter *Letter, skipReceipt bool) {

	chanHost := pub.ConnectionPool.GetChannelFromPool()

	err := chanHost.Channel.PublishWithContext(
		letter.Envelope.Ctx,
		letter.Envelope.Exchange,
		letter.Envelope.RoutingKey,
		letter.Envelope.Mandatory,
		letter.Envelope.Immediate,
		amqp.Publishing{
			ContentType:   letter.Envelope.ContentType,
			Body:          letter.Body,
			Headers:       letter.Envelope.Headers,
			DeliveryMode:  letter.Envelope.DeliveryMode,
			Priority:      letter.Envelope.Priority,
			MessageId:     letter.LetterID.String(),
			CorrelationId: letter.Envelope.CorrelationID,
			Type:          letter.Envelope.Type,
			Timestamp:     time.Now().UTC(),
			AppId:         pub.ConnectionPool.Config.ApplicationName,
		},
	)

	if !skipReceipt {
		pub.publishReceipt(letter, err)

	}

	pub.ConnectionPool.ReturnChannel(chanHost, err != nil)
}

// PublishWithError sends a single message to the address on the letter using a cached ChannelHost.
//
// For proper resilience (at least once delivery guarantee over shaky network) use PublishWithConfirmation
func (pub *Publisher) PublishWithError(letter *Letter, skipReceipt bool) error {

	chanHost := pub.ConnectionPool.GetChannelFromPool()

	err := chanHost.Channel.PublishWithContext(
		letter.Envelope.Ctx,
		letter.Envelope.Exchange,
		letter.Envelope.RoutingKey,
		letter.Envelope.Mandatory,
		letter.Envelope.Immediate,
		amqp.Publishing{
			ContentType:   letter.Envelope.ContentType,
			Body:          letter.Body,
			Headers:       letter.Envelope.Headers,
			DeliveryMode:  letter.Envelope.DeliveryMode,
			Priority:      letter.Envelope.Priority,
			MessageId:     letter.LetterID.String(),
			CorrelationId: letter.Envelope.CorrelationID,
			Type:          letter.Envelope.Type,
			Timestamp:     time.Now().UTC(),
			AppId:         pub.ConnectionPool.Config.ApplicationName,
		},
	)

	if !skipReceipt {
		pub.publishReceipt(letter, err)
	}

	pub.ConnectionPool.ReturnChannel(chanHost, err != nil)
	return err
}

// PublishWithTransient sends a single message to the address on the letter using a transient (new) RabbitMQ channel.
// Subscribe to PublishReceipts to see success and errors.
// For proper resilience (at least once delivery guarantee over shaky network) use PublishWithConfirmation
func (pub *Publisher) PublishWithTransient(letter *Letter) error {

	channel := pub.ConnectionPool.GetTransientChannel(false)
	defer func() {
		defer func() {
			_ = recover()
		}()
		channel.Close()
	}()

	return channel.PublishWithContext(
		letter.Envelope.Ctx,
		letter.Envelope.Exchange,
		letter.Envelope.RoutingKey,
		letter.Envelope.Mandatory,
		letter.Envelope.Immediate,
		amqp.Publishing{
			ContentType:   letter.Envelope.ContentType,
			Body:          letter.Body,
			Headers:       letter.Envelope.Headers,
			DeliveryMode:  letter.Envelope.DeliveryMode,
			Priority:      letter.Envelope.Priority,
			MessageId:     letter.LetterID.String(),
			CorrelationId: letter.Envelope.CorrelationID,
			Type:          letter.Envelope.Type,
			Timestamp:     time.Now().UTC(),
			AppId:         pub.ConnectionPool.Config.ApplicationName,
		},
	)
}

// PublishWithConfirmation sends a single message to the address on the
// letter with confirmation capabilities.
// This is an expensive and slow call, use this when delivery confirmation on
// publish is your highest priority.
//
// A timeout failure drops the letter back in the PublishReceipts. When combined
// with QueueLetter, it automatically gets requeued for re-publish.
// A confirmation failure keeps trying to publish until a timeout failure
// occurs or context got canceled.
// If a nack occurs, it will republish the letter.
// It continuously tries to publish the letter until it receives a confirmation or encounters an error.
//
// The function returns an error if it fails to publish the letter within the specified timeout.
func (pub *Publisher) PublishWithConfirmation(letter *Letter, timeout time.Duration) error {

	for {
	Publish:
		chanHost := pub.ConnectionPool.GetChannelFromPool()
		dConfirmation, err := pub.publishDeferredConfirm(chanHost.Channel, letter)
		pub.ConnectionPool.ReturnChannel(chanHost, err != nil)

		if err != nil {
			if pub.sleepOnError > 0 {
				time.Sleep(pub.sleepOnError)
			}
			continue
		}

		// Wait for very next confirmation on this channel, which should be our confirmation.
		acked, err := waitPublishConfirmation(dConfirmation, letter, timeout)
		if !acked && err == nil {
			goto Publish //nack has occurred, republish
		}

		pub.publishReceipt(letter, err)
		return err
	}
}

// PublishWithConfirmationTransient sends a single message to the address on
// the letter with confirmation capabilities on transient Channels.
// This is an expensive and slow call - use this when delivery confirmation
// on publish is your highest priority.
//
// A timeout failure drops the letter back in the PublishReceipts. When
// combined with QueueLetter, it automatically gets requeued for re-publish.
// A confirmation failure keeps trying to publish until a timeout failure
// occurs or context got canceled.
// If a nack occurs, it will republish the letter.
// It continuously tries to publish the letter until it receives a confirmation or encounters an error.
//
// The function returns an error if it fails to publish the letter within the specified timeout.
func (pub *Publisher) PublishWithConfirmationTransient(letter *Letter, timeout time.Duration) error {

	for {
	Publish:
		// Has to use an Ackable channel for Publish Confirmations.
		channel := pub.ConnectionPool.GetTransientChannel(true)
		dConfirmation, err := pub.publishDeferredConfirm(channel, letter)
		channel.Close()

		if err != nil {
			if pub.sleepOnError > 0 {
				time.Sleep(pub.sleepOnError)
			}
			continue
		}

		// Wait for very next confirmation on this channel, which should be our confirmation.
		acked, err := waitPublishConfirmation(dConfirmation, letter, timeout)
		if !acked && err == nil {
			goto Publish //nack has occurred, republish
		}

		pub.publishReceipt(letter, err)
		return err
	}
}

// publishDeferredConfirm publishes a letter with deferred confirmation.
//
// It sends the letter to the specified exchange and routing key using the provided channel.
// The letter's envelope properties are used to set the message properties.
//
// Returns the deferred confirmation and any error encountered during publishing.
func (pub *Publisher) publishDeferredConfirm(c *amqp.Channel, letter *Letter) (*amqp.DeferredConfirmation, error) {
	dConfirmation, err := c.PublishWithDeferredConfirmWithContext(
		letter.Envelope.Ctx,
		letter.Envelope.Exchange,
		letter.Envelope.RoutingKey,
		letter.Envelope.Mandatory,
		letter.Envelope.Immediate,
		amqp.Publishing{
			ContentType:   letter.Envelope.ContentType,
			Body:          letter.Body,
			Headers:       letter.Envelope.Headers,
			DeliveryMode:  letter.Envelope.DeliveryMode,
			Priority:      letter.Envelope.Priority,
			MessageId:     letter.LetterID.String(),
			CorrelationId: letter.Envelope.CorrelationID,
			Type:          letter.Envelope.Type,
			Timestamp:     time.Now().UTC(),
			AppId:         pub.ConnectionPool.Config.ApplicationName,
		},
	)

	return dConfirmation, err
}

// waitPublishConfirmation waits for the confirmation of a published letter on the given channel.
//
// It returns true if the confirmation is received within the specified timeout, otherwise it returns false.
// If the context of the letter's envelope is canceled, it returns an error with the cancellation reason.
// If the confirmation times out, it returns an error with the timeout duration and the letter ID.
// If the confirmation is received, it returns the acknowledgment status and no error.
func waitPublishConfirmation(dConf *amqp.DeferredConfirmation, letter *Letter, timeout time.Duration) (bool, error) {
	// Wait for very next confirmation on this channel, which should be our confirmation.
	for {
		select {
		case <-letter.Envelope.Ctx.Done():
			err := fmt.Errorf(CONFIRMATION_CANCEL, letter.LetterID.String())
			return false, err

		case <-time.After(timeout):
			err := fmt.Errorf(CONFIRMATION_TIMEOUT, timeout.Milliseconds(), letter.LetterID.String())
			return false, err

		case <-dConf.Done():
			return dConf.Acked(), nil
		}
	}
}

// PublishReceipts yields all the success and failures during all publish events. Highly recommend susbscribing to this.
func (pub *Publisher) PublishReceipts() <-chan *PublishReceipt {
	return pub.publishReceipts
}

// StartAutoPublishing starts the Publisher's auto-publishing capabilities.
func (pub *Publisher) StartAutoPublishing() {
	pub.pubLock.Lock()
	defer pub.pubLock.Unlock()

	if !pub.autoStarted {
		pub.autoStarted = true
		go pub.startAutoPublishingLoop()
	}
}

// StartAutoPublish starts auto-publishing letters queued up - is locking.
func (pub *Publisher) startAutoPublishingLoop() {

AutoPublishLoop:
	for {

	SelectLoop:
		select {
		// Detect if we should stop publishing.
		case stop := <-pub.autoStop:
			if stop {
				break AutoPublishLoop
			}
		default:
			break SelectLoop
		}

		// Deliver letters queued in the publisher, returns true when we are to stop publishing.
		if pub.deliverLetters() {
			break AutoPublishLoop
		}
	}

	pub.pubLock.Lock()
	pub.autoStarted = false
	pub.pubLock.Unlock()
}

func (pub *Publisher) deliverLetters() bool {

	// Allow parallel publishing with transient channels.
	parallelPublishSemaphore := make(chan struct{}, pub.ConnectionPool.Config.MaxCacheChannelCount/2+1)

	for {

		// Publish the letter.
	PublishLoop:
		for {
			select {
			case letter := <-pub.letters:

				parallelPublishSemaphore <- struct{}{}
				go func(letter *Letter) {
					_ = pub.PublishWithConfirmation(letter, pub.PublishTimeout)
					<-parallelPublishSemaphore
				}(letter)

			default:

				if pub.sleepOnIdle > 0 {
					time.Sleep(pub.sleepOnIdle)
				}
				break PublishLoop

			}
		}

	SelectStop:
		select {
		// Detect if we should stop publishing.
		case stop := <-pub.autoStop:
			if stop {
				close(pub.letters)
				return true
			}
		default:
			break SelectStop
		}
	}
}

// stopAutoPublish stops publishing letters queued up.
func (pub *Publisher) stopAutoPublish() {
	pub.pubLock.Lock()
	defer pub.pubLock.Unlock()

	if !pub.autoStarted {
		return
	}

	go func() { pub.autoStop <- true }() // signal auto publish to stop
}

// QueueLetters allows you to bulk queue letters that will be consumed by AutoPublish. By default, AutoPublish uses PublishWithConfirmation as the mechanism for publishing.
func (pub *Publisher) QueueLetters(letters []*Letter) bool {

	for _, letter := range letters {

		if ok := pub.safeSend(letter); !ok {
			return false
		}
	}

	return true
}

// QueueLetter queues up a letter that will be consumed by AutoPublish. By default, AutoPublish uses PublishWithConfirmation as the mechanism for publishing.
func (pub *Publisher) QueueLetter(letter *Letter) bool {

	return pub.safeSend(letter)
}

// safeSend should handle a scenario on publishing to a closed channel.
func (pub *Publisher) safeSend(letter *Letter) (closed bool) {
	defer func() {
		if recover() != nil {
			closed = false
		}
	}()

	pub.letters <- letter
	return true // success
}

// publishReceipt sends the status to the receipt channel.
func (pub *Publisher) publishReceipt(letter *Letter, err error) {

	go func(*Letter, error) {
		publishReceipt := &PublishReceipt{
			LetterID: letter.LetterID,
			Error:    err,
		}

		if err == nil {
			publishReceipt.Success = true
		} else {
			publishReceipt.FailedLetter = letter
		}

		pub.publishReceipts <- publishReceipt
	}(letter, err)
}

// Shutdown cleanly shutdown the publisher and resets it's internal state.
func (pub *Publisher) Shutdown(shutdownPools bool) {

	pub.stopAutoPublish()

	if shutdownPools { // in case the ChannelPool is shared between structs, you can prevent it from shutting down
		pub.ConnectionPool.Shutdown()
	}
}
