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
	// MysteryID is chosen once at "Host a Toast" time and never changes —
	// see MYSTERY_REQUIREMENTS.md. Every quiz question and character secret
	// available in this instance is scoped to this mystery.
	MysteryID uuid.UUID       `gorm:"column:mystery_id;not null" json:"mysteryId"`
	Mystery   *VampireMystery `gorm:"foreignKey:MysteryID" json:"mystery,omitempty"`
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
	// Included and Sigil are DEPRECATED — they were the walk-up "pick a
	// character + sigil" login's mechanisms. Player auth moved to real
	// accounts via invites (VampirePlayerInvite); a character's
	// availability is now implicit (has an accepted invite or not). Left on
	// the struct only so old rows still scan; do not read or write them.
	Included bool   `gorm:"not null;default:true" json:"-"`
	Sigil    string `gorm:"not null;default:''" json:"-"`
	// ImageURL is the player's portrait for this character in this
	// instance — unaffected by the above, still in active use.
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

// VampireMystery is the underlying story an instance's players are trying
// to solve — shared content, edited by super users. See
// MYSTERY_REQUIREMENTS.md. An instance picks exactly one mystery at
// creation and never changes it.
type VampireMystery struct {
	ID        uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
	Name      string    `gorm:"not null" json:"name"`
	Summary   string    `gorm:"not null;default:''" json:"summary"`
	FullLore  string    `gorm:"column:full_lore;not null;default:''" json:"fullLore"`
	// Active hides a mystery from the "Host a Toast" picker without
	// deleting it (and without breaking existing instances that reference
	// it).
	Active bool `gorm:"not null;default:true" json:"active"`
	// IsSubplot distinguishes a subplot from a main mystery. Structurally
	// identical row (summary/full lore/beats/secrets all work the same way
	// against either), but a subplot is selected zero-or-many-at-a-time
	// alongside an instance's one required mystery, doesn't gate invite
	// eligibility, and doesn't have quiz questions — see
	// vampire_instance_subplots and MYSTERY_REQUIREMENTS.md.
	IsSubplot bool `gorm:"column:is_subplot;not null;default:false" json:"isSubplot"`

	// Beats are many-to-many via vampire_mystery_beat_links (a beat can be
	// shared across multiple mysteries/subplots — see VampireMysteryBeatLink),
	// so there's no direct GORM relation here; read them via
	// db.VampireHandle's ListBeatsForMystery instead.
}

func (VampireMystery) TableName() string { return "vampire_mysteries" }

// VampireInstanceSubplot is one subplot selected for one instance —
// zero-or-many per instance, alongside the instance's one required
// mystery_id. No status/ordinal; a subplot is either selected or not.
type VampireInstanceSubplot struct {
	InstanceID uuid.UUID `gorm:"column:instance_id;primaryKey" json:"instanceId"`
	MysteryID  uuid.UUID `gorm:"column:mystery_id;primaryKey" json:"mysteryId"`
	CreatedAt  time.Time `gorm:"not null" json:"createdAt"`
}

func (VampireInstanceSubplot) TableName() string { return "vampire_instance_subplots" }

// VampireMysteryBeat is one discoverable fact — shared, reusable content
// that can be attached to multiple mysteries/subplots at once (see
// VampireMysteryBeatLink), not owned by exactly one. Secrets point at the
// beat they reveal; many secrets (across any mystery that has this beat
// attached) may point at the same beat.
type VampireMysteryBeat struct {
	ID          uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"not null" json:"updatedAt"`
	Title       string    `gorm:"not null;default:''" json:"title"`
	Description string    `gorm:"not null;default:''" json:"description"`
}

func (VampireMysteryBeat) TableName() string { return "vampire_mystery_beats" }

// VampireMysteryBeatLink attaches a beat to one mystery (or subplot), with
// its own ordinal — the same beat can be attached to multiple
// mysteries/subplots, each at its own position in that mystery's list.
// Editing a beat's title/description changes it everywhere it's linked;
// unlinking (removing it from one mystery's list) deletes only this row,
// never the beat itself — see ReplaceMysteryBeats.
type VampireMysteryBeatLink struct {
	MysteryID uuid.UUID `gorm:"column:mystery_id;primaryKey" json:"mysteryId"`
	BeatID    uuid.UUID `gorm:"column:beat_id;primaryKey" json:"beatId"`
	Ordinal   int       `gorm:"not null;default:0" json:"ordinal"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
}

