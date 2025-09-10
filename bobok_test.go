package bobok

import (
	"sync"
	"testing"
	"time"
)

func TestPubSub(t *testing.T) {

	type message struct {
		id  int
		msg string
	}

	b := NewBroadcaster()

	messagesA := []message{
		{1, "Hello"},
		{2, "World"},
		{3, "from"},
		{4, "Bobok"},
		{5, "!"},
	}

	messagesB := []message{
		{1, "Goodbye"},
		{2, "Cruel"},
		{3, "World"},
		{4, "from"},
		{5, "Bobok"},
		{6, "!"},
	}

	messagesC := []message{
		{1, "ShouldntBeReceived"},
		{2, "ShouldntBeReceived"},
		{3, "ShouldntBeReceived"},
		{4, "ShouldntBeReceived"},
		{5, "ShouldntBeReceived"},
		{6, "ShouldntBeReceived"},
	} // no subscribers for C

	messageCount := len(messagesA) + len(messagesB)

	wg := sync.WaitGroup{}
	wg.Add((messageCount) * 2) // each message is listened twice

	// setup listener
	worker := func(label string, expectedMsgs []message) {
		ch, done, cleanup, err := b.Subscribe(label)
		if err != nil {
			t.Fatalf("Failed to subscribe to label %s: %v", label, err)
		}
		defer cleanup()

		for i, expected := range expectedMsgs {

			select {
			case <-done:
				t.Fatalf("Listener for label %s: listener closed unexpectedly at index %d", label, i)
			case raw := <-ch:
				received, ok := raw.(message)
				if !ok {
					t.Fatalf("Listener for label %s: expected message of type %T, got %T at index %d", label, expected, raw, i)
				}
				if received.id != expected.id || received.msg != expected.msg {
					t.Fatalf("Listener for label %s: expected message %v at index %d, got %v", label, expected, i, received)
				}
			case <-time.After(5 * time.Second):
				t.Fatalf("Listener for label %s: timeout waiting for message at index %d", label, i)
			}
			time.Sleep(500 * time.Millisecond) // simulate processing time
			wg.Done()
		}
	}

	go worker("A", messagesA)
	go worker("A", messagesA)
	go worker("B", messagesB)
	go worker("B", messagesB)

	time.Sleep(200 * time.Millisecond) // Give listeners time to subscribe

	// publish messages
	go func() {
		for i := 0; ; i++ {
			stillValuesInA, stillValuesInB, stillValuesInC := i < len(messagesA), i < len(messagesB), i < len(messagesC)
			if stillValuesInA {
				b.Publish("A", messagesA[i])
			}
			if stillValuesInB {
				b.Publish("B", messagesB[i])
			}
			if stillValuesInC {
				b.Publish("C", messagesC[i])
			}
			if !stillValuesInA && !stillValuesInB && !stillValuesInC {
				break
			}
		}
	}()

	wg.Wait()
	time.Sleep(1 * time.Second) // wait a bit to ensure no more messages are processed
}
