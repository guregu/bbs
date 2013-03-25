package bbs

import "fmt"
import "net/http"
import "log"
import "encoding/json"
import "io/ioutil"
import "time"
import "crypto/rand"
import "sync"

var name string
var version int = 0
var options []string
var description string
var server_version string

type Session struct {
	SessionID  string
	UserID     string
	BBS        BBS
	LastAction time.Time
}

type BBS interface {
	LogIn(m *LoginCommand) bool
	LogOut(m *LogoutCommand) *OKMessage
	IsLoggedIn() bool
	Get(m *GetCommand) (*ThreadMessage, *ErrorMessage)
	List(m *ListCommand) (*ListMessage, *ErrorMessage)
	BoardList(m *ListCommand) (*BoardListMessage, *ErrorMessage)
	Reply(m *ReplyCommand) (*OKMessage, *ErrorMessage)
	Post(m *PostCommand) (*OKMessage, *ErrorMessage)
}

var sessions = make(map[string]*Session)
var sessionMutex = &sync.RWMutex{}
var factory func() BBS
var addr string = ":8080"
var hello HelloMessage

var userCommands []string
var guestCommands []string
var defaultBBS BBS

func tryLogin(m *LoginCommand) *Session {
	//try to log in 
	var board BBS
	board = factory()

	if board.LogIn(m) {
		//TODO: better session shit
		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			return nil
		}
		id := fmt.Sprintf("%x", b)
		sesh := Session{id, m.Username, board, time.Now()}

		sessionMutex.Lock()
		sessions[id] = &sesh
		sessionMutex.Unlock()

		return &sesh
	}

	return nil
}

func logout(session string) {
	sessionMutex.Lock()
	delete(sessions, session)
	sessionMutex.Unlock()
}

func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//Display info
		index(w, r)
	case "POST":
		data, _ := ioutil.ReadAll(r.Body)
		incoming := UserCommand{}
		err := json.Unmarshal(data, &incoming)
		if err != nil {
			fmt.Println("JSON Parsing Error!! " + string(data))
			return
		}
		var bbs BBS
		var sesh *Session
		if incoming.Session != "" {
			sesh = getSession(incoming.Session)
			bbs = sesh.BBS
		} else {
			bbs = defaultBBS
		}
		if contains(userCommands, incoming.Command) {
			if sesh == nil {
				// a guest tried to use a user command
				w.Write(jsonify(SessionErrorMessage))
				return
			}
		}
		switch incoming.Command {
		case "hello":
			w.Write(jsonify(&hello))
		case "login":
			m := LoginCommand{}
			json.Unmarshal(data, &m)
			newsesh := tryLogin(&m)
			if newsesh != nil {
				w.Write(jsonify(&WelcomeMessage{"welcome", newsesh.UserID, newsesh.SessionID}))
			} else {
				w.Write(jsonify(&ErrorMessage{"error", "login", "Can't log in!"}))
			}
		case "get":
			m := GetCommand{}
			json.Unmarshal(data, &m)
			success, e := bbs.Get(&m)
			if success != nil {
				w.Write(jsonify(success))
			} else {
				w.Write(jsonify(e))
			}
		case "list":
			m := ListCommand{}
			json.Unmarshal(data, &m)
			if m.Type == "" || m.Type == "thread" {
				success, e := bbs.List(&m)
				if success != nil {
					w.Write(jsonify(success))
				} else {
					w.Write(jsonify(e))
				}
			} else if m.Type == "board" {
				success, e := bbs.BoardList(&m)
				if success != nil {
					w.Write(jsonify(success))
				} else {
					w.Write(jsonify(e))
				}
			}
		case "reply":
			m := ReplyCommand{}
			json.Unmarshal(data, &m)
			success, e := bbs.Reply(&m)
			if success != nil {
				w.Write(jsonify(success))
			} else {
				w.Write(jsonify(e))
			}
		case "post":
			m := PostCommand{}
			json.Unmarshal(data, &m)
			success, e := bbs.Post(&m)
			if success != nil {
				w.Write(jsonify(success))
			} else {
				w.Write(jsonify(e))
			}
		case "logout":
			m := LogoutCommand{}
			json.Unmarshal(data, &m)
			logout(m.Session)
		default:
			w.Write(jsonify(&ErrorMessage{"error", incoming.Command, "Unknown command: " + incoming.Command}))
		}

	default:
		fmt.Println("Weird method used.")
	}
}

func contains(a []string, s string) bool {
	for _, c := range a {
		if c == s {
			return true
		}
	}
	return false
}

func getSession(sesh string) *Session {
	sessionMutex.RLock()
	defer sessionMutex.RUnlock()
	s, ok := sessions[sesh]
	if !ok {
		return nil
	}
	s.LastAction = time.Now()
	return s
}

func jsonify(j interface{}) []byte {
	b, err := json.Marshal(j)
	if err != nil {
		return []byte(`{"cmd": "error", "error": "server error"}`)
	}
	return b
}

func index(w http.ResponseWriter, r *http.Request) {
	f, err := ioutil.ReadFile("index.html")
	if err == nil {
		w.Write(f)
	} else {
		log.Fatal(err)
	}
}

func Serve(address string, path string, hm HelloMessage, fact func() BBS) {
	factory = fact
	addr = address
	hello = hm
	userCommands = hm.Access.UserCommands
	guestCommands = hm.Access.GuestCommands
	defaultBBS = fact()
	http.HandleFunc("/", index)
	http.HandleFunc(path, handle)
	fmt.Printf("Starting server %s at %s%s\n", hm.Name, addr, path)
	http.ListenAndServe(addr, nil)
}
