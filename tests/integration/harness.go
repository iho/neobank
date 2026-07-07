package integration

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	dbmigrate "github.com/iho/neobank/pkg/migrate"
	"github.com/iho/neobank/tests/integration/mockledger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	jwtSecret       = "integration-test-secret"
	defaultPassword = "secret123"
)

// Harness boots Postgres, Redis, a mock ledger, and the service binaries over HTTP.
type Harness struct {
	t *testing.T

	ctx    context.Context
	cancel context.CancelFunc

	DatabaseURL string
	RedisURL    string
	Ledger      *mockledger.Server
	LedgerAddr  string

	UserURL         string
	UserGRPCAddr    string
	PaymentURL      string
	CardURL         string
	NotificationURL string
	KYCURL          string
	CardProcURL     string

	SettlementAccountID string
	TreasuryAccountID   string

	pgContainer    testcontainers.Container
	redisContainer testcontainers.Container
	pool           *pgxpool.Pool
	procs          []*exec.Cmd
	binDir         string
}

func NewHarness(t *testing.T) *Harness {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Harness{t: t, ctx: ctx, cancel: cancel}
}

func (h *Harness) Start() {
	h.t.Helper()

	h.startPostgres()
	h.startRedis()
	h.runMigrations()
	h.startLedger()
	h.buildServiceBinaries()
	h.startServiceProcesses()
}

// Pool returns the shared Postgres pool (available after Start).
func (h *Harness) Pool() *pgxpool.Pool {
	return h.pool
}

func (h *Harness) Cleanup() {
	for _, proc := range h.procs {
		if proc.Process != nil {
			_ = proc.Process.Kill()
			_ = proc.Wait()
		}
	}
	if h.binDir != "" {
		_ = os.RemoveAll(h.binDir)
	}
	if h.pool != nil {
		h.pool.Close()
	}
	if h.Ledger != nil {
		h.Ledger.Stop()
	}
	if h.redisContainer != nil {
		_ = h.redisContainer.Terminate(h.ctx)
	}
	if h.pgContainer != nil {
		_ = h.pgContainer.Terminate(h.ctx)
	}
	h.cancel()
}

