package main

import (
	"fmt"
	"os"

	"github.com/bpineau/katafygio/cmd"
)

var privateExitHandler = os.Exit

// ExitWrapper allow tests on main() exit values
func ExitWrapper(exit int) {
	privateExitHandler(exit)
}

func main() {
	err := cmd.Execute()
	if err != nil {
		fmt.Printf("%+v", err)
		ExitWrapper(1)
	}
}
