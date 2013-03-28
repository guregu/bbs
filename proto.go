package bbs

import "fmt"

//This has all the structs for various commands

type BBSCommand struct {
	Command string `json:"cmd"`
}

type UserCommand struct {
	Command string `json:"cmd"`
	Session string `json:"session"`
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
	Command         string     `json:"cmd"`
	Name            string     `json:"name"`
	ProtocolVersion int        `json:"version"`
	Description     string     `json:"desc"`
	SecureURL       string     `json:"secure,omitempty"` //https URL, if any
	Options         []string   `json:"options,omitempty"`
	Access          AccessInfo `json:"access"`
	Formats         []string   `json:"format"` //formats the server accepts, the first one should be the primary one
	ServerVersion   string     `json:"server"`
}

// guest commands are commands you can use without logging on (e.g. "list", "get") 
// user commands require being logged in first (usually "post" and "reply")
type AccessInfo struct {
	GuestCommands []string `json:"guest,omitempty"`
	UserCommands  []string `json:"user,omitempty"`
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
	Result  string `json:"result,omitempty"`
}

// "login" command (client -> server)
type LoginCommand struct {
	Command         string `json:"cmd"`
	Username        string `json:"username"`
	Password        string `json:"password"`
	ProtocolVersion int    `json:"version"`
}

// "welcome" message (server -> client)
type WelcomeMessage struct {
	Command  string `json:"cmd"`
	Username string `json:"username,omitempty"` //omit for option 'anon'
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
	ThreadID string `json:"id"`
	Board    string `json:"board,omitempty"`  //option: "tags"
	Range    *Range `json:"range"`            //option: "range"
	Filter   string `json:"filter,omitempty"` //option: "filter"
	Format   string `json:"format,omitempty"`
}

// "list" command (client -> server)
type ListCommand struct {
	Command string `json:"cmd"`
	Session string `json:"session,omitempty"`
	Type    string `json:"type"`
	Query   string `json:"query"` //board for "boards", tag expression for "tags" (like "Dogs+Pizza-Anime")
}

// "reply" command (client -> server)
type ReplyCommand struct {
	Command string `json:"cmd"`
	Session string `json:"session,omitempty"`
	To      string `json:"to"`
	Board   string `json:"board,omitempty"` //option: "boards"
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
	Board   string   `json:"board,omitempty"` //option: "boards"
	Tags    []string `json:"tags,omitempty"`  //option: "tags"
}

// "msg" message (server -> client) [response to "get"]
type ThreadMessage struct {
	Command  string     `json:"cmd"`
	ID       string     `json:"id"`
	Title    string     `json:"title,omitempty"`
	Range    *Range     `json:"range,omitempty"`
	Closed   bool       `json:"closed,omitempty"`
	Filter   string     `json:"filter,omitempty"` //option: "filter"
	Board    string     `json:"board,omitempty"`  //option: "boards"
	Tags     []string   `json:"tags,omitempty"`   //option: "tags"
	Format   string     `json:"format,omitempty"`
	Messages []*Message `json:"messages"`
	More     bool       `json:"more,omitempty"`
}

func (t *ThreadMessage) Size() int {
	if t.Messages != nil {
		return len(t.Messages)
	}
	return 0
}

// format for posts used in "msg"
type Message struct {
	ID           string `json:"id"`
	Author       string `json:"user"`
	AuthorID     string `json:"user_id"`
	Date         string `json:"date"`
	Text         string `json:"body"`
	Signature    string `json:"sig"`
	AuthorTitle  string `json:"user_title,omitempty"` //option: "usertitles"
	AvatarURL    string `json:"avatar,omitempty"`     //option: "avatars"
	PictureURL   string `json:"img,omitempty"`        //option: "imageboard"
	ThumbnailURL string `json:"thumb,omitempty"`      //option: "imageboard"
}

type TypedMessage struct {
	Command string `json:"cmd"`
	Type    string `json:"type"`
}

// "list" message where type = "thread" (server -> client)
type ListMessage struct {
	Command string           `json:"cmd"`
	Type    string           `json:"type"`
	Query   string           `json:"query,omitempty"`
	Threads []*ThreadListing `json:"threads"`
}

// "list" message where type = "board" (server -> client)
type BoardListMessage struct {
	Command string          `json:"cmd"`
	Type    string          `json:"type"`
	Query   string          `json:"query,omitempty"`
	Boards  []*BoardListing `json:"boards"`
}

// format for threads in "list"
type ThreadListing struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Author       string   `json:"user"`
	AuthorID     string   `json:"user_id"`
	Date         string   `json:"date"`
	PostCount    int      `json:"posts,omitempty"`
	UnreadPosts  int      `json:"unread_posts"`
	Sticky       bool     `json:"sticky"`          //a sticky (aka pinned) topic
	Closed       bool     `json:"closed"`          //a closed (aka locked) topic
	Tags         []string `json:"tags,omitempty"`  //option: "tags"
	PictureURL   string   `json:"img,omitempty"`   //option: "imageboard"
	ThumbnailURL string   `json:"thumb,omitempty"` //option: "imageboard"
}

// format for boards in "list"
type BoardListing struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"desc,omitempty"`
	ThreadCount int    `json:"threads,omitempty"`
	PostCount   int    `json:"posts,omitempty"`
}
