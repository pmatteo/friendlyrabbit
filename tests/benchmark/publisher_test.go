package benchmark_tests

import (
	"testing"
	"time"

	cmap "github.com/orcaman/concurrent-map"

	fr "github.com/pmatteo/friendlyrabbit"
	"github.com/pmatteo/friendlyrabbit/tests/mock"
)

func BenchmarkPublishWithConfirmationForDuration(b *testing.B) {
	purgeQueue(b)

	timeDuration := time.Minute * 5
	if testing.Short() {
		timeDuration = time.Minute * 1
	}

	b.Logf("Benchmark Starts: %s\r\n", time.Now())
	b.Logf("Est. Benchmark End: %s\r\n", time.Now().Add(timeDuration))

	b.ReportAllocs()

	publisher := fr.NewPublisher(seasoning, connectionPool)
	defer publisher.Shutdown(false)

	publishDone := make(chan bool, 1)
	conMap := cmap.New()
	publishWithConfirmation(b, conMap, publishDone, timeDuration, publisher)
	<-publishDone
}

func BenchmarkPublishForDuration(b *testing.B) {
	purgeQueue(b)

	timeDuration := time.Minute * 5
	if testing.Short() {
		timeDuration = time.Minute * 1
	}

	b.Logf("Benchmark Starts: %s\r\n", time.Now())
	b.Logf("Est. Benchmark End: %s\r\n", time.Now().Add(timeDuration))

	b.ReportAllocs()

	publisher := fr.NewPublisher(seasoning, connectionPool)
	defer publisher.Shutdown(false)

	publishDone := make(chan bool, 1)
	conMap := cmap.New()
	publishLoop(b, conMap, publishDone, timeDuration, publisher)
	<-publishDone
}

func publishLoop(
	b *testing.B,
	conMap cmap.ConcurrentMap,
	done chan bool,
	timeoutDuration time.Duration,
	publisher *fr.Publisher,
) {
	messagesPublished := 0
	messagesFailedToPublish := 0
	publisherErrors := 0

	timeout := time.After(timeoutDuration)

PublishLoop:
	for {
		select {
		case <-timeout:
			break PublishLoop
		default:
			newLetter := mock.CreateMockRandomLetter("TestBenchmarkQueue")
			conMap.Set(newLetter.LetterID.String(), false)
			_ = publisher.Publish(newLetter, true)

			notice := <-publisher.PublishReceipts()
			if notice.Success {
				messagesPublished++
				notice = nil
			} else {
				messagesFailedToPublish++
				notice = nil
			}
		}
	}

	b.Logf("Publisher Errors: %d\r\n", publisherErrors)
	b.Logf("Messages Published: %d\r\n", messagesPublished)
	b.Logf("Messages Failed to Publish: %d\r\n", messagesFailedToPublish)

	done <- true
}
