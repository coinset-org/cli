/*
Copyright © 2022 Cameron Cooper <cameron@coinset.org>

*/
package main

import "github.com/coinset-org/cli/internal/cmd/coinset"

var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
