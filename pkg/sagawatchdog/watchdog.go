//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sagawatchdog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	AlertOpen          = "open"
	AlertInvestigating = "investigating"
	AlertResolved      = "resolved"
)

// AllSchemas scanned by the unified watchdog CLI.
var AllSchemas = []string{"user", "payment", "card"}

var qualifiedSchemas = map[string]string{
	"user":    `"user"`,
	"payment": "payment",
	"card":    "card",
}

// StuckInstance is a saga that has not progressed within the stale window.
type StuckInstance struct {
	UpdatedAt      time.Time
	SagaType       string
	IdempotencyKey string
	Status         string
	CompletedSteps []byte
	Context        []byte
	ID             uuid.UUID
}

// Alert is a persisted stuck-saga record for operator follow-up.
type Alert struct {
	StuckSince     time.Time
	LastSeenAt     time.Time
	Schema         string
	SagaType       string
	IdempotencyKey string
	InstanceStatus string
	AlertStatus    string
	ResolvedBy     string
	Notes          string
	ID             uuid.UUID
	SagaInstanceID uuid.UUID
}

// ScanResult summarizes one watchdog pass for a schema.
type ScanResult struct {
	Schema       string
	StuckFound   int
	AlertsOpen   int
	AutoResolved int
}

// Scanner checks saga_instances and maintains saga_alerts.
type Scanner struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Scanner {
	return &Scanner{pool: pool}
}

// Scan detects stuck sagas, upserts alerts, and auto-resolves alerts whose
// saga instance reached a terminal state.
func (s *Scanner) Scan(ctx context.Context, schema string, staleAfter time.Duration) (ScanResult, error) {
	qual, err := qualifySchema(schema)
	if err != nil {
		return ScanResult{}, err
	}

	if staleAfter <= 0 {
		staleAfter = 15 * time.Minute
	}

	cutoff := time.Now().UTC().Add(-staleAfter)
	now := time.Now().UTC()

	stuck, err := s.listStuck(ctx, qual, cutoff)
	if err != nil {
		return ScanResult{}, err
	}

	for _, inst := range stuck {
		err := s.upsertAlert(ctx, qual, inst, now)
		if err != nil {
			return ScanResult{}, err
		}
	}

	autoResolved, err := s.autoResolve(ctx, qual, now)
	if err != nil {
		return ScanResult{}, err
	}

	openCount, err := s.countOpenAlerts(ctx, qual)
	if err != nil {
		return ScanResult{}, err
	}

	return ScanResult{
		Schema:       schema,
		StuckFound:   len(stuck),
		AlertsOpen:   openCount,
		AutoResolved: autoResolved,
	}, nil
}

// ListOpenAlerts returns active alerts for a schema.
func (s *Scanner) ListOpenAlerts(ctx context.Context, schema string, limit int) ([]Alert, error) {
	qual, err := qualifySchema(schema)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}

	query := fmt.Sprintf(`
SELECT id, saga_instance_id, saga_type, idempotency_key, instance_status, alert_status,
       stuck_since, last_seen_at, COALESCE(resolved_by, ''), COALESCE(notes, '')
FROM %s.saga_alerts
WHERE alert_status IN ('open', 'investigating')
ORDER BY last_seen_at DESC
LIMIT $1`, qual)

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var a Alert

		a.Schema = schema
		err := rows.Scan(
			&a.ID, &a.SagaInstanceID, &a.SagaType, &a.IdempotencyKey, &a.InstanceStatus, &a.AlertStatus,
			&a.StuckSince, &a.LastSeenAt, &a.ResolvedBy, &a.Notes,
		)
		if err != nil {
			return nil, err
		}

		alerts = append(alerts, a)
	}

	return alerts, rows.Err()
}

