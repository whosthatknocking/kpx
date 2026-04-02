package main

import (
	"log"
	"os"

	"github.com/whosthatknocking/kpx/cmd"
)

func main() {
	if err := cmd.GenerateBashCompletion(os.Stdout); err != nil {
		log.Fatal(err)
	}
}
