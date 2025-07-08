package services

import (
	"sync"
)

type SyncService interface {
	InServerEditContext(act Action) error
	InServerUseContext(act Action) error

	InPeerCreateContext(act Action) error
	InPeerEditContext(peerPrint string, act Action) error
	InPeerUseContext(peerPrint string, act Action) error
}

type Action func() error

type Sync struct {
	serverMutex *sync.RWMutex

	peersCreateMutex *sync.Mutex
	peersMutexMap    map[string]*TempMutex
	mapMutex         *sync.RWMutex
}

type TempMutex struct {
	mutex sync.RWMutex
	users int
}

func NewSyncService() *Sync {
	return &Sync{
		serverMutex:      &sync.RWMutex{},
		peersCreateMutex: &sync.Mutex{},
		mapMutex:         &sync.RWMutex{},
		peersMutexMap:    map[string]*TempMutex{},
	}
}

func (s *Sync) InServerEditContext(act Action) error {
	s.serverMutex.Lock()
	defer s.serverMutex.Unlock()

	err := act()

	return err
}

func (s *Sync) InServerUseContext(act Action) error {
	s.serverMutex.RLock()
	defer s.serverMutex.RUnlock()

	err := act()

	return err
}

func (s *Sync) InPeerCreateContext(act Action) error {
	s.peersCreateMutex.Lock()
	defer s.peersCreateMutex.Unlock()

	err := act()

	return err
}

func (s *Sync) InPeerEditContext(peerPrint string, act Action) error {
	err := s.inPeerContext(peerPrint, act, false)
	return err
}

func (s *Sync) InPeerUseContext(peerPrint string, act Action) error {
	err := s.inPeerContext(peerPrint, act, true)
	return err
}

//
// Helpers
//

func (s *Sync) inPeerContext(peerPrint string, act Action, isReadLock bool) error {

	mutex := s.mapGetOrCreate(peerPrint)
	defer s.mapDelete(peerPrint)

	if isReadLock {
		mutex.mutex.RLock()
		defer mutex.mutex.RUnlock()
	} else {
		mutex.mutex.Lock()
		defer mutex.mutex.Unlock()
	}

	err := act()

	return err
}

func (s *Sync) mapGetOrCreate(key string) *TempMutex {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()

	val, ok := s.peersMutexMap[key]
	if !ok {
		val = &TempMutex{mutex: sync.RWMutex{}, users: 0}
		s.peersMutexMap[key] = val
	}

	val.users++

	return val
}

func (s *Sync) mapDelete(key string) {
	s.mapMutex.Lock()
	defer s.mapMutex.Unlock()

	val, ok := s.peersMutexMap[key]
	if !ok {
		return
	}

	if val.users > 0 {
		val.users--
	} else {
		delete(s.peersMutexMap, key)
	}
}
