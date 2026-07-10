package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Vampire Ascendancy (The Crimson Toast) event app models.

// VampireItem is an assignable inventory item (relic, clue, gameplay card).
type VampireItem struct {
	ID            uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt     time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"not null" json:"updatedAt"`
	Code          string    `gorm:"not null;default:''" json:"code"`
	Name          string    `gorm:"not null" json:"name"`
	Description   string    `gorm:"not null;default:''" json:"description"`
	Effect        string    `gorm:"not null;default:''" json:"effect"`
	TargetsPlayer bool      `gorm:"column:targets_player;not null;default:false" json:"targetsPlayer"`
	HFEffect      int       `gorm:"column:hf_effect;not null;default:0" json:"hfEffect"` // HF applied to owner's house at reveal
	// Structured Blood Token effects, resolved into the final tally at the reveal.
	BTSelf         int  `gorm:"column:bt_self;not null;default:0" json:"btSelf"`                  // flat BT to owner
	BTFromTarget   int  `gorm:"column:bt_from_target;not null;default:0" json:"btFromTarget"`     // steal N: +N owner, -N target
	BTDeductTarget int  `gorm:"column:bt_deduct_target;not null;default:0" json:"btDeductTarget"` // deduct N from target
	QuizBTPct      int  `gorm:"column:quiz_bt_pct;not null;default:0" json:"quizBtPct"`           // +pct% of owner's Part 1 BT
	DoubleGameBT   bool `gorm:"column:double_game_bt;not null;default:false" json:"doubleGameBt"` // add a copy of owner's game BT
	Immune         bool `gorm:"column:immune;not null;default:false" json:"immune"`               // cancel incoming steals/deducts
	Reflect        bool `gorm:"column:reflect;not null;default:false" json:"reflect"`             // bounce incoming loss to attacker
	StripResistance bool `gorm:"column:strip_resistance;not null;default:false" json:"stripResistance"` // ignore target immune/reflect
}

func (VampireItem) TableName() string { return "vampire_items" }

// VampirePlayerItem is one item owned by a player, optionally aimed at a target.
type VampirePlayerItem struct {
	ID             uuid.UUID  `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt      time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt      time.Time  `gorm:"not null" json:"updatedAt"`
	PlayerID       uuid.UUID  `gorm:"not null" json:"playerId"`
	ItemID         uuid.UUID  `gorm:"not null" json:"itemId"`
	TargetPlayerID *uuid.UUID `gorm:"column:target_player_id" json:"targetPlayerId"`

	Item *VampireItem `gorm:"foreignKey:ItemID" json:"item,omitempty"`
}

func (VampirePlayerItem) TableName() string { return "vampire_player_items" }

