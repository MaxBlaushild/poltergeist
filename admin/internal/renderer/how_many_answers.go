package renderer

import (
	"fmt"

	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/db"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/MaxBlaushild/poltergeist/admin/internal/templates"
)

// type HowManyAnswer struct {
// 	gorm.Model
// 	HowManyQuestion   HowManyQuestion `json:"howManyQuestion"`
// 	HowManyQuestionID uint            `json:"howManyQuestionId"`
// 	Answer            int             `json:"answer"`
// 	Guess             int             `json:"guess"`
// 	OffBy             int             `json:"offBy"`
// 	Correctness       float64         `json:"correctness"`
// 	User              User            `json:"user"`
// 	UserID            *uint           `json:"userId"`
// 	EphemeralUserID   *string         `json:"ephemeralUserId"`
// }

func (r *renderer) GetHowManyAnswers(ctx *context.Context) table.Table {
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
	info.AddField("Ephemeral User ID", "ephemeral_user_id", db.Varchar)
	info.AddField("Answer", "answer", db.Int)
	info.AddField("Guess", "guess", db.Int)
	info.AddField("Off By", "off_by", db.Int)
	info.AddField("Correctness", "correctness", db.Float)

	info.AddField("User ID", "user_id", db.Int).FieldDisplay(func(model types.FieldModel) interface{} {
		userID := model.Row["user_id"].(int64)
		return templates.Link(
			fmt.Sprint(userID),
			fmt.Sprintf("/info/users/detail?__goadmin_detail_pk=%d", userID),
		)
	}).FieldFilterable()
	info.AddField("Question ID", "how_many_question_id", db.Int).FieldDisplay(func(model types.FieldModel) interface{} {
		questionID := model.Row["hoq_many_question_id"].(int64)
		return templates.Link(
			fmt.Sprint(questionID),
			fmt.Sprintf("/info/how-many-questions/detail?__goadmin_detail_pk=%d", questionID),
		)
	}).FieldFilterable()

	info.
		SetTable("how_many_as").
		SetTitle("How Mnay Answers").
		SetDescription("What they answer with")

	return questionsTable
}
