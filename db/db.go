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

	optsRO = &sql.TxOptions{ReadOnly: true}
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

func CreateSegment(ctx context.Context, name string, percent uint) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if percent < 0 || percent > 100 {
		return errBadPercent
	}

	const qCreate = `insert into segments (name, automatic_percent) values ($1, $2);`
	_, err = tx.ExecContext(ctx, qCreate, name, percent)

	// When we try to write an existing name, the following error is returned:
	//     pq: duplicate key value violates unique constraint "segments_name_key"
	// We parse the error message post factum instead of checking if the name is
	// free beforehand to make one less round trip.
	if err != nil && strings.Contains(err.Error(), "duplicate key") {
		return errNameTaken
	} else if err != nil {
		return err
	}

	if percent > 0 {
		const qRetro = `
with percented(user_id, percent_rank) as ( -- Get user with percent value
   select distinct (user_id), percent_rank() over (order by random())
   from users_to_segments
), sample as ( -- Get $1 % users according to the percent rank
   select user_id from percented
   where percent_rank < ($2 / 100.0) -- .0 required to avoid rounding to zero
), records_to_write as ( -- Prepare new records for insertion
	select user_id, segments.id as segment_id
	from sample
	join segments on true
	where segments.name = $1 and segments.deleted = false
), written_records as ( -- Not all records were written.
   insert into users_to_segments (user_id, segment_id)
	select user_id, segment_id
	from records_to_write
	on conflict do nothing -- Got an explicit entry like that? Whatever, move on.
   returning user_id, segment_id
)
insert into operation_history (user_id, segment_id, operation)
select user_id, segment_id, 'add'
from written_records;
`
		_, err = tx.ExecContext(ctx, qRetro, name, percent)
	}

	return tx.Commit()
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
		qInfect            = `
with random_val as (
   select random() * 100.0 as xi
), bonus_segments as (
   select id
   from segments
   cross join random_val
   where deleted = false and xi <= automatic_percent
), insertions as (
   insert into users_to_segments (user_id, segment_id)
   select $1, id
   from bonus_segments
   on conflict do nothing 
   returning segment_id
)
insert into operation_history (user_id, segment_id, operation)
select $1, segment_id, 'add'
from insertions;
`
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

	if _, err = tx.ExecContext(ctx, qInfect, userId); err != nil {
		return err
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
	tx, err := db.BeginTx(ctx, optsRO)
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
	tx, err := db.BeginTx(ctx, optsRO)
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
where stamp >= $1 and stamp < $2;
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
