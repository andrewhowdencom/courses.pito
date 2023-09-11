package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/andrewhowdencom/courses.pito/delivery-service/problem"
)

const (
	ParamWidth  = "width"
	ParamHeight = "height"
	ParamDepth  = "depth"
	ParamWeight = "weight"
)

// deliveryOptions receives a request for delivery options and returns a series of options, depending on what
// the downstream providers provide.
//
// Note: This method is deliberately verbose, so as to inline all relevant functionality. It would be possible to do
// this is a more "reusable" way, however, that might make the program more challenging for new Go users to follow.
var deliveryOptions = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Create an encoder that writes provided objects to the response body.
	jw := json.NewEncoder(w)

	// Validate that all of the required parameters have actually been provided.
	missing := []string{}
	values := r.URL.Query()
	for _, k := range []string{ParamWidth, ParamHeight, ParamDepth, ParamWeight} {
		if _, ok := values[k]; !ok {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		w.Header().Add("Content-Type", problem.HTTPContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		// Hint: This can fail, but it is ignored.
		jw.Encode(&problem.Problem{
			Type:   "delivery-options.local/problems/missing-parameters",
			Title:  "Missing required parameters",
			Detail: fmt.Sprintf("The following parameters required for this request were not found: %s", strings.Join(missing, ",")),
		})
		return
	}

	// TODO: Implement this method.
})
