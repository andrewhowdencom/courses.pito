package main

import (
	"context"
	"log"
	"time"

	"tracing.local/examples/setup"
)

func main() {
	tp, err := setup.NewTracerProvider()

	if err != nil {
		log.Fatal(err)
	}

	tracer := tp.Tracer("trace.go")

	ctx, firstSpan := tracer.Start(context.Background(), "first")
	time.Sleep(time.Millisecond * 150)

	ctx, secondSpan := tracer.Start(ctx, "second")
	time.Sleep(time.Millisecond * 200)

	secondSpan.End()
	firstSpan.End()
}
