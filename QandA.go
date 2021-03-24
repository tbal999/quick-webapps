package main

/*
quick little 'Q&A' example app - will temporarily upload this to https://fern91.com/qna/ so people can play around with it and ask questions etc if they want to.
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"text/template"
	"time"
)

type page struct {
	Output          [][]string
	FormattedOutput string
}

//variables for the home page
var pagevariables page

//Save snapshot of data to json
func (p page) Save() {
	Base := &p
	output, err := json.MarshalIndent(Base, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ioutil.WriteFile("saveddata.json", output, 0644)
	if err != nil {
		fmt.Println(err)
	}
}

//Load snapshot of data to json
func (p *page) Load() {
	item := *p
	jsonFile, _ := ioutil.ReadFile("saveddata.json")
	_ = json.Unmarshal([]byte(jsonFile), &item)
	*p = item
}

var (
	quit     = make(chan bool, 1) //bool channel to quit the server with a buffer of 1 (we only need 1!)
	homepage =                    //some HTML for the front page
	` 
		<!DOCTYPE html>
		<html>
		<head>
		<title>Q&A</title>
		<style>
		.messages {
			color: brown;
			font-weight: bold;
			font-size: 25px;
			font-family: consolas;
		}
		</style>
		</head>
		<body>
		<form action="/doStuff" method="POST">
		<br><input name="name" type="text" placeholder="name">&nbsp&nbsp<input name="question" type="text" placeholder="question">&nbsp&nbsp<button type="submit" name="submit" value="nameandquestion">Submit</button>&nbsp&nbsp<button type="submit" name="submit" value="reverse">timbuS</button>&nbsp&nbsp<button type="submit" name="submit" value="gopher">Gophers, attack!</button>
		<br>
		<br>
		<div class="messages">
		{{.FormattedOutput}}
		</div>
		</form>
		</body>
		</html>
	`
)

//simple reverse a string function
func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func reveseFormattedOutput() {
	pagevariables.FormattedOutput = ""
	for index := range pagevariables.Output {
		pagevariables.FormattedOutput += reverse(pagevariables.Output[index][0]) + " << " + reverse(pagevariables.Output[index][1]) + "<br><br>"
	}
}

func refreshFormattedOutput() {
	pagevariables.FormattedOutput = ""
	for index := range pagevariables.Output {
		pagevariables.FormattedOutput += pagevariables.Output[index][0] + " >> " + pagevariables.Output[index][1] + "<br><br>"
	}
}

func filterFormattedOutput(filter string) {
	pagevariables.FormattedOutput = ""
	for index := range pagevariables.Output {
		if strings.Contains(pagevariables.Output[index][0], filter) || strings.Contains(pagevariables.Output[index][1], filter) {
			pagevariables.FormattedOutput += pagevariables.Output[index][0] + " >> " + pagevariables.Output[index][1] + "<br><br>"
		}

	}
}

func randomNumber(min, max int, seededRand *rand.Rand) int {
	z := seededRand.Intn(max)
	if z < min {
		z = min
	}
	return z
}

func bloopOutput() {
	var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	pagevariables.FormattedOutput = ""
	for index := range pagevariables.Output {
		var bloopchat string
		for index2 := range pagevariables.Output[index][1] {
			if string(pagevariables.Output[index][1][index2]) == " " {
				switch randomNumber(1, 4, seededRand) {
				case 1:
					bloopchat += ` <img src="https://fern91.com/gophers/go1.jpg"> `
				case 2:
					bloopchat += ` <img src="https://fern91.com/gophers/go2.jpg"> `
				case 3:
					bloopchat += ` <img src="https://fern91.com/gophers/go3.jpg"> `
				case 4:
					bloopchat += ` <img src="https://fern91.com/gophers/go4.jpg"> `
				}
			}
			bloopchat += string(pagevariables.Output[index][1][index2])
		}
		pagevariables.FormattedOutput += pagevariables.Output[index][0] + " >> " + bloopchat + "<br><br>"
	}
}

//collect data from the forms in the http request 
//have a read through this function and figure out how it all works
func collectDataFromForms(r *http.Request) bool {
	var exit bool
	r.ParseForm()                   //we parse the contents of the form
	button := r.FormValue("submit") //grab the button name so we can decide next action
	switch button {
	case "gopher":
		bloopOutput()
	case "reverse":
		if validate(r, "name") && validate(r, "question") {
			name := r.Form["name"][0]
			question := r.Form["question"][0]
			input := []string{name, question}
			pagevariables.Output = append(pagevariables.Output, input)
			reveseFormattedOutput()
			pagevariables.Save()
		} else {
			reveseFormattedOutput()
		}
	case "nameandquestion": 
		if validate(r, "name") { //if there's content in the boxes...
			name := r.Form["name"][0]
			switch r.Form["name"][0] {
			case "filter":
				if validate(r, "question") {
					question := r.Form["question"][0]
					filterFormattedOutput(question)
				}
			default:
				if validate(r, "question") {
					question := r.Form["question"][0]
					log.Print(question, " ", name)
					input := []string{name, question}
					pagevariables.Output = append(pagevariables.Output, input)
					refreshFormattedOutput()
					pagevariables.Save()
				}
			}
		} else {
			refreshFormattedOutput()
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
		quit <- true
	}
}

//validate the user input in the forms on the web page
func validate(r *http.Request, item string) bool {
	/*
		not too savvy with this bit myself - so i'm not sure why the form returns an array, and the data is always in the first index of that array.
		if anybody knows why, feel free to let me know via tom@fern91.com - i would be grateful for the knowledge.
	*/
	if len(r.Form[item]) != 0 && r.Form[item][0] != "" { //if form is not blank & the first array item is not empty..
		return true
	}
	return false
}

//if the error is not nil print error and handle it
func errorExists(err error, str string) bool {
	if err != nil {
		log.Print(str, err) //output error for reference
		return true
	}
	return false
}

//start up a HTTP server
func startHTTPServer(port string) {
	err := http.ListenAndServe(":"+port, nil) //setting up server on listening port
	if errorExists(err, "http server error: ") {
		quit <- true
	}
}

//open up browser/tab dependent on your OS.
func openBrowser(url string) {
	var err error
	switch runtime.GOOS { //open browser/tab dependent on what OS you are on
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform????")
	}
	if errorExists(err, "browser opening error: ") {
		quit <- true //if we pass true via the quit channel, it will cause the app to exit
	}
}

//wait for the message from the quit channel (not actually used on the frontend but kept it in anyway)
func waitForQuit() {
	select {
	case isTrue := <-quit: //if we receive a message from channel
		if isTrue { //if we receive 'true' down the channel...
			fmt.Println("...Shutting down")
			fmt.Println("Goodbye!")

		}
	}
	//if we get here, it will exit to the main goroutine and shut down the app
}

//main function
func main() {
	pagevariables.Load()
	fmt.Println("Server started... (ctrl-c to exit)") //signpost that the server has started
	http.HandleFunc("/", startPage)                   //set up a http handler for the handle of '/' which will call function 'startPage'
	http.HandleFunc("/doStuff", doStuff)              //
	go startHTTPServer("8080")                        //run a local http server on a seperate go routine on port 8080
	openBrowser("http://127.0.0.1:8080")              //open a browser or tab automatically to go to the GUI
	waitForQuit()                                     //block main goroutine from exiting until we've received a message from the quit channel.
}
