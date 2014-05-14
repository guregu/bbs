# BBS Protocol 
(first draft)

The BBS protocol is a generic JSON protocol designed to fit most needs of various kinds of message boards (internet forums, bulletin boards, etc.)
It is designed to be lightweight, and should be as painless as possible to implement.
It could be used for a native forums browsing app. 

#### This project
This project is a simple server that speaks BBS. It also contains a tiny command line client for testing purposes.
Some nice documentation is under construction as well. BSD licensed.

#### Software supporting BBS
* limbo, standalone server software: https://github.com/guregu/limbo
* Relay, a gateway server for ETI and 4chan, and more soon: https://github.com/guregu/relay
* bbs-client, a web client aimed at mobile: https://github.com/guregu/bbs-client

#### Is the protocol stable?
No, but it's getting there.
See: https://github.com/guregu/bbs/issues/1




# Protocol Reference

A message board (or "BBS") has a single endpoint, which we shall refer to as the *BBS endpoint URL*. 
Example: 
```
http://bbs.tiko.jp/bbs
http://somesite.com/boards/bbs-gateway.php
```

Communication between the client and server currently consists of HTTP POSTs. The whole body of the POST should be a JSON object. The entire reply should also be a JSON object. You can put whatever you'd like there for GET requests (for example, a web client). 

Ideally I would rather have a nice RESTful interface, but this allows for:
* No need to impose our URL structure on other sites (just plop down bbs-gateway.php, or whatever)
* We can have the same wire protocol for websockets, which will be included in the next version of the protocol. 

JSON Structure
--------------
Every request and response should take the form of a JSON object like:

```
{
	"cmd": "[command name]"
    ...
}
```

Where [command name] is one of the following commands:

