package main

import (
	"log"
	"net/http"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/bolt"
	"github.com/bobinette/papernet/gin"
)

var md = `# Papernet example

## Markdown

In papernet, a paper summary is written in markdown. It is the rendered pretty well by
the front as you can see.

Markdown allows you to put **some text in bold**, other parts *in italic*. You can even add
lists:

* paper
* rock
* cisor

1. and
2. numbered
3. lists

> You can have some blocks

and some code:
	` + "```" + `python
	a = 1
	b = a + 2
	` + "```" + `
is rendered as
` + "```" + ` python
a = 1
b = a + 2
` + "```" + `

## Equations

The interface enhances basic markdown with latex equation support through code blocks:
	` + "```" + `equation
	\sum_{i=0}^{n}{i}
	` + "```" + `
is rendered as
` + "```" + `equation
\sum_{i=0}^{n}{i}
` + "```"

func main() {
	// Create repository
	repo := bolt.PaperRepository{}
	err := repo.Open("data/papernet.db")
	defer repo.Close()
	if err != nil {
		log.Fatalln("could not open db:", err)
	}

	paper := papernet.Paper{
		ID:      1,
		Title:   "Dev paper",
		Summary: md,
	}
	if err := repo.Upsert(&paper); err != nil {
		log.Fatalln("error inserting dev paper", err)
	}

	// Start web server
	handler, err := gin.New(&repo)
	if err != nil {
		log.Fatalln("could not start server:", err)
	}

	addr := ":1705"
	log.Println("server started, listening on", addr)
	log.Fatal(http.ListenAndServe(addr, handler))
}
