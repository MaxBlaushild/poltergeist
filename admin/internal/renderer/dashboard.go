package renderer

import (
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/gin-gonic/gin"
)

func (r *renderer) GetDashboard(ctx *gin.Context) (types.Panel, error) {
	return types.Panel{
		Title:       "Poltergeist",
		Description: "Data stuff, quick and dirty",
		Content:     `<div>You're not building fast enough if your admin dashboard looks pretty!</div>`,
	}, nil
}
