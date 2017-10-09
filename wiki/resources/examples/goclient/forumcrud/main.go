package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/dgraph-io/dgraph/client"
	"github.com/dgraph-io/dgraph/protos"
	"google.golang.org/grpc"
)

// If omitempty is not set, then edges with empty values (0 for int/float, "" for string, false
// for bool) would be created for values not specified explicitly.

// If the user desires setting empty values, then they should use pointers and set it explicitly
// when they want to.

type Forum struct {
	Id          uint64   `json:"_uid_,omitempty"`
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Threads     []Thread `json:"threads,omitempty"`
}

type Thread struct {
	Id uint64 `json:"_uid_,omitempty"`
	// is an openid-connect id of a subject
	AuthorId string `json:"author_id,omitempty"`
	Title    string `json:"title,omitempty"`
	Preview  string `json:"preview,omitempty"`
	Posts    []Post `json:"posts,omitempty"`
}

type Post struct {
	Id           uint64 `json:"_uid_,omitempty"`
	AuthorId     string `json:"author_id,omitempty"`
	ParentPostId string `json:"parent_post_id,omitempty"`
	Title        string `json:"title,omitempty"`
	Body         string `json:"body,omitempty"`
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type Root struct {
	Forum Forum `json:"forum"`
}

func unmarshalAndPrint(resp *protos.Response) {
	var r Root
	err := client.Unmarshal(resp.N, &r)
	checkErr(err)

	b, err := json.MarshalIndent(r.Forum, "", " ")
	checkErr(err)
	fmt.Printf("Response for forum query\n %s\n\n", b)
}

func main() {
	conn, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	checkErr(err)
	defer conn.Close()

	clientDir, err := ioutil.TempDir("", "client_")
	checkErr(err)
	defer os.RemoveAll(clientDir)

	dgraphClient := client.NewDgraphClient([]*grpc.ClientConn{conn}, client.DefaultOptions, clientDir)

	req := client.Req{}

	fmt.Println("Creating a forum and associated threads and posts.\n")
	// Creating a new forum with associated threads and posts.
	// The forum/thread/post don't set the Id for the objects. So a new forum would be created
	// with a thread and posts.
	f := Forum{
		Name:        "My forum",
		Description: "Forum Description",
		Threads: []Thread{
			Thread{
				AuthorId: "author",
				Title:    "How to build an App?",
				Preview:  "This is how you do it.",
				Posts: []Post{
					Post{
						AuthorId: "author",
						Title:    "Using the Go client",
					},
					Post{
						AuthorId: "author-2",
						Title:    "Using the HTTP API",
					},
				},
			},
		},
	}

	err = req.SetObject(&f)
	checkErr(err)

	resp, err := dgraphClient.Run(context.Background(), &req)
	checkErr(err)

	// The assigned uids for various entities would be returned in the AssignedUids map.
	// First uid would belong to the forum
	forumId := resp.AssignedUids["blank-0"]

	// Lets query for the forum using its uid and everything else associated with it.
	// You can also have indexes on name, description and then use the Term matching functions
	// provided by Dgraph to do the query.
	forumQuery := fmt.Sprintf(`
	{
		forum(func: uid(%d)) {
			_uid_
			name
			description
			threads {
				_uid_
				title
				preview
				author_id
				posts {
					_uid_
					author_id
					title
				}
			}
		}
	}
	`, forumId)

	fmt.Println("Running query to get created forum.")
	req.SetQuery(forumQuery)
	resp, err = dgraphClient.Run(context.Background(), &req)
	checkErr(err)

	unmarshalAndPrint(resp)

	fmt.Printf("Adding another thread with a post to forum with id: %d\n", forumId)
	fmt.Println("Running query to get updated forum.")
	// Lets now add another thread and post to the forum.
	// We are supplying the forum Id, so this would add the thread to the existing forum.
	f = Forum{
		Id: forumId,
		Threads: []Thread{
			Thread{
				AuthorId: "author-2",
				Title:    "What are the two components of Dgraph API?",
				Posts: []Post{
					Post{
						AuthorId: "author-2",
						Title:    "Queries and Mutation",
					},
				},
			},
		},
	}

	// Lets set the object and also the query to get back updated results.
	req.SetObject(f)
	req.SetQuery(forumQuery)
	resp, err = dgraphClient.Run(context.Background(), &req)
	checkErr(err)
	unmarshalAndPrint(resp)
	threadId := resp.AssignedUids["blank-0"]

	fmt.Printf("Updating thread with id: %d\n", threadId)
	// Lets now update the newly created thread.
	t := Thread{
		Id:      threadId,
		Preview: "Preview for the thread",
	}
	req.SetObject(t)
	req.SetQuery(forumQuery)
	resp, err = dgraphClient.Run(context.Background(), &req)
	checkErr(err)
	unmarshalAndPrint(resp)

	// Lets aggregate all the forum, thread and posts nodes and delete them.
	type Node struct {
		Id uint64 `json:"_uid_"`
	}

	var nodes []Node
	nodes = append(nodes, Node{forumId})

	var r Root
	err = client.Unmarshal(resp.N, &r)
	checkErr(err)

	// Recursively get all ids for deletion.
	for _, t := range r.Forum.Threads {
		nodes = append(nodes, Node{t.Id})
		for _, p := range t.Posts {
			nodes = append(nodes, Node{p.Id})
		}
	}

	fmt.Println("Deleting the forum along with its associated threads and posts.")
	req.DeleteObject(nodes)
	req.SetQuery(forumQuery)
	resp, err = dgraphClient.Run(context.Background(), &req)
	checkErr(err)
	unmarshalAndPrint(resp)
}
