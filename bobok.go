package bobok

import (
	"errors"
	"fmt"
	"sync"
)

var bobokSingleton *broadcaster
var once sync.Once

type topic struct {
	mu           sync.Mutex
	cond         *sync.Cond
	messageQueue []any
	subscribers  []chan any
	stopCh       chan struct{}
	started      bool
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
	doneCh := make(chan bool, 1)

	if err = bobokSingleton.appendToTopic(label, ch); err != nil {
		return
	}

	read = ch
	unsubscribe = func() {
		bobokSingleton.removeFromTopic(label, ch)
		if doneCh != nil {
			close(doneCh)
			doneCh = nil
		}
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
	t.messageQueue = append(t.messageQueue, message)
	t.mu.Unlock()
	
	t.cond.Broadcast()

	return nil
}

func (b *broadcaster) getTopic(label string) (*topic, error) {
	if v, exists := b.topicsByLabel.Load(label); !exists {
		t := &topic{
			messageQueue: make([]any, 0),
			subscribers:  make([]chan any, 0),
			stopCh:       make(chan struct{}),
		}
		t.cond = sync.NewCond(&t.mu)
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

func (t *topic) startDistributor() {
	if t.started {
		return
	}
	t.started = true
	
	go func() {
		for {
			t.cond.L.Lock()
			for len(t.messageQueue) == 0 {
				select {
				case <-t.stopCh:
					t.cond.L.Unlock()
					return
				default:
					t.cond.Wait()
				}
			}
			
			// Copy messages and subscribers under lock
			messages := make([]any, len(t.messageQueue))
			copy(messages, t.messageQueue)
			t.messageQueue = t.messageQueue[:0] // Clear the queue
			
			subscribers := make([]chan any, len(t.subscribers))
			copy(subscribers, t.subscribers)
			t.cond.L.Unlock()
			
			// Distribute all messages to all subscribers outside of lock
			for _, msg := range messages {
				for _, ch := range subscribers {
					if ch != nil {
						select {
						case ch <- msg:
						default:
							// Channel full, skip this subscriber
						}
					}
				}
			}
		}
	}()
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
		close(t.stopCh)
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
	t.startDistributor()
	t.mu.Unlock()

	return nil
}