// VampireGame is one of the night's physical contests. Its top-three finishers
// are recorded when the GM scores it; the Blood Token / House Favor awards live
// in the ledgers, not here.
type VampireGame struct {
	ID                uuid.UUID  `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt         time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt         time.Time  `gorm:"not null" json:"updatedAt"`
	Ordinal           int        `gorm:"not null;default:0" json:"ordinal"`
	Name              string     `gorm:"not null" json:"name"`
	Status            string     `gorm:"not null;default:'pending'" json:"status"` // pending | played
	// Finishers per place, as JSON arrays of character id strings (multiple allowed).
	FirstCharacterIDs  datatypes.JSON `gorm:"column:first_character_ids;type:jsonb;default:'[]'" json:"-"`
	SecondCharacterIDs datatypes.JSON `gorm:"column:second_character_ids;type:jsonb;default:'[]'" json:"-"`
	ThirdCharacterIDs  datatypes.JSON `gorm:"column:third_character_ids;type:jsonb;default:'[]'" json:"-"`
	// Schedule within the evening: start/end as minutes-of-day (nil = unscheduled).
	StartMinutes *int   `gorm:"column:start_minutes" json:"startMinutes"`
	EndMinutes   *int   `gorm:"column:end_minutes" json:"endMinutes"`
	Location     string `gorm:"not null;default:''" json:"location"`
}

func (VampireGame) TableName() string { return "vampire_games" }

type VampireHouse struct {
	ID        uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
	Name      string    `gorm:"not null" json:"name"`
	SortOrder int       `gorm:"not null;default:0" json:"sortOrder"`
	// Tagline is the house's motto, e.g. "Order is power".
	Tagline string `gorm:"not null;default:''" json:"tagline"`
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
	// ImageURL is the player's portrait for this character (empty until supplied).
	ImageURL string `gorm:"column:image_url;not null;default:''" json:"imageUrl"`
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
	// Sabotage: when set, verifying this mission deducts SabotageHF House Favor
	// from SabotageHouseID. Rare — most missions just award Blood Tokens.
	SabotageHouseID *uuid.UUID `json:"sabotageHouseId"`
	SabotageHF      int        `gorm:"column:sabotage_hf;not null;default:0" json:"sabotageHf"`
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

type VampireSubmissionPhoto struct {
	ID           uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt    time.Time `gorm:"not null" json:"createdAt"`
	SubmissionID uuid.UUID `gorm:"not null" json:"submissionId"`
	ContentType  string    `gorm:"not null;default:'image/jpeg'" json:"contentType"`
	Data         []byte    `gorm:"type:bytea" json:"-"`
}

func (VampireSubmissionPhoto) TableName() string { return "vampire_submission_photos" }

type VampireHouseFavorLedger struct {
	ID        uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	HouseID   uuid.UUID `gorm:"not null" json:"houseId"`
	// Decimal — Part 2 quiz scoring produces fractional House Favor.
	Delta  float64 `gorm:"not null" json:"delta"`
	Reason string  `gorm:"not null;default:''" json:"reason"`
	GMName string  `gorm:"column:gm_name;not null;default:''" json:"gmName"`
	Source string  `gorm:"not null;default:'manual'" json:"source"` // manual | mission | quiz_part2
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
	CurrentAct           string     `gorm:"not null;default:'pre_event'" json:"currentAct"` // pre_event | act1 | act2 | act3 | quiz_part1 | quiz_part2 | resolved
	ContentUnlocked      bool       `gorm:"not null;default:false" json:"contentUnlocked"`
	QuizPart1Open        bool       `gorm:"not null;default:false" json:"quizPart1Open"`
	QuizPart2Open        bool       `gorm:"not null;default:false" json:"quizPart2Open"`
	QuizPart1OpenedAt    *time.Time `json:"quizPart1OpenedAt"`
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
	ID           uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt    time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"not null" json:"updatedAt"`
	Part         int       `gorm:"not null;default:2" json:"part"` // 1 = open-end (BT), 2 = MC (HF)
	Ordinal      int       `gorm:"not null;default:0" json:"ordinal"`
	Prompt       string    `gorm:"not null;default:''" json:"prompt"`
	QuestionType string    `gorm:"not null;default:'open'" json:"questionType"` // multiple_choice | open
	// Part 1 (open-end, AI-graded)
	Rubric string `gorm:"not null;default:''" json:"rubric"`
	MaxBT  int    `gorm:"column:max_bt;not null;default:0" json:"maxBt"`
	// Part 2 (multiple choice, normalized HF)
	Options       datatypes.JSON `gorm:"type:jsonb;default:'[]'" json:"options"`
	CorrectAnswer string         `gorm:"not null;default:''" json:"correctAnswer"`
	HFValue       float64        `gorm:"column:hf_value;not null;default:0" json:"hfValue"`
	Tier          string         `gorm:"not null;default:''" json:"tier"`
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
	IsCorrect  *bool     `json:"isCorrect"` // Part 2 auto-grade
	AIScore     *float64 `gorm:"column:ai_score" json:"aiScore"`
	AIRationale string   `gorm:"column:ai_rationale;not null;default:''" json:"aiRationale"` // one-line AI note
	AwardedBT   int      `gorm:"column:awarded_bt;not null;default:0" json:"awardedBt"`      // Part 1 BT
	Locked      bool     `gorm:"not null;default:false" json:"locked"`
	// Part 1 grading state machine (async job): '' → queued → grading → graded | failed.
	GradeStatus    string     `gorm:"column:grade_status;not null;default:''" json:"gradeStatus"`
	GradeError     string     `gorm:"column:grade_error;not null;default:''" json:"gradeError"`
	GradeStartedAt *time.Time `gorm:"column:grade_started_at" json:"gradeStartedAt"`
	GradeAttempts  int        `gorm:"column:grade_attempts;not null;default:0" json:"gradeAttempts"`
}

func (VampireQuizSubmission) TableName() string { return "vampire_quiz_submissions" }

// Part 1 grading state machine states.
const (
	QuizGradeStatusQueued  = "queued"  // enqueued, awaiting a worker
	QuizGradeStatusGrading = "grading" // a worker is grading it
	QuizGradeStatusGraded  = "graded"  // graded successfully, BT applied
	QuizGradeStatusFailed  = "failed"  // last attempt errored
)

type VampireGMActionLog struct {
	ID        uuid.UUID      `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time      `gorm:"not null" json:"createdAt"`
	GMName    string         `gorm:"column:gm_name;not null;default:''" json:"gmName"`
	Action    string         `gorm:"not null;default:''" json:"action"`
	Payload   datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"payload"`
}

func (VampireGMActionLog) TableName() string { return "vampire_gm_action_log" }
