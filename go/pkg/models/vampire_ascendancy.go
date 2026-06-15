package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Vampire Ascendancy (The Crimson Toast) event app models.

type VampireHouse struct {
	ID        uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
	Name      string    `gorm:"not null" json:"name"`
	SortOrder int       `gorm:"not null;default:0" json:"sortOrder"`
}

func (VampireHouse) TableName() string { return "vampire_houses" }

type VampireCharacter struct {
	ID              uuid.UUID  `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt       time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"not null" json:"updatedAt"`
	Name            string     `gorm:"not null" json:"name"`
	Title           string     `gorm:"not null;default:''" json:"title"`
	HouseID         *uuid.UUID `json:"houseId"`
	RoleType        string     `gorm:"not null;default:'player'" json:"roleType"` // player | gm | npc
	IsOptional      bool       `gorm:"not null;default:false" json:"isOptional"`
	PreEventInfo    string     `gorm:"not null;default:''" json:"preEventInfo"`
	PostAct1Context string     `gorm:"not null;default:''" json:"postAct1Context"`
	// Per-character sigil. json:"-" so it never leaks through the player /me view.
	Password string `gorm:"not null;default:''" json:"-"`

	House    *VampireHouse    `gorm:"foreignKey:HouseID" json:"house,omitempty"`
	Secrets  []VampireSecret  `gorm:"foreignKey:CharacterID" json:"secrets,omitempty"`
	Missions []VampireMission `gorm:"foreignKey:CharacterID" json:"missions,omitempty"`
}

func (VampireCharacter) TableName() string { return "vampire_characters" }

type VampireSecret struct {
	ID          uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"not null" json:"updatedAt"`
	CharacterID uuid.UUID `gorm:"not null" json:"characterId"`
	Ordinal     int       `gorm:"not null" json:"ordinal"`
	Body        string    `gorm:"not null;default:''" json:"body"`
}

func (VampireSecret) TableName() string { return "vampire_secrets" }

type VampireMission struct {
	ID           uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt    time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"not null" json:"updatedAt"`
	CharacterID  uuid.UUID `gorm:"not null" json:"characterId"`
	Ordinal      int       `gorm:"not null" json:"ordinal"`
	Tier         string    `gorm:"not null;default:'easy'" json:"tier"` // easy | medium | hard
	RewardBT     int       `gorm:"column:reward_bt;not null;default:0" json:"rewardBt"`
	Prompt       string    `gorm:"not null;default:''" json:"prompt"`
	AnswerFormat string    `gorm:"not null;default:''" json:"answerFormat"`
}

func (VampireMission) TableName() string { return "vampire_missions" }

type VampirePlayer struct {
	ID          uuid.UUID  `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"not null" json:"updatedAt"`
	Token       string     `gorm:"not null" json:"token"`
	CharacterID *uuid.UUID `json:"characterId"`
	GuestLabel  string     `gorm:"not null;default:''" json:"guestLabel"`
	Active      bool       `gorm:"not null;default:true" json:"active"`

	Character *VampireCharacter `gorm:"foreignKey:CharacterID" json:"character,omitempty"`
}

func (VampirePlayer) TableName() string { return "vampire_players" }

