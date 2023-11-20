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

	info.AddField("ID", "id", db.Varchar).FieldFilterable()
	info.AddField("Text", "text", db.Varchar)
	info.AddField("Explanation", "explanation", db.Varchar)
	info.AddField("How Many", "how_many", db.Int)
	info.AddField("Valid", "valid", db.Bool)
	info.AddField("Done", "done", db.Bool)

	info.
		SetTable("how_many_questions").
		SetTitle("How Many Questions").
		SetDescription("The questions we ask in guess with us")

	questionsTable.GetNewForm().SetTable("how_many_questions")
	questionsTable.GetActualNewForm().SetTable("how_many_questions")

	return questionsTable
}
