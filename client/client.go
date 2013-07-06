package main

import "github.com/guregu/bbs"
import "net/http"
import "encoding/json"
import "fmt"
import "io/ioutil"
import "bytes"
import "strings"
import "bufio"
import "os"
import "strconv"

var bbsServer = "http://localhost:8080/bbs"
var session = ""
var lastLine = ""
var verbose = true
var nextNext *next
var listNext *listnext

const client_version string = "test-client 0.1" //TODO: use this in User-Agent

type next struct {
	id    string
	token string
}

type listnext struct {
	query string
	token string
}

func main() {
	testClient()
}

func testClient() {
	if len(os.Args) > 1 {
		bbsServer = os.Args[1]
	}

	fmt.Println("Running test client: " + client_version)
	fmt.Printf("Connecting to %s...\n", bbsServer)
	hello, _ := json.Marshal(&bbs.BBSCommand{"hello"})
	send(hello)

	r := bufio.NewReader(os.Stdin)
	fmt.Println("Type help for a list of commands or quit to exit.")
	fmt.Print("> ")
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		input(line)
		fmt.Print("> ")
	}
}

func input(line string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return
	}

	switch fields[0] {
	case ".":
		input(lastLine)
	case "help":
		fmt.Println("COMMANDS\n\tlogin username password\n\tnigol password username\n\tget topicID [filterID]\n\tboards\n\tlist [expression]\n\treply [topicID] [text]\n\tquit\n\thelp\n\t. (repeat last command)")
	case "quit":
		os.Exit(0)
	case "exit":
		os.Exit(0)
	case "register":
		if len(fields) == 3 {
			doRegister(fields[1], fields[2])
		} else {
			fmt.Println("Input error.")
			fmt.Println("usage: register username password")
		}
	case "login":
		args := strings.SplitN(line, " ", 3)
		if strings.Contains(args[2], " ") {
			fmt.Println("WARNING: If you have spaces in your username, you should use 'nigol' instead.")
		}
		if len(fields) != 3 {
			fmt.Println("Input error.")
			fmt.Println("usage: login username password")
		}
		doLogin(fields[1], fields[2])
		lastLine = line
	case "nigol":
		args := strings.SplitN(line, " ", 3)
		if strings.Contains(args[2], " ") {
			fmt.Println("WARNING: If you have spaces in your password, you should use 'login' instead. If you have spaces in both, you're SOL.")
		}
		if len(fields) != 3 {
			fmt.Println("Input error.")
			fmt.Println("usage: nigol password username")
		}
		doLogin(fields[2], fields[1])
		lastLine = line
	case "get":
		if len(fields) == 2 {
			doGet(fields[1], &bbs.Range{1, 50}, "")
		} else if len(fields) == 4 {
			lwr, _ := strconv.Atoi(fields[2])
			hrr, _ := strconv.Atoi(fields[3])
			doGet(fields[1], &bbs.Range{lwr, hrr}, "")
		} else {
			fmt.Println("Input error.")
			fmt.Println("usage: get topicID [lower upper filter]")
			fmt.Println("usage: get board topicID")
		}
		lastLine = line
	case "list":
		if len(fields) == 1 {
			doList("")
		} else if len(fields) == 2 {
			doList(fields[1])
		} else {
			fmt.Println("Input error.")
			fmt.Println("usage: list [expression]")
		}
		lastLine = line
	case "boards":
		doListBoards()
		lastLine = line
	case "reply":
		args := strings.SplitN(line, " ", 3)
		if len(args) < 3 {
			fmt.Println("Input error.")
			fmt.Println("usage: reply topicID text...")
		} else {
			doReply(args[1], strings.Trim(args[2], " \n"))
		}
		lastLine = line
	case "post":
		args := strings.SplitN(line, " ", 3)
		if len(args) < 3 {
			fmt.Println("Input error.")
			fmt.Println("usage: reply title text...")
		} else {
			doPost(args[1], strings.Trim(args[2], " \n"))
		}
		lastLine = line
	case "next":
		if nextNext != nil {
			doGetNext(nextNext.id, nextNext.token)
		}
	case "listnext":
		if listNext != nil {
			doListNext(listNext.query, listNext.token)
		}
	default:
		fmt.Println("What?")
	}
}

func doRegister(u, pw string) {
	reg, _ := json.Marshal(&bbs.RegisterCommand{
		Command:  "register",
		Username: u,
		Password: pw,
	})
	send(reg)
}

func doLogin(u, pw string) {
	login, _ := json.Marshal(&bbs.LoginCommand{"login", u, pw, 0})
	send(login)
}

func doList(exp string) {
	list, _ := json.Marshal(&bbs.ListCommand{"list", session, "thread", exp, ""})
	send(list)
}

func doListBoards() {
	list, _ := json.Marshal(&bbs.ListCommand{"list", session, "board", "", ""})
	send(list)
}

func doListNext(query, token string) {
	nxt, _ := json.Marshal(&bbs.ListCommand{
		Command: "list",
		Session: session,
		Type:    "thread",
		Query:   query,
		Token:   token,
	})
	send(nxt)
}

