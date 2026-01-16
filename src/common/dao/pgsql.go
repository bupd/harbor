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

package dao

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // import pgx v5 driver for migrator
	_ "github.com/golang-migrate/migrate/v4/source/file"     // import local file driver for migrator
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
)

const defaultMigrationPath = "migrations/postgresql/"

type pgsql struct {
	host            string
	port            string
	usr             string
	pwd             string
	database        string
	sslmode         string
	maxIdleConns    int
	maxOpenConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	pool            *pgxpool.Pool
}

// Name returns the name of PostgreSQL
func (p *pgsql) Name() string {
	return "PostgreSQL"
}

// String ...
func (p *pgsql) String() string {
	return fmt.Sprintf("type-%s host-%s port-%s database-%s sslmode-%q",
		p.Name(), p.host, p.port, p.database, p.sslmode)
}

// NewPGSQL returns an instance of postgres
func NewPGSQL(host string, port string, usr string, pwd string, database string, sslmode string, maxIdleConns int, maxOpenConns int, connMaxLifetime time.Duration, connMaxIdleTime time.Duration) Database {
	if len(sslmode) == 0 {
		sslmode = "disable"
	}
	return &pgsql{
		host:            host,
		port:            port,
		usr:             usr,
		pwd:             pwd,
		database:        database,
		sslmode:         sslmode,
		maxIdleConns:    maxIdleConns,
		maxOpenConns:    maxOpenConns,
		connMaxLifetime: connMaxLifetime,
		connMaxIdleTime: connMaxIdleTime,
	}
}

// Register registers pgSQL to orm with the info wrapped by the instance.
// Uses pgxpool for connection pooling with stdlib bridge for Beego ORM compatibility.
func (p *pgsql) Register(alias ...string) error {
	if err := utils.TestTCPConn(net.JoinHostPort(p.host, p.port), 60, 2); err != nil {
		return err
	}

	// Build pgxpool connection string
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=UTC",
		p.host, p.port, p.usr, p.pwd, p.database, p.sslmode)

	// Create pgxpool with configuration
	ctx := context.Background()
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("failed to parse pgxpool config: %w", err)
	}

	// Map configuration - only override pgxpool defaults if explicitly configured
	// pgxpool defaults: MaxConns = max(4, runtime.NumCPU()), MinConns = 0
	// database/sql used 0 to mean "unlimited", so we preserve pgxpool defaults for 0
	if p.maxOpenConns > 0 {
		config.MaxConns = int32(p.maxOpenConns)
	}
	if p.maxIdleConns > 0 {
		config.MinConns = int32(p.maxIdleConns)
	}
	if p.connMaxLifetime > 0 {
		config.MaxConnLifetime = p.connMaxLifetime
	}
	if p.connMaxIdleTime > 0 {
		config.MaxConnIdleTime = p.connMaxIdleTime
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create pgxpool: %w", err)
	}
	p.pool = pool

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Bridge pgxpool to database/sql for Beego ORM compatibility
	sqlDB := stdlib.OpenDBFromPool(pool)

	// Register driver with Beego ORM
	if err := orm.RegisterDriver("pgx", orm.DRPostgres); err != nil {
		pool.Close()
		return err
	}

	an := "default"
	if len(alias) != 0 {
		an = alias[0]
	}

	// Use AddAliasWthDB to register existing sql.DB with Beego
	// Note: Function name has "Wth" not "With" - this is intentional (Beego API quirk)
	if err := orm.AddAliasWthDB(an, "pgx", sqlDB); err != nil {
		pool.Close()
		return err
	}

	log.Infof("pgxpool initialized: MaxConns=%d, MinConns=%d, MaxLifetime=%v, MaxIdleTime=%v",
		config.MaxConns, config.MinConns, config.MaxConnLifetime, config.MaxConnIdleTime)

	return nil
}

// UpgradeSchema calls migrate tool to upgrade schema to the latest based on the SQL scripts.
func (p *pgsql) UpgradeSchema() error {
	port, err := strconv.Atoi(p.port)
	if err != nil {
		return err
	}
	m, err := NewMigrator(&models.PostGreSQL{
		Host:     p.host,
		Port:     port,
		Username: p.usr,
		Password: p.pwd,
		Database: p.database,
		SSLMode:  p.sslmode,
	})
	if err != nil {
		return err
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil || dbErr != nil {
			log.Warningf("Failed to close migrator, source error: %v, db error: %v", srcErr, dbErr)
		}
	}()
	log.Infof("Upgrading schema for pgsql ...")
	err = m.Up()
	if err == migrate.ErrNoChange {
		log.Infof("No change in schema, skip.")
	} else if err != nil { // migrate.ErrLockTimeout will be thrown when another process is doing migration and timeout.
		log.Errorf("Failed to upgrade schema, error: %q", err)
		return err
	}
	return nil
}

// NewMigrator creates a migrator base on the information
func NewMigrator(database *models.PostGreSQL) (*migrate.Migrate, error) {
	dbURL := url.URL{
		Scheme:   "pgx5",
		User:     url.UserPassword(database.Username, database.Password),
		Host:     net.JoinHostPort(database.Host, strconv.Itoa(database.Port)),
		Path:     database.Database,
		RawQuery: fmt.Sprintf("sslmode=%s", database.SSLMode),
	}

	// For UT
	path := os.Getenv("POSTGRES_MIGRATION_SCRIPTS_PATH")
	if len(path) == 0 {
		path = defaultMigrationPath
	}
	srcURL := fmt.Sprintf("file://%s", path)
	m, err := migrate.New(srcURL, dbURL.String())
	if err != nil {
		return nil, err
	}
	m.Log = newMigrateLogger()
	return m, nil
}
