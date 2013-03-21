package bbs

import "fmt"
import "net/http"
import "log"
import "encoding/json"
import "io/ioutil"
import "time"
import "crypto/rand"

var name string
var version int = 0
var options string
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
	Reply(m *ReplyCommand) (*OKMessage, *ErrorMessage)
	Post(m *PostCommand) (*OKMessage, *ErrorMessage)
}

var sessions = make(map[string]*Session)
var factory func() BBS
var addr string = ":8080"

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
		sessions[id] = &sesh

		return &sesh
	}

	return nil
}

func logout(session string) {
	delete(sessions, session)
}

func handle(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		//Display info
		index(w, r)
	case "POST":
		data, _ := ioutil.ReadAll(r.Body)
		incoming := BBSCommand{}
		err := json.Unmarshal(data, &incoming)
		if err != nil {
			fmt.Println("JSON Parsing Error!! " + string(data))
			return
		}
		switch incoming.Command {
		case "hello":
			resp := HelloMessage{"hello", name, version, description, options, server_version}
			w.Write(jsonify(&resp))
		case "login":
			m := LoginCommand{}
			json.Unmarshal(data, &m)
			sesh := tryLogin(&m)
			if sesh != nil {
				w.Write(jsonify(&WelcomeMessage{"welcome", sesh.UserID, sesh.SessionID}))
			} else {
				w.Write(jsonify(&ErrorMessage{"error", "login", "Can't log in!"}))
			}
		case "get":
			m := GetCommand{}
			json.Unmarshal(data, &m)
			sesh := getSession(m.Session)
			if sesh == nil {
				w.Write(jsonify(SessionErrorMessage))
				return
			}
			success, e := sesh.BBS.Get(&m)
			if success != nil {
				w.Write(jsonify(success))
			} else {
				w.Write(jsonify(e))
			}
		case "list":
			m := ListCommand{}
			json.Unmarshal(data, &m)
			sesh := getSession(m.Session)
			if sesh == nil {
				w.Write(jsonify(SessionErrorMessage))
				return
			}
			success, e := sesh.BBS.List(&m)
			if success != nil {
				w.Write(jsonify(success))
			} else {
				w.Write(jsonify(e))
			}
		case "reply":
			m := ReplyCommand{}
			json.Unmarshal(data, &m)
			sesh := getSession(m.Session)
			if sesh == nil {
				w.Write(jsonify(SessionErrorMessage))
				return
			}
			success, e := sesh.BBS.Reply(&m)
			if success != nil {
				w.Write(jsonify(success))
			} else {
				w.Write(jsonify(e))
			}
		case "post":
			m := PostCommand{}
			json.Unmarshal(data, &m)
			sesh := getSession(m.Session)
			if sesh == nil {
				w.Write(jsonify(SessionErrorMessage))
				return
			}
			success, e := sesh.BBS.Post(&m)
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

func getSession(sesh string) *Session {
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

func Serve(address string, path string, n string, opt string, desc string, servname string, fact func() BBS) {
	factory = fact
	addr = address
	name = n
	options = opt
	description = desc
	server_version = servname
	http.HandleFunc("/", index)
	http.HandleFunc(path, handle)
	fmt.Println("Starting server at " + addr)
	http.ListenAndServe(addr, nil)
}
