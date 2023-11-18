package benchmark_tests

import (
	"testing"
	"time"

	cmap "github.com/orcaman/concurrent-map"

	fr "github.com/pmatteo/friendlyrabbit"
)

func BenchmarkPublishForDuration(b *testing.B) {
	purgeQueue(b)

	timeDuration := time.Minute * 2
	if testing.Short() {
		timeDuration = time.Second * 25
	}

	b.Logf("Benchmark Starts: %s\r\n", time.Now())
	b.Logf("Est. Benchmark End: %s\r\n", time.Now().Add(timeDuration))

	b.ReportAllocs()

	publisher := fr.NewPublisherFromConfig(seasoning, connectionPool)
	defer publisher.Shutdown(false)

	publishDone := make(chan bool, 1)
	conMap := cmap.New()
	publishLoop(b, conMap, publishDone, timeDuration, publisher)
	<-publishDone
}
