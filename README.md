# SOP Contest Backend

Примрено за 16 часов рабочего времени (между коммитами ушло много времени на сон) и за 2 недели предварительной подготовки и изучения языка Go на базе материаолов курса с М. А. Бубновой, было реализовано тестовое задание.
Четсно говоря, я во многом пользовался llm, для поиска инфы, для объяснения происходящего, ну и конечно, для изучения данного языка, но за короткий период я покрыл достатчно больой объем информации и останавливаться на достигнутом не собираюсь.
По итогу конкурса был собран REST-сервис для платформы отзывов о преподавателях: пользователи могут регистрироваться, авторизовываться, искать преподавателей, оставлять отзывы и смотреть собственные отзывы. Администратор может просматривать отзывы пользователей и блокировать подозрительные аккаунты.

## 1. Что нужно на машине

Для запуска через Docker:

- Docker
- Docker Compose

Для локальной разработки без Docker:

- Go 1.23+
- PostgreSQL
- Make не обязателен
- `curl` для проверки API
- `jq` опционально, только для удобного извлечения токенов из JSON-ответов



## 2. Быстрый запуск

### 2.1. Клонировать репозиторий

```bash
git clone https://github.com/plastm0oo/sop_contest.git
cd MIEM_contest
```

### 2.2. Создать `.env`

```bash
cp .env.example .env
```

Перед запуском можно оставить значения по умолчанию. Для production-подобного запуска нужно заменить `JWT_SECRET` на сгенерированный код:

```bash
openssl rand -hex 32
```

### 2.3. Запустить приложение

```bash
docker compose up --build
```

После старта Docker Compose поднимает PostgreSQL, ждёт его healthcheck и запускает Go-приложение. На старте приложение подключается к базе и автоматически применяет SQL-миграции из директории `migrations/`.

### 2.4. Первый успешный curl

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ:

```json
{"status":"ok"}
```

Проверка списка преподавателей:

```bash
curl http://localhost:8080/api/teachers
```

Пример ответа:

```json
{
  "items": [
    {
      "id": 1,
      "full_name": "Иванов Иван Иванович",
      "faculty": "Факультет компьютерных наук",
      "reviews_count": 0,
      "avg_rating": 0
    }
  ],
  "total": 3,
  "limit": 20,
  "offset": 0
}
```



## 3. Переменные окружения

Пример `.env.example`:

```env
PORT=8080

DATABASE_URL=postgres://postgres:postgres@postgres:5432/miem_contest?sslmode=disable
JWT_SECRET=01234567890123456789012345678901

ADMIN_EMAIL=
CORS_ALLOWED_ORIGIN=*

ACCESS_TOKEN_DURATION=15m
REFRESH_TOKEN_DURATION=168h

BCRYPT_COST=10

RATE_LIMIT_ATTEMPTS=5
RATE_LIMIT_WINDOW=1m
```


#### `DATABASE_URL`

Строка подключения к PostgreSQL.

Для Docker Compose используется host `postgres`, потому что это имя сервиса внутри docker-сети:

```env
DATABASE_URL=postgres://postgres:postgres@postgres:5432/miem_contest?sslmode=disable
```

#### `JWT_SECRET`

Секрет для подписи access-токенов. Минимальная длина — 32 символа. Если значение короче, приложение завершится на старте с ошибкой.

генерируется командой:

```bash
openssl rand -hex 32
```

#### `PORT`

Порт HTTP-сервера. По умолчанию используется `8080`.

#### `ADMIN_EMAIL`

Email первого администратора. Если переменная задана, пользователь с таким email при регистрации получит роль `admin`. Если переменная пустая, все новые пользователи получают роль `user`.

#### `CORS_ALLOWED_ORIGIN`

Источники, разрешённые CORS middleware. В проде указываем конкретные домены, но для локальной разработки:

```env
CORS_ALLOWED_ORIGIN=*
```

Например для фронтенда на Vite можно указать:

```env
CORS_ALLOWED_ORIGIN=http://localhost:5173
```

#### `ACCESS_TOKEN_DURATION`

Время жизни access-токена (по умолчанию `15m`).

#### `REFRESH_TOKEN_DURATION`

Время жизни refresh-токена (по умолчанию `168h`, то есть 7 дней).

#### `BCRYPT_COST`

Стоимость bcrypt-хэширования пароля. По умолчанию `10`.

