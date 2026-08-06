package server

import (
	"fmt"

	"github.com/MaxBlaushild/poltergeist/pkg/auth"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type server struct {
	authClient  auth.Client
	dbClient    db.DbClient
	asyncClient *asynq.Client // enqueues Part 1 grading jobs onto the job-runner
}

type Server interface {
	ListenAndServe(port string)
	SetupRoutes(r *gin.Engine)
}

func NewServer(
	authClient auth.Client,
	dbClient db.DbClient,
	redisUrl string,
) Server {
	var asyncClient *asynq.Client
	if redisUrl != "" {
		asyncClient = asynq.NewClient(asynq.RedisClientOpt{Addr: util.NormalizeRedisAddr(redisUrl)})
	}
	return &server{
		authClient:  authClient,
		dbClient:    dbClient,
		asyncClient: asyncClient,
	}
}

func (s *server) SetupRoutes(r *gin.Engine) {
	r.GET("/vampire-ascendancy/health", s.GetHealth)

	// Submission photos and item reference photos are served by unguessable
	// id (no token; not secret content).
	r.GET("/vampire-ascendancy/photos/:id", s.getPhoto)
	r.GET("/vampire-ascendancy/items/:id/photo", s.getItemPhoto)

	// Real-account auth — email/password or Google, delegated to the shared
	// authenticator. Unrelated to player sigil auth or the old GM passcode.
	r.POST("/vampire-ascendancy/auth/register", s.registerUser)
	r.POST("/vampire-ascendancy/auth/login", s.loginUser)
	r.POST("/vampire-ascendancy/auth/google", s.loginUserWithGoogle)

	// Platform routes — not scoped to any instance. Creating a Toast,
	// listing "My Toasts", and accepting a Co-Host invite all require a
	// real signed-in account (see withUser).
	r.POST("/vampire-ascendancy/instances", s.withUser, s.createInstance)
	r.GET("/vampire-ascendancy/instances", s.withUser, s.listMyInstances)
	r.POST("/vampire-ascendancy/invites/:token/accept", s.withUser, s.acceptInstanceAdminInvite)

	// Everything else is scoped to one instance ("Toast") via :instanceId.
	inst := r.Group("/vampire-ascendancy/i/:instanceId")

	// Public login routes — pick a character + enter its sigil to get a token.
	inst.GET("/characters", s.listCharactersPublic)
	inst.GET("/characters/:id", s.getCharacterPublic)
	inst.POST("/login", s.login)

	// Public projector feed — standings + games are not secret (players already
	// see them), so the /broadcast screen needs no auth to cast to a TV.
	inst.GET("/broadcast/standings", s.getLeaderboard)
	inst.GET("/broadcast/games", s.getGames)

	// Player routes — authenticated by the per-character token.
	inst.GET("/me", s.withPlayer, s.getMe)
	inst.GET("/state", s.withPlayer, s.getState)
	inst.GET("/leaderboard", s.withPlayer, s.getLeaderboard)
	inst.GET("/houses/:id/overview", s.withPlayer, s.getHouseOverview)
	inst.POST("/missions/:id/submit", s.withPlayer, s.submitMission)
	inst.GET("/quiz", s.withPlayer, s.getQuiz)
	inst.POST("/quiz/part1/submit", s.withPlayer, s.submitQuizPart1)
	inst.POST("/quiz/part2/submit", s.withPlayer, s.submitQuizPart2)
	inst.POST("/quiz/part2/answer", s.withPlayer, s.submitQuizPart2Answer)
	inst.GET("/games", s.withPlayer, s.getGames)
	inst.GET("/inventory", s.withPlayer, s.getInventory)
	inst.POST("/inventory/:id/target", s.withPlayer, s.setInventoryTarget)

	// Admin routes — guarded by a real signed-in Host/Co-Host account
	// (withInstanceAdmin), not the old shared passcode.
	gm := inst.Group("/gm", s.withInstanceAdmin)
	gm.GET("/state", s.gmGetState)
	gm.POST("/unlock", s.gmSetUnlock)
	gm.POST("/act", s.gmSetAct)
	gm.POST("/reset", s.gmResetGame)
	gm.GET("/export", s.gmExportStandings)
	gm.GET("/standings", s.getLeaderboard) // same house standings as the player view
	gm.GET("/standings/breakdown", s.gmStandingsBreakdown)
	gm.GET("/houses", s.getHouses)
	gm.POST("/hf", s.gmAwardHouseFavor)
	gm.POST("/bt", s.gmAwardBloodTokens)
	gm.GET("/submissions", s.gmListSubmissions)
	gm.POST("/submissions/:id/approve", s.gmApproveSubmission)
	gm.POST("/submissions/:id/redeem", s.gmRedeemSubmission)
	gm.POST("/submissions/:id/reject", s.gmRejectSubmission)
	gm.GET("/players", s.gmListPlayers)
	gm.POST("/players", s.gmCreatePlayer)
	gm.PUT("/players/:id", s.gmUpdatePlayer)
	gm.POST("/players/provision", s.gmProvisionPlayers)
	gm.GET("/characters", s.gmListCharacters)
	gm.GET("/characters/:id", s.gmGetCharacter)
	gm.PUT("/characters/:id/portrait", s.gmSetCharacterPortrait)
	gm.GET("/games", s.gmListGames)
	gm.POST("/games", s.gmCreateGame)
	gm.PUT("/games/:id", s.gmUpdateGame)
	gm.DELETE("/games/:id", s.gmDeleteGame)
	gm.POST("/games/:id/result", s.gmRecordGameResult)
	gm.POST("/games/:id/clear", s.gmClearGameResult)
	gm.PUT("/games/:id/schedule", s.gmSetGameSchedule)
	gm.GET("/items", s.gmListItems)
	gm.GET("/player-items", s.gmListPlayerItems)
	gm.POST("/player-items", s.gmAssignItem)
	gm.PUT("/player-items/:id/owner", s.gmTransferPlayerItem)
	gm.DELETE("/player-items/:id", s.gmRemovePlayerItem)
	gm.POST("/notifications", s.gmPushNotification)
	gm.POST("/notifications/clear", s.gmClearNotifications)
	gm.POST("/quiz/part1", s.gmSetPart1Open)
	gm.POST("/quiz/part1/grade", s.gmGradePart1)
	gm.POST("/quiz/part1/regrade", s.gmRegradePart1)
	gm.POST("/quiz/part1/override", s.gmOverridePart1BT)
	gm.POST("/quiz/part2", s.gmSetPart2Open)
	gm.POST("/quiz/part2/rescore", s.gmRescorePart2)
	gm.GET("/quiz/submissions", s.gmListQuizSubmissions)
	gm.GET("/quiz/tally", s.gmQuizTally)

	// Co-Hosts tab: invite/remove/transfer.
	gm.GET("/admins", s.gmListAdmins)
	gm.POST("/admins/invite", s.gmInviteAdmin)
	gm.DELETE("/admins/invites/:id", s.gmDeleteAdminInvite)
	gm.DELETE("/admins/:userId", s.gmRemoveAdmin)
	gm.POST("/admins/transfer", s.gmTransferOwnership)

	// Content tab: include/exclude toggles against the shared library.
	gm.GET("/library/characters", s.gmListLibraryCharacters)
	gm.PUT("/library/characters/:id", s.gmSetCharacterIncluded)
	gm.GET("/library/items", s.gmListLibraryItems)
	gm.PUT("/library/items/:id", s.gmSetItemIncluded)
	gm.GET("/library/quiz-questions", s.gmListLibraryQuizQuestions)
	gm.PUT("/library/quiz-questions/:id", s.gmSetQuizQuestionIncluded)

	// Shared content library editor — characters, houses, items, and quiz
	// questions are global (read by every instance); editing them is
	// restricted to super users, not any instance's Host/Co-Host. Not
	// scoped to an instance at all, unlike everything above.
	admin := r.Group("/vampire-ascendancy/admin", s.withSuperUser)
	admin.GET("/houses", s.adminListHouses)
	admin.PUT("/houses/:id", s.adminUpdateHouse)
	admin.GET("/characters", s.adminListCharacters)
	admin.GET("/characters/:id", s.adminGetCharacter)
	admin.PUT("/characters/:id", s.adminUpdateCharacter)
	admin.GET("/items", s.adminListItems)
	admin.POST("/items", s.adminCreateItem)
	admin.PUT("/items/:id", s.adminUpdateItem)
	admin.DELETE("/items/:id", s.adminDeleteItem)
	admin.POST("/items/:id/photo", s.adminSetItemPhoto)
	admin.DELETE("/items/:id/photo", s.adminDeleteItemPhoto)
	admin.GET("/quiz/questions", s.adminGetQuizQuestions)
	admin.PUT("/quiz/questions", s.adminUpdateQuizQuestions)
	admin.GET("/super-users", s.adminListSuperUsers)
	admin.POST("/super-users", s.adminAddSuperUser)
	admin.DELETE("/super-users/:userId", s.adminRemoveSuperUser)
}

func (s *server) ListenAndServe(port string) {
	r := gin.Default()
	// CORS for the standalone dev server. In production this service is folded
	// into core, which applies its own CORS, so this only matters for local dev.
	r.Use(devCORS)
	s.SetupRoutes(r)
	r.Run(fmt.Sprintf(":%s", port))
}

func devCORS(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Player-Token")
	if ctx.Request.Method == "OPTIONS" {
		ctx.AbortWithStatus(204)
		return
	}
	ctx.Next()
}
