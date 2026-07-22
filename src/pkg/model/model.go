package model

import (
	"math"
	"math/rand"
)

// World is the core migration model with two areas
type World struct {
	areas     [2]*Area
	config    WorldConfig
	iteration int
	nextID    int      // For generating unique agent IDs
	agents    []*Agent // Track all agents
}

// Area represents a single area within the world
type Area struct {
	grid   [][]Cell
	config AreaConfig
	world  *World
}

// NewWorld creates a new world with two areas
func NewWorld(config WorldConfig) *World {
	// Set up random functions for the types package
	SetRandomFunctions(rand.Intn, rand.Float64)

	w := &World{
		config:    config,
		iteration: 0,
		nextID:    1,
		agents:    []*Agent{},
	}

	// Create area 0
	w.areas[0] = w.newArea(config.Area0, 0)
	// Create area 1
	w.areas[1] = w.newArea(config.Area1, 1)

	return w
}

// newArea creates a new area
func (w *World) newArea(config AreaConfig, areaID int) *Area {
	a := &Area{
		grid:   make([][]Cell, config.Height),
		config: config,
		world:  w,
	}
	for i := range a.grid {
		a.grid[i] = make([]Cell, config.Width)
	}
	a.InitGrid(areaID)
	return a
}

// InitGrid populates the area's grid with random agents
func (a *Area) InitGrid(areaID int) {
	for y := 0; y < a.config.Height; y++ {
		for x := 0; x < a.config.Width; x++ {
			if rand.Float64() < a.config.EmptyRatio {
				a.grid[y][x] = Cell{Agent: nil}
			} else {
				agent := NewAgent(a.world.nextID, areaID)
				a.world.nextID++
				a.grid[y][x] = Cell{Agent: agent}
				a.world.agents = append(a.world.agents, agent)
			}
		}
	}
}

// GetState returns the agent at the given position in a specific area
func (a *Area) GetState(x, y int) (*Agent, bool) {
	if x < 0 || x >= a.config.Width || y < 0 || y >= a.config.Height {
		return nil, false
	}
	return a.grid[y][x].Agent, a.grid[y][x].Agent != nil
}

// GetCell returns the cell at the given position in a specific area
func (a *Area) GetCell(x, y int) (*Cell, bool) {
	if x < 0 || x >= a.config.Width || y < 0 || y >= a.config.Height {
		return nil, false
	}
	return &a.grid[y][x], true
}

// SetAgent places an agent at the given position in a specific area
func (a *Area) SetAgent(x, y int, agent *Agent) bool {
	if x < 0 || x >= a.config.Width || y < 0 || y >= a.config.Height {
		return false
	}
	a.grid[y][x].Agent = agent
	if agent != nil {
		agent.CurrentArea = a.getAreaID()
	}
	return true
}

// RemoveAgent removes an agent from the given position in a specific area
func (a *Area) RemoveAgent(x, y int) bool {
	if x < 0 || x >= a.config.Width || y < 0 || y >= a.config.Height {
		return false
	}
	a.grid[y][x].Agent = nil
	return true
}

// getAreaID returns the area ID (0 or 1)
func (a *Area) getAreaID() int {
	if a == a.world.areas[0] {
		return 0
	}
	return 1
}

// GetWidth returns the area width
func (a *Area) GetWidth() int {
	return a.config.Width
}

// GetHeight returns the area height
func (a *Area) GetHeight() int {
	return a.config.Height
}

// countSimilarNeighbors returns the number of similar and total neighbors based on income
func (a *Area) countSimilarNeighbors(x, y int) (similar, total int, avgIncome float64) {
	agent, ok := a.GetState(x, y)
	if !ok {
		return 0, 0, 0
	}

	var totalIncome float64
	var neighborCount int

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			neighbor, ok := a.GetState(nx, ny)
			if ok {
				total++
				totalIncome += neighbor.Income
				neighborCount++

				// Similar if income difference is within 20%
				diff := math.Abs(neighbor.Income - agent.Income)
				if diff/agent.Income < 0.2 {
					similar++
				}
			}
		}
	}

	if neighborCount > 0 {
		avgIncome = totalIncome / float64(neighborCount)
	}
	return similar, total, avgIncome
}

