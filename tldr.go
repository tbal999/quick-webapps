package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"text/template"

	"github.com/JesusIslam/tldr" //Text summarizer for golang using LexRank by Andida Syahendar
)

//variables for the home page - just the one!
var pagevariables = struct {
	Output string
}{}

var (
	//bool channel to quit the server with a buffer of 1 (we only need 1!)
	quit = make(chan bool, 1)
	//the HTML for the front page
	homepage = `
		<!DOCTYPE html><html>
		<head><title>TL;DR</title></head>
		<body>
		<form action="/doStuff" method="POST">
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
func tealDeer(sentencecount int, input string) (paragraphs string) {
	bag := tldr.New()
	output, err := bag.Summarize(input, sentencecount)
	if err != nil {
		fmt.Println(err)
	}
	for index := range output {
		if index < len(output)-1 {
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
func errorExists(err error, str string) bool {
	if err != nil {
		log.Print(str, err) //log it
		return true
	}
	return false
}

//the function that is called if we go to '/' handler.
func startPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("homepage").Parse(homepage)
	if errorExists(err, "template parse error: ") {
		quit <- true
	}
	err = t.Execute(w, pagevariables)
	if errorExists(err, "template execute error: ") {
		quit <- true
	}
}

//the function that is called if we go to '/doStuff' handler.
func doStuff(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("homepage").Parse(homepage)
	if errorExists(err, "template parse error: ") {
		quit <- true
	}
	exit := collectDataFromForms(r) //grab data from the input
	err = t.Execute(w, pagevariables)
	if errorExists(err, "template execute error: ") {
		quit <- true
	}
	if exit {
		quit <- true
	}
}

//collect data from the forms in the http request
func collectDataFromForms(r *http.Request) bool {
	var exit bool
	r.ParseForm()
	button := r.FormValue("submit")
	switch button {
	case "exit":
		exit = true
		pagevariables.Output = `GOODBYE!! Server has shut down... 
you can close this window now!`
	default:
		if validate(r, "entertexthere") && validate(r, "title") {
			integer, _ := strconv.Atoi(string(r.Form["title"][0][0]))
			pagevariables.Output = tealDeer(integer, r.Form["entertexthere"][0])
			log.Print(integer, " paragraph(s) to be generated") //log a text summary attempt
		} else {
			log.Print("? no content received") //log no content received
		}
	}
	return exit
}

//opens your default browser, depending on the OS you are on.
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform????")
	}
	if err != nil {
		log.Print(err)
	}
}

func main() {
	fmt.Println("Server started...")     //signposting the server has started
	http.HandleFunc("/", startPage)      //set up a http handler for the handle of '/' which will call function 'startPage'
	http.HandleFunc("/doStuff", doStuff) //set up a http handler for the handle of '/doStuff' which will call function 'doStuff'
	//run the webserver in a go routine
	go func() {
		err := http.ListenAndServe(":8080", nil) // setting up server on listening port 8080
		if errorExists(err, "http server error: ") {
			quit <- true
		}
	}()
	openBrowser("http://127.0.0.1:8080") //open browser (or tab) for the app automatically
	//block main from exiting until we've received a message from the quit channel.
	select {
	case _, ok := <-quit:
		if ok {
			fmt.Println("...Shutting down")
			fmt.Println("Goodbye!")
		}
	}
}
