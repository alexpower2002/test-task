package repository

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"mkk-luna-test-task/internal/team/member"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Нужен запущенный Docker, чтобы тест работал.

const queryListNonMembers = `
SELECT t.id
FROM tasks t
LEFT JOIN team_members tm
    ON t.team_id = tm.team_id AND t.assignee_id = tm.user_id
WHERE tm.id IS NULL;
`

const queryTop3TaskCreators = `
WITH monthly_counts AS (
  SELECT
    teams.id      AS team_id,
    teams.name    AS team_name,
    users.id      AS user_id,
    users.username,
    COUNT(tasks.id) AS tasks_created
  FROM tasks
  JOIN users  ON users.id  = tasks.creator_id
  JOIN teams  ON teams.id  = tasks.team_id
  WHERE tasks.created_at >= DATE_FORMAT(DATE_SUB(CURDATE(), INTERVAL 1 MONTH), '%Y-%m-01')
    AND tasks.created_at <  DATE_FORMAT(CURDATE(), '%Y-%m-01')
  GROUP BY teams.id, teams.name, users.id, users.username
)
SELECT
  team_id,
  team_name,
  user_id,
  username,
  tasks_created
FROM (
  SELECT
    mc.*,
    ROW_NUMBER() OVER (PARTITION BY team_id ORDER BY tasks_created DESC) AS rn
  FROM monthly_counts mc
) t
WHERE rn <= 3
ORDER BY team_id, rn;
`

const queryDoneTasksLastWeek = `
SELECT
  teams.id            AS team_id,
  teams.name          AS team_name,
  COUNT(DISTINCT tm.user_id) AS members_count,
  COUNT(t.id)         AS done_tasks_last_week
FROM teams
LEFT JOIN team_members tm
  ON tm.team_id = teams.id
LEFT JOIN tasks t
  ON t.team_id = teams.id
  AND t.status = 'done'
  AND t.created_at >= NOW() - INTERVAL 7 DAY
GROUP BY teams.id, teams.name
ORDER BY teams.name;
`

func makeTestMySQLConfig(host string, port int) MySQLConfig {
	return MySQLConfig{
		User:     "testuser",
		Password: "testpass",
		DBName:   "testdb",
		Host:     host,
		Port:     port,
	}
}

func setupMySQLContainer(ctx context.Context, t *testing.T) (*sql.DB, MySQLConfig, func()) {
	mysqlUser := "testuser"
	mysqlPass := "testpass"
	mysqlDB := "testdb"

	req := testcontainers.ContainerRequest{
		Image: "mysql:8.0",
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "rootpass",
			"MYSQL_DATABASE":      mysqlDB,
			"MYSQL_USER":          mysqlUser,
			"MYSQL_PASSWORD":      mysqlPass,
		},
		WaitingFor:   wait.ForLog("port: 3306  MySQL Community Server - GPL"),
		ExposedPorts: []string{"3306/tcp"},
	}

	mysqlC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Fatalf("failed to start MySQL container: %v", err)
	}

	host, err := mysqlC.Host(ctx)

	if err != nil {
		_ = mysqlC.Terminate(ctx)

		t.Fatalf("failed to get container host: %v", err)
	}

	port, err := mysqlC.MappedPort(ctx, "3306")

	if err != nil {
		_ = mysqlC.Terminate(ctx)

		t.Fatalf("failed to get mapped port: %v", err)
	}

	cfg := makeTestMySQLConfig(host, port.Int())

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true", mysqlUser, mysqlPass, host, port.Int(), mysqlDB)

	var db *sql.DB
	var connectErr error

	for i := 0; i < 15; i++ {
		db, connectErr = sql.Open("mysql", dsn)

		if connectErr == nil && db.Ping() == nil {
			break
		}

		time.Sleep(2 * time.Second)
	}

	if connectErr != nil || db == nil || db.Ping() != nil {
		_ = mysqlC.Terminate(ctx)

		t.Fatalf("unable to connect to mysql: %v", connectErr)
	}

	if err := RunMigrations(cfg, "mysql_migrations"); err != nil {
		_ = db.Close()
		_ = mysqlC.Terminate(ctx)

		t.Fatalf("failed to run migrations: %v", err)
	}

	teardown := func() {
		_ = db.Close()
		_ = mysqlC.Terminate(ctx)
	}

	return db, cfg, teardown
}

