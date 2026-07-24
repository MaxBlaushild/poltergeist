package processors

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/MaxBlaushild/job-runner/internal/config"
	"github.com/MaxBlaushild/poltergeist/pkg/db"
	"github.com/MaxBlaushild/poltergeist/pkg/jobs"
	"github.com/MaxBlaushild/poltergeist/pkg/models"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/procexec"
	"github.com/MaxBlaushild/poltergeist/pkg/reef/slice"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// testDB connects to a real Postgres carrying the reef migrations (see
// go/reef-site/INVENTORY.md for how to stand one up locally) and skips the
// test if it can't. This is deliberately a real integration test, not a
// mock: R-10's acceptance criterion is about actual subprocess invocation
// counts, which a mocked DB wouldn't exercise honestly.
func testDB(t *testing.T) db.DbClient {
	t.Helper()
	if _, err := exec.LookPath("openscad"); err != nil {
		t.Skip("openscad not installed, skipping reef processor integration test")
	}

	host := envOr("REEF_TEST_DB_HOST", "localhost")
	port := envOr("REEF_TEST_DB_PORT", "5432")
	user := envOr("REEF_TEST_DB_USER", "db_user")
	name := envOr("REEF_TEST_DB_NAME", "poltergeist")
	password := envOr("REEF_TEST_DB_PASSWORD", "x") // must be non-empty, see go/pkg/db DSN builder

	dbClient, err := db.NewClient(db.ClientConfig{Host: host, Port: port, User: user, Name: name, Password: password})
	if err != nil {
		t.Skipf("no reachable test database (%s:%s/%s), skipping: %v", host, port, name, err)
	}
	return dbClient
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// stubSlice always rejects on weight so the test never needs a real
// slicer binary or an S3 upload — only the generation/caching path (which
// is what this test is actually verifying) touches real subprocesses.
func stubSliceAlwaysOverweight(ctx context.Context, cfg slice.Config, stlPath string) (*slice.Result, error) {
	return &slice.Result{
		WeightG:         999999,
		PrintTimeS:      60,
		SupportRequired: false,
		SlicerVersion:   "stub-test-slicer",
	}, nil
}

// R-10 acceptance criterion: "Configuring identical parameters twice
// performs exactly one generation and one slice; the second is served from
// cache. Verified by a test asserting subprocess invocation count."
//
// The processor must resolve the OpenSCAD version *before* it can even
// compute geometry_hash (R-3.3's hash formula includes openscad_version),
// so a `--version` check runs on every call regardless of cache state —
// that's one procexec invocation, not a "generation". The actual geometry
// render (a second, distinct procexec invocation) is what a cache hit
// skips. This test asserts on that breakdown precisely rather than a single
// opaque "fewer calls" check.
func TestGenerateReefFullProcessor_IdenticalParamsCacheAfterFirstGeneration(t *testing.T) {
	dbClient := testDB(t)
	ctx := context.Background()

	product, err := dbClient.ReefProduct().FindBySlug(ctx, "magnetic-frag-rack")
	if err != nil {
		t.Fatalf("load seeded product: %v", err)
	}

	// Perturb width slightly on every run so repeated test runs don't hit a
	// slice_result left over from a previous run and report a false cache
	// hit on the "first" invocation.
	rand.Seed(time.Now().UnixNano())
	widthMm := 140.0 + float64(rand.Intn(20))

	paramsMap := map[string]interface{}{
		"glassThicknessMm":   10.0,
		"tierCount":          2.0,
		"widthMm":            widthMm,
		"plugHoleDiameterMm": 20.0,
		"holesPerTier":       5.0,
		"color":              "black",
	}
	paramsJSON, err := json.Marshal(paramsMap)
	if err != nil {
		t.Fatal(err)
	}

	cfg1 := mustCreateConfiguration(t, dbClient, product.ID, paramsJSON)
	job1 := mustCreateJob(t, dbClient, cfg1.ID)
	cfg2 := mustCreateConfiguration(t, dbClient, product.ID, paramsJSON)
	job2 := mustCreateJob(t, dbClient, cfg2.ID)

	processor := &GenerateReefFullProcessor{
		dbClient:  dbClient,
		awsClient: nil, // never called: stub slicer always rejects on weight
		cfg: config.PublicConfig{
			ReefOpenSCADBin:          "openscad",
			ReefSubprocessTimeoutSec: 60,
			ReefSubprocessMemoryMB:   1024,
			ReefMaxBboxMm:            210,
			ReefMinWallMm:            2.0,
			ReefMaxPrintTimeS:        4 * 60 * 60,
			ReefMaxWeightG:           250,
			ReefMinDrainPathMm:       4,
		},
		slice: stubSliceAlwaysOverweight,
	}

	before := procexec.Invocations()
	runTask(t, processor, job1.ID, cfg1.ID)
	afterFirst := procexec.Invocations()
	firstDelta := afterFirst - before
	if firstDelta != 2 {
		t.Fatalf("first (uncached) run made %d procexec calls, want 2 (version check + one generation)", firstDelta)
	}

	runTask(t, processor, job2.ID, cfg2.ID)
	afterSecond := procexec.Invocations()
	secondDelta := afterSecond - afterFirst
	if secondDelta != 1 {
		t.Fatalf("second (identical params) run made %d procexec calls, want 1 (version check only — generation must be served from cache)", secondDelta)
	}

	reloaded1, err := dbClient.ReefConfiguration().FindByID(ctx, cfg1.ID)
	if err != nil {
		t.Fatal(err)
	}
	reloaded2, err := dbClient.ReefConfiguration().FindByID(ctx, cfg2.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded1.GeometryHash == nil || reloaded2.GeometryHash == nil || *reloaded1.GeometryHash != *reloaded2.GeometryHash {
		t.Fatalf("expected both configurations to resolve to the same geometry_hash, got %v and %v", reloaded1.GeometryHash, reloaded2.GeometryHash)
	}
	if reloaded1.Status != models.ReefConfigurationStatusRejected || reloaded2.Status != models.ReefConfigurationStatusRejected {
		t.Fatalf("expected both configurations rejected (stub slicer always overweight), got %s and %s", reloaded1.Status, reloaded2.Status)
	}
}

func runTask(t *testing.T, processor *GenerateReefFullProcessor, jobID, configurationID uuid.UUID) {
	t.Helper()
	payload, err := json.Marshal(jobs.GenerateReefFullTaskPayload{ConfigurationID: configurationID, JobID: jobID})
	if err != nil {
		t.Fatal(err)
	}
	task := asynq.NewTask(jobs.GenerateReefFullTaskType, payload)
	if err := processor.ProcessTask(context.Background(), task); err != nil {
		t.Fatalf("ProcessTask: %v", err)
	}
}

func mustCreateConfiguration(t *testing.T, dbClient db.DbClient, productID uuid.UUID, params []byte) *models.ReefConfiguration {
	t.Helper()
	cfg, err := dbClient.ReefConfiguration().Create(context.Background(), &models.ReefConfiguration{
		ProductID: productID,
		Params:    params,
		Status:    models.ReefConfigurationStatusPending,
	})
	if err != nil {
		t.Fatalf("create configuration: %v", err)
	}
	return cfg
}

func mustCreateJob(t *testing.T, dbClient db.DbClient, configurationID uuid.UUID) *models.ReefGenerationJob {
	t.Helper()
	job, err := dbClient.ReefGenerationJob().Create(context.Background(), &models.ReefGenerationJob{
		ConfigurationID: configurationID,
		Kind:            models.ReefGenerationJobKindFull,
		Status:          models.ReefGenerationJobStatusQueued,
	})
	if err != nil {
		t.Fatalf("create generation job: %v", err)
	}
	return job
}
