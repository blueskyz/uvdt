package main

import (
	"log"
	"os"

	_ "flag"
	_ "net/http"
)

func init() {
	log.SetOutput(os.Stdout)
}

func my_func() {
}

func main() {
	log.Printf("hello world %s", "serv")
	my_func()
}
