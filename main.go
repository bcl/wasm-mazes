package main

// NOTE: Copy wasm_exec.js from the current release of go:
// cp /path/to/go/misc/wasm/wasm_exec.js .

/*
Canvas drawing reference - https://www.w3schools.com/tags/ref_canvas.asp
*/

import (
	"fmt"
	"math"
	"syscall/js"
)

// Global channel to keep the application running

var (
	done   chan bool
	width  float64
	height float64
	canvas Canvas
)

// Canvas adds some helper functions to make drawing easier
type Canvas struct {
	ctx js.Value
}

func (c *Canvas) line(x1, y1, x2, y2 float64) {
	c.ctx.Call("beginPath")
	c.ctx.Call("moveTo", x1, y1)
	c.ctx.Call("lineTo", x2, y2)
	c.ctx.Call("stroke")
}

func (c *Canvas) fillRect(x1, y1, x2, y2 float64) {
	c.ctx.Call("fillRect", x1, y1, x2, y2)
}

func (c *Canvas) arc(x, y, r, sAngle, eAngle float64, dir bool) {
	c.ctx.Call("beginPath")
	c.ctx.Call("arc", x, y, r, sAngle, eAngle, dir)
	c.ctx.Call("stroke")
}

func init() {
	done = make(chan bool)
}

func setupCanvas() {
	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "mycanvas")
	width = doc.Get("body").Get("clientWidth").Float()
	height = doc.Get("body").Get("clientHeight").Float()
	canvasEl.Set("width", width)
	canvasEl.Set("height", height)
	canvas = Canvas{canvasEl.Call("getContext", "2d")}
	canvas.ctx.Call("clearRect", 0, 0, width, height)
}

func main() {

	fmt.Println("running...")

	setupCanvas()

	canvas.line(10, 10, width/2, height/2)

	canvas.arc(width*0.75, height*0.75, 33, 0, 2*math.Pi, true)

	canvas.fillRect(100, 100, 125, 125)

	<-done
}