func (h *Harness) startPostgres() {
	pg, err := postgres.Run(h.ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("neobank"),
		postgres.WithUsername("neobank"),
		postgres.WithPassword("neobank"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		h.t.Fatalf("start postgres: %v", err)
	}
	h.pgContainer = pg
	url, err := pg.ConnectionString(h.ctx, "sslmode=disable")
	if err != nil {
		h.t.Fatalf("postgres connection string: %v", err)
	}
	h.DatabaseURL = url

	pool, err := pgxpool.New(h.ctx, url)
	if err != nil {
		h.t.Fatalf("pg pool: %v", err)
	}
	h.pool = pool
}

func (h *Harness) startRedis() {
	rc, err := redis.Run(h.ctx, "redis:7-alpine",
		testcontainers.WithWaitStrategy(wait.ForListeningPort("6379/tcp").WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		h.t.Fatalf("start redis: %v", err)
	}
	h.redisContainer = rc
	url, err := rc.ConnectionString(h.ctx)
	if err != nil {
		h.t.Fatalf("redis connection string: %v", err)
	}
	h.RedisURL = url
}

func (h *Harness) runMigrations() {
	root := repoRoot(h.t)
	initSQL, err := os.ReadFile(filepath.Join(root, "deployments/init-db.sql"))
	if err != nil {
		h.t.Fatalf("read init-db.sql: %v", err)
	}
	if _, err := h.pool.Exec(h.ctx, string(initSQL)); err != nil {
		h.t.Fatalf("apply init-db.sql: %v", err)
	}

	type serviceMigrations struct {
		dir    string
		schema string
	}
	services := []serviceMigrations{
		{filepath.Join(root, "services/user/migrations"), "user"},
		{filepath.Join(root, "services/payment/migrations"), "payment"},
		{filepath.Join(root, "services/card/migrations"), "card"},
		{filepath.Join(root, "services/notification/migrations"), "notification"},
		{filepath.Join(root, "services/simulators/kyc/migrations"), "kyc"},
		{filepath.Join(root, "services/simulators/cardproc/migrations"), "cardproc"},
	}
	for _, svc := range services {
		if err := dbmigrate.Up(h.DatabaseURL, svc.dir, dbmigrate.Config{SchemaName: svc.schema}); err != nil {
			h.t.Fatalf("migrate %s: %v", svc.dir, err)
		}
	}
}

func (h *Harness) startLedger() {
	h.Ledger = mockledger.New()
	addr, err := h.Ledger.Start()
	if err != nil {
		h.t.Fatalf("start mock ledger: %v", err)
	}
	h.LedgerAddr = addr

	treasury, err := h.Ledger.CreateTreasuryAccount()
	if err != nil {
		h.t.Fatalf("treasury account: %v", err)
	}
	h.TreasuryAccountID = treasury.Id
	if err := h.Ledger.CreditAccount(h.TreasuryAccountID, "1000000.00"); err != nil {
		h.t.Fatalf("credit treasury: %v", err)
	}

	settlement, err := h.Ledger.CreateSettlementAccount()
	if err != nil {
		h.t.Fatalf("settlement account: %v", err)
	}
	h.SettlementAccountID = settlement.Id
}

func (h *Harness) buildServiceBinaries() {
	root := repoRoot(h.t)
	dir, err := os.MkdirTemp("", "neobank-integration-bins-*")
	if err != nil {
		h.t.Fatalf("temp bin dir: %v", err)
	}
	h.binDir = dir

	services := map[string]string{
		"notification": "./services/notification/cmd/server",
		"user":         "./services/user/cmd/server",
		"payment":      "./services/payment/cmd/server",
		"card":         "./services/card/cmd/server",
		"kyc":          "./services/simulators/kyc/cmd/server",
		"cardproc":     "./services/simulators/cardproc/cmd/server",
	}
	for name, pkg := range services {
		out := filepath.Join(dir, name)
		cmd := exec.Command("go", "build", "-o", out, pkg)
		cmd.Dir = root
		if outBytes, err := cmd.CombinedOutput(); err != nil {
			h.t.Fatalf("build %s: %v\n%s", name, err, string(outBytes))
		}
	}
}

func (h *Harness) startServiceProcesses() {
	// Ports are reserved by binding a listener and holding it open — not by
	// closing it and remembering the number — because some of these ports
	// are referenced (as URLs baked into other services' env vars) tens of
	// seconds before the owning process actually gets around to binding
	// them. Closing early frees the port back to the OS for that whole
	// window, and on a busy CI runner something else can grab it first,
	// causing the real owner to fail to bind. Each listener is only closed
	// immediately before the process that owns it is started.
	notifRes := reservePort(h.t)
	notifGrpcRes := reservePort(h.t)
	userRes := reservePort(h.t)
	userGrpcRes := reservePort(h.t)
	paymentRes := reservePort(h.t)
	paymentGrpcRes := reservePort(h.t)
	cardRes := reservePort(h.t)
	cardGrpcRes := reservePort(h.t)
	kycRes := reservePort(h.t)
	cardProcRes := reservePort(h.t)

	notifPort, notifGrpcPort := notifRes.port, notifGrpcRes.port
	userPort, userGrpcPort := userRes.port, userGrpcRes.port
	paymentPort, paymentGrpcPort := paymentRes.port, paymentGrpcRes.port
	cardPort, cardGrpcPort := cardRes.port, cardGrpcRes.port
	kycPort := kycRes.port
	cardProcPort := cardProcRes.port

	h.NotificationURL = fmt.Sprintf("http://127.0.0.1:%s", notifPort)
	h.UserURL = fmt.Sprintf("http://127.0.0.1:%s", userPort)
	h.UserGRPCAddr = fmt.Sprintf("127.0.0.1:%s", userGrpcPort)
	h.PaymentURL = fmt.Sprintf("http://127.0.0.1:%s", paymentPort)
	h.CardURL = fmt.Sprintf("http://127.0.0.1:%s", cardPort)
	h.KYCURL = fmt.Sprintf("http://127.0.0.1:%s", kycPort)
	h.CardProcURL = fmt.Sprintf("http://127.0.0.1:%s", cardProcPort)

	notificationIngest := h.NotificationURL + "/api/v1/internal/events"

	// kyc must be up before user: user's SubmitKYC calls out to it
	// synchronously, and it needs user's webhook URL to deliver verdicts
	// back to (see submit_kyc.go / kyc_handlers.go).
	kycRes.release()
	h.startProcess("kyc", map[string]string{
		"DATABASE_URL":            h.DatabaseURL,
		"HTTP_PORT":               kycPort,
		"USER_SERVICE_EVENTS_URL": h.UserURL + "/webhooks/kyc/events",
	})
	waitForHTTP200(h.t, h.KYCURL+"/health")

	userRes.release()
	userGrpcRes.release()
	h.startProcess("user", map[string]string{
		"DATABASE_URL":                     h.DatabaseURL,
		"REDIS_URL":                        h.RedisURL,
		"LEDGER_GRPC_ADDR":                 h.LedgerAddr,
		"HTTP_PORT":                        userPort,
		"GRPC_PORT":                        userGrpcPort,
		"JWT_SECRET":                       jwtSecret,
		"NOTIFICATION_SERVICE_URL":         notificationIngest,
		"DEPOSIT_SOURCE_LEDGER_ACCOUNT_ID": h.TreasuryAccountID,
		"KYC_VENDOR_SERVICE_URL":           h.KYCURL,
	})
	waitForHTTP200(h.t, h.UserURL+"/health")

	notifRes.release()
	notifGrpcRes.release()
	h.startProcess("notification", map[string]string{
		"DATABASE_URL":   h.DatabaseURL,
		"HTTP_PORT":      notifPort,
		"GRPC_PORT":      notifGrpcPort,
		"USER_GRPC_ADDR": h.UserGRPCAddr,
	})
	paymentRes.release()
	paymentGrpcRes.release()
	h.startProcess("payment", map[string]string{
		"DATABASE_URL":             h.DatabaseURL,
		"REDIS_URL":                h.RedisURL,
		"LEDGER_GRPC_ADDR":         h.LedgerAddr,
		"HTTP_PORT":                paymentPort,
		"GRPC_PORT":                paymentGrpcPort,
		"USER_SERVICE_URL":         h.UserURL,
		"USER_GRPC_ADDR":           h.UserGRPCAddr,
		"NOTIFICATION_SERVICE_URL": notificationIngest,
	})
	// cardproc must be up before card: card's IssueCardUseCase calls out to
	// it synchronously for every virtual card, with no direct-call fallback
	// (unlike authorize/capture, which still hit card's own endpoints
	// directly — see docs/vendor-simulators-plan.md Phase 2a "kept as-is").
	cardProcRes.release()
	h.startProcess("cardproc", map[string]string{
		"DATABASE_URL":               h.DatabaseURL,
		"HTTP_PORT":                  cardProcPort,
		"CARD_SERVICE_AUTHORIZE_URL": h.CardURL + "/webhooks/cardproc/authorize",
		"CARD_SERVICE_EVENTS_URL":    h.CardURL + "/webhooks/cardproc/events",
	})
	waitForHTTP200(h.t, h.CardProcURL+"/health")

	cardRes.release()
	cardGrpcRes.release()
	h.startProcess("card", map[string]string{
		"DATABASE_URL":                 h.DatabaseURL,
		"REDIS_URL":                    h.RedisURL,
		"LEDGER_GRPC_ADDR":             h.LedgerAddr,
		"HTTP_PORT":                    cardPort,
		"GRPC_PORT":                    cardGrpcPort,
		"USER_SERVICE_URL":             h.UserURL,
		"USER_GRPC_ADDR":               h.UserGRPCAddr,
		"NOTIFICATION_SERVICE_URL":     notificationIngest,
		"SETTLEMENT_LEDGER_ACCOUNT_ID": h.SettlementAccountID,
		"CARDPROC_SERVICE_URL":         h.CardProcURL,
	})

	waitForHTTP200(h.t, h.NotificationURL+"/health")
	waitForHTTP200(h.t, h.PaymentURL+"/health")
	waitForHTTP200(h.t, h.CardURL+"/health")
}

func (h *Harness) startProcess(name string, env map[string]string) {
	bin := filepath.Join(h.binDir, name)
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), flattenEnv(env)...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		h.t.Fatalf("start %s: %v", name, err)
	}
	h.procs = append(h.procs, cmd)
}

func flattenEnv(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for k, v := range env {
		out = append(out, k+"="+v)
	}
	return out
}

// reservedPort holds a listener on a free port so nothing else on the
// machine can claim it before release() closes it and hands the port
// number to the process meant to bind it.
type reservedPort struct {
	port string
	ln   net.Listener
}

func (r reservedPort) release() {
	_ = r.ln.Close()
}

func reservePort(t *testing.T) reservedPort {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	return reservedPort{port: fmt.Sprintf("%d", port), ln: ln}
}

func waitForHTTP200(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("service not healthy: %s", url)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root (go.work)")
		}
		dir = parent
	}
}

func waitUntil(t *testing.T, timeout time.Duration, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("condition not met before timeout")
}
