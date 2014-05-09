package bbs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"code.google.com/p/go.net/websocket"
)

var name string
var version int = 0
var options []string
var description string
var server_version string

type BBS interface {
	Hello() HelloMessage
	Register(m RegisterCommand) (OKMessage, error)
	LogIn(m LoginCommand) bool
	LogOut(m LogoutCommand) OKMessage
	IsLoggedIn() bool
	Get(m GetCommand) (ThreadMessage, error)
	List(m ListCommand) (ListMessage, error)
	Reply(m ReplyCommand) (OKMessage, error)
	Post(m PostCommand) (OKMessage, error)
}

type Boards interface {
	BoardList(m ListCommand) (BoardListMessage, error)
}

type Realtime interface {
	Listen(Listener)
	Bye()
}

type Bookmarks interface {
	BookmarkList(m ListCommand) (BookmarkListMessage, error)
}

type UnknownHandler interface {
	Unknown(string, []byte) interface{}
}

type Server struct {
	Sessions *SessionHandler
	Name     string
	WS       http.Handler

	factory       func() BBS
	userCommands  []string
	guestCommands []string
	defaultBBS    BBS
}

func NewServer(factory func() BBS) *Server {
	defaultBBS := factory()
	hello := defaultBBS.Hello()
	srv := &Server{
		factory:       factory,
		defaultBBS:    defaultBBS,
		Name:          hello.Name,
		userCommands:  hello.Access.UserCommands,
		guestCommands: hello.Access.GuestCommands,
	}
	srv.Sessions = NewSessionHandler(srv)
	srv.WS = websocket.Handler(srv.ServeWebsocket)
	return srv
}

func (srv *Server) NewBBS() BBS {
	return srv.factory()
}

func (srv *Server) DefaultBBS() BBS {
	return srv.defaultBBS
}

func (srv *Server) do(incoming BBSCommand, data []byte, sesh *Session) interface{} {
	var bbs BBS
	if sesh != nil {
		bbs = sesh.BBS
	} else {
		bbs = srv.DefaultBBS()
		if contains(srv.userCommands, incoming.Command) {
			return SessionErrorMessage
		}
	}
	switch incoming.Command {
	case "hello":
		return bbs.Hello()
	case "login":
		m := LoginCommand{}
		json.Unmarshal(data, &m)
		newsesh := srv.Sessions.TryLogin(m)
		if newsesh == nil {
			return Error("login", "Can't log in!")
		}
		return WelcomeMessage{"welcome", newsesh.UserID, newsesh.SessionID}
	case "register":
		m := RegisterCommand{}
		json.Unmarshal(data, &m)
		ok, err := bbs.Register(m)
		if err != nil {
			return Error("register", err.Error())
		}
		return ok
	case "get":
		m := GetCommand{}
		json.Unmarshal(data, &m)
		ok, err := bbs.Get(m)
		if err != nil {
			return Error("get", err.Error())
		}
		return ok
	case "list":
		m := ListCommand{}
		json.Unmarshal(data, &m)
		switch m.Type {
		case "thread", "":
			ok, err := bbs.List(m)
			if err != nil {
				return Error("list", err.Error())
			}
			return ok
		case "board":
			if b, ok := bbs.(Boards); ok {
				msg, err := b.BoardList(m)
				if err != nil {
					return Error("list", err.Error())
				}
				return msg
			}
		case "bookmark":
			if b, ok := bbs.(Bookmarks); ok {
				msg, err := b.BookmarkList(m)
				if err != nil {
					return Error("list", err.Error())
				}
				return msg
			}
		}
		return Error("list", "unsupported")
	case "reply":
		m := ReplyCommand{}
		json.Unmarshal(data, &m)
		ok, err := bbs.Reply(m)
		if err != nil {
			return err
		}
		return ok
	case "post":
		m := PostCommand{}
		json.Unmarshal(data, &m)
		ok, err := bbs.Post(m)
		if err != nil {
			return err
		}
		return ok
	case "logout":
		m := LogoutCommand{}
		json.Unmarshal(data, &m)
		srv.Sessions.Logout(m.Session)
		return bbs.LogOut(m)
	default:
		if b, ok := bbs.(UnknownHandler); ok {
			result := b.Unknown(incoming.Command, data)
			if result != nil {
				return result
			}
		}
		return Error(incoming.Command, "Unknown command: "+incoming.Command)
	}
	return nil
}

func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		sesh := srv.Sessions.Get(incoming.Session)
		result := srv.do(BBSCommand{incoming.Command}, data, sesh)
		w.Write(jsonify(result))
	default:
		log.Println("Weird method used: " + r.Method)
	}
}

func (srv *Server) ServeWebsocket(socket *websocket.Conn) {
	c := newClient(srv, socket)
	go c.writer()
	c.run()
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
	http.Handle(path, srv)
	http.Handle("/ws", websocket.Handler(srv.ServeWebsocket))
	hm := srv.defaultBBS.Hello()
	log.Printf("Starting BBS %s at %s%s\n", hm.Name, address, path)
	err := http.ListenAndServe(address, nil)
	if err != nil {
		panic(err)
	}
}

func Error(wrt, msg string) ErrorMessage {
	return ErrorMessage{
		Command: "error",
		ReplyTo: wrt,
		Error:   msg,
	}
}

func OK(wrt string) OKMessage {
	return OKMessage{
		Command: "ok",
		ReplyTo: wrt,
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
