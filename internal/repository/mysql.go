package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"mkk-luna-test-task/internal/task"
	"mkk-luna-test-task/internal/task/comment"
	"mkk-luna-test-task/internal/task/history"
	"mkk-luna-test-task/internal/team"
	"mkk-luna-test-task/internal/team/member"
	"mkk-luna-test-task/internal/user"
)

const (
	queryInsertTaskComment = `
		INSERT INTO task_comments (commenter_id, task_id, text, created_at)
		VALUES (?, ?, ?, ?)
	`

	queryInviteMember = `
		INSERT INTO team_members (user_id, team_id, role)
		VALUES (?, ?, ?)
	`

	queryCanUserInvite = `
		SELECT EXISTS(
			SELECT 1
			FROM team_members
			WHERE team_id = ? AND user_id = ? AND role IN ('owner', 'admin')
			LIMIT 1
		)
	`

	queryTeamInsert = `
		INSERT INTO teams (name)
		VALUES (?)
	`

	queryOwnerInsert = `
		INSERT INTO team_members (team_id, user_id, role)
		VALUES (?, ?, 'owner')
	`

	queryListTeams = `
		SELECT t.id, t.name
		FROM teams t
		INNER JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = ?
	`

	queryIsPartOfTeamByTaskId = `
		SELECT COUNT(*) 
		FROM team_members tm
		INNER JOIN tasks t ON tm.team_id = t.team_id
		WHERE tm.user_id = ? AND t.id = ?
	`

	queryCheckUserIsInTeam = `
		SELECT COUNT(*) FROM team_members WHERE user_id = ? AND team_id = ?
	`

	queryRegisterUser = `
		INSERT INTO users (username, password_hashed)
		VALUES (?, ?)
	`

	queryGetUser = `
		SELECT id, username, password_hashed
		FROM users
		WHERE username = ?
	`

	queryListTaskComments = `
		SELECT id, created_at, commenter_id, task_id, text
		FROM task_comments
		WHERE task_id = ? AND id > ?
		ORDER BY id ASC
		LIMIT ?
	`

	queryInsertTaskHistory = `
		INSERT INTO task_history (
			task_id,
			status,
			title,
			description,
			creator_id,
			assignee_id,
			team_id,
			changed_by,
			changed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	queryListHistory = `
		SELECT 
			id,
			task_id,
			status,
			title,
			description,
			creator_id,
			assignee_id,
			team_id,
			changed_by,
			changed_at
		FROM task_history
		WHERE task_id = ? AND id > ?
		ORDER BY id ASC
		LIMIT ?
	`

	queryInsertTask = `
		INSERT INTO tasks (status, title, description, creator_id, assignee_id, team_id, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	queryGetTaskById = `
		SELECT id, status, title, description, creator_id, assignee_id, team_id, created_at
		FROM tasks
		WHERE id = ?
	`

	queryEditTask = `
		UPDATE tasks
		SET status = ?, title = ?, description = ?, assignee_id = ?
		WHERE id = ?
	`
)

type Mysql struct {
	db *sql.DB
}

func NewMysql(db *sql.DB) *Mysql {
	return &Mysql{db: db}
}

func (r *Mysql) CreateTaskComment(
	ctx context.Context,
	commenterId int,
	taskId int,
	text string,
	createdAt time.Time,
) (*comment.Model, error) {
	result, err := r.db.ExecContext(ctx, queryInsertTaskComment, commenterId, taskId, text, createdAt)

	if err != nil {
		return nil, err
	}

	insertedId, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	return &comment.Model{
		Id:          int(insertedId),
		CreatedAt:   createdAt,
		CommenterId: commenterId,
		TaskId:      taskId,
		Text:        text,
	}, nil
}

func (r *Mysql) ListTaskComments(
	ctx context.Context,
	taskId int,
	startFromId int,
	limit int,
) ([]comment.Model, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("invalid limit")
	}

	rows, err := r.db.QueryContext(ctx, queryListTaskComments, taskId, startFromId, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var comments []comment.Model

	for rows.Next() {
		var c comment.Model

		if err := rows.Scan(&c.Id, &c.CreatedAt, &c.CommenterId, &c.TaskId, &c.Text); err != nil {
			return nil, err
		}

		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *Mysql) CreateHistory(
	ctx context.Context,
	t *task.Model,
	changedBy int,
	changedAt time.Time,
) (*history.Model, error) {
	result, err := r.db.ExecContext(
		ctx,
		queryInsertTaskHistory,
		t.Id,
		t.Status,
		t.Title,
		t.Description,
		t.CreatorId,
		t.AssigneeId,
		t.TeamId,
		changedBy,
		changedAt,
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	return &history.Model{
		Id:          int(id),
		TaskId:      t.Id,
		Status:      t.Status,
		Title:       t.Title,
		Description: t.Description,
		CreatorId:   t.CreatorId,
		AssigneeId:  t.AssigneeId,
		TeamId:      t.TeamId,
		ChangedBy:   changedBy,
		ChangedAt:   changedAt,
	}, nil
}

func (r *Mysql) ListHistory(
	ctx context.Context,
	taskId int,
	startFromId int,
	limit int,
) ([]*history.Model, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("invalid limit")
	}

	rows, err := r.db.QueryContext(ctx, queryListHistory, taskId, startFromId, limit)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var histories []*history.Model

	for rows.Next() {
		var h history.Model

		if err := rows.Scan(
			&h.Id,
			&h.TaskId,
			&h.Status,
			&h.Title,
			&h.Description,
			&h.CreatorId,
			&h.AssigneeId,
			&h.TeamId,
			&h.ChangedBy,
			&h.ChangedAt,
		); err != nil {
			return nil, err
		}

		histories = append(histories, &h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return histories, nil
}

func (r *Mysql) CreateTask(
	ctx context.Context,
	status, title, description string,
	creatorId, assigneeId, teamId int,
	createdAt time.Time,
) (*task.Model, error) {
	result, err := r.db.ExecContext(
		ctx,
		queryInsertTask,
		status,
		title,
		description,
		creatorId,
		assigneeId,
		teamId,
		createdAt,
	)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	return &task.Model{
		Id:          int(id),
		Status:      status,
		Title:       title,
		Description: description,
		CreatorId:   creatorId,
		AssigneeId:  assigneeId,
		TeamId:      teamId,
		CreatedAt:   createdAt,
	}, nil
}

func (r *Mysql) GetTaskById(
	ctx context.Context,
	id int,
) (*task.Model, error) {
	var t task.Model

	err := r.db.QueryRowContext(ctx, queryGetTaskById, id).Scan(
		&t.Id,
		&t.Status,
		&t.Title,
		&t.Description,
		&t.CreatorId,
		&t.AssigneeId,
		&t.TeamId,
		&t.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *Mysql) ListTasks(
	ctx context.Context,
	teamId int,
	status string,
	assigneeId int,
	startFromId int,
	limit int,
) ([]*task.Model, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("invalid limit")
	}

	sb := strings.Builder{}

	sb.WriteString("SELECT id, status, title, description, creator_id, assignee_id, team_id, created_at FROM tasks WHERE team_id = ?")

	args := []any{teamId}

	if status != "" {
		sb.WriteString(" AND status = ?")

		args = append(args, status)
	}

	if assigneeId != 0 {
		sb.WriteString(" AND assignee_id = ?")

		args = append(args, assigneeId)
	}

	if startFromId != 0 {
		sb.WriteString(" AND id > ?")

		args = append(args, startFromId)
	}

	sb.WriteString(" ORDER BY id ASC LIMIT ?")

	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, sb.String(), args...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tasks := make([]*task.Model, 0, limit)

	for rows.Next() {
		var t task.Model

		if err := rows.Scan(
			&t.Id,
			&t.Status,
			&t.Title,
			&t.Description,
			&t.CreatorId,
			&t.AssigneeId,
			&t.TeamId,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}

		tasks = append(tasks, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *Mysql) EditTask(
	ctx context.Context,
	id int,
	status, title, description string,
	assigneeId int,
) error {
	result, err := r.db.ExecContext(
		ctx,
		queryEditTask,
		status,
		title,
		description,
		assigneeId,
		id,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *Mysql) InviteMember(
	ctx context.Context,
	userId, teamId int,
	role member.Role,
) (*member.Model, error) {
	_, err := r.db.ExecContext(ctx, queryInviteMember, userId, teamId, role)

	if err != nil {
		return nil, err
	}

	return &member.Model{
		UserId: userId,
		TeamId: teamId,
		Role:   role,
	}, nil
}

func (r *Mysql) CanUserInvite(
	ctx context.Context,
	teamId int,
	userId int,
) (bool, error) {
	exists := false

	err := r.db.QueryRowContext(ctx, queryCanUserInvite, teamId, userId).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *Mysql) CreateTeamAndMakeUserItsOwner(
	ctx context.Context,
	name string,
	userId int,
) (*team.Model, error) {
	tx, err := r.db.BeginTx(ctx, nil)

	if err != nil {
		return nil, err
	}

	result, err := tx.ExecContext(ctx, queryTeamInsert, name)

	if err != nil {
		tx.Rollback()

		return nil, err
	}

	teamID, err := result.LastInsertId()

	if err != nil {
		tx.Rollback()

		return nil, err
	}

	_, err = tx.ExecContext(ctx, queryOwnerInsert, teamID, userId)

	if err != nil {
		tx.Rollback()

		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &team.Model{
		Id:   int(teamID),
		Name: name,
	}, nil
}

func (r *Mysql) ListTeams(
	ctx context.Context,
	userId int,
) ([]*team.Model, error) {
	rows, err := r.db.QueryContext(ctx, queryListTeams, userId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	teams := make([]*team.Model, 0)

	for rows.Next() {
		var t team.Model

		if err := rows.Scan(&t.Id, &t.Name); err != nil {
			return nil, err
		}

		teams = append(teams, &t)
	}

	return teams, nil
}

func (r *Mysql) IsPartOfTeamByTaskId(
	userId, taskId int,
) (bool, error) {
	count := 0

	err := r.db.QueryRow(queryIsPartOfTeamByTaskId, userId, taskId).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Mysql) CheckUserIsInTeam(
	userId, teamId int,
) (bool, error) {
	count := 0

	err := r.db.QueryRow(queryCheckUserIsInTeam, userId, teamId).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Mysql) RegisterUser(
	ctx context.Context,
	username, passwordHashed string,
) (*user.Model, error) {
	result, err := r.db.ExecContext(ctx, queryRegisterUser, username, passwordHashed)

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	return &user.Model{
		Id:             int(id),
		Username:       username,
		PasswordHashed: passwordHashed,
	}, nil
}

func (r *Mysql) GetUser(
	ctx context.Context,
	username string,
) (*user.Model, error) {
	row := r.db.QueryRowContext(ctx, queryGetUser, username)

	var u user.Model

	if err := row.Scan(&u.Id, &u.Username, &u.PasswordHashed); err != nil {
		return nil, err
	}

	return &u, nil
}
