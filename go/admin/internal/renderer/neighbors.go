package renderer

import (
	"fmt"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/MaxBlaushild/poltergeist/admin/internal/templates"
)

func (r *renderer) GetNeighbors(ctx *context.Context) table.Table {
	neighborsTable := table.NewDefaultTable(table.Config{
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

	info := neighborsTable.GetInfo()

	info.SetSortDesc()

	info.AddField("ID", "id", db.Varchar).FieldFilterable()
	info.AddField("Crystal One ID", "crystal_one_id", db.Varchar).FieldDisplay(func(model types.FieldModel) interface{} {
		crystalID := model.Row["crystal_one_id"].(string)
		return templates.Link(
			fmt.Sprint(crystalID),
			fmt.Sprintf("/info/crystals/detail?__goadmin_detail_pk=%s", crystalID),
		)
	}).FieldFilterable()
	info.AddField("Crystal Two ID", "crystal_two_id", db.Varchar).FieldDisplay(func(model types.FieldModel) interface{} {
		crystalID := model.Row["crystal_two_id"].(string)
		return templates.Link(
			fmt.Sprint(crystalID),
			fmt.Sprintf("/info/crystals/detail?__goadmin_detail_pk=%s", crystalID),
		)
	}).FieldFilterable()

	info.
		SetTable("crystal_neighbors").
		SetTitle("Neighbors").
		SetDescription("A path linking two crystals")

	return neighborsTable
}
