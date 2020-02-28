package main

// NOTE: Copy wasm_exec.js from the current release of go:
// cp /path/to/go/misc/wasm/wasm_exec.js .

import (
	"fmt"
	"syscall/js"
)

// Global channel to keep the application running
var c chan bool

func init() {
	c = make(chan bool)
}

func main() {

	fmt.Println("WASM!")

	document := js.Global().Get("document")
	p := document.Call("createElement", "p")
	p.Set("innerHTML", "Hello WASM from Go!")
	document.Get("body").Call("appendChild", p)

	<-c
}
