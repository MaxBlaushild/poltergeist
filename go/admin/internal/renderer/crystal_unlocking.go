package renderer

import (
	"fmt"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/MaxBlaushild/poltergeist/admin/internal/templates"
)

func (r *renderer) GetCrystalUnlockings(ctx *context.Context) table.Table {
	crystalUnlockingsTable := table.NewDefaultTable(table.Config{
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

	info := crystalUnlockingsTable.GetInfo()

	info.SetSortDesc()

	info.AddField("ID", "id", db.Varchar).FieldFilterable()
	info.AddField("Crystal ID", "crystal_id", db.Varchar).FieldDisplay(func(model types.FieldModel) interface{} {
		crystalID := model.Row["crystal_id"].(string)
		return templates.Link(
			fmt.Sprint(crystalID),
			fmt.Sprintf("/info/crystals/detail?__goadmin_detail_pk=%s", crystalID),
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
		SetTable("crystal_unlockings").
		SetTitle("Crystal Unlockings").
		SetDescription("A team unlocks the challenges of a crystal")

	return crystalUnlockingsTable
}