type VampireMissionSubmission struct {
	ID           uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt    time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"not null" json:"updatedAt"`
	PlayerID     uuid.UUID `gorm:"not null" json:"playerId"`
	MissionID    uuid.UUID `gorm:"not null" json:"missionId"`
	Status       string    `gorm:"not null;default:'submitted'" json:"status"` // submitted | verified | rejected
	PlayerAnswer string    `gorm:"not null;default:''" json:"playerAnswer"`
	AwardedBT    int       `gorm:"column:awarded_bt;not null;default:0" json:"awardedBt"`
	VerifiedBy   string    `gorm:"not null;default:''" json:"verifiedBy"`
}

func (VampireMissionSubmission) TableName() string { return "vampire_mission_submissions" }

type VampireHouseFavorLedger struct {
	ID        uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	HouseID   uuid.UUID `gorm:"not null" json:"houseId"`
	Delta     int       `gorm:"not null" json:"delta"`
	Reason    string    `gorm:"not null;default:''" json:"reason"`
	GMName    string    `gorm:"column:gm_name;not null;default:''" json:"gmName"`
	Source    string    `gorm:"not null;default:'manual'" json:"source"` // manual | mission | quiz
}

func (VampireHouseFavorLedger) TableName() string { return "vampire_house_favor_ledger" }

type VampireBloodTokenLog struct {
	ID        uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	PlayerID  uuid.UUID `gorm:"not null" json:"playerId"`
	Delta     int       `gorm:"not null" json:"delta"`
	Reason    string    `gorm:"not null;default:''" json:"reason"`
	Source    string    `gorm:"not null;default:'manual'" json:"source"` // manual | mission | physical_game
	GMName    string    `gorm:"column:gm_name;not null;default:''" json:"gmName"`
}

func (VampireBloodTokenLog) TableName() string { return "vampire_blood_token_log" }

type VampireGameState struct {
	ID                   int        `gorm:"primary_key" json:"id"`
	UpdatedAt            time.Time  `gorm:"not null" json:"updatedAt"`
	CurrentAct           string     `gorm:"not null;default:'pre_event'" json:"currentAct"` // pre_event | act1 | act2 | act3 | quiz | resolved
	ContentUnlocked      bool       `gorm:"not null;default:false" json:"contentUnlocked"`
	QuizOpen             bool       `gorm:"not null;default:false" json:"quizOpen"`
	ActiveNotificationID *uuid.UUID `json:"activeNotificationId"`
}

func (VampireGameState) TableName() string { return "vampire_game_state" }

type VampireNotification struct {
	ID        uuid.UUID  `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time  `gorm:"not null" json:"createdAt"`
	Title     string     `gorm:"not null;default:''" json:"title"`
	Body      string     `gorm:"not null;default:''" json:"body"`
	Scope     string     `gorm:"not null;default:'all'" json:"scope"` // all | house | player
	TargetID  *uuid.UUID `json:"targetId"`
	CreatedBy string     `gorm:"not null;default:''" json:"createdBy"`
	Active    bool       `gorm:"not null;default:true" json:"active"`
}

func (VampireNotification) TableName() string { return "vampire_notifications" }

type VampireQuizQuestion struct {
	ID            uuid.UUID      `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt     time.Time      `gorm:"not null" json:"createdAt"`
	UpdatedAt     time.Time      `gorm:"not null" json:"updatedAt"`
	Ordinal       int            `gorm:"not null;default:0" json:"ordinal"`
	Prompt        string         `gorm:"not null;default:''" json:"prompt"`
	QuestionType  string         `gorm:"not null;default:'open'" json:"questionType"` // multiple_choice | open
	Options       datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"options"`
	CorrectAnswer string         `gorm:"not null;default:''" json:"correctAnswer"`
	HFEffect      datatypes.JSON `gorm:"column:hf_effect;type:jsonb;default:'{}'" json:"hfEffect"`
	Active        bool           `gorm:"not null;default:true" json:"active"`
}

func (VampireQuizQuestion) TableName() string { return "vampire_quiz_questions" }

type VampireQuizSubmission struct {
	ID         uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt  time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"not null" json:"updatedAt"`
	PlayerID   uuid.UUID `gorm:"not null" json:"playerId"`
	QuestionID uuid.UUID `gorm:"not null" json:"questionId"`
	Answer     string    `gorm:"not null;default:''" json:"answer"`
	IsCorrect  *bool     `json:"isCorrect"`
	Locked     bool      `gorm:"not null;default:false" json:"locked"`
}

func (VampireQuizSubmission) TableName() string { return "vampire_quiz_submissions" }

type VampireGMActionLog struct {
	ID        uuid.UUID      `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time      `gorm:"not null" json:"createdAt"`
	GMName    string         `gorm:"column:gm_name;not null;default:''" json:"gmName"`
	Action    string         `gorm:"not null;default:''" json:"action"`
	Payload   datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"payload"`
}

func (VampireGMActionLog) TableName() string { return "vampire_gm_action_log" }
