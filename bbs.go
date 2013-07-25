package bbs

import "fmt"
import "net/http"
import "log"
import "encoding/json"
import "io/ioutil"

var name string
var version int = 0
var options []string
var description string
var server_version string

type BBS interface {
	Hello() HelloMessage
	Register(m *RegisterCommand) (*OKMessage, *ErrorMessage)
	LogIn(m *LoginCommand) bool
	LogOut(m *LogoutCommand) *OKMessage
	IsLoggedIn() bool
	Get(m *GetCommand) (*ThreadMessage, *ErrorMessage)
	List(m *ListCommand) (*ListMessage, *ErrorMessage)
	BoardList(m *ListCommand) (*BoardListMessage, *ErrorMessage)
	BookmarkList(m *ListCommand) (*BookmarkListMessage, *ErrorMessage)
	Reply(m *ReplyCommand) (*OKMessage, *ErrorMessage)
	Post(m *PostCommand) (*OKMessage, *ErrorMessage)
	//Unknown(string, interface{}) interface{}
}

type Server struct {
	HTTP *httpHandler
	//WS   http.Handler
	Sessions *SessionHandler
	Name     string

	factory       func() BBS
	userCommands  []string
	guestCommands []string
	defaultBBS    BBS
}

func NewServer(factory func() BBS) *Server {
	srv := &Server{
		factory:    factory,
		defaultBBS: factory(),
	}
	srv.HTTP = &httpHandler{srv}
	srv.Sessions = NewSessionHandler(srv)
	hello := srv.defaultBBS.Hello()
	srv.Name = hello.Name
	srv.userCommands = hello.Access.UserCommands
	srv.guestCommands = hello.Access.GuestCommands
	return srv
}

func (srv *Server) NewBBS() BBS {
	return srv.factory()
}

func (srv *Server) DefaultBBS() BBS {
	return srv.defaultBBS
}

type httpHandler struct {
	server *Server
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv := h.server
	switch r.Method {
	case "GET":
		//Display info
		index(w, r)
	case "OPTIONS":
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
	case "POST":
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

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
			sesh = srv.Sessions.Get(incoming.Session)
			if sesh != nil {
				bbs = sesh.BBS
			} else {
				bbs = srv.defaultBBS
			}
		} else {
			bbs = srv.defaultBBS
		}
		if contains(srv.userCommands, incoming.Command) {
			if sesh == nil {
				// a guest tried to use a user command
				w.WriteHeader(401) //401 Unauthorized
				w.Write(jsonify(SessionErrorMessage))
				return
			}
		}
		switch incoming.Command {
		case "hello":
			hm := bbs.Hello()
			w.Write(jsonify(&hm))
		case "login":
			m := LoginCommand{}
			json.Unmarshal(data, &m)
			newsesh := srv.Sessions.TryLogin(&m)
			if newsesh != nil {
				w.Write(jsonify(&WelcomeMessage{"welcome", newsesh.UserID, newsesh.SessionID}))
			} else {
				w.WriteHeader(401) //401 Unauthorized
				w.Write(jsonify(&ErrorMessage{"error", "login", "Can't log in!"}))
			}
		case "register":
			m := RegisterCommand{}
			json.Unmarshal(data, &m)
			ok, err := bbs.Register(&m)
			if ok != nil {
				w.Write(jsonify(ok))
			} else {
				w.Write(jsonify(err))
			}
		case "get":
			m := GetCommand{}
			json.Unmarshal(data, &m)
			success, e := bbs.Get(&m)
			if success != nil {
				w.Write(jsonify(success))
			} else {
				w.WriteHeader(404) //404 Not Found
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
					w.WriteHeader(404) //404 Not Found
					w.Write(jsonify(e))
				}
			} else if m.Type == "board" {
				success, e := bbs.BoardList(&m)
				if success != nil {
					w.Write(jsonify(success))
				} else {
					w.WriteHeader(404) //404 Not Found
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
				w.WriteHeader(400) //400 Bad Request
				//TODO: sometimes should be 403 Forbidden
				w.Write(jsonify(e))
			}
		case "post":
			m := PostCommand{}
			json.Unmarshal(data, &m)
			success, e := bbs.Post(&m)
			if success != nil {
				w.Write(jsonify(success))
			} else {
				w.WriteHeader(400) //400 Bad Request
				w.Write(jsonify(e))
			}
		case "logout":
			m := LogoutCommand{}
			json.Unmarshal(data, &m)
			srv.Sessions.Logout(m.Session)
		default:
			w.WriteHeader(500) //500 internal server error
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
		http.NotFound(w, r)
	}
}

func Serve(address string, path string, fact func() BBS) {
	srv := NewServer(fact)
	http.HandleFunc("/", index)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle(path, srv.HTTP)
	hm := srv.defaultBBS.Hello()
	log.Printf("Starting BBS %s at %s%s\n", hm.Name, address, path)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic(err)
	}
}

func Error(wrt, msg string) *ErrorMessage {
	return &ErrorMessage{
		Command: "error",
		ReplyTo: wrt,
		Error:   msg,
	}
}

func OK(wrt string) *OKMessage {
	return &OKMessage{
		Command: "ok",
		ReplyTo: wrt,
	}
}
