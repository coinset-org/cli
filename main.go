/*
Copyright © 2022 Cameron Cooper <cameron@coinset.org>

*/
package main

import "github.com/coinset-org/cli/cmd"

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
