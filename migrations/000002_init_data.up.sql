
-- +goose Up
-- Вставляем тестовые данные с фиксированными UUID
INSERT INTO users (id, email, password, full_name, created_at, updated_at) VALUES
(UUID_TO_BIN('550e8400-e29b-41d4-a716-446655440000'), 'user1@example.com', '$2a$10$examplepasswordhash1', 'User One', NOW(), NOW()),
(UUID_TO_BIN('550e8400-e29b-41d4-a716-446655440001'), 'user2@example.com', '$2a$10$examplepasswordhash2', 'User Two', NOW(), NOW()),
(UUID_TO_BIN('550e8400-e29b-41d4-a716-446655440002'), 'user3@example.com', '$2a$10$examplepasswordhash3', 'User Three', NOW(), NOW());

-- Вставляем тестовые команды с фиксированными UUID
INSERT INTO teams (id, name, created_by, created_at, updated_at) VALUES
(UUID_TO_BIN('550e8400-e29b-41d4-a716-446655440010'), 'Team Alpha', (SELECT id FROM users WHERE email = 'user1@example.com'), NOW(), NOW()),
(UUID_TO_BIN('550e8400-e29b-41d4-a716-446655440011'), 'Team Beta', (SELECT id FROM users WHERE email = 'user2@example.com'), NOW(), NOW());

-- Вставляем участников команд
INSERT INTO team_members (team_id, user_id, role, joined_at, updated_at) VALUES
((SELECT id FROM teams WHERE name = 'Team Alpha'), (SELECT id FROM users WHERE email = 'user1@example.com'), 'owner', NOW(), NOW()),
((SELECT id FROM teams WHERE name = 'Team Alpha'), (SELECT id FROM users WHERE email = 'user2@example.com'), 'member', NOW(), NOW()),
((SELECT id FROM teams WHERE name = 'Team Beta'), (SELECT id FROM users WHERE email = 'user2@example.com'), 'owner', NOW(), NOW()),
((SELECT id FROM teams WHERE name = 'Team Beta'), (SELECT id FROM users WHERE email = 'user3@example.com'), 'member', NOW(), NOW());

-- Вставляем тестовые задачи с фиксированными ID
INSERT INTO tasks (id, team_id, title, description, status, priority, created_by, assignee_id, created_at, updated_at) VALUES
(100, (SELECT id FROM teams WHERE name = 'Team Alpha'), 'Task One', 'Description for Task One', 'done', 'medium', (SELECT id FROM users WHERE email = 'user1@example.com'), (SELECT id FROM users WHERE email = 'user2@example.com'), NOW(), NOW()),
(101, (SELECT id FROM teams WHERE name = 'Team Alpha'), 'Task Two', 'Description for Task Two', 'in_progress', 'high', (SELECT id FROM users WHERE email = 'user2@example.com'), (SELECT id FROM users WHERE email = 'user1@example.com'), NOW(), NOW()),
(102, (SELECT id FROM teams WHERE name = 'Team Beta'), 'Task Three', 'Description for Task Three', 'todo', 'low', (SELECT id FROM users WHERE email = 'user2@example.com'), (SELECT id FROM users WHERE email = 'user3@example.com'), NOW(), NOW());