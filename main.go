package main

// NOTE: Copy wasm_exec.js from the current release of go:
// cp /path/to/go/misc/wasm/wasm_exec.js .

/*
Canvas drawing reference - https://www.w3schools.com/tags/ref_canvas.asp
*/

import (
	"fmt"
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

func setupCanvas() *Canvas {
	doc := js.Global().Get("document")
	canvasEl := doc.Call("getElementById", "mycanvas")
	width = doc.Get("body").Get("clientWidth").Float()
	height = doc.Get("body").Get("clientHeight").Float()
	canvasEl.Set("width", width)
	canvasEl.Set("height", height)
	canvas = Canvas{canvasEl.Call("getContext", "2d")}
	canvas.ctx.Call("clearRect", 0, 0, width, height)
	return &canvas
}

type direction int

// Wall directions for cells
const (
	North direction = iota
	South
	East
	West
)

// Cell is a single maze location
type Cell struct {
	row, col  int
	neighbors map[direction]*Cell
	walls     map[direction]bool
}

// OpenNeighbors?
// ClosedNeighbors?

// Grid holds the details of a maze
type Grid struct {
	rows, cols int
	grid       [][]*Cell
}

func (g *Grid) init(rows, cols int) {
	g.rows = rows
	g.cols = cols
	g.grid = make([][]*Cell, rows)

	// Allocate all the cells
	for row := 0; row < g.rows; row++ {
		g.grid[row] = make([]*Cell, cols)
		for col := 0; col < g.cols; col++ {
			n := make(map[direction]*Cell, 4)
			w := make(map[direction]bool, 4)
			g.grid[row][col] = &Cell{row, col, n, w}
		}
	}

	// Link all the cells to their neighbors
	for row := 0; row < g.rows; row++ {
		for col := 0; col < g.cols; col++ {
			c := g.grid[row][col]
			for _, dir := range []direction{North, South, East, West} {
				n, ok := g.Neighbor(row, col, dir)
				if ok {
					c.neighbors[dir] = n
				}
				c.walls[dir] = true
			}
		}
	}
}

// Neighbor returns the cell in the selected direction from the current row, col position
func (g *Grid) Neighbor(row, col int, dir direction) (*Cell, bool) {
	switch dir {
	case North:
		if row == 0 {
			return nil, false
		}
		return g.grid[row-1][col], true
	case South:
		if row == g.rows-1 {
			return nil, false
		}
		return g.grid[row+1][col], true
	case East:
		if col == g.cols-1 {
			return nil, false
		}
		return g.grid[row][col+1], true

	case West:
		if col == 0 {
			return nil, false
		}
		return g.grid[row][col-1], true

	default:
		return nil, false
	}
}

func binaryTreeMaze(grid *Grid) {
}

// Draw will draw the maze
func Draw(maze Grid, canvas *Canvas) {
	for row := 0; row < maze.rows; row++ {
		for col := 0; col < maze.cols; col++ {
			// cell size is 20x20
			var x, y float64
			x = float64(row) * 20
			y = float64(col) * 20
			fmt.Printf("x = %0.0f, y = %0.0f\n", x, y)

			c := maze.grid[row][col]
			if c.walls[North] {
				canvas.line(x, y, x+20, y)
			}
			if c.walls[South] {
				canvas.line(x, y+20, x+20, y+20)
			}
			if c.walls[East] {
				canvas.line(x+20, y, x+20, y+20)
			}
			if c.walls[West] {
				canvas.line(x, y, x, y+20)
			}
		}
	}
}

func main() {

	fmt.Println("running...")

	canvas := setupCanvas()

	maze := Grid{}
	maze.init(10, 10)

	Draw(maze, canvas)

	<-done
}
