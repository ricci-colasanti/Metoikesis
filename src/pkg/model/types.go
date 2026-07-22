package model

// CellType represents the type of cell (empty or occupied)
type CellType int

const (
	// Empty represents an unoccupied cell
	Empty CellType = iota
	// Occupied represents a cell with an agent
	Occupied
)

// Gender represents the gender of an agent
type Gender int

const (
	Male Gender = iota
	Female
	Other
)

// String returns the name of the gender
func (g Gender) String() string {
	switch g {
	case Male:
		return "Male"
	case Female:
		return "Female"
	case Other:
		return "Other"
	default:
		return "Unknown"
	}
}

// EducationLevel represents the highest education level attained
type EducationLevel int

const (
	NoFormal EducationLevel = iota
	Primary
	Secondary
	Vocational
	Undergraduate
	Postgraduate
	Doctorate
)

// String returns the name of the education level
func (e EducationLevel) String() string {
	switch e {
	case NoFormal:
		return "No Formal"
	case Primary:
		return "Primary"
	case Secondary:
		return "Secondary"
	case Vocational:
		return "Vocational"
	case Undergraduate:
		return "Undergraduate"
	case Postgraduate:
		return "Postgraduate"
	case Doctorate:
		return "Doctorate"
	default:
		return "Unknown"
	}
}

// Agent represents an individual in the migration model
type Agent struct {
	ID               int            // Unique identifier
	Age              int            // Age in years (0-120)
	Gender           Gender         // Male, Female, Other
	Income           float64        // Annual income in local currency
	Education        EducationLevel // Highest education level
	YearsAtResidence int            // How long they've lived at current location
	HomeOwner        bool           // Whether they own their home
	EmploymentStatus string         // "Employed", "Unemployed", "Retired", "Student", etc.
	OriginalArea     int            // Which area the agent started in (0 or 1)
	CurrentArea      int            // Which area the agent is currently in (0 or 1)
	moveHistory      []Position     // Track movement history (optional)
}

// NewAgent creates a new agent with default random attributes
func NewAgent(id int, area int) *Agent {
	return &Agent{
		ID:               id,
		Age:              20 + randInt(0, 60), // 20-80 years old
		Gender:           randomGender(),
		Income:           20000 + randFloat(0, 80000),
		Education:        randomEducation(),
		YearsAtResidence: randInt(0, 30),
		HomeOwner:        randFloat(0, 1) < 0.6, // 60% chance
		EmploymentStatus: randomEmployment(),
		OriginalArea:     area,
		CurrentArea:      area,
		moveHistory:      []Position{},
	}
}

// Helper functions for random generation
func randInt(min, max int) int {
	return min + randIntn(max-min)
}

func randFloat(min, max float64) float64 {
	return min + randFloat64()*(max-min)
}

// These will be implemented with the actual random functions in model.go
var (
	randIntn    func(int) int
	randFloat64 func() float64
)

// SetRandomFunctions sets the random functions (to be called from model.go)
func SetRandomFunctions(intn func(int) int, float64 func() float64) {
	randIntn = intn
	randFloat64 = float64
}

// randomGender returns a random gender
func randomGender() Gender {
	r := randIntn(100)
	if r < 48 {
		return Male
	} else if r < 96 {
		return Female
	}
	return Other
}

// randomEducation returns a random education level
func randomEducation() EducationLevel {
	r := randIntn(100)
	switch {
	case r < 10:
		return NoFormal
	case r < 25:
		return Primary
	case r < 45:
		return Secondary
	case r < 60:
		return Vocational
	case r < 80:
		return Undergraduate
	case r < 95:
		return Postgraduate
	default:
		return Doctorate
	}
}

// randomEmployment returns a random employment status
func randomEmployment() string {
	r := randIntn(100)
	switch {
	case r < 60:
		return "Employed"
	case r < 75:
		return "Unemployed"
	case r < 90:
		return "Retired"
	default:
		return "Student"
	}
}

// AddMoveHistory adds a position to the agent's move history
func (a *Agent) AddMoveHistory(x, y int) {
	if len(a.moveHistory) > 100 { // Limit history to prevent memory bloat
		a.moveHistory = a.moveHistory[1:]
	}
	a.moveHistory = append(a.moveHistory, Position{X: x, Y: y})
}

