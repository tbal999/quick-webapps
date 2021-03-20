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

//variables for the home page
var pagevariables = struct {
	Output string
}{}

var (
	quit     = make(chan bool, 1) //bool channel to quit the server with a buffer of 1 (we only need 1!)
	homepage =                    //some HTML for the front page
	` 
		<!DOCTYPE html>
		<html>
		<head><title>TL;DR</title></head>
		<body>
		<form action="/doStuff" method="POST">
		TL;DR >> paste in large amounts of text (limit 7000 chars) and this will try to summarize the information from that text.<br>
		<br><input name="size" type="text" placeholder="length" value="1"><div>output size: 1 (smallest) - 10 (biggest)</div>
		<br>
		<textarea id="area" maxlength="7000" name="entertexthere" cols="50" rows="25" placeholder="content goes here">{{.Output}}</textarea>
		<br>
		<button type="submit" name="submit" value="submitquery">Submit</button><button type="submit" name="submit" value="exit">Exit</button>
		</form>
		</body>
		</html>
	`
)

//TLDR function that summarizes text and returns a list of paragraphs.
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

//collect data from the forms in the http request
func collectDataFromForms(r *http.Request) bool {
	var exit bool
	r.ParseForm()                   //we parse the contents of the form
	button := r.FormValue("submit") //grab the button name so we can decide next action
	switch button {
	case "exit": //if you click on 'exit' button?
		exit = true
		pagevariables.Output = `GOODBYE!! Server will shut down... 
you can close this window now!`
	default: //there is only one other button so we can use default...
		if validate(r, "entertexthere") && validate(r, "size") { //if there's content in the text boxes...
			integer, _ := strconv.Atoi(string(r.Form["size"][0])) //if it's not a number just refresh page as var will just be 0
			if integer > 10 {                                     //limit the size!
				integer = 10
			}
			pagevariables.Output = tealDeer(integer, r.Form["entertexthere"][0])
			log.Print(integer, " size to be generated") //log a text summary attempt
		} else {
			log.Print("? no content received") //log no content received
		}
	}
	return exit
}

//the function that is called if we go to '/doStuff' handler.
func doStuff(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("homepage").Parse(homepage) //parse the contents of the page variable
	if errorExists(err, "template parse error: ") {
		quit <- true
	}
	exit := collectDataFromForms(r)   //grab data from the input (and return a bool)
	err = t.Execute(w, pagevariables) //execute the template, but passing in the (now updated) page variables
	if errorExists(err, "template execute error: ") {
		quit <- true
	}
	if exit { //if the 'exit' bool is true, we want to quit the server
		quit <- true
	}
}

//the function that is called if we go to '/' handler.
func startPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("homepage").Parse(homepage) //parse page
	if errorExists(err, "template parse error: ") {
		quit <- true
	}
	err = t.Execute(w, pagevariables) //execute with pagevariables (will be blank at start)
	if errorExists(err, "template execute error: ") {
		quit <- true //if we pass true via the quit channel, it will cause the app to exit
	}
}

//open up browser/tab dependent on your OS.
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

//if the error is not nil print error and handle it (tiny bit less errorcode everywhere)
func errorExists(err error, str string) bool {
	if err != nil {
		log.Print(str, err) //log it
		return true
	}
	return false
}

func waitForQuit() {
	select {
	case isTrue := <-quit: //if we receive a message from channel
		if isTrue { //if we receive 'true' down the channel...
			fmt.Println("...Shutting down")
			fmt.Println("Goodbye!")
			return //if we get here, it will exit to the main goroutine and shut down the app
		}
	}
}

func startHTTPServer(port string) {
	err := http.ListenAndServe(":"+port, nil) //setting up server on listening port
	if errorExists(err, "http server error: ") {
		quit <- true
	}
}

//main function
func main() {
	fmt.Println("Server started...")     //signpost that the server has started
	http.HandleFunc("/", startPage)      //set up a http handler for the handle of '/' which will call function 'startPage'
	http.HandleFunc("/doStuff", doStuff) //
	go startHTTPServer("8080")           //run a local http server on a seperate go routine on port 8080
	openBrowser("http://127.0.0.1:8080") //open a browser or tab automatically to go to the GUI
	waitForQuit()                        //block main goroutine from exiting until we've received a message from the quit channel.
}
