package mock

import (
	"math/rand"
	"time"

	fr "github.com/pmatteo/friendlyrabbit"
)

var mockRandomSource = rand.NewSource(time.Now().UnixNano())
var mockRandom = rand.New(mockRandomSource)

// CreateMockLetter creates a mock letter for publishing.
func CreateMockLetter(exchangeName, routingKey string, body []byte) *fr.Letter {

	if body == nil { //   h   e   l   l   o       w   o   r   l   d
		body = []byte("\x68\x65\x6c\x6c\x6f\x20\x77\x6f\x72\x6c\x64")
	}

	l, err := fr.NewLetter(exchangeName, routingKey, body)
	if err != nil {
		panic(err)
	}

	return l
}

// CreateMockRandomLetter creates a mock letter for publishing with random sizes and random Ids.
func CreateMockRandomLetter(routingKey string, optsFuns ...fr.LetterOptsFun) *fr.Letter {

	body := RandomBytes(mockRandom.Intn(randomMax-randomMin) + randomMin)
	h := map[string]interface{}{
		"x-testheader": "HelloWorldHeader",
	}

	optsFuns = append(optsFuns, fr.WithHeaders(h))

	l, err := fr.NewLetter("", routingKey, body, optsFuns...)
	if err != nil {
		panic(err)
	}
	return l
}

type BasicStruct struct {
	N int
	S string
}
