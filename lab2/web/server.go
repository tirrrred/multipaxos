// +build !solution

package web

import (
	"flag"
	"log"
	"net/http"
	"strconv"
)

var httpAddr = flag.String("http", ":8080", "Listen address")

//var countReq int

func main() {
	flag.Parse()
	server := NewServer()
	//Starts the ListenAndServer (HTTP server) based with:
	//var httpAddr which state which listening address (default :8080)
	//var server of "type Server struct", which is a "handler" based on the function ServeHTTP
	log.Fatal(http.ListenAndServe(*httpAddr, server))
	// "type Server struct" --> "ServeHTTP" which implements the Server struct as a handler -->
	// "NewServer" creates a new Server, based on the handler of "type Server struct" -->
	// "main" starts the server on address "var httpAddr" and the handler "server"
}

// Server implements the web server specification found at
// https://github.com/dat520-2020/assignments/blob/master/lab2/README.md#web-server
type Server struct {
	countReq int //counter for http request to the Server
	reqMux   map[string]func(http.ResponseWriter, *http.Request, *Server)
	//Creates a map with a the URL path (string) as key
	//and a function with the handlers functions and a pointer to the Server struct as values
	//Creates as own maps for this so one can
	// TODO(student): Add needed fields

}

// NewServer returns a new Server with all required internal state initialized.
// NOTE: It should NOT start to listen on an HTTP endpoint.
func NewServer() *Server {
	// TODO(student): Implement
	pathMux := make(map[string]func(http.ResponseWriter, *http.Request, *Server))
	pathMux["/"] = rootHandler
	pathMux["/counter"] = counterHandler
	pathMux["/lab2"] = lab2Handler
	pathMux["/fizzbuzz"] = fizzbuzzHandler
	pathMux["/404"] = notFound
	s := &Server{
		0,
		pathMux,
	}
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO(student): Implement
	if reqHandler, ok := s.reqMux[r.URL.Path]; ok {
		reqHandler(w, r, s)
	} else {
		notFound(w, r, s)
		//log.Fatal("Error: Invalid path. Page not found")
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request, s *Server) {
	s.countReq++
	body := []byte("Hello World!\n")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func counterHandler(w http.ResponseWriter, r *http.Request, s *Server) {
	s.countReq++
	body := []byte("counter: " + strconv.Itoa(s.countReq) + "\n") //count request to favicon as well...
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func lab2Handler(w http.ResponseWriter, r *http.Request, s *Server) {
	s.countReq++
	body := []byte("<a href=\"http://www.github.com/dat520-2020/assignments/tree/master/lab2\">Moved Permanently</a>.\n\n")
	w.WriteHeader(http.StatusMovedPermanently)
	w.Write(body)
}

func fizzbuzzHandler(w http.ResponseWriter, r *http.Request, s *Server) {
	s.countReq++
	var fizzValues []string
	if values, ok := r.URL.Query()["value"]; ok { //Query gets the values where the query parametere key is "value", e.g ?value=30
		for _, v := range values {
			n := fizzBuzzGame(v)
			fizzValues = append(fizzValues, n) //stores it in a slice incase several value paramteres is given, e.g ?value=30&value=40. Only the first will be asses, but can be alter later
		}
	}
	body := []byte(fizzValues[0] + "\n")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func notFound(w http.ResponseWriter, r *http.Request, s *Server) {
	s.countReq++
	body := []byte("404 page not found\n")
	w.WriteHeader(http.StatusNotFound)
	w.Write(body)
}

func fizzBuzzGame(valueStr string) string {
	var result string
	if valueStr == "" {
		return "no value provided"
	}
	valueInt, err := strconv.Atoi(valueStr)

	if err != nil {
		return "not an integer"
	}

	div3 := valueInt % 3
	div5 := valueInt % 5

	if div3 == 0 && div5 == 0 {
		result = "fizzbuzz"
	} else if div3 == 0 {
		result = "fizz"
	} else if div5 == 0 {
		result = "buzz"
	} else {
		result = strconv.Itoa(valueInt)
	}
	return result
}
