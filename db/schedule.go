package db

import (
	"context"
	"database/sql"
	"log"
	"time"
)

type removeTask struct {
	ttl       int
	userId    int
	segmentId int
}

var (
	// Actors send many tasks to one queue simultaneously using planRemoval.
	schedule = make(chan removeTask)

	// But we organize them and execute no more than one at any given time. Do not
	// send here yourself.
	queue = make(chan removeTask)
)

func startSchedule() {
	go func() {
		for task := range queue {
			if err := removePerPlan(task.userId, task.segmentId); err != nil {
				log.Println(err)
			}
		}
	}()
	go func() {
		for task := range schedule {
			go func(task removeTask) {
				<-time.After(time.Second * time.Duration(task.ttl))
				queue <- task
			}(task)
		}
	}()
	populateSchedule()
}

// This function is run at start up, so we crash on any error.
func populateSchedule() {
	const qSchedule = `
select stamp, user_id, segment_id
from delayed_removals
`
	rows, err := db.Query(qSchedule)
	if err != nil {
		log.Fatal(err)
	}

	var (
		now   = time.Now()
		tasks []removeTask
	)
	for rows.Next() {
		var stamp time.Time
		var task removeTask
		if err = rows.Scan(&stamp, &task.userId, &task.segmentId); err != nil {
			log.Fatal(err)
		}
		if now.After(stamp) {
			task.ttl = -1
		} else {
			task.ttl = int(stamp.Sub(now).Seconds())
		}
		tasks = append(tasks, task)
	}
	log.Printf("Found %d scheduled tasks\n", len(tasks))

	for _, task := range tasks {
		if task.ttl <= 0 {
			if err := removePerPlan(task.userId, task.segmentId); err != nil {
				log.Fatal(err)
			}
		} else {
			tx, err := db.Begin()
			if err != nil {
				log.Fatal(err)
			}
			if err = planRemoval(context.Background(), tx, task.ttl, task.userId, task.segmentId); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func planRemoval(ctx context.Context, tx *sql.Tx, ttl int, userId int, addToSegmentIds ...int) error {
	const qPlan = `
insert into delayed_removals (stamp, user_id, segment_id)
values ($1, $2, $3)
on conflict do nothing;
`

	eta := time.Now().Add(time.Duration(ttl) * time.Second)

	for _, segmentId := range addToSegmentIds {
		if _, err := tx.ExecContext(ctx, qPlan, eta, userId, segmentId); err != nil {
			return err
		}
	}

	log.Printf("Planned %d tasks at %s", len(addToSegmentIds), eta)

	for _, segmentId := range addToSegmentIds {
		schedule <- removeTask{
			ttl:       ttl,
			userId:    userId,
			segmentId: segmentId,
		}
	}

	return nil
}

func removePerPlan(userId int, segmentId int) error {
	log.Printf("Removing %d from %d as per plan\n", userId, segmentId)
	// Steps:
	// 1. Delete
	// 2. Update operation history
	// 3. Update schedule
	const q = `
with deleted as (
   delete from users_to_segments
   where user_id = $1 and segment_id = $2
   returning user_id, segment_id
), history as (
   insert into operation_history (user_id, segment_id, operation)
	select user_id, segment_id, 'remove'
	from deleted
)
delete from delayed_removals
where user_id = $1 and segment_id = $2;
`
	_, err := db.Exec(q, userId, segmentId)
	return err
}
