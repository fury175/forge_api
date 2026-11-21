package pgdb

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	repo "github.com/fury175/forge_api/internal/adapters/postgresql/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Select(ctx context.Context, dst any, sql string, args ...any) error
	WithTransaction(ctx context.Context, fn func(repo.Querier) error) error
}

type PgxDB struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) DB {
	return &PgxDB{pool: pool}
}

func (d *PgxDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return d.pool.QueryRow(ctx, sql, args...)
}

func (d *PgxDB) Select(ctx context.Context, dst any, sql string, args ...any) error {
	rows, err := d.pool.Query(ctx, sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return err
		}
		return pgx.ErrNoRows
	}

	return scanRow(rows, dst)
}

// scanRow scans the current row into the struct pointed to by dst,
// matching columns to struct fields by column name or json tag.
func scanRow(rows pgx.Rows, dst any) error {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("pgdb: dst must be a non-nil pointer, got %T", dst)
	}
	structVal := rv.Elem()
	if structVal.Kind() != reflect.Struct {
		return fmt.Errorf("pgdb: dst must point to a struct, got %T", dst)
	}

	descriptions := rows.FieldDescriptions()
	fields := make([]any, len(descriptions))
	for i, fd := range descriptions {
		fv := structField(structVal, string(fd.Name))
		if !fv.IsValid() {
			return fmt.Errorf("pgdb: %s has no field for column %q", structVal.Type(), fd.Name)
		}
		fields[i] = fv.Addr().Interface()
	}

	return rows.Scan(fields...)
}

func structField(structVal reflect.Value, column string) reflect.Value {
	t := structVal.Type()
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := f.Name
		if tag, ok := f.Tag.Lookup("json"); ok {
			name = strings.Split(tag, ",")[0]
		}
		if name == "" {
			continue
		}
		if strings.EqualFold(name, column) {
			return structVal.Field(i)
		}
	}
	return reflect.Value{}
}

func (d *PgxDB) WithTransaction(ctx context.Context, fn func(repo.Querier) error) error {
	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := repo.New(tx)
	if err := fn(qtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
