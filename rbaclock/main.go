package main

import (
	"log"
	"rbaclock/cmd"
)

func init() {
	log.SetFlags(0)
}

func main() {
	cmd.Execute()
}


