-- +goose Up

CREATE TABLE users (
    id BINARY(16) NOT NULL,

    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,

    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,

    PRIMARY KEY (id),
    UNIQUE KEY uk_users_email (email)
);

CREATE TABLE teams (
    id BINARY(16) NOT NULL,

    name VARCHAR(255) NOT NULL,

    created_by BINARY(16) NOT NULL,

    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,

    PRIMARY KEY (id),

    KEY idx_teams_created_by (created_by),
    KEY idx_teams_name (name),

    CONSTRAINT fk_teams_created_by
        FOREIGN KEY (created_by)
        REFERENCES users(id)
        ON DELETE RESTRICT
        ON UPDATE CASCADE
);

CREATE TABLE team_members (
    team_id BINARY(16) NOT NULL,
    user_id BINARY(16) NOT NULL,

    role ENUM(
        'owner',
        'admin',
        'member'
    ) NOT NULL,

    joined_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,

    PRIMARY KEY (team_id, user_id),

    KEY idx_team_members_user_team (user_id, team_id),

    CONSTRAINT fk_team_members_team
        FOREIGN KEY (team_id)
        REFERENCES teams(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_team_members_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE tasks (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    team_id BINARY(16) NOT NULL,

    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,

    status ENUM(
        'todo',
        'in_progress',
        'done'
    ) NOT NULL,

    priority ENUM(
        'low',
        'medium',
        'high'
    ) NOT NULL,

    created_by BINARY(16) NOT NULL,
    assignee_id BINARY(16) NULL,

    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,

    PRIMARY KEY (id),

    KEY idx_tasks_team (team_id),
    KEY idx_tasks_assignee (assignee_id),
    KEY idx_tasks_created_by (created_by),

    KEY idx_tasks_filter (
        team_id,
        status,
        assignee_id,
        created_at
    ),

    KEY idx_tasks_team_created (
        team_id,
        created_at
    ),

    CONSTRAINT fk_tasks_team
        FOREIGN KEY (team_id)
        REFERENCES teams(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_tasks_created_by
        FOREIGN KEY (created_by)
        REFERENCES users(id)
        ON DELETE RESTRICT,

    CONSTRAINT fk_tasks_assignee
        FOREIGN KEY (assignee_id)
        REFERENCES users(id)
        ON DELETE SET NULL
);

CREATE TABLE task_history (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    task_id BIGINT UNSIGNED NOT NULL,

    action ENUM(
        'created',
        'updated',
        'assigned',
        'status_changed',
        'comment_added'
    ) NOT NULL,

    field VARCHAR(64) NULL,

    old_value TEXT NULL,
    new_value TEXT NULL,

    changed_by BINARY(16) NOT NULL,

    created_at DATETIME(6) NOT NULL,

    PRIMARY KEY (id),

    KEY idx_task_history_task_created (
        task_id,
        created_at
    ),

    CONSTRAINT fk_history_task
        FOREIGN KEY (task_id)
        REFERENCES tasks(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_history_user
        FOREIGN KEY (changed_by)
        REFERENCES users(id)
        ON DELETE RESTRICT
);

CREATE TABLE task_comments (
    id BINARY(16) NOT NULL,

    task_id BIGINT UNSIGNED NOT NULL,

    user_id BINARY(16) NOT NULL,

    content VARCHAR(1000) NOT NULL,

    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,

    PRIMARY KEY (id),

    KEY idx_comments_task_created (
        task_id,
        created_at
    ),

    KEY idx_comments_user (user_id),

    CONSTRAINT fk_comments_task
        FOREIGN KEY (task_id)
        REFERENCES tasks(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_comments_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- +goose Down
DROP FUNCTION IF EXISTS BIN_TO_UUID;
DROP FUNCTION IF EXISTS UUID_TO_BIN;

DROP TABLE IF EXISTS task_comments;
DROP TABLE IF EXISTS task_history;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users;