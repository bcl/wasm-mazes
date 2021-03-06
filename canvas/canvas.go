/*
Canvas drawing reference - https://www.w3schools.com/tags/ref_canvas.asp
*/

package canvas

import (
	"syscall/js"
)

// Canvas adds some helper functions to make drawing easier
type Canvas struct {
	ctx    js.Value
	width  float64
	height float64
}

func NewCanvas() *Canvas {
	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "mycanvas")
	width := doc.Get("body").Get("clientWidth").Float()
	height := doc.Get("body").Get("clientHeight").Float()
	canvasEl.Set("width", width)
	canvasEl.Set("height", height)
	canvas := Canvas{canvasEl.Call("getContext", "2d"), width, height}
	canvas.ctx.Set("font", "12px Arial")
	canvas.CLS()
	return &canvas
}

func (c *Canvas) CLS() {
	c.ctx.Call("clearRect", 0, 0, c.width, c.height)
}

func (c *Canvas) Line(x1, y1, x2, y2 float64) {
	c.ctx.Call("beginPath")
	c.ctx.Call("moveTo", x1, y1)
	c.ctx.Call("lineTo", x2, y2)
	c.ctx.Set("strokeStyle", "#000000")
	c.ctx.Set("lineWidth", "1.5")
	c.ctx.Call("stroke")
}

func (c *Canvas) FillRect(x, y, w, h float64) {
	c.ctx.Call("fillRect", x, y, w, h)
}

func (c *Canvas) Arc(x, y, r, sAngle, eAngle float64, dir bool) {
	c.ctx.Call("beginPath")
	c.ctx.Call("arc", x, y, r, sAngle, eAngle, dir)
	c.ctx.Call("stroke")
}

func (c *Canvas) Print(x, y float64, msg string) {
	c.ctx.Call("fillText", msg, x, y)
}

func (c *Canvas) Color(color string) {
	c.ctx.Set("fillStyle", color)
}

func (c *Canvas) OnClick(f func(this js.Value, args []js.Value) interface{}) {
	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "mycanvas")
	cb := js.FuncOf(f)
	canvasEl.Call("addEventListener", "click", cb)
}