// GetMoveHistory returns the agent's move history
func (a *Agent) GetMoveHistory() []Position {
	return a.moveHistory
}

// GetMoveCount returns the number of times this agent has moved
func (a *Agent) GetMoveCount() int {
	return len(a.moveHistory)
}

// GetMobilityScore calculates how likely this agent is to move within their area
// Higher score = more likely to move
func (a *Agent) GetMobilityScore() float64 {
	score := 0.5 // Base score

	// Age factor (younger = more mobile)
	if a.Age < 25 {
		score += 0.2
	} else if a.Age < 35 {
		score += 0.1
	} else if a.Age > 60 {
		score -= 0.2
	}

	// Income factor (higher income = more mobile)
	if a.Income > 100000 {
		score += 0.15
	} else if a.Income > 60000 {
		score += 0.05
	} else if a.Income < 20000 {
		score -= 0.1
	}

	// Home ownership (renters move more)
	if !a.HomeOwner {
		score += 0.15
	}

	// Education (higher education = more mobile)
	if a.Education >= Undergraduate {
		score += 0.1
	}

	// Employment (unemployed move more)
	if a.EmploymentStatus == "Unemployed" {
		score += 0.2
	}

	// Years at residence (longer = less likely to move)
	if a.YearsAtResidence > 10 {
		score -= 0.2
	} else if a.YearsAtResidence > 5 {
		score -= 0.1
	}

	// Clamp score between 0 and 1
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

// GetInterAreaMoveProbability calculates likelihood of moving to the other area
// Based on factors like age, income, education, and years at residence
func (a *Agent) GetInterAreaMoveProbability() float64 {
	score := 0.15 // Base probability (15%)

	// Younger people more likely to move between areas
	if a.Age < 30 {
		score += 0.2
	} else if a.Age < 45 {
		score += 0.1
	} else if a.Age > 60 {
		score -= 0.1
	}

	// Higher income = more likely to move between areas
	if a.Income > 100000 {
		score += 0.15
	} else if a.Income > 60000 {
		score += 0.05
	}

	// Higher education = more likely to move
	if a.Education >= Undergraduate {
		score += 0.1
	}

	// Home owners less likely to move between areas
	if a.HomeOwner {
		score -= 0.15
	}

	// Long-term residents less likely to move
	if a.YearsAtResidence > 10 {
		score -= 0.1
	} else if a.YearsAtResidence > 5 {
		score -= 0.05
	}

	// Unemployed more likely to move
	if a.EmploymentStatus == "Unemployed" {
		score += 0.15
	}

	// Clamp score between 0 and 0.8 (max 80% chance)
	if score < 0 {
		return 0
	}
	if score > 0.8 {
		return 0.8
	}
	return score
}

// Cell represents a cell in the grid that may contain an agent
type Cell struct {
	Agent *Agent // nil if empty
}

// Position represents an (x,y) coordinate in the grid
type Position struct {
	X, Y int
}

// AreaConfig holds configuration for a single area
type AreaConfig struct {
	Width      int     // Grid width (number of columns)
	Height     int     // Grid height (number of rows)
	EmptyRatio float64 // Proportion of empty cells (0.0 - 1.0)
	Threshold  float64 // Similarity threshold (0.0 - 1.0)
}

// WorldConfig holds configuration for the entire world (two areas)
type WorldConfig struct {
	Area0           AreaConfig // Configuration for area 0
	Area1           AreaConfig // Configuration for area 1
	InterAreaFactor float64    // Multiplier for inter-area migration (0.0 - 1.0)
}

// DefaultConfig returns a default world configuration
func DefaultConfig() WorldConfig {
	return WorldConfig{
		Area0: AreaConfig{
			Width:      50,
			Height:     50,
			EmptyRatio: 0.15,
			Threshold:  0.3,
		},
		Area1: AreaConfig{
			Width:      50,
			Height:     50,
			EmptyRatio: 0.15,
			Threshold:  0.3,
		},
		InterAreaFactor: 0.5,
	}
}
