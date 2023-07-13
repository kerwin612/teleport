// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pgbk

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gravitational/trace"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgtype/zeronull"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jonboulle/clockwork"
	"github.com/sirupsen/logrus"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/api/utils"
	"github.com/gravitational/teleport/lib/backend"
	pgcommon "github.com/gravitational/teleport/lib/backend/pgbk/common"
)

const (
	Name    = "postgresql"
	AltName = "postgres"

	// componentName is the component name used for logging.
	componentName = "pgbk"
)

const (
	defaultChangeFeedBatchSize = 10000
	defaultChangeFeedInterval  = backend.DefaultPollStreamPeriod

	defaultExpiryBatchSize = 1000
	defaultExpiryInterval  = 30 * time.Second
)

// AuthMode determines if we should use some environment-specific authentication
// mechanism or credentials.
type AuthMode string

const (
	// StaticAuth uses the static credentials as defined in the connection
	// string.
	StaticAuth AuthMode = ""
	// AzureADAuth gets a connection token from Azure and uses it as the
	// password when connecting.
	AzureADAuth AuthMode = "azure"
)

// Check returns an error if the AuthMode is invalid.
func (a AuthMode) Check() error {
	switch a {
	case StaticAuth, AzureADAuth:
		return nil
	default:
		return trace.BadParameter("invalid authentication mode %q, should be %q or %q", a, StaticAuth, AzureADAuth)
	}
}

// Config is the configuration struct for [Backend]; outside of tests or custom
// code, it's usually generated by converting the [backend.Params] from the
// Teleport configuration file.
type Config struct {
	ConnString string `json:"conn_string"`

	AuthMode      AuthMode `json:"auth_mode"`
	AzureClientID string   `json:"azure_client_id"`

	ChangeFeedPollInterval types.Duration `json:"change_feed_poll_interval"`
	ChangeFeedBatchSize    int            `json:"change_feed_batch_size"`

	DisableExpiry   bool           `json:"disable_expiry"`
	ExpiryInterval  types.Duration `json:"expiry_interval"`
	ExpiryBatchSize int            `json:"expiry_batch_size"`
}

func (c *Config) CheckAndSetDefaults() error {
	if err := c.AuthMode.Check(); err != nil {
		return trace.Wrap(err)
	}

	if c.AzureClientID != "" && c.AuthMode != AzureADAuth {
		return trace.BadParameter("azure client ID requires azure auth mode")
	}

	if c.ChangeFeedPollInterval < 0 {
		return trace.BadParameter("change feed poll interval must be non-negative")
	}
	if c.ChangeFeedPollInterval == 0 {
		c.ChangeFeedPollInterval = types.Duration(defaultChangeFeedInterval)
	}
	if c.ChangeFeedBatchSize < 0 {
		return trace.BadParameter("change feed batch size must be non-negative")
	}
	if c.ChangeFeedBatchSize == 0 {
		c.ChangeFeedBatchSize = defaultChangeFeedBatchSize
	}

	if c.ExpiryInterval < 0 {
		return trace.BadParameter("expiry interval must be non-negative")
	}
	if c.ExpiryInterval == 0 {
		c.ExpiryInterval = types.Duration(defaultExpiryInterval)
	}
	if c.ExpiryBatchSize < 0 {
		return trace.BadParameter("expiry batch size must be non-negative")
	}
	if c.ExpiryBatchSize == 0 {
		c.ExpiryBatchSize = defaultExpiryBatchSize
	}

	return nil
}

// NewFromParams starts and returns a [*Backend] with the given params
// (generally read from the Teleport configuration file).
func NewFromParams(ctx context.Context, params backend.Params) (*Backend, error) {
	var cfg Config
	if err := utils.ObjectToStruct(params, &cfg); err != nil {
		return nil, trace.Wrap(err)
	}

	bk, err := NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return bk, nil
}

