package main

import (
	"log"

	"github.com/metagram-net/firehose/cmd/drift/cmd"
)

func main() {
	if err := cmd.Main(); err != nil {
		log.Fatal(err.Error())
	}
}