// IsHappy checks if an agent at the given position is happy
func (a *Area) IsHappy(x, y int) bool {
	_, ok := a.GetState(x, y)
	if !ok {
		return true
	}
	similar, total, _ := a.countSimilarNeighbors(x, y)
	if total == 0 {
		return true
	}
	return float64(similar)/float64(total) >= a.config.Threshold
}

// GetUnhappyAgents returns a list of positions of unhappy agents
func (a *Area) GetUnhappyAgents() []Position {
	var unhappy []Position
	for y := 0; y < a.config.Height; y++ {
		for x := 0; x < a.config.Width; x++ {
			if !a.IsHappy(x, y) {
				unhappy = append(unhappy, Position{X: x, Y: y})
			}
		}
	}
	return unhappy
}

// GetEmptyPositions returns a list of all empty positions
func (a *Area) GetEmptyPositions() []Position {
	var empty []Position
	for y := 0; y < a.config.Height; y++ {
		for x := 0; x < a.config.Width; x++ {
			if a.grid[y][x].Agent == nil {
				empty = append(empty, Position{X: x, Y: y})
			}
		}
	}
	return empty
}

// FindBestEmptySpot finds the most compatible empty spot for an agent
func (a *Area) FindBestEmptySpot(agent *Agent, startX, startY int) (Position, float64) {
	emptySpots := a.GetEmptyPositions()
	if len(emptySpots) == 0 {
		return Position{X: -1, Y: -1}, 0
	}

	var bestPos Position
	bestScore := -1.0

	for _, pos := range emptySpots {
		incomeScore := a.calculateIncomeCompatibility(agent.Income, pos.X, pos.Y)
		distanceScore := a.calculateDistanceScore(startX, startY, pos.X, pos.Y)
		score := (incomeScore * 0.7) + (distanceScore * 0.3)

		if score > bestScore {
			bestScore = score
			bestPos = pos
		}
	}

	return bestPos, bestScore
}

// calculateIncomeCompatibility calculates how compatible an agent's income is with neighbors at a position
func (a *Area) calculateIncomeCompatibility(income float64, x, y int) float64 {
	cell, ok := a.GetCell(x, y)
	if !ok || cell.Agent != nil {
		return 0
	}

	var neighborIncomes []float64
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			neighbor, ok := a.GetState(nx, ny)
			if ok {
				neighborIncomes = append(neighborIncomes, neighbor.Income)
			}
		}
	}

	if len(neighborIncomes) == 0 {
		return 0.5
	}

	var total float64
	for _, inc := range neighborIncomes {
		total += inc
	}
	avgNeighborIncome := total / float64(len(neighborIncomes))

	diff := math.Abs(income - avgNeighborIncome)
	if diff == 0 {
		return 1.0
	}

	ratio := diff / avgNeighborIncome
	score := math.Exp(-ratio)

	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

// calculateDistanceScore returns a score based on distance (closer = higher score)
func (a *Area) calculateDistanceScore(x1, y1, x2, y2 int) float64 {
	distance := math.Sqrt(float64((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1)))
	maxDistance := math.Sqrt(float64(a.config.Width*a.config.Width + a.config.Height*a.config.Height))

	if distance == 0 {
		return 1.0
	}

	score := 1.0 - (distance / maxDistance)
	if score < 0 {
		return 0
	}
	return score
}

// StepArea performs one iteration of the migration model for a single area
func (a *Area) StepArea() int {
	unhappy := a.GetUnhappyAgents()
	if len(unhappy) == 0 {
		return 0
	}

	rand.Shuffle(len(unhappy), func(i, j int) {
		unhappy[i], unhappy[j] = unhappy[j], unhappy[i]
	})

	if len(a.GetEmptyPositions()) == 0 {
		return 0
	}

	moved := 0
	for _, pos := range unhappy {
		agent, _ := a.GetState(pos.X, pos.Y)
		if agent == nil {
			continue
		}

		bestSpot, score := a.FindBestEmptySpot(agent, pos.X, pos.Y)

		if bestSpot.X >= 0 && bestSpot.Y >= 0 && score > 0.3 {
			a.RemoveAgent(pos.X, pos.Y)
			a.SetAgent(bestSpot.X, bestSpot.Y, agent)

			agent.AddMoveHistory(bestSpot.X, bestSpot.Y)
			agent.YearsAtResidence = 0

			moved++
		}
	}
	return moved
}

