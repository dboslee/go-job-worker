package core

import (
	"io"
	"sync"
)

// MergeStreams sends the combined output of two streams to a given channel
func MergeStreams(ch chan []byte, s1, s2 *Stream) {
	ch1 := make(chan []byte, 64)
	ch2 := make(chan []byte, 64)
	var data []byte
	var ok bool

	s1.Subscribe(ch1)
	s2.Subscribe(ch2)

	// Loop until both channels close
	for ch1 != nil || ch2 != nil {
		select {
		case data, ok = <-ch1:
			if !ok {
				ch1 = nil
			}
		case data, ok = <-ch2:
			if !ok {
				ch2 = nil
			}
		}
		if data != nil {
			ch <- data
		}
	}
	close(ch)
}

// Stream allows for sending one to many channels.
type Stream struct {
	subs   []chan []byte
	mu     sync.RWMutex
	closed bool
}

// NewStream creats a new Stream instance.
func NewStream() *Stream {
	return &Stream{subs: make([]chan []byte, 0)}
}

// Subscribe adds a channel to receive published messages.
func (s *Stream) Subscribe(ch chan []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Prevent subscribing more than once
	for _, sub := range s.subs {
		if sub == ch {
			return
		}
	}

	// Close the channel immediately if the stream is closed
	if s.closed {
		close(ch)
		return
	}
	s.subs = append(s.subs, ch)
}

// Unsubscribe removes and closed a given channel.
func (s *Stream) Unsubscribe(ch chan []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, sub := range s.subs {
		if ch != sub {
			continue
		}
		// Reslice and close
		s.subs = append(s.subs[:i], s.subs[i+1:]...)
		close(ch)
		break
	}
}

// Publish sends the message to all subscribers
func (s *Stream) Publish(b []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ch := range s.subs {
		select {
		case ch <- b:
			// Send bytes
		default:
			// Drop bytes
		}
	}
}

// Pipe publishes bytes to the stream from a reader
func (s *Stream) Pipe(r io.Reader) {
	for {
		b := make([]byte, 1024)
		n, err := r.Read(b)

		// Close the stream on EOF
		if err != nil {
			s.Close()
			return
		}
		s.Publish(b[:n])
	}
}

// Close closes all channels still subscribed
func (s *Stream) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true

	// Clean up subs
	for _, ch := range s.subs {
		close(ch)
	}
	s.subs = make([]chan []byte, 0)
}
