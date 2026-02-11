CREATE INDEX idx_tasks_team_id ON tasks(team_id);
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX idx_tasks_team_id_id ON tasks(team_id, id);

CREATE INDEX idx_task_comments_task_id_id ON task_comments(task_id, id);

CREATE INDEX idx_task_history_task_id_id ON task_history(task_id, id);

CREATE INDEX idx_team_members_user_id ON team_members(user_id);
CREATE INDEX idx_team_members_team_id ON team_members(team_id);

CREATE INDEX idx_users_username ON users(username);
