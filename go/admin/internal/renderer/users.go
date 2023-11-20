package renderer

import (
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
)

func (r *renderer) GetUsers(ctx *context.Context) table.Table {
	usersTable := table.NewDefaultTable(table.Config{
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

	info := usersTable.GetInfo()

	info.SetSortDesc()

	info.AddField("ID", "id", db.Varchar).FieldFilterable()
	info.AddField("Name", "name", db.Varchar)
	info.AddField("Phone Number", "phone_number", db.Varchar)

	info.
		SetTable("users").
		SetTitle("Users").
		SetDescription("Yep. They're users.")

	return usersTable
}