func (VampireMysteryBeatLink) TableName() string { return "vampire_mystery_beat_links" }

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
	ID           uuid.UUID  `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt    time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt    time.Time  `gorm:"not null" json:"updatedAt"`
	Name         string     `gorm:"not null" json:"name"`
	Title        string     `gorm:"not null;default:''" json:"title"`
	HouseID      *uuid.UUID `json:"houseId"`
	RoleType     string     `gorm:"not null;default:'player'" json:"roleType"` // player | gm | npc
	IsOptional   bool       `gorm:"not null;default:false" json:"isOptional"`
	PreEventInfo string     `gorm:"not null;default:''" json:"preEventInfo"`
	// Bio is the character's actual biography — separate from PreEventInfo
	// ("Pre-event bio" in the admin UI), which predates mystery-scoping and
	// doubles as GM/player prep material. Backfilled from PreEventInfo when
	// added (see migration 000470); free to diverge from there.
	Bio string `gorm:"not null;default:''" json:"bio"`
	// Tags are free-text personality/trait labels ("musical", "gambler",
	// "risk taker") — shared content, edited by super users (by hand, or
	// generated by the LLM job below). Used by the Invites tab to
	// filter/search the character picker.
	Tags StringArray `gorm:"type:jsonb;default:'[]'" json:"tags"`
	// TagsGenerationStatus/Error track the AI tag-generation job for this
	// character, mirroring VampireQuizSubmission's grade_status/grade_error
	// (see CharacterTagsStatus* below).
	TagsGenerationStatus string `gorm:"column:tags_generation_status;not null;default:''" json:"tagsGenerationStatus"`
	TagsGenerationError  string `gorm:"column:tags_generation_error;not null;default:''" json:"tagsGenerationError"`
	// ImageURL and Password are DEPRECATED — a character is now shared across
	// every instance that includes it, so its portrait and sigil moved to
	// VampireInstanceCharacter (one per instance). Left on this struct only so
	// old rows still scan; do not read or write them going forward.
	ImageURL string `gorm:"column:image_url;not null;default:''" json:"-"`
	Password string `gorm:"not null;default:''" json:"-"`

	House    *VampireHouse    `gorm:"foreignKey:HouseID" json:"house,omitempty"`
	Secrets  []VampireSecret  `gorm:"foreignKey:CharacterID" json:"secrets,omitempty"`
	Missions []VampireMission `gorm:"foreignKey:CharacterID" json:"missions,omitempty"`
	// PostAct1Contexts spans every mystery this character has ever been
	// cast in — same "unfiltered, across every mystery" shape as Secrets
	// above. Used only by the tag-generation prompt, which wants the full
	// picture; player- and GM-facing reads go through
	// GetCharacterMysteryContext, scoped to one mystery.
	PostAct1Contexts []VampireCharacterMysteryContext `gorm:"foreignKey:CharacterID" json:"-"`
}

func (VampireCharacter) TableName() string { return "vampire_characters" }

// VampireCharacterMysteryContext is a character's pre- and post-Act-1
// context for one specific mystery — the mystery-scoped equivalent of the
// old VampireCharacter.PostAct1Context column, joined by PreAct1Context
// (mystery-scoped from the start; the old character-global equivalent was
// PreEventInfo — see migration 000471). One row per (character, mystery);
// unlike secrets these are single strings, so they're upserted, not
// wholesale-replaced.
type VampireCharacterMysteryContext struct {
	ID              uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt       time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt       time.Time `gorm:"not null" json:"updatedAt"`
	CharacterID     uuid.UUID `gorm:"column:character_id;not null" json:"characterId"`
	MysteryID       uuid.UUID `gorm:"column:mystery_id;not null" json:"mysteryId"`
	PreAct1Context  string    `gorm:"column:pre_act1_context;not null;default:''" json:"preAct1Context"`
	PostAct1Context string    `gorm:"column:post_act1_context;not null;default:''" json:"postAct1Context"`
}

func (VampireCharacterMysteryContext) TableName() string {
	return "vampire_character_mystery_contexts"
}

type VampireSecret struct {
	ID          uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt   time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"not null" json:"updatedAt"`
	CharacterID uuid.UUID `gorm:"not null" json:"characterId"`
	Ordinal     int       `gorm:"not null" json:"ordinal"`
	Body        string    `gorm:"not null;default:''" json:"body"`
	// MysteryID/BeatID scope this secret to one mystery and the beat it
	// reveals. Nullable indefinitely — not just during backfill — since a
	// secret with no mystery is simply unusable by any instance (see
	// MYSTERY_REQUIREMENTS.md's eligibility gating). BeatID is a
	// form-required field in the Super Admin editor, not a hard DB
	// constraint, so an in-progress secret can be saved before its beat is
	// decided.
	MysteryID *uuid.UUID `gorm:"column:mystery_id" json:"mysteryId"`
	BeatID    *uuid.UUID `gorm:"column:beat_id" json:"beatId"`
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
	// MysteryID scopes this mission to one mystery (or subplot) — same
	// rationale and nullability posture as VampireSecret.MysteryID: a
	// character can have a different set of missions per mystery/subplot,
	// and a mission with no mystery is simply unusable by any instance.
	MysteryID *uuid.UUID `gorm:"column:mystery_id" json:"mysteryId"`
}

