package main

import "github.com/justestif/qry/cmd"

// version is set at build time via -ldflags="-X main.version=<tag>"
var version = "dev"

func main() {
	cmd.Execute(version)
}
