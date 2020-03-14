package main

// NOTE: Copy wasm_exec.js from the current release of go:
// cp /path/to/go/misc/wasm/wasm_exec.js .

/*
Canvas drawing reference - https://www.w3schools.com/tags/ref_canvas.asp
*/

import (
	"fmt"
	"math/rand"
	"syscall/js"
	"time"
)

var (
	done   chan bool // Global channel to keep the application running
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

// Link opens the walls in the direction indicated
func (c *Cell) Link(dir direction) bool {
	if _, ok := c.neighbors[dir]; !ok {
		return false
	}

	switch dir {
	case North:
		c.walls[North] = false
		c.neighbors[North].walls[South] = false
		return true
	case South:
		c.walls[South] = false
		c.neighbors[South].walls[North] = false
		return true
	case East:
		c.walls[East] = false
		c.neighbors[East].walls[West] = false
		return true
	case West:
		c.walls[West] = false
		c.neighbors[West].walls[East] = false
		return true
	}
	return false
}

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

func binaryTreeMaze(maze *Grid) {
	// Visit all the cells
	for row := 0; row < maze.rows; row++ {
		for col := 0; col < maze.cols; col++ {
			// Get the north and east neighbors, if allowed
			var neighbors []direction
			if _, ok := maze.Neighbor(row, col, North); ok {
				neighbors = append(neighbors, North)
			}
			if _, ok := maze.Neighbor(row, col, East); ok {
				neighbors = append(neighbors, East)
			}

			if len(neighbors) == 0 {
				continue
			}
			i := rand.Intn(len(neighbors))
			d := neighbors[i]

			// Link the current cell to the neighbor
			maze.grid[row][col].Link(d)
		}
	}
}

// Draw will draw the maze
func Draw(maze Grid, canvas *Canvas) {
	for row := 0; row < maze.rows; row++ {
		for col := 0; col < maze.cols; col++ {
			// cell size is 20x20
			var x, y float64
			x = float64(col) * 20
			y = float64(row) * 20

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
	rand.Seed(time.Now().UnixNano())
	canvas := setupCanvas()

	maze := Grid{}
	maze.init(10, 10)

	binaryTreeMaze(&maze)
	Draw(maze, canvas)

	<-done
}
