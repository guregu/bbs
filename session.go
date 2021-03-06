package bbs

import (
	"crypto/rand"
	"fmt"
	"log"
	"sync"
	"time"
)

type Session struct {
	SessionID  string
	UserID     string
	BBS        BBS
	LastAction time.Time
}

type SessionHandler struct {
	Server *Server

	sessions     map[string]*Session
	sessionMutex sync.RWMutex
}

func NewSessionHandler(srv *Server) *SessionHandler {
	return &SessionHandler{
		Server:       srv,
		sessions:     make(map[string]*Session),
		sessionMutex: sync.RWMutex{},
	}
}

func (sh *SessionHandler) Get(sesh string) *Session {
	sh.sessionMutex.RLock()
	defer sh.sessionMutex.RUnlock()
	s, ok := sh.sessions[sesh]
	if !ok {
		return nil
	}
	// update last active time in a diff thread
	go sh.Touch(sesh)
	return s
}

func (sh *SessionHandler) Touch(sesh string) {
	sh.sessionMutex.Lock()
	s := sh.sessions[sesh]
	if s != nil {
		s.LastAction = time.Now()
	}
	sh.sessionMutex.Unlock()
}

func (sh *SessionHandler) Add(sesh *Session) {
	sh.sessionMutex.Lock()
	defer sh.sessionMutex.Unlock()

	if _, exists := sh.sessions[sesh.SessionID]; exists {
		log.Printf("Warning: replaced already-existing session %s", sesh.SessionID)
	}

	sh.sessions[sesh.SessionID] = sesh
}

func (sh *SessionHandler) TryLogin(m LoginCommand) *Session {
	//try to log in
	var board BBS
	board = sh.Server.NewBBS()
	if board.LogIn(m) {
		sesh := &Session{
			SessionID:  sessionKey(),
			UserID:     m.Username,
			BBS:        board,
			LastAction: time.Now(),
		}
		sh.Add(sesh)
		return sesh
	}
	return nil
}

func (sh *SessionHandler) Upgrade(sesh *Session, m LoginCommand) bool {
	if sesh.BBS.LogIn(m) {
		sesh.SessionID = sessionKey()
		sesh.UserID = m.Username
		sesh.LastAction = time.Now()

		sh.Add(sesh)
		return true
	}
	return false
}

func (sh *SessionHandler) Copy(from *Session, to *Session) {
	to.SessionID = from.SessionID
	to.UserID = from.UserID
	to.BBS = from.BBS
	to.LastAction = from.LastAction
}

func (sh *SessionHandler) Logout(sesh string) {
	sh.sessionMutex.Lock()
	delete(sh.sessions, sesh)
	sh.sessionMutex.Unlock()
}

func sessionKey() string {
	//TODO: make this better
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}
