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
	gmPasscode  string
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
		authClient: authClient,
		dbClient:   dbClient,
		// GM admin passcode. In prod this comes from the ECS task secrets; for
		// local dev set GM_PASSCODE in local.env.
		gmPasscode:  "bloodmoon",
		asyncClient: asyncClient,
	}
}

func (s *server) SetupRoutes(r *gin.Engine) {
	r.GET("/vampire-ascendancy/health", s.GetHealth)

	// Submission photos are served by unguessable id (no token; not secret content).
	r.GET("/vampire-ascendancy/photos/:id", s.getPhoto)

	// Public login routes — pick a character + enter its sigil to get a token.
	r.GET("/vampire-ascendancy/characters", s.listCharactersPublic)
	r.GET("/vampire-ascendancy/characters/:id", s.getCharacterPublic)
	r.POST("/vampire-ascendancy/login", s.login)

	// Player routes — authenticated by the per-character token.
	r.GET("/vampire-ascendancy/me", s.withPlayer, s.getMe)
	r.GET("/vampire-ascendancy/state", s.withPlayer, s.getState)
	r.GET("/vampire-ascendancy/leaderboard", s.withPlayer, s.getLeaderboard)
	r.GET("/vampire-ascendancy/houses/:id/overview", s.withPlayer, s.getHouseOverview)
	r.POST("/vampire-ascendancy/missions/:id/submit", s.withPlayer, s.submitMission)
	r.GET("/vampire-ascendancy/quiz", s.withPlayer, s.getQuiz)
	r.POST("/vampire-ascendancy/quiz/part1/submit", s.withPlayer, s.submitQuizPart1)
	r.POST("/vampire-ascendancy/quiz/part2/submit", s.withPlayer, s.submitQuizPart2)
	r.POST("/vampire-ascendancy/quiz/part2/answer", s.withPlayer, s.submitQuizPart2Answer)
	r.GET("/vampire-ascendancy/games", s.withPlayer, s.getGames)
	r.GET("/vampire-ascendancy/inventory", s.withPlayer, s.getInventory)
	r.POST("/vampire-ascendancy/inventory/:id/target", s.withPlayer, s.setInventoryTarget)

	// GM admin routes — guarded by the shared passcode.
	gm := r.Group("/vampire-ascendancy/gm", s.withGM)
	gm.GET("/state", s.gmGetState)
	gm.POST("/unlock", s.gmSetUnlock)
	gm.POST("/act", s.gmSetAct)
	gm.POST("/reset", s.gmResetGame)
	gm.GET("/export", s.gmExportStandings)
	gm.GET("/standings", s.getLeaderboard) // same house standings as the player view
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
	gm.GET("/characters", s.gmListCharacters)
	gm.GET("/characters/:id", s.gmGetCharacter)
	gm.PUT("/characters/:id", s.gmUpdateCharacter)
	gm.PUT("/houses/:id", s.gmUpdateHouse)
	gm.GET("/games", s.gmListGames)
	gm.POST("/games", s.gmCreateGame)
	gm.PUT("/games/:id", s.gmUpdateGame)
	gm.DELETE("/games/:id", s.gmDeleteGame)
	gm.POST("/games/:id/result", s.gmRecordGameResult)
	gm.POST("/games/:id/clear", s.gmClearGameResult)
	gm.GET("/items", s.gmListItems)
	gm.POST("/items", s.gmCreateItem)
	gm.PUT("/items/:id", s.gmUpdateItem)
	gm.DELETE("/items/:id", s.gmDeleteItem)
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
	gm.GET("/quiz/questions", s.gmGetQuizQuestions)
	gm.PUT("/quiz/questions", s.gmUpdateQuizQuestions)
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
	ctx.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Player-Token, X-GM-Passcode, X-GM-Name")
	if ctx.Request.Method == "OPTIONS" {
		ctx.AbortWithStatus(204)
		return
	}
	ctx.Next()
}
