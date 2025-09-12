package bobok

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestPubSub(t *testing.T) {

	type message struct {
		id  int
		msg string
	}

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

	// waitgroup to wait for all subscriptions to be done
	subsWg := sync.WaitGroup{}
	subsWg.Add(4) // we will create 4 subscriptions

	// waitgroup to wait for all messages to be received
	wg := sync.WaitGroup{}
	wg.Add((messageCount) * 2) // each message is listened twice

	// setup listener
	worker := func(label string, expectedMsgs []message) {
		ch, unsubscribe, err := Subscribe(label)
		if err != nil {
			t.Fatalf("Failed to subscribe to label %s: %v", label, err)
		}
		defer unsubscribe()

		subsWg.Done() // signal that subscription is done

		for i, expected := range expectedMsgs {

			select {
			case raw, ok := <-ch:
				if !ok {
					t.Fatalf("Listener for label %s: channel closed unexpectedly at index %d", label, i)
				}
				received, ok := raw.(message)
				fmt.Println("Listener for label", label, "received message:", received)
				if !ok {
					t.Fatalf("Listener for label %s: expected message of type %T, got %T at index %d", label, expected, raw, i)
				}
				if received.id != expected.id || received.msg != expected.msg {
					t.Fatalf("Listener for label %s: expected message %v at index %d, got %v", label, expected, i, received)
				}
			case <-time.After(5 * time.Second):
				t.Fatalf("Listener for label %s: timeout waiting for message at index %d", label, i)
			}
			wg.Done()
		}
	}

	go worker("A", messagesA)
	go worker("A", messagesA)
	go worker("B", messagesB)
	go worker("B", messagesB)

	subsWg.Wait() // wait for all subscriptions to be ready

	// publish messages
	go func() {
		for i := 0; ; i++ {
			stillValuesInA, stillValuesInB, stillValuesInC := i < len(messagesA), i < len(messagesB), i < len(messagesC)
			if stillValuesInA {
				Publish("A", messagesA[i])
			}
			if stillValuesInB {
				Publish("B", messagesB[i])
			}
			if stillValuesInC {
				Publish("C", messagesC[i])
			}
			if !stillValuesInA && !stillValuesInB && !stillValuesInC {
				break
			}
		}
	}()

	wg.Wait()
}
