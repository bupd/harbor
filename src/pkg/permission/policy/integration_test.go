// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package policy_test

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/permission/policy"
)

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func dsn(t *testing.T) string {
	t.Helper()
	host := os.Getenv("POSTGRESQL_HOST")
	if host == "" {
		t.Skip("POSTGRESQL_HOST is not set")
	}
	port, err := strconv.Atoi(envOr("POSTGRESQL_PORT", "5432"))
	require.NoError(t, err)

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(envOr("POSTGRESQL_USR", "postgres"), os.Getenv("POSTGRESQL_PWD")),
		Host:     net.JoinHostPort(host, strconv.Itoa(port)),
		Path:     envOr("POSTGRESQL_DATABASE", "registry"),
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

func open(t *testing.T, conn string, origin string) *sql.DB {
	t.Helper()
	cfg, err := pgx.ParseConfig(conn)
	require.NoError(t, err)
	if origin != "" {
		cfg.RuntimeParams["options"] = "-c harbor.origin=" + origin
	}
	db := stdlib.OpenDB(*cfg)
	db.SetMaxOpenConns(4)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// replica is a core that listens. deaf is one that does not, standing for the
// replica whose connection died or which was paused past the message.
func replica(t *testing.T, conn string, origin string) *policy.Store {
	t.Helper()
	s := deaf(t, conn, origin)
	require.NoError(t, s.Watch(conn))
	return s
}

func deaf(t *testing.T, conn string, origin string) *policy.Store {
	t.Helper()
	s, err := policy.New(context.Background(), open(t, conn, origin), origin)
	require.NoError(t, err)
	t.Cleanup(s.Close)
	return s
}

func waitFor(t *testing.T, limit time.Duration, done func() bool) {
	t.Helper()
	deadline := time.Now().Add(limit)
	for time.Now().Before(deadline) {
		if done() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal(fmt.Sprintf("condition not met within %v", limit))
}

// The roles Harbor ships carried their grants in the binary, so the first boot
// after this change is the one that writes them down. Every boot after that has
// nothing to do.
func TestIntegrationSeedingIsIdempotent(t *testing.T) {
	conn := dsn(t)
	store := deaf(t, conn, "test-seed")

	require.NoError(t, store.EnsureSeeded(context.Background()))
	first := store.Generation()

	granted, err := store.Enforce(policy.Subject(1), policy.Object("repository"), "push")
	require.NoError(t, err)
	assert.True(t, granted, "projectAdmin pushes, from the database now")

	granted, err = store.Enforce(policy.Subject(3), policy.Object("repository"), "push")
	require.NoError(t, err)
	assert.False(t, granted, "guest still does not")

	// A second boot writes nothing, so it does not reload either.
	require.NoError(t, store.EnsureSeeded(context.Background()))
	assert.Equal(t, first, store.Generation(), "nothing to seed, nothing to reload")
}

// One replica writes, another finds out. No restart, no Redis, nothing polling.
func TestIntegrationAWriteReachesAnotherReplica(t *testing.T) {
	conn := dsn(t)
	writer := open(t, conn, "test-writer")
	reader := replica(t, conn, "test-reader")
	require.NoError(t, reader.EnsureSeeded(context.Background()))

	before := reader.Generation()
	start := time.Now()

	// permission_policy is shared, so this takes the row that is already there
	// and links the role to it, which is what the API does too.
	var policyID int64
	require.NoError(t, writer.QueryRow(
		`INSERT INTO permission_policy (scope, resource, action, effect)
		 VALUES ('/project/*', 'repository', 'push', 'allow')
		 ON CONFLICT ON CONSTRAINT unique_rbac_policy
		 DO UPDATE SET scope = EXCLUDED.scope
		 RETURNING id`).Scan(&policyID))
	_, err := writer.Exec(
		`INSERT INTO role_permission (role_type, role_id, permission_policy_id) VALUES ($1, 3, $2)`,
		policy.RoleType, policyID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = writer.Exec(
			`DELETE FROM role_permission WHERE role_id = 3 AND permission_policy_id = $1`, policyID)
	})

	waitFor(t, 5*time.Second, func() bool { return reader.Generation() > before })
	t.Logf("the second replica reloaded %v after the write", time.Since(start).Round(time.Millisecond))

	granted, err := reader.Enforce(policy.Subject(3), policy.Object("repository"), "push")
	require.NoError(t, err)
	assert.True(t, granted, "the second replica answers from the new policy")
}

// The case that decides whether holding the policy in memory is safe: a replica
// that is never told anything.
func TestIntegrationAReplicaThatWasNeverToldStillConverges(t *testing.T) {
	conn := dsn(t)
	writer := open(t, conn, "test-converge-writer")
	quiet := deaf(t, conn, "test-converge-deaf")
	require.NoError(t, quiet.EnsureSeeded(context.Background()))

	before := quiet.AppliedVersion()
	require.NotZero(t, before, "a replica knows which policy version it loaded")

	// permission_policy is shared, so this takes the row that is already there
	// and links the role to it, which is what the API does too.
	var policyID int64
	require.NoError(t, writer.QueryRow(
		`INSERT INTO permission_policy (scope, resource, action, effect)
		 VALUES ('/project/*', 'repository', 'delete', 'allow')
		 ON CONFLICT ON CONSTRAINT unique_rbac_policy
		 DO UPDATE SET scope = EXCLUDED.scope
		 RETURNING id`).Scan(&policyID))
	_, err := writer.Exec(
		`INSERT INTO role_permission (role_type, role_id, permission_policy_id) VALUES ($1, 3, $2)`,
		policy.RoleType, policyID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = writer.Exec(
			`DELETE FROM role_permission WHERE role_id = 3 AND permission_policy_id = $1`, policyID)
	})

	granted, err := quiet.Enforce(policy.Subject(3), policy.Object("repository"), "delete")
	require.NoError(t, err)
	assert.False(t, granted, "it has not heard, so it is serving the old policy")

	changed, err := quiet.Resync(context.Background())
	require.NoError(t, err)
	assert.True(t, changed, "the version in the database is ahead, so it reloads")
	assert.Greater(t, quiet.AppliedVersion(), before)

	granted, err = quiet.Enforce(policy.Subject(3), policy.Object("repository"), "delete")
	require.NoError(t, err)
	assert.True(t, granted, "it converged without ever being told")

	changed, err = quiet.Resync(context.Background())
	require.NoError(t, err)
	assert.False(t, changed, "nothing changed, so nothing is reloaded")
}
