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

	offers := []*carriers.DeliveryOption{}
	for _, c := range srv.carriers {

		// Here, we do not want to _fail_ the request if a single provider fails. Instead, we just want to return
		// whatever providers are available. Otherwise, we'd be only as available as a the worst downstream provider!
		// However, that creates a dilemma: How do we know when we need to intervene with a provider?
		next, _ := c.Query(pkg)
		offers = append(offers, next...)
	}

	// If there are no delivery methods, we want to indicate this to the user.
	if len(offers) == 0 {
		w.Header().Add("Content-Type", problem.HTTPContentTypeJSON)
		w.WriteHeader(http.StatusNotFound)

		// Hint: This can fail, but it is ignored.
		jw.Encode(&problem.Problem{
			Type:   "delivery-options.local/problems/no-options",
			Title:  "There are no delivery options available",
			Detail: fmt.Sprintf("Despite querying %d providers, there are no options provided", len(srv.carriers)),
		})
		return
	}

	w.Header().Add("Content-Type", "application/json")
	jw.Encode(offers)
}