// Step performs one iteration of the migration model for the entire world
func (w *World) Step() int {
	w.iteration++
	totalMoved := 0

	// Process each area
	for i := 0; i < 2; i++ {
		totalMoved += w.areas[i].StepArea()
	}

	// Inter-area migration
	interMoved := w.processInterAreaMoves()
	totalMoved += interMoved

	return totalMoved
}

// processInterAreaMoves handles movement between the two areas
func (w *World) processInterAreaMoves() int {
	moved := 0

	for _, agent := range w.agents {
		// Skip if agent is already in their original area (optional)
		// or if they've moved recently (can add cooldown)

		// Calculate probability of moving between areas
		prob := agent.GetInterAreaMoveProbability() * w.config.InterAreaFactor

		if rand.Float64() < prob {
			// Find the current area and position
			currentAreaIdx := agent.CurrentArea
			currentArea := w.areas[currentAreaIdx]

			// Find the agent's position in the current area
			var currentX, currentY int
			found := false
			for y := 0; y < currentArea.config.Height; y++ {
				for x := 0; x < currentArea.config.Width; x++ {
					if currentArea.grid[y][x].Agent == agent {
						currentX = x
						currentY = y
						found = true
						break
					}
				}
				if found {
					break
				}
			}

			if !found {
				continue
			}

			// Determine target area (switch to the other area)
			targetAreaIdx := 1 - currentAreaIdx
			targetArea := w.areas[targetAreaIdx]

			// Check if target area has an empty spot
			emptySpots := targetArea.GetEmptyPositions()
			if len(emptySpots) == 0 {
				continue
			}

			// Find the best spot in the target area
			bestSpot, score := targetArea.FindBestEmptySpot(agent,
				rand.Intn(targetArea.config.Width), rand.Intn(targetArea.config.Height))

			if bestSpot.X >= 0 && bestSpot.Y >= 0 && score > 0.2 {
				// Remove from current area
				currentArea.RemoveAgent(currentX, currentY)

				// Add to target area
				targetArea.SetAgent(bestSpot.X, bestSpot.Y, agent)

				// Update agent info
				agent.CurrentArea = targetAreaIdx
				agent.YearsAtResidence = 0

				// Record the move
				agent.AddMoveHistory(bestSpot.X, bestSpot.Y)

				moved++
			}
		}
	}

	return moved
}

// GetArea returns a reference to a specific area
func (w *World) GetArea(areaID int) *Area {
	if areaID < 0 || areaID > 1 {
		return nil
	}
	return w.areas[areaID]
}

// GetConfig returns the world configuration
func (w *World) GetConfig() WorldConfig {
	return w.config
}

// GetIteration returns the current iteration count
func (w *World) GetIteration() int {
	return w.iteration
}

// GetAgentCount returns the total number of agents in the world
func (w *World) GetAgentCount() int {
	return len(w.agents)
}

// CalculateHappiness returns the proportion of happy agents in the world
func (w *World) CalculateHappiness() float64 {
	var totalAgents, happyAgents int
	for _, area := range w.areas {
		for y := 0; y < area.config.Height; y++ {
			for x := 0; x < area.config.Width; x++ {
				if area.grid[y][x].Agent != nil {
					totalAgents++
					if area.IsHappy(x, y) {
						happyAgents++
					}
				}
			}
		}
	}
	if totalAgents == 0 {
		return 1.0
	}
	return float64(happyAgents) / float64(totalAgents)
}

// IsStable returns true if the world has reached a stable state
func (w *World) IsStable() bool {
	for _, area := range w.areas {
		if len(area.GetUnhappyAgents()) > 0 {
			return false
		}
	}
	return true
}

// Reset reinitializes the world with a new random grid
func (w *World) Reset() {
	w.iteration = 0
	w.nextID = 1
	w.agents = []*Agent{}
	w.areas[0].InitGrid(0)
	w.areas[1].InitGrid(1)
}
