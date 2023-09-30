package renderer

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
)

func (r *renderer) GetHowManyQuestions(ctx *context.Context) table.Table {
	questionsTable := table.NewDefaultTable(table.Config{
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

	info := questionsTable.GetInfo()

	info.SetSortDesc()

	info.AddField("ID", "id", db.Int).FieldFilterable()
	info.AddField("Text", "text", db.Varchar)
	info.AddField("Explanation", "explanation", db.Varchar)
	info.AddField("How Many", "how_many", db.Int)
	info.AddField("Valid", "valid", db.Bool)
	info.AddField("Done", "done", db.Bool)

	info.
		SetTable("how_many_qs").
		SetTitle("How Mnay Questions").
		SetDescription("The questions we ask in guess with us")

	return questionsTable
}
