package renderer

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/gin-gonic/gin"
)

type renderer struct{}

type Renderer interface {
	GetCrystalUnlockings(ctx *context.Context) table.Table
	GetCrystals(ctx *context.Context) table.Table
	GetHowManyAnswers(ctx *context.Context) table.Table
	GetHowManyQuestions(ctx *context.Context) table.Table
	GetNeighbors(ctx *context.Context) table.Table
	GetTeams(ctx *context.Context) table.Table
	GetUserTeams(ctx *context.Context) table.Table
	GetUsers(ctx *context.Context) table.Table
	GetDashboard(ctx *gin.Context) (types.Panel, error)
}

func NewRenderer() Renderer {
	return &renderer{}
}
