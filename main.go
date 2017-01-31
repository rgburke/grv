package main

import (
	"log"
)

func main() {
	grv := NewGRV()

	if err := grv.Initialise("."); err != nil {
		log.Fatal(err)
	}

	grv.Run()

	grv.Free()
}
