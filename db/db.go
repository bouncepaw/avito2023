package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host   = "localhost"
	port   = 5432
	dbname = "bouncepaw"
)

var (
	db *sql.DB

	errNameTaken      = errors.New("name taken")
	errNameFree       = errors.New("name free")
	errAlreadyDeleted = errors.New("already deleted")
)

func init() {
	psqlInfo := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=disable", host, port, dbname)
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
}

func Close() {
	db.Close()
}

func segmentNameTaken(ctx context.Context, name string) (bool, error) {
	const q = `
select exists(
	select 1
	from segments
	where name = $1 
);
`

	rows, err := db.QueryContext(ctx, q, name)
	if err != nil {
		return false, err
	}

	rows.Next()
	var result bool

	err = rows.Scan(&result)
	if err != nil {
		return false, err
	}

	err = rows.Close()
	if err != nil {
		return false, err
	}

	return result, nil
}

func CreateSegment(ctx context.Context, name string, percent uint) error {
	if nameTaken, err := segmentNameTaken(ctx, name); nameTaken {
		return errNameTaken
	} else if err != nil {
		return err
	}

	// The name is free to use, no conflicts.

	const q = `insert into segments (name, automatic_percent) values ($1, $2);`
	_, err := db.ExecContext(ctx, q, name, percent)
	return err
}

func DeleteSegment(ctx context.Context, name string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if segment is actually deletable.
	const qInfo = `select id, deleted from segments where name = $1;`
	row := tx.QueryRowContext(ctx, qInfo, name)

	var (
		id      int
		deleted bool
	)
	err = row.Scan(&id, &deleted)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return errNameFree
	case err != nil:
		return err
	case deleted:
		return errAlreadyDeleted
	}

	// OK. We can delete now.
	const qMarkDeletion = `update segments set deleted = true where id = $1;`
	_, err = tx.ExecContext(ctx, qMarkDeletion, id)
	if err != nil {
		return err
	}

	const qRemoveUsers = `delete from users_to_segments where segment_id = $1;`
	_, err = tx.ExecContext(ctx, qRemoveUsers, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}