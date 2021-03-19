package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/JesusIslam/tldr" //Text summarizer for golang using LexRank by Andida Syahendar
)

//variables for the home page - just the one for today!
type webPageVariables struct {
	Output string
}

var (
	//channel to quit the server
	quit = make(chan int, 1)
	//the HTML for the front page
	homepage = `
		<!DOCTYPE html><html>
		<head><title>TL;DR</title></head>
		<body>
		<form action="/" method="POST">
		TL;DR >> paste in large amounts of text (limit 7000 chars) and this will try to summarize the information from that text.<br>
		<br><input name="title" type="text" placeholder="length" value="1"><div>1 for smallest length, 2+ for more.</div>
		<br>
		<textarea id="area" maxlength="7000" name="entertexthere" cols="50" rows="25" placeholder="content goes here">{{.Output}}</textarea>
		<br>
		<button type="submit" name="submit" value="submitquery">Submit</button><button type="submit" name="submit" value="exit">Exit</button>
		</form>
		</body>
		</html>
`
)

//function that summarizes text and returns a list of paragraphs.
func tealdeer(sentencecount int, input string) (paragraphs string) {
	bag := tldr.New()
	output, err := bag.Summarize(input, sentencecount)
	if err != nil {
		fmt.Println(err)
	}
	for index := range output {
		if index != len(output) {
			paragraphs += output[index] + "\n\n"
		} else {
			paragraphs += output[index]
		}
	}
	return
}

//validate the user input in the forms on the web page
func validate(r *http.Request, item string) bool {
	if len(r.Form[item]) != 0 && r.Form[item][0] != "" {
		return true
	}
	return false
}

//if the error is not nil print error and handle
func errorexists(err error, str string) bool {
	if err != nil {
		log.Print(str, err) //log it
		return true
	}
	return false
}

//the function that is invoked if we go to '/' handler on server.
func startPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("homepage").Parse(homepage)
	if errorexists(err, "template parse error: ") {
		quit <- 1
	}
	pagevariables, exit := collectDataFromForms(r)
	err = t.Execute(w, pagevariables)
	if errorexists(err, "template execute error: ") {
		quit <- 1
	}
	if exit {
		quit <- 1
	}
}

//collect data from the forms in the http request
func collectDataFromForms(r *http.Request) (webPageVariables, bool) {
	var exit bool
	blogpost := webPageVariables{}
	r.ParseForm()
	button := r.FormValue("submit")
	switch button {
	case "exit":
		exit = true
		blogpost.Output = "GOODBYE!! Server has shut down... you can close this window!"
	default:
		if validate(r, "entertexthere") && validate(r, "title") {
			integer, _ := strconv.Atoi(string(r.Form["title"][0][0]))
			blogpost.Output = tealdeer(integer, r.Form["entertexthere"][0])
			log.Print(integer, " paragraph to be generated") //log a text summary attempt
		} else {
			log.Print("? no content received") //log no content received
		}
	}
	return blogpost, exit
}

func main() {
	fmt.Println("Server started...") //signposting the server has started
	http.HandleFunc("/", startPage)  //set up a http handler for the handle of '/' which will call function 'startPage'
	//run the webserver in a go routine
	go func() {
		err := http.ListenAndServe(":8080", nil) // setting up server on listening port 8080
		if errorexists(err, "http server error: ") {
			quit <- 1
		}
	}()
	//block main from exiting until we've received a message from the quit channel.
	select {
	case _, ok := <-quit:
		if ok {
			fmt.Println("Goodbye!")
		}
	}
}