#### `RATE_LIMIT_ATTEMPTS` и `RATE_LIMIT_WINDOW`

Ограничение попыток для auth endpoints:

- `POST /api/auth/register`
- `POST /api/auth/login`

Например:

```env
RATE_LIMIT_ATTEMPTS=5
RATE_LIMIT_WINDOW=1m
```

не более 5 попыток в минуту с одного IP на каждый auth endpoint.



## 4. API и примеры запросов

Все ответы возвращаются в JSON, кроме endpoint с ответом `204 No Content`.

Формат ошибки:

```json
{"error":"понятный текст ошибки"}
```

Для ошибок валидации:

```json
{
  "error": "validation failed",
  "details": {
    "field": "reason"
  }
}
```



### 4.1. Healthcheck

#### `GET /health`

```bash
curl http://localhost:8080/health
```

Ответ:

```json
{"status":"ok"}
```



### 4.2. Регистрация

#### `POST /api/auth/register`

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@hse.ru","password":"Secret123"}'
```

Успешный ответ:

```json
{
  "user": {
    "id": 1,
    "email": "user@hse.ru",
    "role": "user"
  },
  "access_token": "eyJhbGci...",
  "refresh_token": "long-random-string",
  "access_expires_in": 900,
  "refresh_expires_in": 604800
}
```

Правила пароля прописанные в usecase:

- минимум 8 символов;
- хотя бы одна буква;
- хотя бы одна цифра.

Пример ошибки валидации:

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"bad","password":"123"}'
```

```json
{
  "error": "validation failed",
  "details": {
    "email": "must be a valid email",
    "password": "минимум 8 символов, одна буква и одна цифра"
  }
}
```

Если пользователь уже существует:

```json
{"error":"пользователь с таким email уже существует"}
```



### 4.3. Логин

#### `POST /api/auth/login`

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@hse.ru","password":"Secret123"}'
```

Успешный ответ имеет ту же структуру, что и регистрация:

```json
{
  "user": {
    "id": 1,
    "email": "user@hse.ru",
    "role": "user"
  },
  "access_token": "eyJhbGci...",
  "refresh_token": "long-random-string",
  "access_expires_in": 900,
  "refresh_expires_in": 604800
}
```

При неверном email или пароле возвращается одинаковый ответ:

```json
{"error":"неверный email или пароль"}
```

Это делается, чтобы не раскрывать, какие email зарегистрированы в системе.



### 4.4. Refresh token rotation

#### `POST /api/auth/refresh`

Каждый refresh-запрос отзывает старый refresh-токен и выдаёт новую пару access/refresh-токенов.

```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

Успешный ответ:

```json
{
  "user": {
    "id": 1,
    "email": "user@hse.ru",
    "role": "user"
  },
  "access_token": "new-access-token",
  "refresh_token": "new-refresh-token",
  "access_expires_in": 900,
  "refresh_expires_in": 604800
}
```

Повторное использование старого refresh-токена вернёт:

```json
{"error":"недействительный refresh token"}
```



### 4.5. Logout

#### `POST /api/auth/logout`

Отзывает конкретный refresh-токен.

```bash
curl -i -X POST http://localhost:8080/api/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"<refresh_token>"}'
```

Успешный ответ:

```text
HTTP/1.1 204 No Content
```

Access-токен при logout не отзывается: он живёт короткое время и перестаёт работать после истечения срока действия.



### 4.6. Получить список преподавателей

#### `GET /api/teachers`

Публичный endpoint, авторизация для которого не требуется.

| Параметр | Тип | Описание | Значение по умолчанию |
|---|---:|---|---:|
| `q` | string | Поиск по `full_name` через `ILIKE` | пусто |
| `faculty` | string | Фильтр по факультету, точное совпадение | пусто |
| `limit` | int | Количество записей, от 1 до 100 | 20 |
| `offset` | int | Смещение, от 0 | 0 |

Пример:

```bash
curl "http://localhost:8080/api/teachers?limit=10&offset=0"
```

Поиск:

```bash
curl "http://localhost:8080/api/teachers?q=Иванов"
```

Фильтр по факультету:

```bash
curl "http://localhost:8080/api/teachers?faculty=Факультет%20компьютерных%20наук"
```

Пример ответа:

