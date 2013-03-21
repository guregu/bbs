package bbs

import "fmt"

//This has all the structs for various commands

type BBSCommand struct {
	Command string `json:"cmd"`
}

//From start to end inclusive, starting from 1. 
type Range struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

func (r *Range) String() string {
	return fmt.Sprintf("%d-%d", r.Start, r.End)
}

func (r *Range) Validate() bool {
	if r.Start > r.End {
		return false
	}
	return true
}

// "hello" message (server -> client)
type HelloMessage struct {
	Command         string `json:"cmd"`
	Name            string `json:"name"`
	ProtocolVersion int    `json:"version"`
	Description     string `json:"description"`
	Options         string `json:"options"`
	ServerVersion   string `json:"server"`
}

// "error" message (server -> client)
type ErrorMessage struct {
	Command string `json:"cmd"`
	ReplyTo string `json:"wrt"`
	Error   string `json:"error"`
}

// session expired or invalid? use this
var SessionErrorMessage *ErrorMessage = &ErrorMessage{"error", "session", "Invalid session."}

// "ok" message (server -> client)
type OKMessage struct {
	Command string `json:"cmd"`
	ReplyTo string `json:"wrt"`
	Message string `json:"msg,omitempty"`
}

// "login" command (client -> server)
type LoginCommand struct {
	Command         string `json:"cmd"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	ProtocolVersion int    `json:"version"`
	ClientVersion   string `json:"client"`
}

// "welcome" message (server -> client)
type WelcomeMessage struct {
	Command  string `json:"cmd"`
	Username string `json:"username"`
	Session  string `json:"session"`
}

// "logout" command (client -> server)
type LogoutCommand struct {
	Command string `json:"cmd"`
	Session string `json:"session"`
}

// "get" command (client -> server)
type GetCommand struct {
	Command  string `json:"cmd"`
	Session  string `json:"session,omitempty"`
	ThreadID string `json:"thread"`
	Range    *Range `json:"range"`
	Filter   string `json:"filter,omitempty"`
}

// "list" command (client -> server)
type ListCommand struct {
	Command string `json:"cmd"`
	Session string `json:"session,omitempty"`
	Type    string `json:"type"`
	Query   string `json:"query"`
}

// "reply" command (client -> server)
type ReplyCommand struct {
	Command string `json:"cmd"`
	Session string `json:"session,omitempty"`
	To      string `json:"to"`
	Text    string `json:"body"`
	Format  string `json:"format,omitempty"`
}

// "post" command (client -> server)
type PostCommand struct {
	Command string   `json:"cmd"`
	Session string   `json:"session,omitempty"`
	Title   string   `json:"title"`
	Text    string   `json:"body"`
	Format  string   `json:"format,omitempty"`
	Tags    []string `json:"tags,omitempty"`
}

// "msg" message (server -> client) [response to "get"]
type ThreadMessage struct {
	Command  string     `json:"cmd"`
	ID       string     `json:"id"`
	Title    string     `json:"title"`
	Range    *Range     `json:"range"`
	Filter   string     `json:"filter,omitempty"`
	Tags     []string   `json:"tags,omitempty"`
	Messages []*Message `json:"messages"`
}

func (t *ThreadMessage) Size() int {
	if t.Messages != nil {
		return len(t.Messages)
	}
	return 0
}

// format for posts used in "msg"
type Message struct {
	ID                  string `json:"id"`
	Author              string `json:"user"`
	AuthorID            string `json:"user_id"`
	Date                string `json:"date"`
	Text                string `json:"body"`
	AuthorTitle         string `json:"user_title,omitempty"`
	AvatarURL           string `json:"avatar,omitempty"`
	PictureURL          string `json:"img,omitempty"`
	PictureThumbnailURL string `json:"thumb,omitempty"`
}

// "list" message (server -> client)
type ListMessage struct {
	Command string           `json:"cmd"`
	Type    string           `json:"type"`
	Query   string           `json:"query"`
	Threads []*ThreadListing `json:"threads"`
}

// format for threads in "list"
type ThreadListing struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Author    string   `json:"user"`
	AuthorID  string   `json:"user_id"`
	Date      string   `json:"date"`
	PostCount int      `json:"posts"`
	Tags      []string `json:"tags,omitempty"`
}
