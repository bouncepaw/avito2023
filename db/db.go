package db

import (
	"context"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

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
	err := db.Close()
	if err != nil {
		log.Fatal(err)
	}
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

func removeFromSegmentsByIds(ctx context.Context, userId int, segmentIds []int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const (
		qRemove = `delete from users_to_segments where user_id = $1 and segment_id = $2;`
		qRecord = `insert into operation_history (user_id, segment_id, operation) values ($1, $2, 'remove');`
	)

	for _, id := range segmentIds {
		if _, err = tx.ExecContext(ctx, qRemove, userId, id); err != nil {
			return err
		}

		if _, err = tx.ExecContext(ctx, qRecord, userId, id); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func UpdateUser(ctx context.Context, userId int, addTo []string, removeFrom []string, ttl int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	const (
		qSegmentIdByName   = `select id from segments where name = $1;`
		qAddToSegment      = `insert into users_to_segments (user_id, segment_id) values ($1, $2);`
		qRecordAdd         = `insert into operation_history (user_id, segment_id, operation) values ($1, $2, 'add');`
		qRemoveFromSegment = `delete from users_to_segments where user_id = $1 and segment_id = $2;`
		qRecordRemove      = `insert into operation_history (user_id, segment_id, operation) values ($1, $2, 'remove');`
	)

	var addToSegmendIds []int
	var segmentId int
	for _, name := range addTo {
		// Get id for name
		row := tx.QueryRowContext(ctx, qSegmentIdByName, name)
		switch err = row.Scan(&segmentId); {
		case errors.Is(err, sql.ErrNoRows):
			log.Printf("Didn't find id for segment %s\n", name)
			return errNameFree
		case err != nil:
			return err
		}
		addToSegmendIds = append(addToSegmendIds, segmentId)

		// Save relation
		if _, err = tx.ExecContext(ctx, qAddToSegment, userId, segmentId); err != nil {
			return err
		}

		if _, err = tx.ExecContext(ctx, qRecordAdd, userId, segmentId); err != nil {
			return err
		}
	}

	if ttl > 0 {
		schedule <- removeTask{
			ttl:        ttl,
			userId:     userId,
			segmentIds: addToSegmendIds,
		}
	}

	for _, name := range removeFrom {
		// Get id for name
		row := tx.QueryRowContext(ctx, qSegmentIdByName, name)
		switch err = row.Scan(&segmentId); {
		case errors.Is(err, sql.ErrNoRows):
			log.Printf("Didn't find id for segment %s\n", name)
			return errNameFree
		case err != nil:
			return err
		}

		// Save relation
		if _, err = tx.ExecContext(ctx, qRemoveFromSegment, userId, segmentId); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, qRecordRemove, userId, segmentId); err != nil {
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

func GetHistory(ctx context.Context, year, month int) (string, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	// The first moment of the given month and the first moment
	// of the month after the given one.
	thisMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	var nextMonth time.Time
	if month == 12 {
		nextMonth = time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	} else {
		nextMonth = time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)
	}

	const qHistory = `
select stamp, user_id, segments.name, operation
from operation_history
join segments on segment_id = id
where segment_id = segments.id;
`
	rows, err := tx.QueryContext(ctx, qHistory, thisMonth, nextMonth)
	if err != nil {
		return "", err
	}

	var (
		buf    strings.Builder
		csvDoc = csv.NewWriter(&buf)
	)
	csvDoc.Comma = ';'

	for rows.Next() {
		var (
			stamp       time.Time
			userId      int
			segmentName string
			operation   string
		)

		err = rows.Scan(&stamp, &userId, &segmentName, &operation)
		if err != nil {
			return "", err
		}

		// See swagger.yml to learn about field order.
		err = csvDoc.Write([]string{
			strconv.Itoa(userId),
			segmentName,
			operation,
			stamp.String(),
		})
		if err != nil {
			return "", err
		}
	}
	csvDoc.Flush()
	return buf.String(), nil
}
