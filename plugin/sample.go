package main

import "fmt"

type archivePrinter func()

func (p archivePrinter) Transform(input []map[string]interface{}) []map[string]interface{} {
	fmt.Printf("%#v", input)
	return input
}

var Transformer archivePrinter
