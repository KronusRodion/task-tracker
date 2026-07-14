# Task-tracker
This is repository with task-group tracking application.
Project uses MySQL like main persistence data storage and Redis like cache. 

# Tests
All usecases was covered with unit-tests in code
The main usecases (ex. Create team, invite user, create/update task), repositories + handlers was covered by integration tests

To run tests:
```go
go test ./...
```

# Environments
Пример .env файла в корне проекта
```bash
# .env - переменные окружения для docker-compose

# Приложение
APP_PORT=8080
# Путь до конфига внутри приложения
CONF_PATH=/app/config.yaml

# База данных MySQL
DB_HOST=mysql
DB_PORT=3306
DB_USER=root
DB_PASSWORD=password
DB_NAME=task_tracker
DB_SSL_MODE=disable
DB_TIMEOUT=5s
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=5

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Аутентификация
AUTH_ACCESS_SECRET=test-access-secret-key-1234567890
AUTH_REFRESH_SECRET=test-refresh-secret-key-0987654321
AUTH_ISSUER=task-tracker
AUTH_ACCESS_TTL=15m
AUTH_REFRESH_TTL=720h
```

# Metrics
The application collects runtime, endpoints and network stats and provide /metrics endpoint in Prometheus format. 

### SQL Queries
- **✅ Complex SQL Queries with JOINs, Aggregations, and Window Functions**
  - **a) Query with JOIN 3+ tables + aggregation**
    *"Получить для каждой команды: название, количество участников, количество задач в статусе done за последние 7 дней"*
    **Status:** Implemented in the database layer with optimized queries and indexed tables.

  - **b) Recursive Query or Subquery with Window Function**
    *"Получить топ-3 пользователя по количеству созданных задач в каждой команде за месяц"*
    **Status:** Implemented using window functions in SQL queries.

  - **c) Query with condition on related tables**
    *"Найти задачи, где assignee не является членом команды этой задачи" (валидация целостности)*
    **Status:** Implemented in the `FindInvalidAssigneeTasks` method in `usecase.go` and tested.

---

### Optimization
- **✅ Database Indexes**
  - **a) Indexes for important SQL queries**
    **Status:** Indexes are added to frequently queried columns (e.g., `team_id`, `status`, `created_at`) to optimize performance.

- **✅ Redis Caching**
  - **b) Caching task lists for teams (TTL 5 min)**
    **Status:** Implemented in `usecase.go` with Redis caching for task lists, invalidated on task creation/update.

---

### Protection
- **✅ Rate Limiting**
  - **a) Rate limiting on endpoints (100 requests/min per user)**
    **Status:** Implemented in `middleware/rate_limiter.go` and applied to all API endpoints.

- **✅ Circuit Breaker for Notification Service**
  - **b) Mock notification service with Circuit Breaker**
    **Status:** Implemented using `gobreaker` library in `usecase.go` and `module.go`. The `NotificationService` is wrapped with a circuit breaker to handle failures gracefully.

---

### Additional API Features
- **✅ Task Comments API**
  - **a) Create a comment for a task**
    **Status:** Not yet implemented (planned for future updates).

  - **b) Delete your own comment for a task**
    **Status:** Not yet implemented (planned for future updates).

  - **c) List comments for a task**
    **Status:** Not yet implemented (planned for future updates).

  - **d) Create `task_comments` table**
    **Status:** Not yet implemented (planned for future updates).

---