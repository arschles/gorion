package synctest

import "sync"

// NotifyingLocker is an implementation of sync.Locker that notifies callers when
// locks and unlocks happen. otherwise, it behaves identically as a sync.Mutex.
//
// Example usage:
//  nl := NewNotifyingLocker()
//  lch := nl.NotifyLock()
//  uch := nl.NotifyUnlock()
//  go func() {
//    nl.Lock()
//  }()
//  go func() {
//    <-lch // wait for nl.Lock() to be called in the goroutine above
//    nl.Unlock()
//  }()
//  <-uch // wait for nl.Unlock to be called
type NotifyingLocker struct {
	unlockChans     []chan struct{}
	unlockChansLock *sync.Mutex
	lockChans       []chan struct{}
	lockChansLock   *sync.Mutex
	lck             *sync.Mutex
}

// NewNotifyingLocker creates a new NotifyingLocker ready for use
func NewNotifyingLocker() *NotifyingLocker {
	return &NotifyingLocker{
		unlockChans:     nil,
		unlockChansLock: &sync.Mutex{},
		lockChans:       nil,
		lockChansLock:   &sync.Mutex{},
		lck:             &sync.Mutex{},
	}
}

// NotifyLock returns a channel that will close when n is locked. The channel
// never sends and will be closed immediately if n is already locked
func (n *NotifyingLocker) NotifyLock() <-chan struct{} {
	n.lockChansLock.Lock()
	defer n.lockChansLock.Unlock()
	ch := make(chan struct{})
	n.lockChans = append(n.lockChans, ch)
	return ch
}

// NotifyUnlock returns a channel that will close when n is unlocked. The channel
// never sends and will be closed immediately if n is already unlocked.
func (n *NotifyingLocker) NotifyUnlock() <-chan struct{} {
	n.unlockChansLock.Lock()
	defer n.unlockChansLock.Unlock()
	ch := make(chan struct{})
	n.unlockChans = append(n.unlockChans, ch)
	return ch
}

// Lock locks n and closes all unclosed channels returned previously by NotifyLock
func (n *NotifyingLocker) Lock() {
	n.lck.Lock()
	n.lockChansLock.Lock()
	defer n.lockChansLock.Unlock()
	for _, lck := range n.lockChans {
		close(lck)
	}
	n.lockChans = nil
}

// Unlock unlocks n and closes all unclosed channels returned previously by NotifyUnlock
func (n *NotifyingLocker) Unlock() {
	n.lck.Unlock()
	n.unlockChansLock.Lock()
	defer n.unlockChansLock.Unlock()
	for _, lck := range n.unlockChans {
		close(lck)
	}
	n.unlockChans = nil
}
