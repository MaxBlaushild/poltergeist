package renderer

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
)

func (r *renderer) GetCrystals(ctx *context.Context) table.Table {
	crystalsTable := table.NewDefaultTable(table.Config{
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

	info := crystalsTable.GetInfo()

	info.SetSortDesc()

	info.AddField("ID", "id", db.Varchar).FieldFilterable()
	info.AddField("Clue", "clue", db.Varchar)
	info.AddField("Capture Challenge", "capture_challenge", db.Varchar)
	info.AddField("Attune Challenge", "attune_challenge", db.Varchar)
	info.AddField("Captured", "captured", db.Boolean)
	info.AddField("Attuned", "attuned", db.Boolean)
	info.AddField("Attuned", "attuned", db.Boolean)
	info.AddField("Lat", "lat", db.Varchar)
	info.AddField("Lng", "lng", db.Varchar)

	info.
		SetTable("crystals").
		SetTitle("Crystals").
		SetDescription("The points of interest on the map")

	return crystalsTable
}
