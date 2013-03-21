package main

import "github.com/tiko-chan/bbs"
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

const client_version string = "test-client 0.1"

func main() {
	testClient()
}

func testClient() {
	if len(os.Args) > 1 {
		bbsServer = os.Args[1]
	}

	fmt.Println("Running test client.")
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
	if len(line) == 0 {
		return
	}
	fields := strings.Fields(line)

	switch fields[0] {
	case "help":
		fmt.Println("COMMANDS\n\tlogin username password\n\tnigol password username\n\tget topicID [filterID]\n\tlist [expression]\n\treply [topicID] [text]\n\tquit\n\thelp")
	case "quit":
		os.Exit(0)
	case "exit":
		os.Exit(0)
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
	case "get":
		if len(fields) == 2 {
			doGet(fields[1], &bbs.Range{1, 50}, "")
		} else if len(fields) == 4 {
			lwr, _ := strconv.Atoi(fields[2])
			hrr, _ := strconv.Atoi(fields[3])
			doGet(fields[1], &bbs.Range{lwr, hrr}, "")
		} else {
			fmt.Println("Input error.")
			fmt.Println("usage: get topicID [filterID]")
		}
	case "list":
		if len(fields) == 1 {
			doList("")
		} else if len(fields) == 2 {
			doList(fields[1])
		} else {
			fmt.Println("Input error.")
			fmt.Println("usage: list [expression]")
		}
	case "reply":
		args := strings.SplitN(line, " ", 3)
		if len(args) < 3 {
			fmt.Println("Input error.")
			fmt.Println("usage: reply topicID text...")
		} else {
			doReply(args[1], args[2])
		}
	default:
		fmt.Println("What?")
	}
}

func doLogin(u, pw string) {
	login, _ := json.Marshal(&bbs.LoginCommand{"login", u, pw, 0, client_version})
	send(login)
}

func doList(exp string) {
	list, _ := json.Marshal(&bbs.ListCommand{"list", session, "thread", exp})
	send(list)
}

func doGet(t string, r *bbs.Range, filter string) {
	get, _ := json.Marshal(&bbs.GetCommand{"get", session, t, r, filter})
	send(get)
}

func doReply(id, text string) {
	reply, _ := json.Marshal(&bbs.ReplyCommand{"reply", session, id, text, "html"})
	send(reply)
}

func getURL(url string) string {
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func parse(js []byte) {
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
		fmt.Println("(OK: " + m.ReplyTo + ") " + m.Message)
	case "list":
		m := bbs.ListMessage{}
		json.Unmarshal(js, &m)
		onList(&m)
	}
}

func onError(msg *bbs.ErrorMessage) {
	prettyPrint("Error on: "+msg.ReplyTo, msg.Error)
}

func onHello(msg *bbs.HelloMessage) {
	prettyPrint("Connected", fmt.Sprintf("Name: %s\nDesc: %s\nOptions: %s\nVersion: %d\nServer: %s\n", msg.Name, msg.Description, msg.Options, msg.ProtocolVersion, msg.ServerVersion))
}

func onWelcome(msg *bbs.WelcomeMessage) {
	prettyPrint("Welcome "+msg.Username, "Session: "+msg.Session)
	session = msg.Session
}

func onMsg(msg *bbs.ThreadMessage) {
	fmt.Printf("Thread: %s [%d] \n  Tags: %s \n", msg.Title, len(msg.Messages), strings.Join(msg.Tags, ", "))
	for i, m := range msg.Messages {
		fmt.Printf("#%d User: %s | Date: %s | UserID: %s \n", i+msg.Range.Start, m.Author, m.Date, m.AuthorID)
		fmt.Println(m.Text + "\n")
	}
}

func onList(msg *bbs.ListMessage) {
	prettyPrint("Threads", msg.Query)
	for _, t := range msg.Threads {
		fmt.Printf("#%s [%s] %s | %d posts | %s | %s\n", t.ID, t.Author, t.Title, t.PostCount, t.Date, strings.Join(t.Tags, ", "))
	}
}

func send(js []byte) {
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
