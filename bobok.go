package bobok

import (
	"errors"
	"fmt"
	"sync"
)

var bobokSingleton *broadcaster
var once sync.Once

type broadcaster struct {
	pipesByLabel sync.Map // map[string][]chan any
}

func Subscribe(label string) (read <-chan any, unsubscribe func(), err error) {
	once.Do(func() {
		// initialize singleton
		bobokSingleton = &broadcaster{
			pipesByLabel: sync.Map{},
		}
	})

	if bobokSingleton == nil {
		err = errors.New("bobok not initialized")
		return
	}

	ch := make(chan any, 10)
	doneCh := make(chan bool, 1)

	if err = bobokSingleton.appendToChannels(label, ch); err != nil {
		return
	}

	read = ch
	unsubscribe = func() {
		bobokSingleton.removeFromChannels(label, ch)
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
			pipesByLabel: sync.Map{},
		}
	})

	if bobokSingleton == nil {
		return errors.New("bobok not initialized")
	}

	channels, err := bobokSingleton.getChannels(label)

	if err != nil {
		return fmt.Errorf("failed to publsh to label %s: %w", label, err)
	}

	for _, ch := range channels {
		if ch == nil {
			continue
		}
		ch <- message
	}

	return nil
}

func (b *broadcaster) getChannels(label string) (pipe []chan any, err error) {
	if v, exists := b.pipesByLabel.Load(label); !exists {
		pipe = make([]chan any, 0)
		b.pipesByLabel.Store(label, pipe)
		return pipe, nil
	} else {
		var isOk bool
		pipe, isOk = v.([]chan any)
		if !isOk {
			err = errors.New("invalid pipe type")
		}
	}
	return
}

func (b *broadcaster) removeFromChannels(label string, ch chan any) error {
	channels, err := b.getChannels(label)
	if err != nil {
		return fmt.Errorf("failed to remove channel from label %s: %w", label, err)
	}

	for i, c := range channels {
		if c == ch {
			channels = append(channels[:i], channels[i+1:]...)
			close(c)
			break
		}
	}

	if len(channels) == 0 {
		b.pipesByLabel.Delete(label)
	} else {
		b.pipesByLabel.Store(label, channels)
	}

	return nil
}

func (b *broadcaster) appendToChannels(label string, ch chan any) error {
	channels, err := b.getChannels(label)
	if err != nil {
		return fmt.Errorf("failed to append channel to label %s: %w", label, err)
	}

	channels = append(channels, ch)
	b.pipesByLabel.Store(label, channels)

	return nil
}
