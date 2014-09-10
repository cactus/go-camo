package gologit_test

import "github.com/cactus/gologit"

func ExampleNew() {
	logger := gologit.New(true)
	logger.Debug("It works!")
}
