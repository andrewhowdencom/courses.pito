package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/andrewhowdencom/courses.pito/delivery-service/carriers"
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
func (srv *Server) deliveryOptions(w http.ResponseWriter, r *http.Request) {
	// Create an encoder that writes provided objects to the response body.
	jw := json.NewEncoder(w)

	// Validate that all of the required parameters have actually been provided and are well formed. If they are, add
	// them to a map for later access
	pMissing := []string{}
	pBroken := []string{}
	pOK := map[string]int64{}

	values := r.URL.Query()
	for _, k := range []string{ParamWidth, ParamHeight, ParamDepth, ParamWeight} {
		// Hint: How will we know about these later?
		if !values.Has(k) {
			pMissing = append(pMissing, k)
			continue
		}

		// Convert the string input parameter into a number.
		i64, err := strconv.ParseInt(values.Get(k), 10, 64)

		// Hint: How will we know about these later?
		if err != nil {
			pBroken = append(pBroken, k)
			continue
		}

		pOK[k] = i64
	}

	// Here, we handle if there are any of the parameters are missing. We return a helpful "problem" object, to point
	// users in the right direction.
	if len(pMissing) > 0 {
		w.Header().Add("Content-Type", problem.HTTPContentTypeJSON)
		w.WriteHeader(http.StatusBadRequest)

		// Hint: This can fail, but it is ignored.
		jw.Encode(&problem.Problem{
			Type:  "delivery-options.local/problems/bad-parameters",
			Title: "Missing or malformed input parameters",
			Detail: fmt.Sprintf(
				"The following parameters required for this request were not found: %s. The following were malformed: %s",
				strings.Join(pMissing, ","),
				strings.Join(pBroken, ","),
			),
		})
		return
	}

	// Here, we are querying all of the providers for their delivery options.
	//
	// In production, you'd probably want to do this in parallel. However, that makes the error handling a little more
	// challenging to read, so we'll just do this serially for now.
	pkg := &carriers.Package{
		Width:  pOK[ParamWidth],
		Height: pOK[ParamHeight],
		Depth:  pOK[ParamDepth],
		Weight: pOK[ParamWeight],
	}

	offers, err := srv.carriers.Query(pkg)

	switch err {
	case nil:
		w.Header().Add("Content-Type", "application/json")
		jw.Encode(offers)

	case carriers.ErrNoOffersFound:
		w.Header().Add("Content-Type", problem.HTTPContentTypeJSON)
		w.WriteHeader(http.StatusNotFound)

		// Hint: This can fail, but it is ignored.
		jw.Encode(&problem.Problem{
			Type:   "delivery-options.local/problems/no-options",
			Title:  "There are no delivery options available",
			Detail: "Despite querying all providers, there are no options provided",
		})
	default:
		w.Header().Add("Content-Type", problem.HTTPContentTypeJSON)
		w.WriteHeader(http.StatusInternalServerError)

		// Hint: This can fail, but it is ignored.
		jw.Encode(&problem.Problem{
			Type:   "delivery-options.local/server/internal-server-error",
			Title:  "An unexpected server error has occurred",
			Detail: "An error that is not handled within the software has occurred. Please check telemetry for details",
		})
	}
}
