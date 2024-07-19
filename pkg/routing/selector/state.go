package selector

import (
	"fmt"
	"sync"

	"github.com/livekit/protocol/logger"
)

type State int

const (
	StateBuilding State = iota
	StateStarted
	StateRunning
	StateEOS
	StateStopping
	StateFinished
)

type StateManager struct {
	lock  sync.RWMutex
	state State
}

func (s *StateManager) GetState() State {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.state
}

func (s *StateManager) GetStateLocked() State {
	return s.state
}

func (s *StateManager) LockState() {
	s.lock.Lock()
}

func (s *StateManager) UnlockState() {
	s.lock.Unlock()
}

func (s *StateManager) LockStateShared() {
	s.lock.RLock()
}

func (s *StateManager) UnlockStateShared() {
	s.lock.RUnlock()
}

func (s *StateManager) UpgradeState(state State) (State, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	old := s.state
	if old >= state {
		return old, false
	} else {
		logger.Debugw(fmt.Sprintf("pipeline state %v -> %v", old, state))
		s.state = state
		return old, true
	}
}

func (s State) String() string {
	switch s {
	case StateBuilding:
		return "building"
	case StateStarted:
		return "starting"
	case StateRunning:
		return "running"
	case StateEOS:
		return "eos"
	case StateStopping:
		return "stopping"
	case StateFinished:
		return "finished"
	default:
		return "unknown"
	}
}
