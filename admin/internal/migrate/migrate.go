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
	cmd := exec.Command("psql", "-h", cfg.Public.DbHost, "-U", cfg.Public.DbUser, "-f", "./internal/migrate/admin.pgsql", cfg.Public.DbName)

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%s", cfg.Secret.DbPassword))
	cmd.Stderr = errText

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, errText.String())
	}

	return nil
}
