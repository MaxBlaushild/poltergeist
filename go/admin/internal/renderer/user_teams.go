package renderer

import (
	"fmt"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/MaxBlaushild/poltergeist/admin/internal/templates"
)

func (r *renderer) GetUserTeams(ctx *context.Context) table.Table {
	userteamsTable := table.NewDefaultTable(table.Config{
		Driver:     db.DriverPostgresql,
		CanAdd:     true,
		Editable:   true,
		Deletable:  true,
		Exportable: true,
		Connection: table.DefaultConnectionName,
		PrimaryKey: table.PrimaryKey{
			Type: db.Int,
			Name: table.DefaultPrimaryKeyName,
		},
	})

	info := userteamsTable.GetInfo()

	info.SetSortDesc()

	info.AddField("ID", "id", db.Varchar).FieldFilterable()
	info.AddField("User ID", "user_id", db.Varchar).FieldDisplay(func(model types.FieldModel) interface{} {
		userID := model.Row["user_id"].(string)
		return templates.Link(
			fmt.Sprint(userID),
			fmt.Sprintf("/info/users/detail?__goadmin_detail_pk=%s", userID),
		)
	}).FieldFilterable()
	info.AddField("Team ID", "team_id", db.Varchar).FieldDisplay(func(model types.FieldModel) interface{} {
		teamID := model.Row["team_id"].(string)
		return templates.Link(
			fmt.Sprint(teamID),
			fmt.Sprintf("/info/teams/detail?__goadmin_detail_pk=%s", teamID),
		)
	}).FieldFilterable()

	info.
		SetTable("user_teams").
		SetTitle("User Teams").
		SetDescription("Users in teams")

	return userteamsTable
}
