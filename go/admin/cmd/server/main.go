package main

import (
	"context"
	"fmt"
	"io"

	"github.com/GoAdminGroup/themes/adminlte"
	"github.com/MaxBlaushild/poltergeist/admin/internal/config"
	"github.com/MaxBlaushild/poltergeist/admin/internal/migrate"
	"github.com/MaxBlaushild/poltergeist/admin/internal/renderer"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	ada "github.com/GoAdminGroup/go-admin/adapter/gin"
	"github.com/GoAdminGroup/go-admin/engine"
	adminConfig "github.com/GoAdminGroup/go-admin/modules/config"
)

func main() {
	ctx := context.Background()

	cfg, err := config.ParseFlagsAndGetConfig()
	if err != nil {
		panic(err)
	}

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	if err := migrate.Migrate(ctx, cfg); err != nil {
		fmt.Println("HSAKJHSJKHAJKS")
		fmt.Println(err.Error())
	}

	ginEngine := gin.New()
	adminEngine := engine.Default()
	rndrer := renderer.NewRenderer()

	if err := adminEngine.AddConfig(&adminConfig.Config{
		Env: cfg.Public.AdminEnv,
		Databases: adminConfig.DatabaseList{
			"default": adminConfig.Database{
				MaxIdleConns: 5,
				MaxOpenConns: 5,
				Driver:       adminConfig.DriverPostgresql,
				Dsn:          fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", cfg.Public.DbUser, cfg.Secret.DbPassword, cfg.Public.DbHost, cfg.Public.DbPort, cfg.Public.DbName, cfg.Public.SslMode),
			},
		},
		UrlPrefix:   "/admin",
		Language:    "en",
		Debug:       false,
		IndexUrl:    "/",
		ColorScheme: adminlte.ColorschemeSkinBlack,
		Animation: adminConfig.PageAnimation{
			Type: "fadeInUp",
		},
	}).
		AddGenerator("crystals", rndrer.GetCrystals).
		AddGenerator("crystal-unlockings", rndrer.GetCrystalUnlockings).
		AddGenerator("how-many-answers", rndrer.GetHowManyAnswers).
		AddGenerator("how-many-questions", rndrer.GetHowManyQuestions).
		AddGenerator("neighbors", rndrer.GetNeighbors).
		AddGenerator("teams", rndrer.GetTeams).
		AddGenerator("user-teams", rndrer.GetUserTeams).
		AddGenerator("users", rndrer.GetUserTeams).
		Use(ginEngine); err != nil {
		panic(err)
	}

	ginEngine.GET("/admin", ada.Content(rndrer.GetDashboard))

	if err := ginEngine.Run(":9093"); err != nil {
		panic(err)
	}
}