// NewWithConfig starts and returns a [*Backend] with the given [Config].
func NewWithConfig(ctx context.Context, cfg Config) (*Backend, error) {
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.ConnString)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	log := logrus.WithField(trace.Component, componentName)

	if cfg.AuthMode == AzureADAuth {
		bc, err := pgcommon.AzureBeforeConnect(cfg.AzureClientID, log)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		poolConfig.BeforeConnect = bc
	}

	poolConfig.AfterConnect = func(ctx context.Context, c *pgx.Conn) error {
		_, err := c.Exec(ctx, "SET default_transaction_isolation TO serializable", pgx.QueryExecModeExec)
		return trace.Wrap(err)
	}

	log.Info("Setting up backend.")

	pgcommon.TryEnsureDatabase(ctx, poolConfig, log)

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	ctx, cancel := context.WithCancel(ctx)
	b := &Backend{
		cfg:    cfg,
		log:    log,
		pool:   pool,
		buf:    backend.NewCircularBuffer(),
		cancel: cancel,
	}

	if err := b.setupAndMigrate(ctx); err != nil {
		b.Close()
		return nil, trace.Wrap(err)
	}

	if !cfg.DisableExpiry {
		b.wg.Add(1)
		go func() {
			defer b.wg.Done()
			b.backgroundExpiry(ctx)
		}()
	}

	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		b.backgroundChangeFeed(ctx)
	}()

	return b, nil
}

// Backend is a PostgreSQL-backed [backend.Backend].
type Backend struct {
	cfg  Config
	log  logrus.FieldLogger
	pool *pgxpool.Pool
	buf  *backend.CircularBuffer

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func (b *Backend) Close() error {
	b.cancel()
	b.wg.Wait()
	b.buf.Close()
	b.pool.Close()
	return nil
}

var schemas = []string{
	`CREATE TABLE kv (
		key bytea NOT NULL,
		value bytea NOT NULL,
		expires timestamptz,
		revision uuid NOT NULL,
		CONSTRAINT kv_pkey PRIMARY KEY (key)
	);
	CREATE INDEX kv_expires_idx ON kv (expires) WHERE expires IS NOT NULL;`,
}

// setupAndMigrate sets up the database schema, applying migrations in order if
// necessary.
func (b *Backend) setupAndMigrate(ctx context.Context) error {
	var version int32
	var migrateErr error

	if err := pgcommon.Retry0(ctx, b.log, func() error {
		if _, err := b.pool.Exec(ctx,
			`CREATE TABLE IF NOT EXISTS backend_version (
				version integer PRIMARY KEY CHECK (version > 0),
				created timestamptz NOT NULL DEFAULT now()
			)`, pgx.QueryExecModeExec,
		); err != nil && !pgcommon.IsCode(err, pgerrcode.InsufficientPrivilege) {
			return trace.Wrap(err)
		}
		return nil
	}); err != nil {
		return trace.Wrap(err)
	}

	if err := pgcommon.RetryTx(ctx, b.log, b.pool, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	}, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx,
			"SELECT COALESCE(max(version), 0) FROM backend_version",
			pgx.QueryExecModeExec,
		).Scan(&version); err != nil {
			return trace.Wrap(err)
		}

		if int(version) > len(schemas) {
			migrateErr = trace.BadParameter("unsupported schema version %v", version)
			return nil
		}

		if int(version) == len(schemas) {
			return nil
		}

		for _, s := range schemas[version:] {
			if _, err := tx.Exec(ctx, s, pgx.QueryExecModeExec); err != nil {
				return trace.Wrap(err)
			}
		}

		if _, err := tx.Exec(ctx,
			"INSERT INTO backend_version (version) VALUES ($1)",
			pgx.QueryExecModeExec, len(schemas),
		); err != nil {
			return trace.Wrap(err)
		}

		return nil
	}); err != nil {
		return trace.Wrap(err)
	}

	if migrateErr != nil {
		return migrateErr
	}

	if int(version) != len(schemas) {
		b.log.WithFields(logrus.Fields{
			"previous_version": version,
			"current_version":  len(schemas),
		}).Info("Migrated database schema.")
	}

	return nil
}

var _ backend.Backend = (*Backend)(nil)

