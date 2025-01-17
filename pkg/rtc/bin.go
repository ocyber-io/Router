package rtc

import (
	"sync"
	"time"

	"github.com/go-gst/go-gst/gst"
)

type Callbacks struct {
	mu       sync.RWMutex
	GstReady chan struct{}

	// upstream callbacks
	onError func(error)
	onStop  []func() error

	// source callbacks
	//onTrackAdded   []func(*config.TrackSource)
	onTrackMuted   []func(string)
	onTrackUnmuted []func(string, time.Duration)
	onTrackRemoved []func(string)

	// internal
	addBin    func(bin *gst.Bin)
	removeBin func(bin *gst.Bin)
}

func (c *Callbacks) SetOnError(f func(error)) {
	c.mu.Lock()
	c.onError = f
	c.mu.Unlock()
}

func (c *Callbacks) OnError(err error) {
	c.mu.RLock()
	onError := c.onError
	c.mu.RUnlock()

	if onError != nil {
		onError(err)
	}
}

func (c *Callbacks) AddOnStop(f func() error) {
	c.mu.Lock()
	c.onStop = append(c.onStop, f)
	c.mu.Unlock()
}

func (c *Callbacks) AddOnTrackMuted(f func(string)) {
	c.mu.Lock()
	c.onTrackMuted = append(c.onTrackMuted, f)
	c.mu.Unlock()
}

func (c *Callbacks) OnTrackMuted(trackID string) {
	c.mu.RLock()
	onTrackMuted := c.onTrackMuted
	c.mu.RUnlock()

	for _, f := range onTrackMuted {
		f(trackID)
	}
}

func (c *Callbacks) AddOnTrackUnmuted(f func(string, time.Duration)) {
	c.mu.Lock()
	c.onTrackUnmuted = append(c.onTrackUnmuted, f)
	c.mu.Unlock()
}

func (c *Callbacks) OnTrackUnmuted(trackID string, pts time.Duration) {
	c.mu.RLock()
	onTrackUnmuted := c.onTrackUnmuted
	c.mu.RUnlock()

	for _, f := range onTrackUnmuted {
		f(trackID, pts)
	}
}

func (c *Callbacks) AddOnTrackRemoved(f func(string)) {
	c.mu.Lock()
	c.onTrackRemoved = append(c.onTrackRemoved, f)
	c.mu.Unlock()
}

func (c *Callbacks) OnTrackRemoved(trackID string) {
	c.mu.RLock()
	onTrackRemoved := c.onTrackRemoved
	c.mu.RUnlock()

	for _, f := range onTrackRemoved {
		f(trackID)
	}
}
