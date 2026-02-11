Приложение запускается через Docker compose:

```sh
docker compose up
```

После запуска приложения можно выполнять следующие запросы к API — все эндпоинты работают согласно спецификации:

### 1. Аутентификация (JWT)

#### Зарегистрировать нового пользователя
```sh
curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass"}'
```

#### Войти и получить JWT-токен
```sh
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass"}'
```

### 2. Управление командами

#### Создать команду (стать owner)
```sh
curl -X POST http://localhost:8080/api/v1/teams \
  -H "Content-Type: application/json" \
  -H "jwt-token: <token>" \
  -d '{"name": "team_name"}'
```

#### Список команд, где пользователь состоит
```sh
curl -X GET http://localhost:8080/api/v1/teams \
  -H "jwt-token: <token>"
```

#### Пригласить пользователя в команду (только owner/admin)
```sh
curl -X POST http://localhost:8080/api/v1/teams/{id}/invite \
  -H "Content-Type: application/json" \
  -H "jwt-token: <token>" \
  -d '{"user_id": <user_id>, "role": "member"}'
```

### 3. Управление задачами

#### Создать задачу (только член команды)
```sh
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "jwt-token: <token>" \
  -d '{"status": "todo", "title": "Задача", "description": "Описание", "assignee_id": 5, "team_id": 1}'
```

#### Фильтрация с пагинацией
```sh
curl -X GET "http://localhost:8080/api/v1/tasks?team_id=1&status=todo&assignee_id=5" \
  -H "jwt-token: <token>"
```

#### Обновить задачу
```sh
curl -X PUT http://localhost:8080/api/v1/tasks/{id} \
  -H "Content-Type: application/json" \
  -H "jwt-token: <token>" \
  -d '{"status": "in_progress", "title": "Обновленный заголовок", "description": "Новое описание", "assignee_id": 6}'
```

#### История изменений задачи (с пагинацией)
```sh
curl -X GET http://localhost:8080/api/v1/tasks/{id}/history \
  -H "jwt-token: <token>"
```

### 4. Комментарии к задачам

#### Добавить комментарий к задаче
```sh
curl -X POST http://localhost:8080/api/v1/tasks/{id}/comments \
  -H "Content-Type: application/json" \
  -H "jwt-token: <token>" \
  -d '{"text": "Ваш комментарий"}'
```

#### Получить список комментариев к задаче (с пагинацией)
```sh
curl -X GET "http://localhost:8080/api/v1/tasks/{id}/comments" \
  -H "jwt-token: <token>"
```

---

#### ПРИМЕЧАНИЕ

Так как не было предоставлено конкретных use cases (сценариев использования), невозможно точно определить, какую именно систему ожидал автор тестового задания, кроме того, что можно было предположить по структуре базы данных, требованиям к API и т.д. Также в техническом задании не были указаны полные контракты — например, не было подробно описано, как должен передаваться JWT, какие ответы ожидаются и т.п. Поэтому я реализовал всё исходя из собственных предположений. Не было представлено нефункциональных требований - насколько важна консистентность, какую нагрузку ожидать, так что это я тоже предположил сам.