```json
{
  "items": [
    {
      "id": 1,
      "full_name": "Иванов Иван Иванович",
      "faculty": "Факультет компьютерных наук",
      "reviews_count": 1,
      "avg_rating": 5
    }
  ],
  "total": 3,
  "limit": 20,
  "offset": 0
}
```



### 4.7. Получить карточку преподавателя

#### `GET /api/teachers/{id}`

Публичный endpoint, опять-таки авторизация не требуется.

```bash
curl http://localhost:8080/api/teachers/1
```

Пример ответа:

```json
{
  "id": 1,
  "full_name": "Иванов Иван Иванович",
  "faculty": "Факультет компьютерных наук",
  "email": "ivanov@hse.ru",
  "reviews_count": 1,
  "avg_rating": 5,
  "rating_distribution": {
    "1": 0,
    "2": 0,
    "3": 0,
    "4": 0,
    "5": 1
  }
}
```

Если преподаватель не найден:

```json
{"error":"преподаватель не найден"}
```



### 4.8. Создать отзыв

#### `POST /api/feedbacks`

Уже защищённый endpoint, который требует access-токен:

```http
Authorization: Bearer <access_token>
```

Пример получения токена через `jq`:

```bash
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@hse.ru","password":"Secret123"}')

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
```

Создание отзыва:

```bash
curl -X POST http://localhost:8080/api/feedbacks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{"teacher_id":1,"rating":5,"comment":"Отличные лекции, все понятно"}'
```

Успешный ответ:

```json
{
  "id": 1,
  "teacher_id": 1,
  "rating": 5,
  "comment": "Отличные лекции, все понятно",
  "created_at": "2026-05-08T10:30:00Z"
}
```

При повторном отзыве того же пользователя на того же преподавателя:

```json
{"error":"вы уже оставляли отзыв на этого преподавателя"}
```

Или ошибка валидации:

```json
{
  "error": "validation failed",
  "details": {
    "rating": "must be between 1 and 5",
    "comment": "length must be between 10 and 2000 characters"
  }
}
```



### 4.9. Получить свои отзывы

#### `GET /api/feedbacks/me`

Защищённый endpoint.

```bash
curl http://localhost:8080/api/feedbacks/me \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Пример ответа:

```json
{
  "items": [
    {
      "id": 1,
      "teacher_id": 1,
      "teacher_name": "Иванов Иван Иванович",
      "rating": 5,
      "comment": "Отличные лекции, всё понятно",
      "created_at": "2026-05-08T10:30:00Z"
    }
  ]
}
```



### 4.10. Admin: получить список отзывов

#### `GET /api/admin/feedbacks`

Только для пользователя с ролью `admin`.

Чтобы создать admin-пользователя, укажите в `.env` ADMIN_EMAIL=:

```env
ADMIN_EMAIL=admin@hse.ru
```

Затем зарегистрируйте пользователя с этим email:

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@hse.ru","password":"Admin123"}'
```

Получите admin access token:

```bash
ADMIN_LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@hse.ru","password":"Admin123"}')

ADMIN_ACCESS_TOKEN=$(echo "$ADMIN_LOGIN_RESPONSE" | jq -r '.access_token')
```

Запрос:

```bash
curl http://localhost:8080/api/admin/feedbacks \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"
```

Пример фильтра по пользователю:

```bash
curl "http://localhost:8080/api/admin/feedbacks?user_id=1" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"
```

Фильтра по преподавателю:

```bash
curl "http://localhost:8080/api/admin/feedbacks?teacher_id=1" \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"
```

Пример ответа:

```json
{
  "items": [
    {
      "id": 1,
      "teacher_id": 1,
      "teacher_name": "Иванов Иван Иванович",
      "user_id": 1,
      "user_email": "user@hse.ru",
      "rating": 5,
      "comment": "Отличные лекции, всё понятно",
      "created_at": "2026-05-08T10:30:00Z"
    }
  ],
  "total": 1,
  "limit": 20,
  "offset": 0
}
```

Если пользователь не admin:

```json
{"error":"недостаточно прав"}
```



### 4.11. Admin: заблокировать пользователя

#### `POST /api/admin/users/{id}/block`

Только для пользователя с ролью `admin`.

```bash
curl -i -X POST http://localhost:8080/api/admin/users/1/block \
  -H "Authorization: Bearer $ADMIN_ACCESS_TOKEN"
```

Успешный ответ:

