package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type RWMutex struct {
	mu        sync.Mutex
	readCond  *sync.Cond
	writeCond *sync.Cond

	readers        uint64
	writers        uint64
	waitingWriters uint64
	writing        bool

	once sync.Once
}

func (m *RWMutex) init() {
	m.readCond = sync.NewCond(&m.mu)
	m.writeCond = sync.NewCond(&m.mu)
}

func (m *RWMutex) Lock() {
	m.once.Do(m.init)
	m.mu.Lock()

	m.waitingWriters++

	for m.readers > 0 || m.writing {
		m.writeCond.Wait()
	}

	m.waitingWriters--
	m.writing = true
	m.mu.Unlock()
}

func (m *RWMutex) Unlock() {
	m.once.Do(m.init)

	m.mu.Lock()
	m.writing = false
	if m.waitingWriters > 0 {
		m.writeCond.Signal()
	} else {
		m.readCond.Broadcast()
	}
	m.mu.Unlock()
}

func (m *RWMutex) RLock() {
	m.once.Do(m.init)

	m.mu.Lock()
	for m.waitingWriters > 0 || m.writing {
		m.readCond.Wait()
	}

	m.readers++
	m.mu.Unlock()
}

func (m *RWMutex) RUnlock() {
	m.once.Do(m.init)

	m.mu.Lock()
	m.readers--

	if m.readers == 0 && m.waitingWriters > 0 {
		m.writeCond.Signal()
	}
	m.mu.Unlock()
}

func TestRWMutexWithWriter(t *testing.T) {
	var mutex RWMutex
	mutex.Lock() // writer

	var mutualExlusionWithWriter atomic.Bool
	mutualExlusionWithWriter.Store(true)
	var mutualExlusionWithReader atomic.Bool
	mutualExlusionWithReader.Store(true)

	go func() {
		mutex.Lock() // another writer
		mutualExlusionWithWriter.Store(false)
	}()

	go func() {
		mutex.RLock() // another reader
		mutualExlusionWithReader.Store(false)
	}()

	time.Sleep(time.Second)
	assert.True(t, mutualExlusionWithWriter.Load())
	assert.True(t, mutualExlusionWithReader.Load())
}

func TestRWMutexWithReaders(t *testing.T) {
	var mutex RWMutex
	mutex.RLock() // reader

	var mutualExlusionWithWriter atomic.Bool
	mutualExlusionWithWriter.Store(true)

	go func() {
		mutex.Lock() // another writer
		mutualExlusionWithWriter.Store(false)
	}()

	time.Sleep(time.Second)
	assert.True(t, mutualExlusionWithWriter.Load())
}

func TestRWMutexMultipleReaders(t *testing.T) {
	var mutex RWMutex
	mutex.RLock() // reader

	var readersCount atomic.Int32
	readersCount.Add(1)

	go func() {
		mutex.RLock() // another reader
		readersCount.Add(1)
	}()

	go func() {
		mutex.RLock() // another reader
		readersCount.Add(1)
	}()

	time.Sleep(time.Second)
	assert.Equal(t, int32(3), readersCount.Load())
}

func TestRWMutexWithWriterPriority(t *testing.T) {
	var mutex RWMutex
	mutex.RLock() // reader

	var mutualExlusionWithWriter atomic.Bool
	mutualExlusionWithWriter.Store(true)
	var readersCount atomic.Int32
	readersCount.Add(1)

	go func() {
		mutex.Lock() // another writer is waiting for reader
		mutualExlusionWithWriter.Store(false)
	}()

	time.Sleep(time.Second)

	go func() {
		mutex.RLock() // another reader is waiting for a higher priority writer
		readersCount.Add(1)
	}()

	go func() {
		mutex.RLock() // another reader is waiting for a higher priority writer
		readersCount.Add(1)
	}()

	time.Sleep(time.Second)

	assert.True(t, mutualExlusionWithWriter.Load())
	assert.Equal(t, int32(1), readersCount.Load())
}
