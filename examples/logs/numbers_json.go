package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {

	w := json.NewEncoder(os.Stderr)

	if len(os.Args[1:]) == 0 {
		fmt.Fprintf(os.Stderr, "%s: %s", time.Now(), "No arguments supplied")
		return
	}

	total := 1
	for _, in := range os.Args[1:] {
		i, err := strconv.Atoi(in)

		if err != nil {
			w.Encode(map[string]interface{}{
				"time":  time.Now(),
				"error": "failed to create number",
				"context": map[string]interface{}{
					"input":         in,
					"strconv_error": err.Error(),
				},
			})

			return
		}

		total = total * i
	}

	fmt.Println(total)
}
