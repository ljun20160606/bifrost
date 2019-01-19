package bridge

import (
	"github.com/pkg/errors"
	"time"
)

// 包含超时的channel
type Channel struct {
	ch      chan interface{}
	timeout time.Duration
}

func NewChannel(timeout time.Duration) *Channel {
	return &Channel{
		ch:      make(chan interface{}),
		timeout: timeout,
	}
}

func (c *Channel) Get() interface{} {
	select {
	case data := <-c.ch:
		return data
	case <-time.After(c.timeout):
		return nil
	}
}

func (c *Channel) Set(data interface{}) error {
	select {
	case c.ch <- data:
		return nil
	default:
		return errors.New("timeout")
	}
}
