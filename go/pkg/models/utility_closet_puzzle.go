package models

import (
	"time"

	"github.com/google/uuid"
)

// UtilityClosetPuzzle represents the state of the utility closet puzzle
// This is a singleton - only one instance should exist
type UtilityClosetPuzzle struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null" json:"updatedAt"`

	// Hue light IDs for each button (0-5)
	Button0HueLightID *int `gorm:"column:button_0_hue_light_id" json:"button0HueLightId,omitempty"`
	Button1HueLightID *int `gorm:"column:button_1_hue_light_id" json:"button1HueLightId,omitempty"`
	Button2HueLightID *int `gorm:"column:button_2_hue_light_id" json:"button2HueLightId,omitempty"`
	Button3HueLightID *int `gorm:"column:button_3_hue_light_id" json:"button3HueLightId,omitempty"`
	Button4HueLightID *int `gorm:"column:button_4_hue_light_id" json:"button4HueLightId,omitempty"`
	Button5HueLightID *int `gorm:"column:button_5_hue_light_id" json:"button5HueLightId,omitempty"`

	// Current hue state for each button (color index 0-5)
	Button0CurrentHue int `gorm:"column:button_0_current_hue;not null" json:"button0CurrentHue"`
	Button1CurrentHue int `gorm:"column:button_1_current_hue;not null" json:"button1CurrentHue"`
	Button2CurrentHue int `gorm:"column:button_2_current_hue;not null" json:"button2CurrentHue"`
	Button3CurrentHue int `gorm:"column:button_3_current_hue;not null" json:"button3CurrentHue"`
	Button4CurrentHue int `gorm:"column:button_4_current_hue;not null" json:"button4CurrentHue"`
	Button5CurrentHue int `gorm:"column:button_5_current_hue;not null" json:"button5CurrentHue"`

	// Base hue state for each button (color index 0-5)
	Button0BaseHue int `gorm:"column:button_0_base_hue;not null" json:"button0BaseHue"`
	Button1BaseHue int `gorm:"column:button_1_base_hue;not null" json:"button1BaseHue"`
	Button2BaseHue int `gorm:"column:button_2_base_hue;not null" json:"button2BaseHue"`
	Button3BaseHue int `gorm:"column:button_3_base_hue;not null" json:"button3BaseHue"`
	Button4BaseHue int `gorm:"column:button_4_base_hue;not null" json:"button4BaseHue"`
	Button5BaseHue int `gorm:"column:button_5_base_hue;not null" json:"button5BaseHue"`

	// AllGreensAchieved tracks whether all lights have been red at least once
	// This is required to unlock blue color from yellow
	AllGreensAchieved bool `gorm:"column:all_greens_achieved;not null;default:false" json:"allGreensAchieved"`

	// AllPurplesAchieved tracks whether all lights have been purple at least once
	// This is required to unlock white color
	AllPurplesAchieved bool `gorm:"column:all_purples_achieved;not null;default:false" json:"allPurplesAchieved"`
}

func (UtilityClosetPuzzle) TableName() string {
	return "utility_closet_puzzle"
}

// GetButtonHueLightID returns the Hue light ID for the given button slot (0-5)
func (p *UtilityClosetPuzzle) GetButtonHueLightID(slot int) *int {
	switch slot {
	case 0:
		return p.Button0HueLightID
	case 1:
		return p.Button1HueLightID
	case 2:
		return p.Button2HueLightID
	case 3:
		return p.Button3HueLightID
	case 4:
		return p.Button4HueLightID
	case 5:
		return p.Button5HueLightID
	default:
		return nil
	}
}

// GetButtonCurrentHue returns the current hue (color index) for the given button slot (0-5)
func (p *UtilityClosetPuzzle) GetButtonCurrentHue(slot int) int {
	switch slot {
	case 0:
		return p.Button0CurrentHue
	case 1:
		return p.Button1CurrentHue
	case 2:
		return p.Button2CurrentHue
	case 3:
		return p.Button3CurrentHue
	case 4:
		return p.Button4CurrentHue
	case 5:
		return p.Button5CurrentHue
	default:
		return 0
	}
}

// SetButtonCurrentHue sets the current hue (color index) for the given button slot (0-5)
func (p *UtilityClosetPuzzle) SetButtonCurrentHue(slot int, hue int) {
	if hue < 0 {
		hue = 0
	}
	if hue > 6 {
		hue = 6
	}

	switch slot {
	case 0:
		p.Button0CurrentHue = hue
	case 1:
		p.Button1CurrentHue = hue
	case 2:
		p.Button2CurrentHue = hue
	case 3:
		p.Button3CurrentHue = hue
	case 4:
		p.Button4CurrentHue = hue
	case 5:
		p.Button5CurrentHue = hue
	}
}

