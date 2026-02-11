package history

import "time"

type Model struct {
	Id          int       `json:"id"`
	TaskId      int       `json:"task_id"`
	Status      string    `json:"status"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatorId   int       `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	AssigneeId  int       `json:"assignee_id"`
	TeamId      int       `json:"team_id"`
	ChangedBy   int       `json:"changed_by"`
	ChangedAt   time.Time `json:"changed_at"`
}
