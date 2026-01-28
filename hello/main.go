package main

import (
	"fmt"
	"net/http"
	"strconv"
)

func main() {
	http.HandleFunc("/", handler)
	fmt.Println("Server running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	input := r.FormValue("sides")

	html := pageHTML()

	if input != "" {
		sides, err := strconv.Atoi(input)
		if err == nil {
			html += "<h3>Shape: " + shapeName(sides) + "</h3>"
			html += "<pre>" + drawShape(sides) + "</pre>"
		}
	}

	w.Write([]byte(html))
}
