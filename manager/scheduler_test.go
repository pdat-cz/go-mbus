package manager

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewScheduler(t *testing.T) {
	s := newScheduler(30*time.Second, 500*time.Millisecond)

	if s.interval != 30*time.Second {
		t.Errorf("interval = %v, want 30s", s.interval)
	}
	if s.staggerDelay != 500*time.Millisecond {
		t.Errorf("staggerDelay = %v, want 500ms", s.staggerDelay)
	}
	if s.IsRunning() {
		t.Error("Scheduler should not be running initially")
	}
}

func TestSchedulerAddRemoveDevice(t *testing.T) {
	s := newScheduler(30*time.Second, 0)

	// Add device
	s.AddDevice(DeviceConfig{Address: 1, Name: "Meter 1", Enabled: true})

	schedules := s.GetDeviceSchedules()
	if len(schedules) != 1 {
		t.Errorf("Device count = %v, want 1", len(schedules))
	}
	if schedules[1].name != "Meter 1" {
		t.Errorf("Device name = %v, want 'Meter 1'", schedules[1].name)
	}

	// Remove device
	s.RemoveDevice(1)
	schedules = s.GetDeviceSchedules()
	if len(schedules) != 0 {
		t.Errorf("Device count after remove = %v, want 0", len(schedules))
	}
}

func TestSchedulerUpdateDevice(t *testing.T) {
	s := newScheduler(30*time.Second, 0)
	s.AddDevice(DeviceConfig{Address: 1, Name: "Meter 1", Enabled: true})

	// Update device
	s.UpdateDevice(DeviceConfig{Address: 1, Name: "Updated Meter", Enabled: false})

	schedules := s.GetDeviceSchedules()
	if schedules[1].name != "Updated Meter" {
		t.Errorf("Device name = %v, want 'Updated Meter'", schedules[1].name)
	}
	if schedules[1].enabled {
		t.Error("Device should be disabled")
	}
}

func TestSchedulerDeviceInterval(t *testing.T) {
	s := newScheduler(30*time.Second, 0)

	// Add device with custom interval
	customInterval := 60 * time.Second
	s.AddDevice(DeviceConfig{
		Address:  1,
		Name:     "Meter 1",
		Enabled:  true,
		Interval: &customInterval,
	})

	schedules := s.GetDeviceSchedules()
	if schedules[1].interval != 60*time.Second {
		t.Errorf("Device interval = %v, want 60s", schedules[1].interval)
	}
}

func TestSchedulerEnableDisableDevice(t *testing.T) {
	s := newScheduler(30*time.Second, 0)
	s.AddDevice(DeviceConfig{Address: 1, Name: "Meter 1", Enabled: true})

	// Disable device
	s.DisableDevice(1)
	schedules := s.GetDeviceSchedules()
	if schedules[1].enabled {
		t.Error("Device should be disabled")
	}

	// Enable device
	s.EnableDevice(1)
	schedules = s.GetDeviceSchedules()
	if !schedules[1].enabled {
		t.Error("Device should be enabled")
	}
}

func TestSchedulerTriggerPoll(t *testing.T) {
	s := newScheduler(30*time.Second, 0)
	s.AddDevice(DeviceConfig{Address: 1, Name: "Meter 1", Enabled: true})

	// Mark as polled to set nextPoll to future
	s.MarkPolled(1)

	// Get next poll time (should be ~30s in the future)
	nextPoll, ok := s.GetNextPollTime(1)
	if !ok {
		t.Fatal("Should get next poll time")
	}

	// Verify it's in the future
	if !nextPoll.After(time.Now()) {
		t.Error("Next poll time should be in the future after MarkPolled")
	}

	// Trigger poll (should reset next poll time to now)
	result := s.TriggerPoll(1)
	if !result {
		t.Error("TriggerPoll should return true")
	}

	newNextPoll, _ := s.GetNextPollTime(1)
	// After trigger, next poll should be very close to now (within 1 second)
	if time.Until(newNextPoll) > time.Second {
		t.Error("Next poll time should be reset to approximately now")
	}

	// Trigger poll for non-existent device
	result = s.TriggerPoll(99)
	if result {
		t.Error("TriggerPoll should return false for non-existent device")
	}
}

func TestSchedulerStartStop(t *testing.T) {
	s := newScheduler(100*time.Millisecond, 0)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.Start(ctx)
	if !s.IsRunning() {
		t.Error("Scheduler should be running after Start")
	}

	// Starting again should be no-op
	s.Start(ctx)

	s.Stop()
	if s.IsRunning() {
		t.Error("Scheduler should not be running after Stop")
	}

	// Stopping again should be no-op
	s.Stop()
}

func TestSchedulerPolling(t *testing.T) {
	s := newScheduler(50*time.Millisecond, 0)

	var mu sync.Mutex
	polledAddresses := make([]int, 0)

	s.SetOnPoll(func(address int) {
		mu.Lock()
		polledAddresses = append(polledAddresses, address)
		mu.Unlock()
	})

	s.AddDevice(DeviceConfig{Address: 1, Name: "Meter 1", Enabled: true})
	s.AddDevice(DeviceConfig{Address: 2, Name: "Meter 2", Enabled: true})
	s.AddDevice(DeviceConfig{Address: 3, Name: "Meter 3", Enabled: false}) // Disabled

	ctx, cancel := context.WithCancel(context.Background())
	s.Start(ctx)

	// Wait for at least one polling cycle
	time.Sleep(200 * time.Millisecond)

	cancel()
	s.Stop()

	mu.Lock()
	count := len(polledAddresses)
	mu.Unlock()

	if count < 2 {
		t.Errorf("Expected at least 2 polls, got %v", count)
	}

	// Check that disabled device was not polled
	mu.Lock()
	for _, addr := range polledAddresses {
		if addr == 3 {
			t.Error("Disabled device should not be polled")
		}
	}
	mu.Unlock()
}

func TestSchedulerMarkPolled(t *testing.T) {
	s := newScheduler(30*time.Second, 0)
	s.AddDevice(DeviceConfig{Address: 1, Name: "Meter 1", Enabled: true})

	beforePoll, _ := s.GetNextPollTime(1)

	s.MarkPolled(1)

	afterPoll, _ := s.GetNextPollTime(1)

	if !afterPoll.After(beforePoll) {
		t.Error("Next poll time should be after mark polled")
	}
}

func TestSchedulerSetInterval(t *testing.T) {
	s := newScheduler(30*time.Second, 0)

	s.SetInterval(60 * time.Second)

	if s.interval != 60*time.Second {
		t.Errorf("interval = %v, want 60s", s.interval)
	}
}

func TestSchedulerGetNextPollTimeNotFound(t *testing.T) {
	s := newScheduler(30*time.Second, 0)

	_, ok := s.GetNextPollTime(99)
	if ok {
		t.Error("GetNextPollTime should return false for non-existent device")
	}
}
