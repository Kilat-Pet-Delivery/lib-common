package kafka

import (
	"fmt"
	"sync"
	"testing"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// TestGetWriterConcurrent is a regression test for KPD-66. Before the mutex,
// getWriter read and wrote p.writers with no synchronisation, and Publish is
// reached from HTTP handlers -- so concurrent publishes to different topics
// raced on the map. In Go that is not a benign race: the runtime throws
// "fatal error: concurrent map writes" and the process dies.
//
// Run with -race to see the old code fail. Building a kafka.Writer does not
// dial the broker, so this needs no Kafka.
func TestGetWriterConcurrent(t *testing.T) {
	p := NewProducer([]string{"localhost:9092"}, zap.NewNop())

	// Every goroutine gets its own topic, so nearly all of them take the miss
	// path and write to the map at the same time. That is what makes the old
	// unguarded version fail reliably rather than occasionally: with only a
	// handful of misses the write window is too small for the runtime to catch.
	const goroutines = 512

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			<-start // release them together, to widen the overlap
			if w := p.getWriter(fmt.Sprintf("topic.%d", i)); w == nil {
				t.Errorf("getWriter returned nil")
			}
		}(i)
	}
	close(start)
	wg.Wait()

	if got := len(p.writers); got != goroutines {
		t.Errorf("expected %d writers, got %d -- a lost write means the map was mutated concurrently", goroutines, got)
	}
}

// TestGetWriterConcurrentSameTopic pins the double-check on the miss path: many
// goroutines racing for one topic must still end up sharing a single writer.
func TestGetWriterConcurrentSameTopic(t *testing.T) {
	p := NewProducer([]string{"localhost:9092"}, zap.NewNop())

	const goroutines = 256
	start := make(chan struct{})
	seen := make([]*kafka.Writer, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			seen[i] = p.getWriter("booking.events")
		}(i)
	}
	close(start)
	wg.Wait()

	for i, w := range seen {
		if w != seen[0] {
			t.Fatalf("goroutine %d got a different writer -- the miss path leaked a duplicate", i)
		}
	}
	if got := len(p.writers); got != 1 {
		t.Errorf("expected 1 writer, got %d", got)
	}
}

// TestGetWriterReturnsSameWriter pins the caching behaviour: one writer per
// topic, reused across calls.
func TestGetWriterReturnsSameWriter(t *testing.T) {
	p := NewProducer([]string{"localhost:9092"}, zap.NewNop())

	first := p.getWriter("booking.events")
	second := p.getWriter("booking.events")
	if first != second {
		t.Errorf("getWriter returned a different writer for the same topic")
	}
	if other := p.getWriter("payment.events"); other == first {
		t.Errorf("getWriter returned the same writer for two different topics")
	}
}

// TestWriterAllowsAutoTopicCreation is a regression test for KPD-64. Without
// AllowAutoTopicCreation, kafka-go refuses to publish to a topic that does not
// exist yet -- even with auto-creation enabled on the broker -- and every event
// to a never-consumed topic is lost.
func TestWriterAllowsAutoTopicCreation(t *testing.T) {
	p := NewProducer([]string{"localhost:9092"}, zap.NewNop())

	if !p.getWriter("booking.events").AllowAutoTopicCreation {
		t.Errorf("writer must set AllowAutoTopicCreation, or publishing to a topic nobody has consumed fails with Unknown Topic Or Partition")
	}
}
