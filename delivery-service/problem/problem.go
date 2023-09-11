package problem

// HTTPContentTypeJSON is how the API indicates that the content type is a problem, rather than the normal JSON.
const HTTPContentTypeJSON = "application/problem+json"

// Problem is an implementation of the problem definitions for RFC7807. It is designed to be an easy way to communicate
// user facing errors through the program, and provides helper methods for communicating it across transports.
type Problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Instance string `json:"instance,omitempty"`
}
