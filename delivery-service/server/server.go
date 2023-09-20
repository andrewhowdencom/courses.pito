package server

import (
	"context"
	"net/http"

	"github.com/andrewhowdencom/courses.pito/delivery-service/carriers"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Server struct {
	srv *http.Server

	// carriers are the carriers that can provide the shipping method.
	carriers []carriers.Carrier
}

// New generates a new server, appropriately configured
func New() *Server {
	srv := &Server{}
	mux := http.NewServeMux()

	// The binding of the method to the routes includes the "instrumentation middleware". The first example is
	// explained in depth, with the subsequent handlers repeated.
	mux.Handle(
		// The first argument is the route. In this case, "healthz" following Googles conventions.
		// See
		// 1. https://stackoverflow.com/a/43381061
		"/healthz",

		// The second argument is the handler. However, the actual handler is wrapped in the middleware handler
		// that instruments the request and response.
		otelhttp.NewHandler(
			// The first argument converts the "handler func" to a "handler"
			http.HandlerFunc(srv.healthz),

			// The second argument will be used as the operation name in the metrics and traces.
			"healthz",
		),
	)

	mux.Handle("/delivery-options", otelhttp.NewHandler(http.HandlerFunc(srv.deliveryOptions), "delivery-options"))
	return &Server{
		srv: &http.Server{
			Handler: mux,
		},
	}
}

func (s *Server) Listen(addr string) error {
	s.srv.Addr = addr

	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown() error {
	return s.srv.Shutdown(context.Background())
}
