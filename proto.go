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

func (r Range) String() string {
	return fmt.Sprintf("%d-%d", r.Start, r.End)
}

func (r Range) Validate() bool {
	if r.Start > r.End {
		return false
	}
	return true
}

func (r Range) Empty() bool {
	return r.End == 0
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
	Lists           []string   `json:"lists"`
	ServerVersion   string     `json:"server"`
	IconURL         string     `json:"icon"`
	// for option "range"
	DefaultRange Range `json:"default_range,omitempty"`
	// for option "realtime"
	RealtimeURL string `json:"realtime"`
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
var SessionErrorMessage ErrorMessage = Error("session", "bad session")

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

	// for "re-logins" only
	Session string `json:"session,omitempty"`
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

// "register" command (client -> server)
type RegisterCommand struct {
	Command  string `json:"cmd"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
}

// "get" command (client -> server)
type GetCommand struct {
	Command  string `json:"cmd"`
	Session  string `json:"session,omitempty"`
	ThreadID string `json:"id"`
	Range    Range  `json:"range"`            //option: "range"
	Filter   string `json:"filter,omitempty"` //option: "filter"
	Format   string `json:"format,omitempty"`
	Token    string `json:"token,omitempty"`
}

// "list" command (client -> server)
type ListCommand struct {
	Command string `json:"cmd"`
	Session string `json:"session,omitempty"`
	Type    string `json:"type"`
	Query   string `json:"query"` //board for "boards", tag expression for "tags" (like "Dogs+Pizza-Anime")
	Token   string `json:"token,omitempty"`
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
	Board   string   `json:"board,omitempty"` //option: "boards"
	Tags    []string `json:"tags,omitempty"`  //option: "tags"
}

// "msg" message (server -> client) [response to "get"]
type ThreadMessage struct {
	Command   string    `json:"cmd"`
	ID        string    `json:"id" bson:"_id"`
	Title     string    `json:"title,omitempty"`
	Range     Range     `json:"range,omitempty"`
	Closed    bool      `json:"closed,omitempty"`
	Filter    string    `json:"filter,omitempty"` //option: "filter"
	Board     string    `json:"board,omitempty"`  //option: "boards"
	Tags      []string  `json:"tags,omitempty"`   //option: "tags"
	Format    string    `json:"format,omitempty"`
	Messages  []Message `json:"messages"`
	More      bool      `json:"more,omitempty"`
	NextToken string    `json:"next,omitempty"`
}

func (t ThreadMessage) Size() int {
	if t.Messages != nil {
		return len(t.Messages)
	}
	return 0
}

// format for posts used in "msg"
type Message struct {
	ID                 string `json:"id"`
	Author             string `json:"user"`
	AuthorID           string `json:"user_id,omitempty"`
	Date               string `json:"date,omitempty"`
	Text               string `json:"body"`
	Signature          string `json:"sig,omitempty"`
	AuthorTitle        string `json:"user_title,omitempty"`   //option: "usertitles"
	AvatarURL          string `json:"avatar,omitempty"`       //option: "avatars"
	AvatarThumbnailURL string `json:"avatar_thumb,omitempty"` //option: "avatars"
	PictureURL         string `json:"img,omitempty"`          //option: "imageboard"
	ThumbnailURL       string `json:"thumb,omitempty"`        //option: "imageboard"
}

type TypedMessage struct {
	Command string `json:"cmd"`
	Type    string `json:"type"`
}

// "list" message where type = "thread" (server -> client)
type ListMessage struct {
	Command   string          `json:"cmd"`
	Type      string          `json:"type"`
	Query     string          `json:"query,omitempty"`
	Threads   []ThreadListing `json:"threads"`
	NextToken string          `json:"next,omitempty"`
}

// "list" message where type = "board" (server -> client)
type BoardListMessage struct {
	Command string         `json:"cmd"`
	Type    string         `json:"type"`
	Query   string         `json:"query,omitempty"`
	Boards  []BoardListing `json:"boards"`
}

type BookmarkListMessage struct {
	Command   string     `json:"cmd"`
	Type      string     `json:"type"`
	Bookmarks []Bookmark `json:"bookmarks"`
}

type Bookmark struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Query string `json:"query"`
}

// format for threads in "list"
type ThreadListing struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Author       string   `json:"user,omitempty"`
	AuthorID     string   `json:"user_id,omitempty"`
	Date         string   `json:"date,omitempty"`
	PostCount    int      `json:"posts,omitempty"`
	UnreadPosts  int      `json:"unread_posts,omitempty"`
	Sticky       bool     `json:"sticky,omitempty"` //a sticky (aka pinned) topic
	Closed       bool     `json:"closed,omitempty"` //a closed (aka locked) topic
	Tags         []string `json:"tags,omitempty"`   //option: "tags"
	PictureURL   string   `json:"img,omitempty"`    //option: "imageboard"
	ThumbnailURL string   `json:"thumb,omitempty"`  //option: "imageboard"
}

// format for boards in "list"
type BoardListing struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"desc,omitempty"`
	ThreadCount int    `json:"threads,omitempty"`
	PostCount   int    `json:"posts,omitempty"`
	Date        string `json:"date,omitempty"`
}