// Create implements [backend.Backend].
func (b *Backend) Create(ctx context.Context, i backend.Item) (*backend.Lease, error) {
	revision := newRevision()
	created, err := pgcommon.Retry(ctx, b.log, func() (bool, error) {
		tag, err := b.pool.Exec(ctx,
			"INSERT INTO kv (key, value, expires, revision) VALUES ($1, $2, $3, $4)"+
				" ON CONFLICT (key) DO UPDATE SET"+
				" value = excluded.value, expires = excluded.expires, revision = excluded.revision"+
				" WHERE kv.expires IS NOT NULL AND kv.expires <= now()",
			i.Key, i.Value, zeronull.Timestamptz(i.Expires.UTC()), revision)
		if err != nil {
			return false, trace.Wrap(err)
		}
		return tag.RowsAffected() > 0, nil
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if !created {
		return nil, trace.AlreadyExists("key %q already exists", i.Key)
	}
	return newLease(i), nil
}

// Put implements [backend.Backend].
func (b *Backend) Put(ctx context.Context, i backend.Item) (*backend.Lease, error) {
	revision := newRevision()
	if err := pgcommon.Retry0(ctx, b.log, func() error {
		_, err := b.pool.Exec(ctx,
			"INSERT INTO kv (key, value, expires, revision) VALUES ($1, $2, $3, $4)"+
				" ON CONFLICT (key) DO UPDATE SET"+
				" value = excluded.value, expires = excluded.expires, revision = excluded.revision",
			i.Key, i.Value, zeronull.Timestamptz(i.Expires.UTC()), revision)
		return trace.Wrap(err)
	}); err != nil {
		return nil, trace.Wrap(err)
	}

	return newLease(i), nil
}

// CompareAndSwap implements [backend.Backend].
func (b *Backend) CompareAndSwap(ctx context.Context, expected backend.Item, replaceWith backend.Item) (*backend.Lease, error) {
	if !bytes.Equal(expected.Key, replaceWith.Key) {
		return nil, trace.BadParameter("expected and replaceWith keys should match")
	}

	revision := newRevision()
	swapped, err := pgcommon.Retry(ctx, b.log, func() (bool, error) {
		tag, err := b.pool.Exec(ctx,
			"UPDATE kv SET value = $1, expires = $2, revision = $3"+
				" WHERE key = $4 AND value = $5 AND (expires IS NULL OR expires > now())",
			replaceWith.Value, zeronull.Timestamptz(replaceWith.Expires.UTC()), revision,
			replaceWith.Key, expected.Value)
		if err != nil {
			return false, trace.Wrap(err)
		}
		return tag.RowsAffected() > 0, nil
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if !swapped {
		return nil, trace.CompareFailed("key %q does not exist or does not match expected", replaceWith.Key)
	}
	return newLease(replaceWith), nil
}

// Update implements [backend.Backend].
func (b *Backend) Update(ctx context.Context, i backend.Item) (*backend.Lease, error) {
	revision := newRevision()
	updated, err := pgcommon.Retry(ctx, b.log, func() (bool, error) {
		tag, err := b.pool.Exec(ctx,
			"UPDATE kv SET value = $1, expires = $2, revision = $3"+
				" WHERE key = $4 AND (expires IS NULL OR expires > now())",
			i.Value, zeronull.Timestamptz(i.Expires.UTC()), revision, i.Key)
		if err != nil {
			return false, trace.Wrap(err)
		}
		return tag.RowsAffected() > 0, nil
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if !updated {
		return nil, trace.NotFound("key %q does not exist", i.Key)
	}
	return newLease(i), nil
}

// Get implements [backend.Backend].
func (b *Backend) Get(ctx context.Context, key []byte) (*backend.Item, error) {
	item, err := pgcommon.Retry(ctx, b.log, func() (*backend.Item, error) {
		batch := new(pgx.Batch)
		batch.Queue("SET transaction_read_only TO on")

		var item *backend.Item
		batch.Queue("SELECT value, expires, revision FROM kv"+
			" WHERE key = $1 AND (expires IS NULL OR expires > now())", key,
		).QueryRow(func(row pgx.Row) error {
			var value []byte
			var expires zeronull.Timestamptz
			var revision pgtype.UUID
			if err := row.Scan(&value, &expires, &revision); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return nil
				}
				return trace.Wrap(err)
			}

			item = &backend.Item{
				Key:     key,
				Value:   value,
				Expires: time.Time(expires).UTC(),
				// Revision: crockfordLowercaseBase32.EncodeToString(revision.Bytes[:]),
			}
			return nil
		})

		if err := b.pool.SendBatch(ctx, batch).Close(); err != nil {
			return nil, trace.Wrap(err)
		}

		return item, nil
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if item == nil {
		return nil, trace.NotFound("key %q does not exist", key)
	}
	return item, nil
}

// GetRange implements [backend.Backend].
func (b *Backend) GetRange(ctx context.Context, startKey []byte, endKey []byte, limit int) (*backend.GetResult, error) {
	if limit <= 0 {
		limit = backend.DefaultRangeLimit
	}

	items, err := pgcommon.Retry(ctx, b.log, func() ([]backend.Item, error) {
		batch := new(pgx.Batch)
		batch.Queue("SET transaction_read_only TO on")
		// TODO(espadolini): figure out if we want transaction_deferred enabled
		// for GetRange

		var items []backend.Item
		batch.Queue(
			"SELECT key, value, expires, revision FROM kv"+
				" WHERE key BETWEEN $1 AND $2 AND (expires IS NULL OR expires > now())"+
				" ORDER BY key LIMIT $3",
			startKey, endKey, limit,
		).Query(func(rows pgx.Rows) error {
			var err error
			items, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (backend.Item, error) {
				var key, value []byte
				var expires zeronull.Timestamptz
				var revision pgtype.UUID
				if err := row.Scan(&key, &value, &expires, &revision); err != nil {
					return backend.Item{}, err
				}
				return backend.Item{
					Key:     key,
					Value:   value,
					Expires: time.Time(expires).UTC(),
					// Revision: crockfordLowercaseBase32.EncodeToString(revision.Bytes[:]),
				}, nil
			})
			return trace.Wrap(err)
		})

		if err := b.pool.SendBatch(ctx, batch).Close(); err != nil {
			return nil, trace.Wrap(err)
		}

		return items, nil
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &backend.GetResult{Items: items}, nil
}

// Delete implements [backend.Backend].
func (b *Backend) Delete(ctx context.Context, key []byte) error {
	deleted, err := pgcommon.Retry(ctx, b.log, func() (bool, error) {
		tag, err := b.pool.Exec(ctx,
			"DELETE FROM kv WHERE key = $1 AND (expires IS NULL OR expires > now())", key)
		if err != nil {
			return false, trace.Wrap(err)
		}
		return tag.RowsAffected() > 0, nil
	})
	if err != nil {
		return trace.Wrap(err)
	}

	if !deleted {
		return trace.NotFound("key %q does not exist", key)
	}
	return nil
}

// DeleteRange implements [backend.Backend].
func (b *Backend) DeleteRange(ctx context.Context, startKey []byte, endKey []byte) error {
	// this is the only backend operation that might affect a disproportionate
	// amount of rows at the same time; in actual operation, DeleteRange hardly
	// ever deletes more than dozens of items at once, and logical decoding
	// starts having performance issues when a transaction affects _thousands_
	// of rows at once, so we're good here (but see [Backend.backgroundExpiry])
	if err := pgcommon.Retry0(ctx, b.log, func() error {
		_, err := b.pool.Exec(ctx,
			"DELETE FROM kv WHERE key BETWEEN $1 AND $2",
			startKey, endKey,
		)
		return trace.Wrap(err)
	}); err != nil {
		return trace.Wrap(err)
	}

	return nil
}

// KeepAlive implements [backend.Backend].
func (b *Backend) KeepAlive(ctx context.Context, lease backend.Lease, expires time.Time) error {
	revision := newRevision()
	updated, err := pgcommon.Retry(ctx, b.log, func() (bool, error) {
		tag, err := b.pool.Exec(ctx,
			"UPDATE kv SET expires = $1, revision = $2 WHERE key = $3 AND (expires IS NULL OR expires > now())",
			zeronull.Timestamptz(expires.UTC()), revision, lease.Key)
		if err != nil {
			return false, trace.Wrap(err)
		}
		return tag.RowsAffected() > 0, nil
	})
	if err != nil {
		return trace.Wrap(err)
	}

	if !updated {
		return trace.NotFound("key %q does not exist", lease.Key)
	}
	return nil
}

// NewWatcher implements [backend.Backend].
func (b *Backend) NewWatcher(ctx context.Context, watch backend.Watch) (backend.Watcher, error) {
	return b.buf.NewWatcher(ctx, watch)
}

// CloseWatchers implements [backend.Backend].
func (b *Backend) CloseWatchers() { b.buf.Clear() }

// Clock implements implements [backend.Backend].
func (b *Backend) Clock() clockwork.Clock { return clockwork.NewRealClock() }
