package manager

import (
	"context"
	"sync"
	"time"
)

// scheduler manages polling of M-Bus devices.
type scheduler struct {
	mu sync.RWMutex

	// Configuration
	interval     time.Duration
	staggerDelay time.Duration

	// Device schedules
	devices map[int]*deviceSchedule

	// Callbacks
	onPoll func(address int)

	// State
	running bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// deviceSchedule tracks per-device scheduling.
type deviceSchedule struct {
	address      int
	name         string
	interval     time.Duration
	lastPoll     time.Time
	nextPoll     time.Time
	enabled      bool
	inProgress   bool
}

// newScheduler creates a new polling scheduler.
func newScheduler(interval, staggerDelay time.Duration) *scheduler {
	return &scheduler{
		interval:     interval,
		staggerDelay: staggerDelay,
		devices:      make(map[int]*deviceSchedule),
	}
}

// SetOnPoll sets the callback for device polling.
func (s *scheduler) SetOnPoll(fn func(address int)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onPoll = fn
}

// AddDevice adds a device to the polling schedule.
func (s *scheduler) AddDevice(config DeviceConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	interval := s.interval
	if config.Interval != nil {
		interval = *config.Interval
	}

	s.devices[config.Address] = &deviceSchedule{
		address:  config.Address,
		name:     config.Name,
		interval: interval,
		enabled:  config.Enabled,
		nextPoll: time.Now(), // Poll immediately on start
	}
}

// RemoveDevice removes a device from the polling schedule.
func (s *scheduler) RemoveDevice(address int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.devices, address)
}

// UpdateDevice updates a device's schedule.
func (s *scheduler) UpdateDevice(config DeviceConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if dev, ok := s.devices[config.Address]; ok {
		dev.name = config.Name
		dev.enabled = config.Enabled
		if config.Interval != nil {
			dev.interval = *config.Interval
		}
	}
}

// SetInterval updates the global polling interval.
func (s *scheduler) SetInterval(interval time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.interval = interval
}

// Start begins the polling loop.
func (s *scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}

	ctx, s.cancel = context.WithCancel(ctx)
	s.running = true
	s.mu.Unlock()

	s.wg.Add(1)
	go s.pollLoop(ctx)
}

// Stop halts the polling loop.
func (s *scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}

	s.running = false
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()

	s.wg.Wait()
}

// pollLoop is the main polling goroutine.
func (s *scheduler) pollLoop(ctx context.Context) {
	defer s.wg.Done()

	// Use a ticker for the minimum check interval
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkAndPoll(ctx)
		}
	}
}

// checkAndPoll checks for devices due for polling and polls them.
func (s *scheduler) checkAndPoll(ctx context.Context) {
	s.mu.RLock()
	onPoll := s.onPoll
	stagger := s.staggerDelay

	// Collect devices that need polling
	var toPoll []*deviceSchedule
	now := time.Now()

	for _, dev := range s.devices {
		if dev.enabled && !dev.inProgress && now.After(dev.nextPoll) {
			toPoll = append(toPoll, dev)
		}
	}
	s.mu.RUnlock()

	if onPoll == nil || len(toPoll) == 0 {
		return
	}

	// Poll devices with stagger delay
	for i, dev := range toPoll {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Mark as in progress
		s.mu.Lock()
		dev.inProgress = true
		s.mu.Unlock()

		// Call the poll handler
		onPoll(dev.address)

		// Update schedule
		s.mu.Lock()
		dev.inProgress = false
		dev.lastPoll = time.Now()
		dev.nextPoll = dev.lastPoll.Add(dev.interval)
		s.mu.Unlock()

		// Stagger delay between devices (except for last one)
		if i < len(toPoll)-1 && stagger > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(stagger):
			}
		}
	}
}

// MarkPolled updates the last poll time for a device.
func (s *scheduler) MarkPolled(address int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if dev, ok := s.devices[address]; ok {
		dev.lastPoll = time.Now()
		dev.nextPoll = dev.lastPoll.Add(dev.interval)
	}
}

// GetNextPollTime returns when a device is next scheduled for polling.
func (s *scheduler) GetNextPollTime(address int) (time.Time, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if dev, ok := s.devices[address]; ok {
		return dev.nextPoll, true
	}
	return time.Time{}, false
}

// GetDeviceSchedules returns all device schedules.
func (s *scheduler) GetDeviceSchedules() map[int]deviceSchedule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[int]deviceSchedule)
	for addr, dev := range s.devices {
		result[addr] = *dev
	}
	return result
}

// TriggerPoll forces an immediate poll of a device.
func (s *scheduler) TriggerPoll(address int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if dev, ok := s.devices[address]; ok && dev.enabled {
		dev.nextPoll = time.Now()
		return true
	}
	return false
}

// EnableDevice enables polling for a device.
func (s *scheduler) EnableDevice(address int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if dev, ok := s.devices[address]; ok {
		dev.enabled = true
		dev.nextPoll = time.Now() // Poll immediately
	}
}

// DisableDevice disables polling for a device.
func (s *scheduler) DisableDevice(address int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if dev, ok := s.devices[address]; ok {
		dev.enabled = false
	}
}

// IsRunning returns whether the scheduler is running.
func (s *scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}
