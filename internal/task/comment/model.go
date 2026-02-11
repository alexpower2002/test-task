package comment

import (
	"time"
)

type Model struct {
	Id          int
	CreatedAt   time.Time
	CommenterId int
	TaskId      int
	Text        string
}