// ResolveAlert marks an alert investigated or resolved.
func (s *Scanner) ResolveAlert(ctx context.Context, schema string, alertID uuid.UUID, status, resolvedBy, notes string) (bool, error) {
	qual, err := qualifySchema(schema)
	if err != nil {
		return false, err
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status != AlertInvestigating && status != AlertResolved {
		return false, errors.New("status must be investigating or resolved")
	}

	now := time.Now().UTC()

	query := fmt.Sprintf(`
UPDATE %s.saga_alerts
SET alert_status = $2,
    resolved_by = $3,
    notes = COALESCE(NULLIF($4, ''), notes),
    resolved_at = CASE WHEN $2 = 'resolved' THEN $5 ELSE resolved_at END,
    updated_at = $5
WHERE id = $1 AND alert_status != 'resolved'`, qual)

	tag, err := s.pool.Exec(ctx, query, alertID, status, resolvedBy, notes, now)
	if err != nil {
		return false, err
	}

	return tag.RowsAffected() > 0, nil
}

func (s *Scanner) listStuck(ctx context.Context, qual string, cutoff time.Time) ([]StuckInstance, error) {
	query := fmt.Sprintf(`
SELECT id, saga_type, idempotency_key, status, completed_steps, context, updated_at
FROM %s.saga_instances
WHERE status IN ('running', 'compensating')
  AND updated_at < $1`, qual)

	rows, err := s.pool.Query(ctx, query, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stuck []StuckInstance
	for rows.Next() {
		var inst StuckInstance
		err := rows.Scan(
			&inst.ID, &inst.SagaType, &inst.IdempotencyKey, &inst.Status,
			&inst.CompletedSteps, &inst.Context, &inst.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		stuck = append(stuck, inst)
	}

	return stuck, rows.Err()
}

func (s *Scanner) upsertAlert(ctx context.Context, qual string, inst StuckInstance, now time.Time) error {
	query := fmt.Sprintf(`
INSERT INTO %s.saga_alerts (
    id, saga_instance_id, saga_type, idempotency_key, instance_status,
    alert_status, stuck_since, last_seen_at, completed_steps, context, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, 'open', $6, $7, $8, $9, $7, $7)
ON CONFLICT (saga_instance_id) WHERE (alert_status IN ('open', 'investigating'))
DO UPDATE SET
    instance_status = EXCLUDED.instance_status,
    last_seen_at = EXCLUDED.last_seen_at,
    completed_steps = EXCLUDED.completed_steps,
    context = EXCLUDED.context,
    updated_at = EXCLUDED.updated_at`, qual)

	_, err := s.pool.Exec(ctx, query,
		uuid.New(),
		inst.ID,
		inst.SagaType,
		inst.IdempotencyKey,
		inst.Status,
		inst.UpdatedAt,
		now,
		inst.CompletedSteps,
		inst.Context,
	)

	return err
}

func (s *Scanner) autoResolve(ctx context.Context, qual string, now time.Time) (int, error) {
	query := fmt.Sprintf(`
UPDATE %s.saga_alerts AS a
SET alert_status = 'resolved',
    resolved_by = 'saga-watchdog',
    notes = COALESCE(a.notes, '') || ' auto-resolved: saga reached terminal state',
    resolved_at = $1,
    updated_at = $1
FROM %s.saga_instances AS s
WHERE a.saga_instance_id = s.id
  AND a.alert_status IN ('open', 'investigating')
  AND s.status IN ('completed', 'failed')`, qual, qual)

	tag, err := s.pool.Exec(ctx, query, now)
	if err != nil {
		return 0, err
	}

	return int(tag.RowsAffected()), nil
}

func (s *Scanner) countOpenAlerts(ctx context.Context, qual string) (int, error) {
	query := fmt.Sprintf(`
SELECT COUNT(*) FROM %s.saga_alerts WHERE alert_status IN ('open', 'investigating')`, qual)

	var count int

	err := s.pool.QueryRow(ctx, query).Scan(&count)

	return count, err
}

func qualifySchema(schema string) (string, error) {
	qual, ok := qualifiedSchemas[schema]
	if !ok {
		return "", fmt.Errorf("unsupported schema %q", schema)
	}

	return qual, nil
}

// Ping verifies database connectivity.
func (s *Scanner) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}
