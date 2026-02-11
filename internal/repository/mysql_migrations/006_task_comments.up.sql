CREATE TABLE IF NOT EXISTS task_comments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME NOT NULL,
    commenter_id INT NOT NULL,
    task_id INT NOT NULL,
    text TEXT NOT NULL,
    FOREIGN KEY (commenter_id) REFERENCES users(id),
    FOREIGN KEY (task_id) REFERENCES tasks(id)
);