```text
HTTP/1.1 204 No Content
```

При блокировке пользователя:

- `users.is_blocked` устанавливается в `true`;
- все активные refresh-токены пользователя отзываются;
- операция выполняется в одной транзакции.

После блокировки логин пользователя вернёт:

```json
{"error":"аккаунт заблокирован"}
```



## 5. Миграции

Миграции лежат в директории:

```text
migrations
 └── 001_table.sql
```

На старте приложения запускается собственный migration runner из пакета:

```text
internal/migrator
```

Runner делает следующее:

1. Создаёт служебную таблицу `schema_migrations`, если её ещё нет.
2. Читает `.sql` файлы из директории `migrations/`.
3. Сортирует файлы по имени.
4. Проверяет, применялась ли миграция ранее.
5. Выполняет новую миграцию внутри транзакции.
6. Записывает имя применённой миграции в `schema_migrations`.

Проверить применённые миграции можем так:

```bash
docker compose exec postgres psql -U postgres -d miem_contest
```

```sql
SELECT * FROM schema_migrations;
```

Пример результата:

```text
 version      | applied_at
--------------+-------------------------------
 001_init.sql | 2026-05-08 10:30:00+00
```

Подход с `/docker-entrypoint-initdb.d` не использую, потому что он работает только при первом создании volume PostgreSQL. Для задания же удобнее миграции на старте приложения. Потому что при разработке мы часто меняем структуру БД, и хочется, чтобы миграции применялись автоматически при каждом запуске, а не только при первом.



## 6. Структура проекта

Текущая структура:

```text
MIEM_contest
├── Dockerfile
├── cmd
│   └── server
│       └── main.go
├── docker-compose.yml
├── go.mod
├── go.sum
├── migrations
│   └── 001_init.sql
├── internal
│   ├── auth
│   │   ├── auth.go
│   │   └── auth_test.go
│   ├── config
│   │   └── config.go
│   ├── middleware
│   │   ├── middleware.go
│   │   └── middleware_test.go
│   ├── migrator
│   │   └── migrator.go
│   └── service
│       ├── auth_context.go
│       ├── delivery
│       │   └── http
│       │       ├── handler.go
│       │       └── middleware.go
│       ├── errors.go
│       ├── interfaces.go
│       ├── models.go
│       ├── repository
│       │   └── repository.go
│       └── usecase
│           ├── usecase.go
│           └── usecase_test.go
├── .env.example
├── .gitignore
└── README.md
```



### `cmd/server/main.go`

Точка входа приложения, основные задачи которой:

- загрузить конфигурацию;
- подключиться к PostgreSQL;
- применить миграции;
- собрать зависимости вручную;
- создать repository, usecase и handler;
- зарегистрировать маршруты;
- подключить middleware;
- запустить HTTP-сервер;
- выполнить graceful shutdown при `SIGINT` или `SIGTERM`.

Dependency Injection сделан вручную, без DI-фреймворков. Для такого размера проекта это проще и прозрачнее.



### `internal/config`

Пакет отвечает за чтение и валидацию переменных окружения.

В нём проверяется:

- что `DATABASE_URL` задан;
- что `JWT_SECRET` не короче 32 символов;
- что duration-переменные имеют корректный формат;
- что числовые переменные действительно являются числами.

Приложение падает на старте, если критичная конфигурация некорректна, чтобы не запускаться в неопределённом состоянии.



### `internal/auth`

Пакет с низкоуровневыми auth-функциями:

- bcrypt-хэширование паролей;
- проверка пароля через bcrypt;
- генерация JWT access-токена;
- проверка JWT access-токена;
- генерация refresh-токена;
- SHA-256-хэширование refresh-токена перед записью в БД.

Access-токен содержит:

- `user_id`;
- `email`;
- `role`;
- `exp`;
- `iat`.

Пароли и refresh-токены не хранятся в открытом виде.



### `internal/migrator`

Собственный минимальный runner миграций. Не поддерживает rollback, потому что в задании откаты не являются обязательными. При этом он решает главную задачу: автоматически применяет новые SQL-файлы при старте приложения и фиксирует применённые версии.



### `internal/middleware`

Глобальные middleware HTTP-сервера:

- CORS;
- rate limiting для `/api/auth/register` и `/api/auth/login`.

Rate limiter хранит счётчики в памяти, Redis был бы излишним усложнением.