// SetButtonHueLightID sets the Hue light ID for the given button slot (0-5)
func (p *UtilityClosetPuzzle) SetButtonHueLightID(slot int, lightID *int) {
	switch slot {
	case 0:
		p.Button0HueLightID = lightID
	case 1:
		p.Button1HueLightID = lightID
	case 2:
		p.Button2HueLightID = lightID
	case 3:
		p.Button3HueLightID = lightID
	case 4:
		p.Button4HueLightID = lightID
	case 5:
		p.Button5HueLightID = lightID
	}
}

// GetButtonBaseHue returns the base hue (color index) for the given button slot (0-5)
func (p *UtilityClosetPuzzle) GetButtonBaseHue(slot int) int {
	switch slot {
	case 0:
		return p.Button0BaseHue
	case 1:
		return p.Button1BaseHue
	case 2:
		return p.Button2BaseHue
	case 3:
		return p.Button3BaseHue
	case 4:
		return p.Button4BaseHue
	case 5:
		return p.Button5BaseHue
	default:
		return 0
	}
}

// SetButtonBaseHue sets the base hue (color index) for the given button slot (0-5)
func (p *UtilityClosetPuzzle) SetButtonBaseHue(slot int, hue int) {
	if hue < 0 {
		hue = 0
	}
	if hue > 6 {
		hue = 6
	}

	switch slot {
	case 0:
		p.Button0BaseHue = hue
	case 1:
		p.Button1BaseHue = hue
	case 2:
		p.Button2BaseHue = hue
	case 3:
		p.Button3BaseHue = hue
	case 4:
		p.Button4BaseHue = hue
	case 5:
		p.Button5BaseHue = hue
	}
}

// ResetToBaseState resets all buttons to their base hue values
func (p *UtilityClosetPuzzle) ResetToBaseState() {
	p.Button0CurrentHue = p.Button0BaseHue
	p.Button1CurrentHue = p.Button1BaseHue
	p.Button2CurrentHue = p.Button2BaseHue
	p.Button3CurrentHue = p.Button3BaseHue
	p.Button4CurrentHue = p.Button4BaseHue
	p.Button5CurrentHue = p.Button5BaseHue
	p.AllGreensAchieved = false
	p.AllPurplesAchieved = false
}

// PuzzleColor represents a color for the utility closet puzzle
type PuzzleColor int

const (
	PuzzleColorOff    PuzzleColor = iota // Off (grey)
	PuzzleColorBlue                      // Blue
	PuzzleColorRed                       // Red
	PuzzleColorWhite                     // White
	PuzzleColorYellow                    // Yellow
	PuzzleColorPurple                    // Purple
	PuzzleColorGreen                     // Green (success state)
)

// String returns the string representation of the color
func (c PuzzleColor) String() string {
	switch c {
	case PuzzleColorOff:
		return "Off"
	case PuzzleColorBlue:
		return "Blue"
	case PuzzleColorRed:
		return "Red"
	case PuzzleColorWhite:
		return "White"
	case PuzzleColorYellow:
		return "Yellow"
	case PuzzleColorPurple:
		return "Purple"
	case PuzzleColorGreen:
		return "Green"
	default:
		return "Unknown"
	}
}

// ToInt returns the integer representation of the color (0-6)
func (c PuzzleColor) ToInt() int {
	return int(c)
}

// PuzzleColorFromInt returns a PuzzleColor from an integer (0-6)
func PuzzleColorFromInt(i int) PuzzleColor {
	if i < 0 || i > 6 {
		return PuzzleColorOff
	}
	return PuzzleColor(i)
}

// ColorIndexToRGB converts a color index (0-6) to RGB values
// Color mapping: 0=Off (grey), 1=Blue, 2=Red, 3=White, 4=Yellow, 5=Purple, 6=Green
func ColorIndexToRGB(colorIndex int) (r, g, b uint8) {
	switch colorIndex {
	case 0: // Off (grey)
		return 128, 128, 128
	case 1: // Blue
		return 0, 0, 255
	case 2: // Red
		return 255, 0, 0
	case 3: // White
		return 255, 255, 255
	case 4: // Yellow
		return 255, 200, 0 // Warmer, more golden yellow
	case 5: // Purple
		return 128, 0, 128
	case 6: // Green (success state)
		return 0, 255, 0
	default:
		return 128, 128, 128 // Default to grey (off)
	}
}
