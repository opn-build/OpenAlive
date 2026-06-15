// Package engine runs the activity loop, a Go port of core/activity_engine.py.
//
// It runs on its own goroutine and stops instantly via context cancellation
// (the Python version's threading.Event), instead of waiting out a poll
// interval. UI callbacks are invoked from the goroutine; the caller is
// responsible for marshalling them onto the GUI thread.
package engine

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"openalive/internal/config"
	"openalive/internal/i18n"
	"openalive/internal/schedule"
	"openalive/internal/winput"
)

const (
	pollInterval = 2 * time.Second       // status poll cadence
	moveDwell    = 80 * time.Millisecond // pause between out/back move
)

// Engine drives mouse/keyboard activity according to the scheduler.
type Engine struct {
	cfg   *config.Store
	sched *schedule.Scheduler

	onAction       func(kind, msg string)
	onStatusChange func(status string)

	mu         sync.Mutex
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	lastStatus string
}

// New builds an Engine. Callbacks may be nil.
func New(cfg *config.Store, sched *schedule.Scheduler, onAction func(kind, msg string), onStatusChange func(status string)) *Engine {
	return &Engine{cfg: cfg, sched: sched, onAction: onAction, onStatusChange: onStatusChange}
}

// Start launches the loop. Calling Start while running is a no-op.
func (e *Engine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	e.cancel = cancel
	e.wg.Add(1)
	go e.loop(ctx)
}

// Stop cancels the loop and waits for it to exit. Returns almost immediately.
func (e *Engine) Stop() {
	e.mu.Lock()
	cancel := e.cancel
	e.cancel = nil
	e.mu.Unlock()
	if cancel != nil {
		cancel()
		e.wg.Wait()
	}
}

func (e *Engine) loop(ctx context.Context) {
	defer e.wg.Done()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	var nextActionAt time.Time // zero => act on first active tick

	check := func() {
		status := e.sched.Status()
		if status != e.lastStatus {
			e.lastStatus = status
			if e.onStatusChange != nil {
				e.onStatusChange(status)
			}
		}
		if status == schedule.StatusActive && !time.Now().Before(nextActionAt) {
			e.performAction()
			interval := e.cfg.Snapshot().MouseIntervalSeconds
			if interval < 5 {
				interval = 5
			}
			nextActionAt = time.Now().Add(time.Duration(interval) * time.Second)
		}
	}

	check() // act on startup without waiting a full tick
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			check()
		}
	}
}

func (e *Engine) performAction() {
	if !winput.Available() {
		e.emit("simulado", i18n.T("engine.sim_unavailable"))
		return
	}

	c := e.cfg.Snapshot()
	pixels := c.MouseMovePixels
	if pixels < 1 {
		pixels = 1
	}
	dx := pixels
	if rand.Intn(2) == 0 {
		dx = -pixels
	}
	dy := pixels
	if rand.Intn(2) == 0 {
		dy = -pixels
	}

	winput.MoveMouse(dx, dy)
	time.Sleep(moveDwell)
	winput.MoveMouse(-dx, -dy)

	keyInfo := ""
	if c.KeystrokeEnabled {
		if vk, ok := winput.VKCodes[c.KeystrokeKey]; ok {
			winput.TapKey(vk)
			keyInfo = " " + i18n.T("engine.key_action", map[string]string{"key": c.KeystrokeKey})
		}
	}

	e.emit("action", i18n.T("engine.mouse_action", map[string]string{
		"px":    fmt.Sprintf("%d", pixels),
		"extra": keyInfo,
	}))
}

func (e *Engine) emit(kind, msg string) {
	if e.onAction != nil {
		e.onAction(kind, msg)
	}
}