func TestMysqlRepositoryIntegration(t *testing.T) {
	ctx := context.Background()

	db, _, teardown := setupMySQLContainer(ctx, t)
	defer teardown()

	repo := NewMysql(db)
	now := time.Now()
	utcBeginThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	utcBeginLastMonth := utcBeginThisMonth.AddDate(0, -1, 0)

	userLead, err := repo.RegisterUser(ctx, "teamlead", "пароль1")
	assert.NoError(t, err)
	assert.NotZero(t, userLead.Id)

	userDev1, err := repo.RegisterUser(ctx, "dev1", "пароль2")
	assert.NoError(t, err)
	assert.NotZero(t, userDev1.Id)

	userDev2, err := repo.RegisterUser(ctx, "dev2", "пароль3")
	assert.NoError(t, err)
	assert.NotZero(t, userDev2.Id)

	userQa, err := repo.RegisterUser(ctx, "qauser", "пароль4")
	assert.NoError(t, err)
	assert.NotZero(t, userQa.Id)

	teamBackend, err := repo.CreateTeamAndMakeUserItsOwner(ctx, "Бэкенд команда", userLead.Id)
	assert.NoError(t, err)
	assert.NotZero(t, teamBackend.Id)

	_, err = repo.InviteMember(ctx, userDev1.Id, teamBackend.Id, member.NormalRole)
	assert.NoError(t, err)
	_, err = repo.InviteMember(ctx, userDev2.Id, teamBackend.Id, member.NormalRole)
	assert.NoError(t, err)

	_, err = repo.InviteMember(ctx, userQa.Id, teamBackend.Id, member.NormalRole)
	assert.NoError(t, err)

	teamFrontend, err := repo.CreateTeamAndMakeUserItsOwner(ctx, "Фронтенд команда", userDev1.Id)
	assert.NoError(t, err)
	assert.NotZero(t, teamFrontend.Id)

	_, err = repo.InviteMember(ctx, userLead.Id, teamFrontend.Id, member.NormalRole)
	assert.NoError(t, err)

	_, err = repo.InviteMember(ctx, userQa.Id, teamFrontend.Id, member.NormalRole)
	assert.NoError(t, err)

	teamsDev1, err := repo.ListTeams(ctx, userDev1.Id)
	assert.NoError(t, err)
	assert.Len(t, teamsDev1, 2)

	var foundBackend, foundFrontend bool

	for _, tm := range teamsDev1 {
		if tm.Name == "Бэкенд команда" {
			foundBackend = true
		}

		if tm.Name == "Фронтенд команда" {
			foundFrontend = true
		}
	}
	assert.True(t, foundBackend)
	assert.True(t, foundFrontend)

	task1, err := repo.CreateTask(ctx, "todo", "Реализовать вход", "API эндпоинт для входа пользователей", userLead.Id, userDev1.Id, teamBackend.Id, utcBeginLastMonth.Add(24*time.Hour))
	assert.NoError(t, err)
	assert.NotZero(t, task1.Id)
	assert.Equal(t, "Реализовать вход", task1.Title)

	task2, err := repo.CreateTask(ctx, "todo", "Написать юнит-тесты", "Добавить тесты для jwt-аутентификации", userDev1.Id, userDev2.Id, teamBackend.Id, utcBeginLastMonth.Add(48*time.Hour))
	assert.NoError(t, err)

	task3, err := repo.CreateTask(ctx, "todo", "Провести тестирование функций", "Функциональное тестирование для QA", userDev1.Id, userQa.Id, teamBackend.Id, utcBeginLastMonth.Add(72*time.Hour))
	assert.NoError(t, err)

	task4Done, err := repo.CreateTask(ctx, "done", "Доработка тестов", "Последние правки", userDev2.Id, userDev2.Id, teamBackend.Id, now.AddDate(0, 0, -3))
	assert.NoError(t, err)
	assert.NotZero(t, task4Done.Id)

	_, err = repo.CreateTask(ctx, "todo", "Frontend Feature #1", "Task 1", userDev1.Id, userDev1.Id, teamFrontend.Id, utcBeginLastMonth.Add(24*time.Hour))
	assert.NoError(t, err)

	_, err = repo.CreateTask(ctx, "todo", "Frontend Feature #2", "Task 2", userDev1.Id, userLead.Id, teamFrontend.Id, utcBeginLastMonth.Add(48*time.Hour))
	assert.NoError(t, err)

	_, err = repo.CreateTask(ctx, "todo", "баг скинов на фронтэнде баннеров", "баннерная реклама не влазит в айфрейм 300x250", userLead.Id, userDev1.Id, teamFrontend.Id, utcBeginLastMonth.Add(90*time.Hour))
	assert.NoError(t, err)

	// QA в бэкенде, но может проверять и фронтэнд.
	_, err = repo.CreateTask(ctx, "done", "багфикс чего-то там", "Что-то там важное", userDev1.Id, userQa.Id, teamFrontend.Id, now.AddDate(0, 0, -6))
	assert.NoError(t, err)

	assert.NotEqual(t, task1.Id, task2.Id)
	assert.NotEqual(t, task2.Id, task3.Id)

	retrieved1, err := repo.GetTaskById(ctx, task1.Id)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved1)
	assert.Equal(t, userDev1.Id, retrieved1.AssigneeId)

	retrieved2, err := repo.GetTaskById(ctx, task2.Id)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved2)
	assert.Equal(t, userDev2.Id, retrieved2.AssigneeId)

	retrieved3, err := repo.GetTaskById(ctx, task3.Id)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved3)
	assert.Equal(t, userQa.Id, retrieved3.AssigneeId)

	comment1, err := repo.CreateTaskComment(ctx, userDev1.Id, task1.Id, "Начал работу над задачей.", now)
	assert.NoError(t, err)
	assert.NotZero(t, comment1.Id)

	_, err = repo.CreateTaskComment(ctx, userLead.Id, task1.Id, "какой прогресс по задаче", now)
	assert.NoError(t, err)

	comments, err := repo.ListTaskComments(ctx, task1.Id, 0, 10)
	assert.NoError(t, err)
	assert.Len(t, comments, 2)

	hist1, err := repo.CreateHistory(ctx, retrieved1, userDev1.Id, now)
	assert.NoError(t, err)
	assert.NotZero(t, hist1.Id)

	histories, err := repo.ListHistory(ctx, task1.Id, 0, 10)
	assert.NoError(t, err)
	assert.NotEmpty(t, histories)

	err = repo.EditTask(ctx, task2.Id, "in-progress", "Юнит-тесты", "Расширить покрытие тестами", userDev2.Id)
	assert.NoError(t, err)

	edited2, err := repo.GetTaskById(ctx, task2.Id)
	assert.NoError(t, err)
	assert.Equal(t, "Юнит-тесты — этап 2", edited2.Title)
	assert.Equal(t, "in-progress", edited2.Status)

	userOutsider, err := repo.RegisterUser(ctx, "outsider", "пароль123")
	assert.NoError(t, err)

	_, err = repo.CreateTask(ctx, "todo", "Ошибочная задача", "Задача назначена не-участнику команды", userDev1.Id, userOutsider.Id, teamFrontend.Id, now)
	assert.NoError(t, err)

	allTasks, err := repo.ListTasks(ctx, teamBackend.Id, "", 0, 0, 10)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(allTasks), 4)

	rows, err := db.QueryContext(ctx, "SELECT user_id, team_id FROM team_members WHERE team_id = ?", teamFrontend.Id)
	assert.NoError(t, err)

	var membersFront []struct {
		UserId int
		TeamId int
	}

	for rows.Next() {
		var userId, teamId int

		assert.NoError(t, rows.Scan(&userId, &teamId))

		membersFront = append(membersFront, struct {
			UserId int
			TeamId int
		}{
			UserId: userId,
			TeamId: teamId,
		})
	}

	rows.Close()

	assert.NoError(t, err)

	foundLead := false

	for _, m := range membersFront {
		if m.UserId == userLead.Id {
			foundLead = true
			break
		}
	}

	assert.True(t, foundLead)

	rows, err = db.QueryContext(ctx, queryListNonMembers)

	assert.NoError(t, err)

	var outlierTasksCount int

	for rows.Next() {
		var id int64

		err := rows.Scan(&id)

		if err != nil {
			break
		}

		outlierTasksCount++
	}

	rows.Close()

	assert.Equal(t, 1, outlierTasksCount)

	rows, err = db.QueryContext(ctx, queryTop3TaskCreators)

	assert.NoError(t, err)

	type TopCreator struct {
		TeamId       int64
		TeamName     string
		UserId       int64
		Username     string
		TasksCreated int64
	}

	var creators []TopCreator

	for rows.Next() {
		var c TopCreator

		err := rows.Scan(&c.TeamId, &c.TeamName, &c.UserId, &c.Username, &c.TasksCreated)

		assert.NoError(t, err)

		creators = append(creators, c)
	}

	rows.Close()

	assert.NotEmpty(t, creators)

	var hasDev1, hasLead bool

	for _, x := range creators {
		if x.Username == "dev1" {
			hasDev1 = true
		}

		if x.Username == "teamlead" {
			hasLead = true
		}
	}
	assert.True(t, hasDev1)
	assert.True(t, hasLead)

	rows, err = db.QueryContext(ctx, queryDoneTasksLastWeek)

	assert.NoError(t, err)

	var foundBack, foundFront bool

	for rows.Next() {
		var teamId int64
		var teamName string
		var membersCount int64
		var doneTasksCount int64

		err := rows.Scan(&teamId, &teamName, &membersCount, &doneTasksCount)

		assert.NoError(t, err)

		if teamName == "Бэкенд команда" {
			assert.GreaterOrEqual(t, doneTasksCount, int64(1))

			foundBack = true
		}
		if teamName == "Фронтенд команда" {
			assert.GreaterOrEqual(t, doneTasksCount, int64(1))

			foundFront = true
		}
	}

	rows.Close()

	assert.True(t, foundBack)
	assert.True(t, foundFront)
}