func (VampireMission) TableName() string { return "vampire_missions" }

type VampirePlayer struct {
	ID         uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt  time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"not null" json:"updatedAt"`
	InstanceID uuid.UUID `gorm:"column:instance_id;not null" json:"instanceId"`
	// Token is DEPRECATED — player auth moved to real accounts (UserID).
	// Left on the struct only so old rows still scan; do not read or write
	// it going forward.
	Token       string     `gorm:"not null" json:"-"`
	UserID      *uuid.UUID `gorm:"column:user_id" json:"userId"`
	CharacterID *uuid.UUID `json:"characterId"`
	GuestLabel  string     `gorm:"not null;default:''" json:"guestLabel"`
	Active      bool       `gorm:"not null;default:true" json:"active"`

	Character *VampireCharacter `gorm:"foreignKey:CharacterID" json:"character,omitempty"`
}

func (VampirePlayer) TableName() string { return "vampire_players" }

// VampirePlayerInvite is a Host/Co-Host's invitation of a specific real
// person (by phone number) to join one instance as a player. Character-
// agnostic — accepting it (after signing in/up with a real account, same
// as Hosts) creates the VampirePlayer row with no character yet; the
// player picks their own from the instance's curated pool afterward (see
// VampireInstanceCharacterPool). Replaces the old walk-up "pick your
// character + sigil" login, and the version of this flow where a Host
// assigned the character up front.
type VampirePlayerInvite struct {
	ID             uuid.UUID  `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt      time.Time  `gorm:"not null" json:"createdAt"`
	UpdatedAt      time.Time  `gorm:"not null" json:"updatedAt"`
	InstanceID     uuid.UUID  `gorm:"column:instance_id;not null" json:"instanceId"`
	GuestName      string     `gorm:"column:guest_name;not null;default:''" json:"guestName"`
	PhoneNumber    string     `gorm:"column:phone_number;not null;default:''" json:"phoneNumber"`
	Token          string     `gorm:"not null" json:"-"`
	Status         string     `gorm:"not null;default:'pending'" json:"status"` // pending | accepted | declined
	InvitedBy      *uuid.UUID `gorm:"column:invited_by" json:"invitedBy"`
	AcceptedUserID *uuid.UUID `gorm:"column:accepted_user_id" json:"acceptedUserId"`
	RespondedAt    *time.Time `gorm:"column:responded_at" json:"respondedAt"`
}

func (VampirePlayerInvite) TableName() string { return "vampire_player_invites" }

const (
	PlayerInviteStatusPending  = "pending"
	PlayerInviteStatusAccepted = "accepted"
	PlayerInviteStatusDeclined = "declined"
)

// VampireInstanceCharacterPool is one character a Host has made selectable
// by players in one instance — curated from the full mystery-eligible set
// (see MYSTERY_REQUIREMENTS.md's secrets-based eligibility). A character
// not in the pool can't be self-selected at RSVP-accept time even if
// otherwise eligible.
type VampireInstanceCharacterPool struct {
	InstanceID  uuid.UUID `gorm:"column:instance_id;primaryKey" json:"instanceId"`
	CharacterID uuid.UUID `gorm:"column:character_id;primaryKey" json:"characterId"`
	CreatedAt   time.Time `gorm:"not null" json:"createdAt"`
}

func (VampireInstanceCharacterPool) TableName() string { return "vampire_instance_character_pool" }

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
	ID uuid.UUID `gorm:"primary_key;default:uuid_generate_v4()" json:"id"`
	// MysteryID scopes this question to one mystery — the quiz is no
	// longer a single global set trimmed per instance (see
	// MYSTERY_REQUIREMENTS.md); an instance's quiz is simply every
	// question whose MysteryID matches its own.
	MysteryID    uuid.UUID `gorm:"column:mystery_id;not null" json:"mysteryId"`
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

// Character AI tag-generation state machine states — same shape as the
// quiz-grading ones above, one LLM job per character instead of per answer.
const (
	CharacterTagsStatusQueued     = "queued"     // enqueued, awaiting a worker
	CharacterTagsStatusGenerating = "generating" // a worker is generating tags
	CharacterTagsStatusGenerated  = "generated"  // generated successfully, tags saved
	CharacterTagsStatusFailed     = "failed"     // last attempt errored
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
