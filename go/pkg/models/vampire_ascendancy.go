package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Vampire Ascendancy (The Crimson Toast) event app models.
//
// Multi-tenancy: an "instance" is one group's run of the event ("hosting a
// Toast" in user-facing copy — see go/vampire-ascendancy/docs/
// MULTI_TENANT_REQUIREMENTS.md). The content library below (houses,
// characters, secrets, missions, items, quiz questions) stays global and
// shared across every instance; everything else — game state, players,
// ledgers, submissions, notifications, the audit log — belongs to exactly
// one instance via InstanceID.

// VampireInstance is one group's run of the event.
type VampireInstance struct {
	ID        uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
	Name      string    `gorm:"not null;default:''" json:"name"`
	Status    string    `gorm:"not null;default:'active'" json:"status"` // active | archived
	// Nullable: the legacy instance backfilled from the pre-multi-tenant
	// event has no single creator on record.
	CreatedBy *uuid.UUID `gorm:"column:created_by" json:"createdBy"`
}

func (VampireInstance) TableName() string { return "vampire_instances" }

// VampireInstanceAdmin is a user's administrative membership on an instance
// ("Host" for role=owner, "Co-Host" for role=admin in user-facing copy).
type VampireInstanceAdmin struct {
	ID         uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt  time.Time `gorm:"not null" json:"createdAt"`
	InstanceID uuid.UUID `gorm:"not null" json:"instanceId"`
	UserID     uuid.UUID `gorm:"not null" json:"userId"`
	Role       string    `gorm:"not null;default:'admin'" json:"role"` // owner | admin

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (VampireInstanceAdmin) TableName() string { return "vampire_instance_admins" }

const (
	InstanceAdminRoleOwner = "owner"
	InstanceAdminRoleAdmin = "admin"
)

// VampireInstanceAdminInvite is a pending "invite a Co-Host" by email. The
// invitee may not have a User account yet, so this is keyed by email + a
// bearer token until accepted.
type VampireInstanceAdminInvite struct {
	ID         uuid.UUID  `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt  time.Time  `gorm:"not null" json:"createdAt"`
	InstanceID uuid.UUID  `gorm:"not null" json:"instanceId"`
	Email      string     `gorm:"not null" json:"email"`
	InvitedBy  uuid.UUID  `gorm:"column:invited_by;not null" json:"invitedBy"`
	Token      string     `gorm:"not null" json:"-"`
	ExpiresAt  time.Time  `gorm:"column:expires_at;not null" json:"expiresAt"`
	AcceptedAt *time.Time `gorm:"column:accepted_at" json:"acceptedAt"`
}

func (VampireInstanceAdminInvite) TableName() string { return "vampire_instance_admin_invites" }

// VampireInstanceCharacter is one instance's inclusion toggle for a library
// character, plus the per-instance secret (sigil) and portrait that used to
// live on VampireCharacter before an event could run more than one at a
// time.
type VampireInstanceCharacter struct {
	InstanceID  uuid.UUID `gorm:"primary_key;column:instance_id" json:"instanceId"`
	CharacterID uuid.UUID `gorm:"primary_key;column:character_id" json:"characterId"`
	CreatedAt   time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"not null" json:"updatedAt"`
	Included    bool      `gorm:"not null;default:true" json:"included"`
	// Sigil (PIN) validates a player landed on the right character. json:"-"
	// so it never leaks through a player-facing response.
	Sigil string `gorm:"not null;default:''" json:"-"`
	// ImageURL is the player's portrait for this character in this instance.
	ImageURL string `gorm:"column:image_url;not null;default:''" json:"imageUrl"`
}

func (VampireInstanceCharacter) TableName() string { return "vampire_instance_characters" }

// VampireInstanceItem is one instance's inclusion toggle for a library item.
type VampireInstanceItem struct {
	InstanceID uuid.UUID `gorm:"primary_key;column:instance_id" json:"instanceId"`
	ItemID     uuid.UUID `gorm:"primary_key;column:item_id" json:"itemId"`
	CreatedAt  time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"not null" json:"updatedAt"`
	Included   bool      `gorm:"not null;default:true" json:"included"`
}

func (VampireInstanceItem) TableName() string { return "vampire_instance_items" }

// VampireInstanceQuizQuestion is one instance's inclusion toggle for a
// library quiz question.
type VampireInstanceQuizQuestion struct {
	InstanceID uuid.UUID `gorm:"primary_key;column:instance_id" json:"instanceId"`
	QuestionID uuid.UUID `gorm:"primary_key;column:question_id" json:"questionId"`
	CreatedAt  time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"not null" json:"updatedAt"`
	Included   bool      `gorm:"not null;default:true" json:"included"`
}

func (VampireInstanceQuizQuestion) TableName() string { return "vampire_instance_quiz_questions" }

// VampireSuperUser is a user allowed to edit the shared content library
// (characters, houses, items, quiz questions) across every instance — a
// separate, higher tier than Host/Co-Host, which only toggles inclusion
// within one instance. See MULTI_TENANT_REQUIREMENTS.md.
type VampireSuperUser struct {
	UserID    uuid.UUID  `gorm:"primary_key;column:user_id" json:"userId"`
	CreatedAt time.Time  `gorm:"not null" json:"createdAt"`
	CreatedBy *uuid.UUID `gorm:"column:created_by" json:"createdBy"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (VampireSuperUser) TableName() string { return "vampire_super_users" }

// VampireSuperUserActionLog audits edits to the shared content library —
// the global-content equivalent of VampireGMActionLog, which is per-instance.
type VampireSuperUserActionLog struct {
	ID        uuid.UUID      `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time      `gorm:"not null" json:"createdAt"`
	UserID    *uuid.UUID     `gorm:"column:user_id" json:"userId"`
	Action    string         `gorm:"not null;default:''" json:"action"`
	Payload   datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"payload"`
}

func (VampireSuperUserActionLog) TableName() string { return "vampire_super_user_action_log" }

// VampireItem is an assignable inventory item (relic, clue, gameplay card).
type VampireItem struct {
	ID            uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt     time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt     time.Time `gorm:"not null" json:"updatedAt"`
	Code          string    `gorm:"not null;default:''" json:"code"`
	Name          string    `gorm:"not null" json:"name"`
	Category      string    `gorm:"not null;default:''" json:"category"` // e.g. War / Glory / Protection
	Description   string    `gorm:"not null;default:''" json:"description"`
	Effect        string    `gorm:"not null;default:''" json:"effect"`
	TargetsPlayer bool      `gorm:"column:targets_player;not null;default:false" json:"targetsPlayer"`
	HFEffect      int       `gorm:"column:hf_effect;not null;default:0" json:"hfEffect"` // HF applied to owner's house at reveal
	// Structured Blood Token effects, resolved into the final tally at the reveal.
	BTSelf          int  `gorm:"column:bt_self;not null;default:0" json:"btSelf"`                       // flat BT to owner
	BTFromTarget    int  `gorm:"column:bt_from_target;not null;default:0" json:"btFromTarget"`          // steal N: +N owner, -N target
	BTDeductTarget  int  `gorm:"column:bt_deduct_target;not null;default:0" json:"btDeductTarget"`      // deduct N from target
	QuizBTPct       int  `gorm:"column:quiz_bt_pct;not null;default:0" json:"quizBtPct"`                // +pct% of owner's Part 1 BT
	DoubleGameBT    bool `gorm:"column:double_game_bt;not null;default:false" json:"doubleGameBt"`      // add a copy of owner's game BT
	Immune          bool `gorm:"column:immune;not null;default:false" json:"immune"`                    // cancel incoming steals/deducts
	Reflect         bool `gorm:"column:reflect;not null;default:false" json:"reflect"`                  // bounce incoming loss to attacker
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
	ID         uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt  time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"not null" json:"updatedAt"`
	InstanceID uuid.UUID `gorm:"column:instance_id;not null" json:"instanceId"`
	Ordinal    int       `gorm:"not null;default:0" json:"ordinal"`
	Name       string    `gorm:"not null" json:"name"`
	Status     string    `gorm:"not null;default:'pending'" json:"status"` // pending | played
	// Finishers per place, as JSON arrays of character id strings (multiple allowed).
	FirstCharacterIDs  datatypes.JSON `gorm:"column:first_character_ids;type:jsonb;default:'[]'" json:"-"`
	SecondCharacterIDs datatypes.JSON `gorm:"column:second_character_ids;type:jsonb;default:'[]'" json:"-"`
	ThirdCharacterIDs  datatypes.JSON `gorm:"column:third_character_ids;type:jsonb;default:'[]'" json:"-"`
	// Schedule within the evening: start/end as minutes-of-day (nil = unscheduled).
	StartMinutes *int   `gorm:"column:start_minutes" json:"startMinutes"`
	EndMinutes   *int   `gorm:"column:end_minutes" json:"endMinutes"`
	Location     string `gorm:"not null;default:''" json:"location"`
	AssignedGM   string `gorm:"column:assigned_gm;not null;default:''" json:"assignedGm"` // GM running it
	RunNotes     string `gorm:"column:run_notes;not null;default:''" json:"-"`            // GM-only how-to
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
	// ImageURL and Password are DEPRECATED — a character is now shared across
	// every instance that includes it, so its portrait and sigil moved to
	// VampireInstanceCharacter (one per instance). Left on this struct only so
	// old rows still scan; do not read or write them going forward.
	ImageURL string `gorm:"column:image_url;not null;default:''" json:"-"`
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
	InstanceID  uuid.UUID  `gorm:"column:instance_id;not null" json:"instanceId"`
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
	InstanceID   uuid.UUID `gorm:"column:instance_id;not null" json:"instanceId"`
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

// VampireItemPhoto is a single GM-facing reference photo for a catalog item.
type VampireItemPhoto struct {
	ItemID      uuid.UUID `gorm:"primary_key;column:item_id" json:"itemId"`
	ContentType string    `gorm:"column:content_type;not null;default:'image/jpeg'" json:"contentType"`
	Data        []byte    `gorm:"column:data;type:bytea" json:"-"`
	UpdatedAt   time.Time `gorm:"not null" json:"updatedAt"`
}

func (VampireItemPhoto) TableName() string { return "vampire_item_photos" }

type VampireHouseFavorLedger struct {
	ID         uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt  time.Time `gorm:"not null" json:"createdAt"`
	InstanceID uuid.UUID `gorm:"column:instance_id;not null" json:"instanceId"`
	HouseID    uuid.UUID `gorm:"not null" json:"houseId"`
	// Decimal — Part 2 quiz scoring produces fractional House Favor.
	Delta  float64 `gorm:"not null" json:"delta"`
	Reason string  `gorm:"not null;default:''" json:"reason"`
	GMName string  `gorm:"column:gm_name;not null;default:''" json:"gmName"`
	Source string  `gorm:"not null;default:'manual'" json:"source"` // manual | mission | quiz_part2
}

func (VampireHouseFavorLedger) TableName() string { return "vampire_house_favor_ledger" }

type VampireBloodTokenLog struct {
	ID         uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt  time.Time `gorm:"not null" json:"createdAt"`
	InstanceID uuid.UUID `gorm:"column:instance_id;not null" json:"instanceId"`
	PlayerID   uuid.UUID `gorm:"not null" json:"playerId"`
	Delta      int       `gorm:"not null" json:"delta"`
	Reason     string    `gorm:"not null;default:''" json:"reason"`
	Source     string    `gorm:"not null;default:'manual'" json:"source"` // manual | mission | physical_game
	GMName     string    `gorm:"column:gm_name;not null;default:''" json:"gmName"`
}

func (VampireBloodTokenLog) TableName() string { return "vampire_blood_token_log" }

// VampireGameState was a singleton row (id always 1) before multi-tenancy;
// it is now one row per instance, keyed by InstanceID.
type VampireGameState struct {
	InstanceID           uuid.UUID  `gorm:"primary_key;column:instance_id" json:"instanceId"`
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
	ID         uuid.UUID  `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt  time.Time  `gorm:"not null" json:"createdAt"`
	InstanceID uuid.UUID  `gorm:"column:instance_id;not null" json:"instanceId"`
	Title      string     `gorm:"not null;default:''" json:"title"`
	Body       string     `gorm:"not null;default:''" json:"body"`
	Scope      string     `gorm:"not null;default:'all'" json:"scope"` // all | house | player
	TargetID   *uuid.UUID `json:"targetId"`
	CreatedBy  string     `gorm:"not null;default:''" json:"createdBy"`
	Active     bool       `gorm:"not null;default:true" json:"active"`
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
	ID          uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"not null" json:"updatedAt"`
	InstanceID  uuid.UUID `gorm:"column:instance_id;not null" json:"instanceId"`
	PlayerID    uuid.UUID `gorm:"not null" json:"playerId"`
	QuestionID  uuid.UUID `gorm:"not null" json:"questionId"`
	Answer      string    `gorm:"not null;default:''" json:"answer"`
	IsCorrect   *bool     `json:"isCorrect"` // Part 2 auto-grade
	AIScore     *float64  `gorm:"column:ai_score" json:"aiScore"`
	AIRationale string    `gorm:"column:ai_rationale;not null;default:''" json:"aiRationale"` // one-line AI note
	AwardedBT   int       `gorm:"column:awarded_bt;not null;default:0" json:"awardedBt"`      // Part 1 BT
	Locked      bool      `gorm:"not null;default:false" json:"locked"`
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
	ID         uuid.UUID      `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt  time.Time      `gorm:"not null" json:"createdAt"`
	InstanceID uuid.UUID      `gorm:"column:instance_id;not null" json:"instanceId"`
	GMName     string         `gorm:"column:gm_name;not null;default:''" json:"gmName"`
	Action     string         `gorm:"not null;default:''" json:"action"`
	Payload    datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"payload"`
}

func (VampireGMActionLog) TableName() string { return "vampire_gm_action_log" }
