package main

// NOTE: Copy wasm_exec.js from the current release of go:
// cp /path/to/go/misc/wasm/wasm_exec.js .

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"syscall/js"

	"github.com/bcl/wasm-mazes/canvas"
)

var (
	done          chan bool // Global channel to keep the application running
	mazeAlgorithm algorithm
)

type algorithm int

const (
	BinaryMaze algorithm = iota
	SidewinderMaze
)

const (
	CellWidth  = 30
	CellHeight = 30
)

func init() {
	done = make(chan bool)

	mazeAlgorithm = SidewinderMaze
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
// neighbors holds pointers to the neighboring cells
// walls entries are true if there is a wall, and false if there is an opening to the neighbor
type Cell struct {
	row, col  int
	neighbors map[direction]*Cell
	walls     map[direction]bool
	distance  int // distance from entrance
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

func (c *Cell) Linked() (linked []*Cell) {
	for _, d := range []direction{North, South, East, West} {
		if c.walls[d] == false {
			linked = append(linked, c.neighbors[d])
		}
	}

	return linked
}

// Grid holds the details of a maze
type Grid struct {
	rows, cols   int
	grid         [][]*Cell
	maxDistance  int
	farthestCell *Cell
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
			g.grid[row][col] = &Cell{row, col, n, w, -1}
		}
	}

	// Link all the cells to their neighbors, and close all the walls
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

// TODO This seems like it should be a Cell function, ask it for it's neighbor
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

func (g *Grid) CellColor(row, col int) string {
	intensity := (float64(g.maxDistance) - float64(g.grid[row][col].distance)) / float64(g.maxDistance)
	dark := int(255 * intensity)
	bright := 128 + int(127*intensity)
	return fmt.Sprintf("#%02X%02X%02X", dark, bright, dark)
}

func (g *Grid) ClearDistance() {
	for row := 0; row < g.rows; row++ {
		for col := 0; col < g.cols; col++ {
			g.grid[row][col].distance = -1
		}
	}
}

// Draw will draw the maze
func Draw(maze Grid, canvas *canvas.Canvas, showDistance, showColor bool) {
	canvas.Color("#000000")
	for row := 0; row < maze.rows; row++ {
		for col := 0; col < maze.cols; col++ {
			var x, y float64
			x = float64(col) * CellWidth
			y = float64(row) * CellHeight

			c := maze.grid[row][col]
			if c.walls[North] {
				canvas.Line(x, y, x+CellWidth, y)
			}
			if c.walls[South] {
				canvas.Line(x, y+CellHeight, x+CellWidth, y+CellHeight)
			}
			if c.walls[East] {
				canvas.Line(x+CellWidth, y, x+CellWidth, y+CellHeight)
			}
			if c.walls[West] {
				canvas.Line(x, y, x, y+CellHeight)
			}

			// Color the cells based on the distance from the start
			if showColor {
				canvas.Color(maze.CellColor(row, col))
				canvas.FillRect(x+1, y+1, CellWidth-2, CellHeight-2)
				canvas.Color("#000000")
			}

			// Print distance value
			if showDistance {
				canvas.Print(x+2, y+14, fmt.Sprintf("%d", maze.grid[row][col].distance))
			}
		}
	}
}

func DrawSolution(maze Grid, path []*Cell, canvas *canvas.Canvas, showDistance, showColor bool) {
	Draw(maze, canvas, showDistance, showColor)
	for _, c := range path {
		var x, y float64
		x = float64(c.col) * CellWidth
		y = float64(c.row) * CellHeight
		canvas.Color("#00F0FF")
		canvas.FillRect(x+1, y+1, CellWidth-2, CellHeight-2)
		canvas.Color("#000000")
		canvas.Print(x+2, y+14, fmt.Sprintf("%d", c.distance))

	}
}

func BinaryTreeMaze(maze *Grid) {
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

// RunSidewinder executes the Sidewinder algorithm
// Sidewinder visits each location, flips a coin to open the East wall or to
// open the North wall in a random cell from the last 'run' of cells
// Book says to start in the SW corner, but I don't think that matters as long as
// you run it row by row
func RunSidewinder(maze *Grid) {
	var run []*Cell

	// Visit all the cells
	for row := 0; row < maze.rows; row++ {
		run = []*Cell{}
		for col := 0; col < maze.cols; col++ {
			// Top row can only open east, not north
			if row == 0 {
				maze.grid[row][col].Link(East)
				continue
			}
			// Add this cell to the run of cells
			run = append(run, maze.grid[row][col])

			// Flip coin
			i := rand.Intn(2)

			// True or Right Column (cannot open East), so end the run
			if i == 1 || col == maze.cols-1 {
				rm := rand.Intn(len(run))
				maze.grid[run[rm].row][run[rm].col].Link(North)
				run = []*Cell{}
			} else {
				maze.grid[row][col].Link(East)
			}
		}
	}
}

// CalculateDijkstra calculates the distance from the entrance to each cell
func CalculateDijkstra(maze *Grid, row, col int) {
	maze.maxDistance = 0
	frontier := []*Cell{maze.grid[row][col]}
	frontier[0].distance = 0
	for {
		// Keep track of the largest distance
		if frontier[0].distance > maze.maxDistance {
			maze.maxDistance = frontier[0].distance
			maze.farthestCell = frontier[0]
		}

		// Get the cell's accessable neighbors
		neighbors := frontier[0].Linked()
		for _, n := range neighbors {
			if n.distance == -1 {
				n.distance = frontier[0].distance + 1
				frontier = append(frontier, n)
			}
		}

		// Pop the current cell off the frontier list
		frontier = frontier[1:]

		if len(frontier) == 0 {
			break
		}
	}
}

// FindExit finds the path to the exit (in 0,0) when starting at a given point in the maze
// It returns a slice of Cells to follow.
func FindExit(maze *Grid, row, col int) (path []*Cell) {

	c := maze.grid[row][col]
	path = append(path, c)
	for c.distance != 0 {
		var next *Cell
		for _, n := range c.Linked() {
			if n.distance < c.distance {
				next = n
			}
		}
		if next == nil {
			fmt.Printf("STUCK! at %d,%d\n", c.row, c.col)
			return path
		}
		c = next
		path = append(path, c)
	}

	return path
}

type Solver struct {
	maze         *Grid
	canvas       *canvas.Canvas
	showDistance bool
	showColor    bool
}

func (s *Solver) Display() {
	s.maze.init(20, 20)

	switch mazeAlgorithm {
	case BinaryMaze:
		BinaryTreeMaze(s.maze)
	case SidewinderMaze:
		RunSidewinder(s.maze)
	}

	// Find longest path thru maze
	CalculateDijkstra(s.maze, 0, 0)
	row := s.maze.farthestCell.row
	col := s.maze.farthestCell.col
	s.maze.ClearDistance()
	CalculateDijkstra(s.maze, row, col)

	s.canvas.CLS()
	Draw(*s.maze, s.canvas, s.showDistance, s.showColor)
}

func (s *Solver) SolveMaze(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return nil
	}
	x := args[0].Get("offsetX").Int()
	y := args[0].Get("offsetY").Int()

	// Is it inside the maze?
	if x >= s.maze.cols*CellWidth || y >= s.maze.rows*CellHeight {
		return nil
	}

	col := x / 30
	row := y / 30
	fmt.Printf("Click @ %d, %d / %d, %d\n", x, y, row, col)

	path := FindExit(s.maze, row, col)
	s.canvas.CLS()
	DrawSolution(*s.maze, path, s.canvas, s.showDistance, s.showColor)

	return nil
}

// InitButtons sets the initial state from the html button states
func (s *Solver) InitButtons() {
	doc := js.Global().Get("document")
	s.showDistance = doc.Call("getElementById", "distance").Get("checked").Bool()
	s.showColor = doc.Call("getElementById", "color").Get("checked").Bool()
}

func (s *Solver) InitMazeSelection() {
	doc := js.Global().Get("document")
	mazes := doc.Call("getElementById", "mazes")

	for i, name := range []string{"Binary", "Sidewinder"} {
		opt := doc.Call("createElement", "option")
		opt.Set("value", i)
		opt.Set("innerHTML", name)
		mazes.Call("appendChild", opt)
	}
}

func (s *Solver) ToggleDistance(this js.Value, args []js.Value) interface{} {
	s.showDistance = !s.showDistance
	s.canvas.CLS()
	Draw(*s.maze, s.canvas, s.showDistance, s.showColor)

	return nil
}

func (s *Solver) ToggleColor(this js.Value, args []js.Value) interface{} {
	s.showColor = !s.showColor
	s.canvas.CLS()
	Draw(*s.maze, s.canvas, s.showDistance, s.showColor)

	return nil
}

func (s *Solver) ChangeMaze(this js.Value, args []js.Value) interface{} {
	doc := js.Global().Get("document")
	maze, err := strconv.Atoi(doc.Call("getElementById", "mazes").Get("value").String())
	if err != nil {
		fmt.Printf("Error converting value: %s", err)
		return nil
	}
	switch maze {
	case 0:
		mazeAlgorithm = BinaryMaze
	case 1:
		mazeAlgorithm = SidewinderMaze
	}

	s.Display()
	return nil
}

func OnEvent(event, id string, f func(this js.Value, args []js.Value) interface{}) {
	doc := js.Global().Get("document")
	btn := doc.Call("getElementById", id)
	cb := js.FuncOf(f)
	btn.Call("addEventListener", event, cb)
}

func main() {

	fmt.Println("running...")
	rand.Seed(time.Now().UnixNano())
	canvas := canvas.NewCanvas()

	maze := Grid{}

	solver := Solver{&maze, canvas, false, false}
	solver.InitButtons()
	solver.InitMazeSelection()

	// Setup events
	OnEvent("change", "mazes", solver.ChangeMaze)
	OnEvent("click", "distance", solver.ToggleDistance)
	OnEvent("click", "color", solver.ToggleColor)
	canvas.OnClick(solver.SolveMaze)

	solver.Display()

	<-done
}
