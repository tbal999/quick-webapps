package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
)

var pagevariables = struct {
	Pageoutput string
}{}

var page = `
		<!DOCTYPE html>
		<html>
		<body>
		<form action="/" method="POST">
		<input name="input" type="text" placeholder="type here!">
		<button type="submit" name="submit" value="submitquery">Submit</button>
		</form>
		<br>
		Output text goes here: {{.Pageoutput}}
		</body>
		</html>
`

//validate the user input in the form on the web page
func validate(r *http.Request, nameOfForm string) (string, bool) {
	if len(r.Form[nameOfForm]) != 0 && r.Form[nameOfForm][0] != "" {
		return r.Form[nameOfForm][0], true
	}
	return "", false
}

//the function that is invoked if we go to '/' handler on server.
func frontPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("index").Parse(page)
	if err != nil {
		log.Print(err)
	}
	r.ParseForm()
	if input, ok := validate(r, "input"); ok {
		pagevariables.Pageoutput = input
	}
	err = t.Execute(w, pagevariables)
	if err != nil {
		log.Print(err)
	}
}

func main() {
	fmt.Println("Starting server...")
	http.HandleFunc("/", frontPage)          //set up a http handler for the handle of '/' which will call function 'startPage'
	err := http.ListenAndServe(":8080", nil) // setting up server on listening port 8080
	if err != nil {
		fmt.Println(err)
	}
}
