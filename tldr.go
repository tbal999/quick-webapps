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

//collect data from the forms in the http request
func collectDataFromForms(r *http.Request) bool {
	var exit bool
	r.ParseForm() //we parse the contents of the form
	button := r.FormValue("submit")
	switch button {
	case "exit":
		exit = true
		pagevariables.Output = `GOODBYE!! Server has shut down... 
you can close this window now!`
	default:
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
	t, err := template.New("homepage").Parse(homepage) //go templating package will parse the contents of the page variable
	if errorExists(err, "template parse error: ") {
		quit <- true
	}
	exit := collectDataFromForms(r)   //grab data from the input
	err = t.Execute(w, pagevariables) //execute the template, but passing in the (now updated) page variables
	if errorExists(err, "template execute error: ") {
		quit <- true
	}
	if exit {
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

//main goroutine
func main() {
	fmt.Println("Server started...")     //signposting the server has started
	http.HandleFunc("/", startPage)      //set up a http handler for the handle of '/' which will call function 'startPage'
	http.HandleFunc("/doStuff", doStuff) //set up a http handler for the handle of '/doStuff' which will call function 'doStuff'
	//run the webserver in a seperate go routine
	go func() {
		err := http.ListenAndServe(":8080", nil) // setting up server on listening port 8080
		if errorExists(err, "http server error: ") {
			quit <- true
		}
	}()
	openBrowser("http://127.0.0.1:8080") //open browser (or tab) for the app automatically
	//block main goroutine from exiting until we've received a message from the quit channel.
	select {
	case isTrue := <-quit: //if at any time we receive 'true' down the channel...
		if isTrue {
			fmt.Println("...Shutting down")
			fmt.Println("Goodbye!")
			//if we get here, it will now exit the main goroutine and shut down the app
		}
	}
}
