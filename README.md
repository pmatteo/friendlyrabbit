# friendlyrabbit

A user friendly RabbitMQ library written in Golang. Fork of [turbocookedrabbit](https://github.com/houseofcat/turbocookedrabbit).

[![Go Report Card](https://goreportcard.com/report/github.com/pmatteo/friendlyrabbit)](https://goreportcard.com/report/github.com/pmatteo/friendlyrabbit)

[![Release](https://img.shields.io/github/release/pmatteo/friendlyrabbit.svg?style=flat-square)](https://github.com/pmatteo/friendlyrabbit/releases)

## Installation

``` terminal
go get github.com/pmatteo/friendlyrabbit
```

### Getting Started

#### RabbitMQ Connection

`friendlyrabbit` expose a convinient struct called `ConnectionPool` which - as the name suggest - offers an easy way to established a pool of connections.

Configuration for `ConnectionPool` can be loaded from a json/yml file or manually defining a new `PoolConfig`.

```go
poolConfig := &PoolConfig{}
 err := ConvertJsonToStruct("pool_config.json", config)

connectionPool, err := friendlyrabbit.NewConnectionPool(poolConfig)
```

#### Create a simple Publisher

```go
publisher := friendlyrabbit.NewPublisher(config, connectionPool)

// Create the letter's Envelope containing all the address details of where a letter is going.
env := fr.NewEnvelope(context.Background(), exchangeName, routingKey, amqpTableHeader)

// Create and send a simple message
letter := friendlyrabbit.NewLetter(
    uuid.New(),
    env,
    []bytes("message"),
)
err := publisher.Publish(letter)
```

It's also possible to publish a message and wait for its publishing confirmation. This is obviously a bit slower operation so be careful.

```go
publisher.PublishWithConfirmation(letter, time.Millisecond*500)

receipt := <-publisher.PublishReceipts():
if !receipt.Success {
// log? requeue? break WaitLoop?
}
```

Checkout the wiki for more information about the Publisher.

#### Create a simple Consumer

```go
publisher := friendlyrabbit.NewPublisher(config, connectionPool)

// Create the letter's Envelope containing all the address details of where a letter is going.
env := fr.NewEnvelope(context.Background(), exchangeName, routingKey, amqpTableHeader)

// Create and send a simple message
letter := friendlyrabbit.NewLetter(
    uuid.New(),
    env,
    []bytes("message"),
)
err := publisher.Publish(letter)
```

It's also possible to publish a message and wait for its publishing confirmation. This is obviously a bit slower operation so be careful.

```go
publisher.PublishWithConfirmation(letter, time.Millisecond*500)

receipt := <-publisher.PublishReceipts():
if !receipt.Success {
// log? requeue? break WaitLoop?
}
```

Checkout the wiki for more information about the Publisher.