IP определяется так:

1. если есть `X-Real-IP`, используется он;
2. иначе IP берётся из `r.RemoteAddr`;
3. порт отрезается через `net.SplitHostPort`, чтобы разные TCP-порты не считались разными клиентами.



### `internal/service`

Основной доменный пакет приложения, который содержит:

- модели запросов и ответов;
- интерфейсы слоёв;
- доменные ошибки;
- context helpers для авторизованного пользователя.


### `internal/service/delivery/http`

HTTP delivery-слой.

- принять HTTP-запрос;
- проверить HTTP-метод;
- распарсить JSON или query-параметры;
- достать авторизованного пользователя из context;
- вызвать usecase;
- преобразовать доменные ошибки в HTTP-статусы;
- вернуть JSON-ответ.

В этом же пакете лежат route-level middleware:

- `authMiddleware` — проверяет `Authorization: Bearer <access_token>`;
- `adminMiddleware` — проверяет, что роль пользователя равна `admin`.

Эти middleware находятся в delivery-слое, потому что они завязаны на handler и HTTP-ответы.



### `internal/service/usecase`

Слой логики приложения.

- нормализовать email;
- валидировать входные данные;
- определить роль пользователя при регистрации;
- хэшировать пароль;
- проверять пароль при логине;
- проверять блокировку пользователя;
- выдавать access/refresh-токены;
- выполнять refresh-token rotation;
- валидировать отзывы;
- вызывать repository через интерфейс.

Usecase не знает деталей SQL-запросов и не зависит от конкретной реализации базы данных, работая через интерфейс repository.



### `internal/service/repository`

Слой работы с PostgreSQL через `database/sql`.

- выполнять SQL-запросы;
- создавать и искать пользователей;
- сохранять и отзывать refresh-токены;
- получать преподавателей и статистику;
- создавать отзывы;
- получать отзывы пользователя;
- выполнять admin-запросы;
- блокировать пользователя и отзывать его refresh-токены в одной транзакции.

Все пользовательские значения передаются в SQL через параметры (`$1`, `$2` и т.д.), чтобы защититься от SQL-инъекций, а не через string concatenation.



## 7. База данных

В первой миграции создаются таблицы:

- `users`
- `teachers`
- `feedbacks`
- `refresh_tokens`
- `schema_migrations`

### `users`

Пользователи, которые могут быть обычными пользователями (`user`) или администраторами (`admin`).

- `id`;
- `email`;
- `password_hash`;
- `role`;
- `is_blocked`;
- `created_at`;
- `updated_at`.

Email хранится в нижнем регистре. Пароль хранится только в виде bcrypt-хэша.

### `teachers`

Преподаватели.

- `id`;
- `full_name`;
- `faculty`;
- `email`.

### `feedbacks`

Отзывы, где ключевые ограничения:

- `rating` от 1 до 5;
- длина `comment` от 10 до 2000 символов;
- `UNIQUE (teacher_id, user_id)` — один отзыв от одного пользователя на одного преподавателя.

### `refresh_tokens`

Refresh-токены. В таблице хранится не сам токен, а SHA-256-хэш токена:

- `token_hash`;
- `expires_at`;
- `revoked_at`.

Это снижает риск при утечке базы: plain refresh-токены не лежат в БД.



## 8. Безопасность

### bcrypt для паролей

Пароли хэшируются через bcrypt. В БД никогда не сохраняется plain password, только его хэш. При логине пароль проверяется через `bcrypt.CompareHashAndPassword`.

### Refresh-токены хранятся в виде SHA-256-хэша

Клиент получает refresh-токен только один раз в ответе API. В базе хранится только его хэш (через SHA-256). При использовании refresh-токена клиент отправляет его, сервер хэширует и сравнивает с хэшами в БД. Это снижает риск при утечке базы данных, так как plain refresh-токены не сохраняются.

### Защита от user enumeration

При логине не различаются ошибки “email не найден” и “пароль неверный” — в обоих случаях возвращается одинаковое сообщение об ошибке, чтобы предотвратить возможность нежелатеьным пользователям узнать, какие email зарегистрированы в системе.

```json
{"error":"неверный email или пароль"}
```

### Проверка JWT signing method

При разборе JWT проверяется ожидаемый алгоритм подписи. Токены с неожиданным signing method отклоняются чтобы предотвратить атаки с подменой алгоритма.

