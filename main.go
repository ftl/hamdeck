package main

import "github.com/ftl/hamdeck/cmd"

var version string = "develop"

func main() {
	cmd.Execute(version)
}
