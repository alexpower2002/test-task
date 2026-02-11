package task

import "time"

type Model struct {
	Id          int
	Status      string
	Title       string
	Description string
	CreatorId   int
	CreatedAt   time.Time
	AssigneeId  int
	TeamId      int
}
