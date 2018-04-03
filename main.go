package main

import (
	"fmt"
	"os"

	"github.com/bpineau/katafygio/cmd"
)

//var privateExitHandler func(code int) = os.Exit
var privateExitHandler = os.Exit

// ExitWrapper allow unit tests on main() exit values
func ExitWrapper(exit int) {
	privateExitHandler(exit)
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Printf("%+v", err)
		ExitWrapper(1)
	}
}
