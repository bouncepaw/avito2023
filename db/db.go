package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host   = "postgres"
	port   = 5432
	dbname = "postgres"
)

var (
	db *sql.DB

	errNameTaken      = errors.New("name taken")
	errNameFree       = errors.New("name free")
	errAlreadyDeleted = errors.New("already deleted")
	errBadPercent     = errors.New("bad percent")
)

func init() {
	psqlInfo := fmt.Sprintf("host=%s port=%d dbname=%s user=postgres sslmode=disable password=password", host, port, dbname)
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
	if percent < 0 || percent > 100 {
		return errBadPercent
	}
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

func UpdateUser(ctx context.Context, userId int, addTo []string, removeFrom []string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const (
		qSegmentIdByName   = `select id from segments where name = $1;`
		qAddToSegment      = `insert into users_to_segments (user_id, segment_id) values ($1, $2);`
		qRemoveFromSegment = `delete from users_to_segments where user_id = $1 and segment_id = $2;`
	)

	var segmentId int
	for _, name := range addTo {
		// Get id for name
		row := tx.QueryRowContext(ctx, qSegmentIdByName, name)
		switch err = row.Scan(&segmentId); {
		case errors.Is(err, sql.ErrNoRows):
			return errNameFree
		case err != nil:
			return err
		}

		// Save relation
		_, err = tx.ExecContext(ctx, qAddToSegment, userId, segmentId)
		if err != nil {
			return err
		}
	}

	for _, name := range removeFrom {
		// Get id for name
		row := tx.QueryRowContext(ctx, qSegmentIdByName, name)
		switch err = row.Scan(&segmentId); {
		case errors.Is(err, sql.ErrNoRows):
			return errNameFree
		case err != nil:
			return err
		}

		// Save relation
		_, err = tx.ExecContext(ctx, qRemoveFromSegment, userId, segmentId)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func GetSegments(ctx context.Context, userId int) ([]string, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	const (
		qSegments = `
select name from segments
join users_to_segments uts
on segments.id = uts.segment_id
where uts.user_id = $1;
`
	)

	var segments []string
	rows, err := tx.QueryContext(ctx, qSegments, userId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var s string
		err = rows.Scan(&s)
		if err != nil {
			return nil, err
		}
		segments = append(segments, s)
	}

	return segments, tx.Commit()
}
