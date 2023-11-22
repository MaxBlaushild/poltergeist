package migrate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/MaxBlaushild/poltergeist/admin/internal/config"
	"github.com/pkg/errors"
)

func Migrate(ctx context.Context, cfg *config.Config) error {
	errText := new(bytes.Buffer)
	cmd := exec.Command("psql", "sslmode=required", "-h", cfg.Public.DbHost, "-U", cfg.Public.DbUser, "-p", "5432", "-d", cfg.Public.DbName, "<", "./internal/migrate/admin.pgsql")
	fmt.Println("SHITASS!")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", cfg.Secret.DbPassword))
	cmd.Stderr = errText
	if err := cmd.Run(); err != nil {
		fmt.Println("THIS DONT WORK")
		fmt.Println(err)
		return errors.Wrap(err, errText.String())
	}

	fmt.Println("YARP")

	return nil
}
