package server

import "net/http"

// healthz provids a handler that returns whether or not the application is "alive"
//
// see https://stackoverflow.com/a/43381061
var healthz = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
})
