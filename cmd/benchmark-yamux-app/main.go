package main

import (
	"bytes"
	"fmt"
	"os"
	"time"
)

func main() {
	data := bytes.Repeat([]byte{'A'}, 200)
	for {
		fmt.Fprintf(os.Stderr, "%s\n", data)
		time.Sleep(100 * time.Microsecond)
	}
}