func doGet(t string, r *bbs.Range, filter string) {
	get, _ := json.Marshal(&bbs.GetCommand{"get", session, t, r, filter, "text", ""})
	send(get)
}

func doGetNext(t string, n string) {
	nxt, _ := json.Marshal(&bbs.GetCommand{
		Command:   "get",
		Session:   session,
		ThreadID:  t,
		NextToken: n,
	})
	send(nxt)
}

func doReply(id, text string) {
	reply, _ := json.Marshal(&bbs.ReplyCommand{"reply", session, id, text, "text"})
	send(reply)
}

func doPost(title, text string) {
	post, _ := json.Marshal(&bbs.PostCommand{
		Command: "post",
		Session: session,
		Title:   title,
		Text:    text,
		Format:  "text"})
	send(post)
}

func getURL(url string) string {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func parse(js []byte) {
	if verbose {
		fmt.Println("server -> client")
		fmt.Println(string(js))
	}

	bbscmd := new(bbs.BBSCommand)
	err := json.Unmarshal(js, bbscmd)
	if err != nil {
		fmt.Println("JSON Parsing Error!! " + string(js))
		return
	}

	switch bbscmd.Command {
	case "error":
		m := bbs.ErrorMessage{}
		json.Unmarshal(js, &m)
		onError(&m)
	case "hello":
		m := bbs.HelloMessage{}
		json.Unmarshal(js, &m)
		onHello(&m)
	case "welcome":
		m := bbs.WelcomeMessage{}
		json.Unmarshal(js, &m)
		onWelcome(&m)
	case "msg":
		m := bbs.ThreadMessage{}
		json.Unmarshal(js, &m)
		onMsg(&m)
	case "ok":
		m := bbs.OKMessage{}
		json.Unmarshal(js, &m)
		fmt.Println("(OK: " + m.ReplyTo + ") " + m.Result)
	case "list":
		t := bbs.TypedMessage{}
		json.Unmarshal(js, &t)
		if t.Type == "thread" {
			m := bbs.ListMessage{}
			json.Unmarshal(js, &m)
			onList(&m)
		} else if t.Type == "board" {
			m := bbs.BoardListMessage{}
			json.Unmarshal(js, &m)
			onBoardList(&m)
		}
	}
}

func onError(msg *bbs.ErrorMessage) {
	prettyPrint("Error on: "+msg.ReplyTo, msg.Error)
}

func onHello(msg *bbs.HelloMessage) {
	prettyPrint("Connected", fmt.Sprintf("Name: %s\nDesc: %s\nOptions: %v\nVersion: %d\nServer: %s\n", msg.Name, msg.Description, msg.Options, msg.ProtocolVersion, msg.ServerVersion))
}

func onWelcome(msg *bbs.WelcomeMessage) {
	prettyPrint("Welcome "+msg.Username, "Session: "+msg.Session)
	session = msg.Session
}

func onMsg(msg *bbs.ThreadMessage) {
	fmt.Printf("Thread: %s [%d] \n  Tags: %s \n", msg.Title, len(msg.Messages), strings.Join(msg.Tags, ", "))
	if msg.Closed {
		fmt.Println("(Closed)")
	}
	for _, m := range msg.Messages {
		fmt.Printf("#%s User: %s | Date: %s | UserID: %s \n", m.ID, m.Author, m.Date, m.AuthorID)
		fmt.Println(m.Text + "\n")
	}
	if msg.More {
		nextNext = &next{msg.ID, msg.NextToken}
	}
}

func onList(msg *bbs.ListMessage) {
	prettyPrint("Threads", msg.Query)
	for _, t := range msg.Threads {
		info := ""
		if t.Closed && t.Sticky {
			info = "(Sticky, Closed)"
		} else if t.Closed {
			info = "(Closed)"
		} else if t.Sticky {
			info = "(Sticky)"
		}
		newposts := " "
		if t.UnreadPosts > 0 {
			newposts = fmt.Sprintf(" (unread: %d) ", t.UnreadPosts)
		}
		if msg.NextToken != "" {
			listNext = &listnext{msg.Query, msg.NextToken}
		}
		fmt.Printf("#%s [%s] %s %s | %d posts%s| %s | %s\n", t.ID, t.Author, info, t.Title, t.PostCount, newposts, t.Date, strings.Join(t.Tags, ", "))
	}
}

func onBoardList(msg *bbs.BoardListMessage) {
	prettyPrint("Boards", msg.Query)
	for i, b := range msg.Boards {
		fmt.Printf("#%d (%s) %s\n", i, b.ID, b.Name)
		if b.Description != "" {
			fmt.Println(b.Description)
		}
	}
}

func send(js []byte) {
	if verbose {
		fmt.Println("client -> server")
		fmt.Println(string(js))
	}
	resp, err := http.Post(bbsServer, "application/json", bytes.NewReader(js))
	if err != nil {
		fmt.Printf("Could not send data. %v\n", err)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	parse(body)
}

func prettyPrint(s1, s2 string) {
	fmt.Printf("### %s\n%s\n", s1, s2)
}
