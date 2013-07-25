package bbs

import "sync"
import "fmt"
import "time"
import "crypto/rand"

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

func (sh *SessionHandler) TryLogin(m *LoginCommand) *Session {
	//try to log in
	var board BBS
	board = sh.Server.NewBBS()
	if board.LogIn(m) {
		//TODO: better session shit
		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			return nil
		}
		id := fmt.Sprintf("%x", b)
		sesh := &Session{
			SessionID:  id,
			UserID:     m.Username,
			BBS:        board,
			LastAction: time.Now(),
		}
		sh.sessionMutex.Lock()
		sh.sessions[id] = sesh
		sh.sessionMutex.Unlock()
		return sesh
	}
	return nil
}

func (sh *SessionHandler) Logout(sesh string) {
	sh.sessionMutex.Lock()
	delete(sh.sessions, sesh)
	sh.sessionMutex.Unlock()
}

func (sh *SessionHandler) Get(sesh string) *Session {
	sh.sessionMutex.RLock()
	defer sh.sessionMutex.RUnlock()
	s, ok := sh.sessions[sesh]
	if !ok {
		return nil
	}
	//TODO redo everything
	//s.LastAction = time.Now()
	return s
}
