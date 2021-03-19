package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
)

//struct where we will store any variables we want to place on the page
var pagevariables = struct {
	Pageoutput string
}{}

//the html for the page (including formatting
var page = `
		<!DOCTYPE html>
		<html>
		<body>
		<form action="/" method="POST">
		<input name="input" type="text" placeholder="type here!"> ` + //the name lets you validate the input of this text input box
	`
		<button type="submit" name="submit" value="submitquery">Submit</button>
		</form>
		<br>` + ` {{.Pageoutput}}` + //this is where the variable is from the struct
	`</body>
		</html>`

//validate the user input in the form on the web page
//the validation is literally 'if the length is 0 and if the string of the first index isn't empty'
func validate(r *http.Request, nameOfForm string) (string, bool) {
	if len(r.Form[nameOfForm]) != 0 && r.Form[nameOfForm][0] != "" {
		return r.Form[nameOfForm][0], true
	}
	return "", false
}

//the function that is called if we go to '/' handler on server.
func frontPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("index").Parse(page) //go templating package will parse the contents of the page variable
	if err != nil {
		log.Print(err)
	}
	r.ParseForm()                              //now we parse the contents of the form
	if input, ok := validate(r, "input"); ok { //if there's content in the text box named 'input'...
		pagevariables.Pageoutput = "You typed in: " + input
	}
	err = t.Execute(w, pagevariables) //execute the template, but passing in the (now updated) page variable
	if err != nil {
		log.Print(err)
	}
}

func main() {
	fmt.Println("Starting server...")
	http.HandleFunc("/", frontPage)          //set up a handle for '/' which will call function 'frontPage'
	err := http.ListenAndServe(":8080", nil) // setting up server on listening port 8080
	if err != nil {
		fmt.Println(err)
	}
}