### Параметризованные SQL-запросы

Пользовательские значения передаются в SQL через placeholders. Это защищает от SQL-инъекций.

### Rate limiting

На `POST /api/auth/register` и `POST /api/auth/login` действует ограничение попыток по IP чтобы предотвратить brute-force атаки. Например, не более 5 попыток в минуту с одного IP на каждый auth endpoint.

### Блокировка пользователя

Admin endpoint блокирует пользователя и отзывает все активные refresh-токены в одной транзакции чтобы гарантировать, что после блокировки пользователь не сможет получить новые access-токены через refresh.



## 9. Логирование

На каждый HTTP-запрос пишется одна строка в stdout:

```text
method=GET path=/api/teachers status=200 duration=2.1ms remote=127.0.0.1
```

Логируются:

- HTTP method;
- path;
- status;
- duration;
- remote IP.

Для тестового задания используется стандартный `log.Printf`, без zap/zerolog.


# BONUS
## 10. Graceful shutdown

Приложение обрабатывает сигналы:

- `SIGINT`
- `SIGTERM`

При получении сигнала сервер прекращает принимать новые соединения и даёт активным запросам до 10 секунд на завершение через `http.Server.Shutdown`.

Это нужно, чтобы контейнер завершался корректно и не обрывал активные HTTP-запросы резко.


## 11. Тесты и проверки

### Запуск тестов

```bash
go test ./...
```

Тестами покрыты:

- bcrypt-хэширование и проверка пароля;
- генерация и проверка access-токена;
- генерация и хэширование refresh-токена;
- rate limiter;
- CORS preflight;
- валидация пароля.

### go vet

```bash
go vet ./...
```

### Форматирование

```bash
gofmt -w .
```

### Docker-проверка

```bash
docker compose down -v
docker compose up --build
```

После запуска:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/api/teachers
```



## 12. Спорные решения

### Собственный migration runner вместо `golang-migrate`

Для проекта реализован небольшой собственный runner миграций, потому что он в целом закрывает требования задания, не добавляя внешнюю зависимость и остаётся понятным для проверки.
Runner поддерживает применение up-миграций и фиксацию применённых версий в `schema_migrations`. Rollback не реализован, потому что для тестового задания откаты не обязательны.

### `net/http` вместо роутера типа chi/gorilla

Проект использует стандартный `net/http`. Этого достаточно для текущего набора endpoint-ов. Path-параметры разбираются вручную.

### Rate limiter в памяти

Rate limiter хранит состояние в памяти процесса. Для такого небольшого приложения этого достаточно. Для production и горизонтального масштабирования лучше вынести rate limiting в Redis.

### Access-токены не отзываются при logout

Logout отзывает refresh-токен. Access-токен остаётся действительным до истечения короткого срока жизни, что является компромиссным решением, ведь для мгновенного отзыва access-токенов потребовался бы JTI blocklist или дополнительная проверка состояния пользователя в БД на каждый защищённый запрос.

### Проверка `is_blocked` не выполняется на каждый protected request

Сейчас блокировка влияет на login и refresh. Уже выданный access-токен может работать до истечения срока действия, что соответствует условию - access-токен живёт коротко, а усложнение через blocklist или постоянный SELECT в auth middleware не добавлялось.

### Ручная сборка зависимостей

Зависимости собираются вручную в `cmd/server/main.go`. Для проекта такого размера это проще, чем DI-контейнер.

### SQL без ORM

Repository использует прямой SQL через `sqlx`. Это даёт контроль над запросами, JOIN-ами, агрегациями и транзакциями. ORM здесь был бы избыточен.



## 13. Что бы сделал дальше за ещё один день

1. Я бы реализовал swagger или OpenAPI файл с описанием всех эндпоинтовов и примерами ответов, потому что это удобно.
2. Audit log, потому что для админ-операций полезно иметь запись о том, кто и когда выполнял критичные действия.
   - login;
   - refresh;
   - logout;
   - block user;
   - revoke tokens.
3. Более строгую email-валидацию для регистрации, например через регулярное выражение, чтобы отсеивать явно некорректные email.
4. Проверку `is_blocked` в auth middleware или JTI blocklist для мгновенного отзыва access-токенов.
5. Makefile с командами:
    - `make test`;
    - `make vet`;
    - `make run`;
    - `make docker-up`.