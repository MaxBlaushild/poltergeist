package renderer

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
)

func (r *renderer) GetTeams(ctx *context.Context) table.Table {
	teamsTable := table.NewDefaultTable(table.Config{
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

	info := teamsTable.GetInfo()

	info.SetSortDesc()

	info.AddField("ID", "id", db.Int).FieldFilterable()
	info.AddField("Name", "name", db.Varchar)
	info.
		SetTable("teams").
		SetTitle("Teams").
		SetDescription("The teams competing")

	return teamsTable
}
