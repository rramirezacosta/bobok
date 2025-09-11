package bobok

import (
	"errors"
	"fmt"
	"sync"
)

var bobokSingleton *broadcaster
var once sync.Once

type topic struct {
	mu          sync.Mutex
	subscribers []chan any
}

type broadcaster struct {
	topicsByLabel sync.Map // map[string]*topic
}

func Subscribe(label string) (read <-chan any, unsubscribe func(), err error) {
	once.Do(func() {
		// initialize singleton
		bobokSingleton = &broadcaster{
			topicsByLabel: sync.Map{},
		}
	})

	if bobokSingleton == nil {
		err = errors.New("bobok not initialized")
		return
	}

	ch := make(chan any, 10)

	if err = bobokSingleton.appendToTopic(label, ch); err != nil {
		return
	}

	read = ch
	unsubscribe = func() {
		bobokSingleton.removeFromTopic(label, ch)
	}

	return
}

func Publish(label string, message any) error {
	once.Do(func() {
		// initialize singleton
		bobokSingleton = &broadcaster{
			topicsByLabel: sync.Map{},
		}
	})

	if bobokSingleton == nil {
		return errors.New("bobok not initialized")
	}

	t, err := bobokSingleton.getTopic(label)
	if err != nil {
		return fmt.Errorf("failed to publish to label %s: %w", label, err)
	}

	t.mu.Lock()
	for _, ch := range t.subscribers {
		if ch != nil {
			select {
			case ch <- message:
			default:
				// Channel full, skip this subscribers
			}
		}
	}
	t.mu.Unlock()

	return nil
}

func (b *broadcaster) getTopic(label string) (*topic, error) {
	if v, exists := b.topicsByLabel.Load(label); !exists {
		t := &topic{
			mu:          sync.Mutex{},
			subscribers: make([]chan any, 0),
		}
		b.topicsByLabel.Store(label, t)
		return t, nil
	} else {
		if t, ok := v.(*topic); ok {
			return t, nil
		} else {
			return nil, errors.New("invalid topic type")
		}
	}
}

func (b *broadcaster) removeFromTopic(label string, ch chan any) error {
	t, err := b.getTopic(label)
	if err != nil {
		return fmt.Errorf("failed to remove channel from label %s: %w", label, err)
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	for i, c := range t.subscribers {
		if c == ch {
			t.subscribers = append(t.subscribers[:i], t.subscribers[i+1:]...)
			close(c)
			break
		}
	}

	if len(t.subscribers) == 0 {
		b.topicsByLabel.Delete(label)
	}

	return nil
}

func (b *broadcaster) appendToTopic(label string, ch chan any) error {
	t, err := b.getTopic(label)
	if err != nil {
		return fmt.Errorf("failed to append channel to label %s: %w", label, err)
	}

	t.mu.Lock()
	t.subscribers = append(t.subscribers, ch)
	t.mu.Unlock()

	return nil
}
