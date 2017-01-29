package main

import (
	"log"
)

func main() {
	grv := NewGRV()

	if err := grv.Initialise("."); err != nil {
		log.Fatal(err)
	}

	viewDimension := grv.ui.ViewDimension()

	wins, err := grv.view.Render(viewDimension)
	if err != nil {
		log.Fatal(err)
	}

	if err := grv.ui.Update(wins); err != nil {
		log.Fatal(err)
	}

	grv.ui.GetInput()

	grv.ui.End()
}
