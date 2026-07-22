package main

import (
	"image"
	"image/color"
	"math/rand"
	"time"

	"metoikesis/pkg/model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

const (
	maxIterations = 1000
)

// DisplayMode defines what the grid shows
type DisplayMode int

const (
	DisplayType DisplayMode = iota
	DisplayIncome
)

// GUI handles the application UI and simulation control
type GUI struct {
	world        *model.World
	imageWidget0 *canvas.Image
	imageWidget1 *canvas.Image
	stopChan     chan bool
	running      bool
	displayMode  DisplayMode
}

// NewGUI creates a new GUI controller
func NewGUI(world *model.World, img0, img1 *canvas.Image) *GUI {
	return &GUI{
		world:        world,
		imageWidget0: img0,
		imageWidget1: img1,
		stopChan:     make(chan bool),
		running:      false,
		displayMode:  DisplayType,
	}
}

// RenderGridArea renders a single area with origin circles
func (g *GUI) RenderGridArea(area *model.Area) *image.RGBA {
	width := area.GetWidth()
	height := area.GetHeight()
	cellSize := 10

	img := image.NewRGBA(image.Rect(0, 0, width*cellSize, height*cellSize))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			agent, ok := area.GetState(x, y)
			var col color.Color

			if !ok {
				col = color.RGBA{240, 240, 240, 255} // Empty cell
			} else {
				// Base color based on display mode
				switch g.displayMode {
				case DisplayType:
					col = g.getTypeColor(agent)
				case DisplayIncome:
					col = g.getIncomeColor(agent.Income)
				default:
					col = color.RGBA{240, 240, 240, 255}
				}
			}

			// Fill the cell background
			for dy := 0; dy < cellSize; dy++ {
				for dx := 0; dx < cellSize; dx++ {
					img.Set(x*cellSize+dx, y*cellSize+dy, col)
				}
			}

			// Draw origin circle if there's an agent
			if ok && agent != nil {
				// Origin color: White for Area 0, Black for Area 1
				var originColor color.Color
				if agent.OriginalArea == 0 {
					originColor = color.RGBA{255, 255, 255, 255} // Solid white
				} else {
					originColor = color.RGBA{0, 0, 0, 255} // Solid black
				}
				g.drawSolidCircle(img, x*cellSize, y*cellSize, cellSize, originColor)
			}
		}
	}
	return img
}

// drawSolidCircle draws a solid circle showing the agent's origin
func (g *GUI) drawSolidCircle(img *image.RGBA, cellX, cellY, cellSize int, col color.Color) {
	centerX := cellX + cellSize/2
	centerY := cellY + cellSize/2
	radius := cellSize / 4

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			if dx*dx+dy*dy <= radius*radius {
				px := centerX + dx
				py := centerY + dy
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, col)
				}
			}
		}
	}
}

// getTypeColor returns color based on gender
func (g *GUI) getTypeColor(agent *model.Agent) color.Color {
	switch agent.Gender {
	case model.Male:
		return color.RGBA{50, 50, 200, 255} // Blue
	case model.Female:
		return color.RGBA{200, 50, 50, 255} // Red
	default:
		return color.RGBA{100, 100, 100, 255} // Gray
	}
}

// getIncomeColor returns color based on income
func (g *GUI) getIncomeColor(income float64) color.Color {
	minIncome := 0.0
	maxIncome := 200000.0

	normalized := (income - minIncome) / (maxIncome - minIncome)
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}

	red := uint8(255 - normalized*255)
	green := uint8(normalized * 255)
	blue := uint8(50)

	return color.RGBA{red, green, blue, 255}
}

// UpdateImages refreshes both area images
func (g *GUI) UpdateImages() {
	fyne.Do(func() {
		g.imageWidget0.Image = g.RenderGridArea(g.world.GetArea(0))
		g.imageWidget0.Refresh()
		g.imageWidget1.Image = g.RenderGridArea(g.world.GetArea(1))
		g.imageWidget1.Refresh()
	})
}

// Run starts the simulation loop
func (g *GUI) Run() {
	g.running = true
	for g.running {
		select {
		case <-g.stopChan:
			g.running = false
			return
		default:
			moved := g.world.Step()
			g.UpdateImages()

			if moved == 0 || g.world.GetIteration() >= maxIterations {
				g.running = false
			}

			time.Sleep(50 * time.Millisecond)
		}
	}
}

// Stop halts the simulation
func (g *GUI) Stop() {
	if g.running {
		g.stopChan <- true
		g.running = false
	}
}

// Reset resets the world and re-renders
func (g *GUI) Reset() {
	g.Stop()
	g.world.Reset()
	g.UpdateImages()
}

// SetDisplayMode changes the display mode
func (g *GUI) SetDisplayMode(mode DisplayMode) {
	g.displayMode = mode
	g.UpdateImages()
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Create the world with default configuration
	config := model.DefaultConfig()
	world := model.NewWorld(config)

	// Create the Fyne application
	app := app.NewWithID("com.metoikesis.model")
	window := app.NewWindow("Metoikesis - Two Area Migration Model")

	// Fixed window size
	window.Resize(fyne.NewSize(1100, 500))

	// Create image widgets for both areas
	imageWidget0 := canvas.NewImageFromImage(nil)
	imageWidget0.FillMode = canvas.ImageFillStretch

	imageWidget1 := canvas.NewImageFromImage(nil)
	imageWidget1.FillMode = canvas.ImageFillStretch

	// Create GUI controller
	gui := NewGUI(world, imageWidget0, imageWidget1)

	// Initial render
	gui.UpdateImages()

	// Create a container without layout to manually position images
	content := container.NewWithoutLayout(imageWidget0, imageWidget1)

	// Function to update image positions and sizes
	layoutAreas := func(size fyne.Size) {
		gap := float32(10)
		imgWidth := (size.Width - gap) / 2
		imgHeight := size.Height

		imageWidget0.Move(fyne.NewPos(0, 0))
		imageWidget0.Resize(fyne.NewSize(imgWidth, imgHeight))

		imageWidget1.Move(fyne.NewPos(imgWidth+gap, 0))
		imageWidget1.Resize(fyne.NewSize(imgWidth, imgHeight))
	}

	// Perform initial layout
	layoutAreas(window.Canvas().Size())

	// Set the content
	window.SetContent(content)

	// Create menu
	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Type", func() {
			gui.SetDisplayMode(DisplayType)
		}),
		fyne.NewMenuItem("Income", func() {
			gui.SetDisplayMode(DisplayIncome)
		}),
	)

	runMenu := fyne.NewMenu("Run",
		fyne.NewMenuItem("Start", func() {
			if !gui.running {
				go gui.Run()
			}
		}),
		fyne.NewMenuItem("Stop", func() {
			gui.Stop()
		}),
		fyne.NewMenuItem("Reset", func() {
			gui.Reset()
			go gui.Run()
		}),
	)

	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File"),
		viewMenu,
		runMenu,
	)
	window.SetMainMenu(mainMenu)

	// Auto-start the simulation
	go gui.Run()

	window.ShowAndRun()
}
