package synctest

import (
	"testing"
	"time"
)

const (
	timeout = 100 * time.Millisecond
)

func TestNotifyLock(t *testing.T) {
	nl := NewNotifyingLocker()
	ch := nl.NotifyLock()
	nl.Lock()
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatalf("didn't receive on channel after %s", timeout)
	}
}

func TestNotifyLockAlreadyLocked(t *testing.T) {
	nl := NewNotifyingLocker()
	nl.Lock()
	ch := nl.NotifyLock()
	select {
	case <-ch:
		t.Fatalf("received on channel")
	case <-time.After(timeout):
	}
}

func TestNotifyLockOnUnlock(t *testing.T) {
	nl := NewNotifyingLocker()
	nl.Lock()
	ch := nl.NotifyLock()
	nl.Unlock()
	select {
	case <-ch:
		t.Fatalf("got a lock notification on unlock")
	case <-time.After(timeout):
	}
}

func TestNotifyUnlock(t *testing.T) {
	nl := NewNotifyingLocker()
	ch := nl.NotifyUnlock()
	nl.Lock()
	nl.Unlock()
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Fatalf("didn't get unlock notification after %s", timeout)
	}
}

func TestNotifyUnlockAlreadyUnlocked(t *testing.T) {
	nl := NewNotifyingLocker()
	nl.Lock()
	nl.Unlock()
	ch := nl.NotifyUnlock()
	select {
	case <-ch:
		t.Fatalf("got unlock notification")
	case <-time.After(timeout):
	}
}

func TestNofifyUnlockOnLock(t *testing.T) {
	nl := NewNotifyingLocker()
	ch := nl.NotifyUnlock()
	nl.Lock()
	select {
	case <-ch:
		t.Fatalf("got unlock notification on lock")
	case <-time.After(timeout):
	}
}
