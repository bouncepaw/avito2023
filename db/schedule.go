package db

import (
	"context"
	"log"
	"time"
)

/*
This is the simplest scheduler.
*/

type removeTask struct {
	ttl        int
	userId     int
	segmentIds []int
}

var schedule = make(chan removeTask)

func init() {
	go func() {
		for task := range schedule {
			log.Println("Received a task in ", task.ttl)
			go func(task removeTask) {
				<-time.After(time.Second * time.Duration(task.ttl))
				if err := removeFromSegmentsByIds(context.Background(), task.userId, task.segmentIds); err != nil {
					log.Println(err)
				}
			}(task)
		}
	}()
}
