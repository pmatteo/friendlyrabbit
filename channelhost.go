package friendlyrabbit

import (
	"errors"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ChannelHost is an internal representation of amqp.Channel.
type ChannelHost struct {
	Channel      *amqp.Channel
	ID           uint64
	ConnectionID uint64
	Ackable      bool
	transient    bool
	Errors       chan *amqp.Error
	connHost     *ConnectionHost
	chanLock     *sync.Mutex
}

// NewChannelHost creates a simple ChannelHost wrapper for management by end-user developer.
func NewChannelHost(
	connHost *ConnectionHost,
	id uint64,
	connectionID uint64,
	ackable, transient bool,
) (*ChannelHost, error) {

	if connHost.Connection.IsClosed() {
		return nil, errors.New("can't open a channel - connection is already closed")
	}

	chanHost := &ChannelHost{
		ID:           id,
		ConnectionID: connectionID,
		Ackable:      ackable,
		transient:    transient,
		connHost:     connHost,
		chanLock:     &sync.Mutex{},
	}

	err := chanHost.MakeChannel()
	if err != nil {
		return nil, err
	}

	return chanHost, nil
}

// Close allows for manual close of Amqp Channel kept internally.
func (ch *ChannelHost) Close() {
	ch.Channel.Close()
}

// MakeChannel tries to create (or re-create) the channel from the ConnectionHost its attached to.
func (ch *ChannelHost) MakeChannel() (err error) {
	ch.chanLock.Lock()
	defer ch.chanLock.Unlock()

	ch.Channel, err = ch.connHost.Connection.Channel()
	if err != nil {
		return err
	}

	if ch.Ackable {
		err = ch.Channel.Confirm(false)
		if err != nil {
			return err
		}
	}

	ch.Errors = make(chan *amqp.Error, 100)
	ch.Channel.NotifyClose(ch.Errors)

	return nil
}

// PauseForFlowControl allows you to wait and sleep while receiving flow control messages.
func (ch *ChannelHost) PauseForFlowControl() {

	ch.connHost.PauseOnFlowControl()
}