| Client commands | Server commands |
| --------------- | --------------- |
| [hello](#hello-command-client--server) | [hello](#hello-command-server--client) |
| [login](#login-command-client--server) | [welcome](#welcome-command-server--client) |
| [logout](#logout-command-client--server) | [msg](#msg-command-server--client) |
| [get](#get-command-client--server) | [list](#list-command-server--client) |
| [list](#list-command-client--server) | [ok](#ok-command-server--client) |
| [post](#post-command-client--server) | [error](#error-command-server--client) |
| [reply](#reply-command-client--server) | |

The rest of the object's contents depend on what kind of command it is.
Since this is an extensible protocol, *clients should silently ignore fields they don't understand*.


Request Flow
------------
| Client request command | Possible server responses |
| ---------------------- | ------------------------- |
| [hello](#hello-command-client--server) | [hello](#hello-command-server--client) |
| [login](#login-command-client--server) | [welcome](#welcome-command-server--client), [error](#error-command-server--client) |
| [logout](#logout-command-client--server) | [ok](#ok-command-server--client) |
| [get](#get-command-client--server) | [msg](#msg-command-server--client), [error](#error-command-server--client) |
| [list](#list-command-client--server) | [list](#list-command-server--client), [error](#error-command-server--client) |
| [post](#post-command-client--server) | [ok](#ok-command-server--client), [error](#error-command-server--client) |
| [reply](#reply-command-client--server) | [ok](#ok-command-server--client), [error](#error-command-server--client) |


## "hello" command (client → server)
Politely greets the server, requesting some general information.
The reply should be a "hello" command.

### Fields 
None.

### Example
```json
{
	"cmd": "hello"
}
```

### Notes
Your client will (probably) have a list of BBSs the user has added. You can 'hello' each one to figure out the server name, options, etc., also check if a server is up.

## "hello" command (server → client)
Responds with information about the BBS.

### Fields
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| name | string | required | | Server name. |
| version | int | required | | Highest supported BBS protocol version (right now: 0) |
| desc | string | required | | Server description. |
| secure | string | optional | | HTTPS URL to a secure connection for this BBS |
| options | string array | optional | | The options this server supports. (See Options section) |
| access | object | required | | Describes which commands require login (see below) |
| format | string array | required | | Formats this server understands, with the preferred format first. (See Formats section) |
| lists | string array | required | | Describes the lists available (see "list" command)
| server | string | required | | Server version string, can be anything. |

#### `access` object
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| guest | string array | required | | Commands that don't require logging in. |
| user | string array | required | | Commands that require logging in. |

### Example
```json
{
    "cmd": "hello",
    "name": "ETI Gateway",
    "version": 0,
    "desc": "ETI -> BBS Gateway",
    "options": [
        "tags",
        "avatars",
        "usertitles",
        "filter"
    ],
    "access": {
        "guest": [
            "hello",
            "login",
            "logout"
        ],
        "user": [
            "get",
            "list",
            "post",
            "reply"
        ]
    },
    "format": [
        "html",
        "text"
    ],
    "lists": [
        "thread",
        "tag"
    ],
    "server": "eti-relay 0.1"
}
```

### Notes
Every server is required to respond to this command, it help clients set up their layout and know what your server does.


## "login" command (client → server)

Used to log in. The reply will be a "welcome" command on success or an "error" command on error. The "welcome" reply should contain a "session" string, which the client will attach to commands from then on out.
There is currently no way for servers to specify whether they need a username and password, just one, or none. Expect this command to expand soon. For now, usernames and passwords are required.

### Fields
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| username | string | required | | User name. |
| password | string | required | | Password. | 
| version | int | required | | Client's requested protocol version, ideally the server should honor this. | 

### Example
```json
{
    "cmd": "login",
    "username": "Llamaguy",
    "password": "hunter2",
    "version": 0
}
```

### Notes
For servers that require you to log in to post or even read messages, this is a useful command. For read only public servers or anonymous boards, you won't need this. Clients should use a secure (HTTPS) connection to log in if the server has them. Servers specify their secure URL in the "hello" command.

## "welcome" command (server → client)
Sent to clients as a response to the "login" command when their log in is successful. It contains a `session` token that the client should include in requests from now on. 

### Fields
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| session | string | required | | Session token. |
| username | string | optional | | The user's properly capitalized username. Optional for anonymous boards. |

### Example
```json
{
	"cmd": "welcome",
	"session": "3a53192bcdca028d285692a731b041e1",
	"username": "LlamaGuy"
}
```

### Notes
You must include this `session` token in further requests to preform them as a logged in user. The `username` field is useful to get a properly capitalized username, or to inform people of their username for websites that use e-mail for log in, etc.


## "error" command (server → client)
Sent as a response to client commands that did not successfuly do what they were supposed to. In this case, the `wrt` field should be the client command that failed. There is a special case that if the client sends an invalid (expired, etc.) `session` token, that `wrt` should be "session" instead.

### Fields
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| wrt | string | required | | The name of the command that failed, or "session" for session token issues. |
| error | string | optional | | A description of the error. |

### Notes
The "error" command is often sent when a client requests to do something that the server doesn't allow for. In this case, the client should generally display the error message. If `wrt` is "session", that means the client sent a bad `session` token, and should log in again.

## "logout" command (client → server)
Requests a log out. Currently, there is no confirmation of logging out actually happening. So, if someone not logged in logs out, the server is totally OK with that.

### Fields
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| session | string | required | | Session token. |

### Example
```json
{
	"cmd": "logout",
	"session": "3a53192bcdca028d285692a731b041e1"
}
```

### Notes
Servers should expire people's sessions after inactivity regardless. 


## "ok" command (server → client)
This command is sent to clients as a response to a successful command. Many commands, such as "get" have a specific command the server sends as a response (in this example: "msg") and this command is not for them. The "ok" command is sent as a response to commands that do not need more than a simple string in return. 

### Fields
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| wrt | string | required | | The name of the command that succeeded. |
| result | string | optional | | Any additional information, can be ignored or omitted. |

### Example
```json
{
	"cmd": "ok",
	"wrt": "logout"
}
```

### Notes
This command is sent as a response to many different client commands to indicate success. For "post" and "reply", `result` should be the new thread ID or post ID, but it is allowed to be missing.

## "get" command (client → server)
This command is used to request messages from the server. It is used for viewing threads. Servers respond with a "msg" command or an "error" command.

### Fields
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| id | string | required | | Desired thread's ID. |
| session | string | optional | | Session token, if logged in. |
| range | object | optional | range | Desired message range, see below. |
| filter | string | optional | filter | The user ID whose messages you want. |
| format | string | optional | | The desired format. Omit for server default. |

#### `range` object
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| start | int | required | | Post number to start from. 1-indexed (1 is the OP) |
| end | int | required | | Post number to end at, inclusive. 

### Example
```json
{
    "cmd": "get",
    "id": "8382679",
    "session": "3a53192bcdca028d285692a731b041e1",
    "range": {
        "start": 1,
        "end": 50
    },
    "format": "text"
}
```

### Notes
For servers that don't support the "range" option, this command will get every post. For servers that do, it will get the default range if `range` is omitted. 


## "msg" command (server → client)
This command is used to send messages (posts) to the client. It is used to display a thread. It is the response to a "get" command. In the future, it will be expanded with real-time features.

### Fields
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| id | string | required | | Thread ID |
| title | string | optional | | Thread title, if any. |
| range | object | optional | range | Range describing the enclosed messages, if any. |
| closed | boolean | optional | | If true, it means the topic can't be posted in. |
| filter | string | optional | filter | The user ID filtered by, if any. |
| board | string | required* | boards | The board this topic belongs to. |
| tags | string array | optional | tags | The tags this thread is associated with, if any. |
| format | string | optional | | The format the following posts are in. Default format if omitted. |
| messages | object array | required | | Posts. See below. |
| more | boolean | optional | | Are there more posts available? |

#### `messages` objects
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| id | string | required | | Message ID |
| user | string | required? | | Username | 
| user_id | string | optional | | User ID used for filtering threads, getting profiles, etc. |
| date | string | optional | | The date, in no particular format. |
| body | string | required | | The post body. Message text. The content |
| sig | string | optional | | Post signature. Think USENET signatures, not file signatures. |
| user_title | string | optional | usertitles | A little blurb of text. "Custom Title" on SA. |
| avatar | string | optional | avatars | URL to user's full sized avatar. |
| avatar_thumb | string | optional | avatars | URL to user's thumbnail avatar. |
| img | string | optional | imageboard | This post's attached image. |
| thumb | string | optional | imageboard | This post's attached image's thumbnail. | 

### Example
```json
{
    "cmd": "msg",
    "id": "8382679",
    "title": "Who's excited for pizza tomorrow?",
    "range": {
        "start": 1,
        "end": 4
    },
    "tags": [
        "Pizza"
    ],
    "format": "text",
    "messages": [
        {
            "id": "m122677227",
            "user": "scofflaw",
            "user_id": "22490",
            "date": "3/22/2013 09:12:32 AM",
            "body": "i wonder who's gonna win the weekly pizza lottery",
            "sig": "- Scofflaw"
        },
        {
            "id": "m122678244",
            "user": "zekachu",
            "user_id": "21735",
            "date": "3/22/2013 09:33:10 AM",
            "body": ":o i cant wait",
            "sig": "http://zeke.fm\nbirds flying high you know how i feel"
        },
        {
            "id": "m122678304",
            "user": "Johnnybigone",
            "user_id": "9063",
            "date": "3/22/2013 09:34:11 AM",
            "body": "oh man oh man oh man\n"
        },
        {
            "id": "m122678349",
            "user": "Big the Cat500",
            "user_id": "6315",
            "date": "3/22/2013 09:34:53 AM",
            "body": "spoilers zekachu is gonna win it",
            "sig": "Eternal Pirate Gai\nSkies of Arcadia >>>>>> Final Fantasy. Try playing it, sometime.",
            "user_title": "Eternal Pirate"
        }
    ]
}
```

### Notes
This is the meat of your BBS. You can omit nearly everything. I'm leaning towards names being required. Even on anonymous boards, you can set the name as "Anonymous" for everyone. 
There is no way to request single posts right now, without using `range`. 


## "list" command (client → server)
Asks for lists. Like a thread list or a board list.
The lists a server supports are given in the "hello" command (`lists`). 

### Fields
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| type | string | required | | The kind of list requested. ("thread", "board", "tag"...)  |
| query | string | optional | | For thread lists: the board ID/tag expression. Or blank/missing. | 
| session | string | optional | | Session token. | 

### Example
```json
{
	"cmd": "list",
	"type":"thread",
	"query":"Pizza&Crime",
	"session":"3a53192bcdca028d285692a731b041e1"
}
```

## "list" command (server → client)
The response to a client's "list" command. 

### Fields
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| type | string | required | | List type (see client "list" command) |
| query | string | optional | | The client's query, if any. |
| threads | object array | required* | | The thread list. Required when `type` is "thread". See below. |
| boards | object array | required* | boards | Board list. Required when `type` is "board". See below. |

#### `threads` object (thread listing)
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| id | string | required | | Thread ID |
| title | string | required | | Thread title. Can be blank. |
| user | string | optional | | Username of the thread author. |
| user_id | string | optional | | User ID of the thread author. |
| date | string | optional | | Some kind of date. Could be the last post date. Could be the thread creation date, whichever works for you. |
| posts | int | optional | | Post count. |
| unread_posts | int | optional | | Unread post count. | 
| sticky | bool | optional | | True if this thread is stickied (pinned), false or omitted otherwise. |
| closed | bool | optional | | True if this thread is closed for posting, false or omitted otherwise. |
| tags | string array | optional | tags | Tags associated with this topic, if any. |
| img | string | optional | imageboard | URL for the image attached to the thread, if any. |
| thumb | string | optional | imageboard | Thumbnail URL for the image attached to the thread. |

#### `boards` object (board listing)
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| id | string | required | boards | Board ID |
| name | string | optional | boards | Board name |
| desc | string | optional | boards | Board description |
| threads | int | optional | boards | Thread count |
| date | string | optional | boards | Some kind of date (last post, usually). |


### Examples
Thread list:
```json
{
    "cmd": "list",
    "type": "thread",
    "query": "Pizza&Crime",
    "threads": [
        {
            "id": "8331156",
            "title": "ITT Wolfpac comments on the Christopher Dorner manifesto while eating pizza",
            "user": "Wolfpac",
            "user_id": "282",
            "date": "2/12/2013 16:20",
            "posts": 9,
            "tags": [
                "LUE",
                "Anarchism"
            ]
        }
    ]
}
```

Board list:
```json
{
    "cmd": "list",
    "type": "board",
    "boards": [
        {
            "id": "3",
            "name": "/3/ - 3DCG"
        },
        {
            "id": "a",
            "name": "/a/ - Anime & Manga"
        },
        {
            "id": "adv",
            "name": "/adv/ - Advice"
        },
        {
            "id": "an",
            "name": "/an/ - Animals & Nature"
        },
        {
            "id": "asp",
            "name": "/asp/ - Alternative Sports"
        },
        {
            "id": "b",
            "name": "/b/ - Random (NWS)"
        }
    ]
}
```

## "post" command (client → server)
Used for posting new threads. 

### Fields
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| title | string | required | | New thread title. |
| body | string | required | | New thread body (post content) |
| format | string | optional | | Format this is in, or default format is omitted. |
| board | string | required* | boards | The board to post to. Required for "boards" option servers. |
| tags | string array | optional | tags | The tags to associate with the new thread. |
| session | string | optional | | Session token. |

### Example
```json
{
	"cmd": "post",
	"title": "Hello everyone.",
	"body": "Hey guys I'm new here XD",
	"format": "text",
	"session": "3a53192bcdca028d285692a731b041e1"
}
```
### Notes
None.

## "reply" command (client → server)
Used for replying to existing threads. 

### Fields
| Field name | Type | Required? | Option | Description |
| ---------- | ---- | --------- | ------ | ----------- |
| to | string | required | | Thread ID to reply to |
| body | string | required | | New thread body (post content) |
| format | string | optional | | Format this is in, or default format is omitted. |
| session | string | optional | | Session token. |

### Example
```json
{
	"cmd": "reply",
	"to": "q/2312321",
	"body": "Why u guys delete my post?????",
	"format": "text",
	"session": "3a53192bcdca028d285692a731b041e1"
}
```

### Notes
None.
