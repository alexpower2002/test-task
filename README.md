# gRPC сервис курсов USDT

Сервис реализует:

- gRPC метод `GetRates` для получения `ask`, `bid` и времени получения курса USDT
- сохранение курса в PostgreSQL при каждом вызове `GetRates`
- gRPC метод `HealthCheck`
- graceful shutdown

В реализации используются:

- Go `1.25`
- HTTP-клиент на `resty`
- `PostgreSQL`
- миграции из директории `migrations/`
- `golangci-lint`
- `OpenTelemetry`
- `Prometheus`
- `zap`

## Конфигурация

Параметры можно задавать и флагами, и переменными окружения.

| Флаг | Переменная окружения | Значение по умолчанию                                               |
| --- | --- |---------------------------------------------------------------------|
| `-grpc-port` | `APP_GRPC_PORT` | `9001`                                                              |
| `-metrics-port` | `APP_METRICS_PORT` | `9090`                                                              |
| `-postgres-dsn` | `APP_POSTGRES_DSN` | `postgres://postgres:postgres@localhost:5432/rates?sslmode=disable` |
| `-grinex-url` | `APP_GRINEX_URL` | `https://grinex.io/api/v1/spot/depth?symbol=usdta7a5`               |
| `-http-timeout` | `APP_HTTP_TIMEOUT` | `5s`                                                                |
| `-shutdown-timeout` | `APP_SHUTDOWN_TIMEOUT` | `10s`                                                               |
| `-db-connect-attempts` | `APP_DB_CONNECT_ATTEMPTS` | `10`                                                                |
| `-db-connect-delay` | `APP_DB_CONNECT_DELAY` | `1s`                                                                |
| `-log-level` | `APP_LOG_LEVEL` | `info`                                                              |
| `-service-name` | `APP_SERVICE_NAME` | `rates-service`                                                     |
| `-trace-exporter` | `APP_TRACE_EXPORTER` | `stdout`                                                            |
| `-ask-method` | `APP_ASK_METHOD` | `topN`                                                              |
| `-ask-n` | `APP_ASK_N` | `1`                                                                 |
| `-ask-m` | `APP_ASK_M` | `2`                                                                 |
| `-bid-method` | `APP_BID_METHOD` | `topN`                                                              |
| `-bid-n` | `APP_BID_N` | `1`                                                                 |
| `-bid-m` | `APP_BID_M` | `2`                                                                 |

Правила расчёта:

- `topN` возвращает цену из позиции `N`
- `avgNM` возвращает среднее значение на диапазоне `[N; M]`

Я не до конца понял, требовалось ли выбирать режим расчёта заранее через конфигурацию приложения или передавать его отдельно для каждого запроса.
Я выбрал вариант с конфигурацией через переменные окружения.

## Команды

```bash
make build
make test
make docker-build
make run
make lint
```

## Запуск

Перед запуском через Docker создайте файл `.env` в корне проекта. В него можно вынести не только настройки PostgreSQL, но и параметры приложения, например, чтобы проверить другой режим расчёта:

```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=rates
APP_POSTGRES_DSN=postgres://postgres:postgres@postgres:5432/rates?sslmode=disable
APP_GRPC_PORT=9001
APP_METRICS_PORT=9090
APP_TRACE_EXPORTER=none
APP_ASK_METHOD=avgNM
APP_ASK_N=0
APP_ASK_M=2
APP_BID_METHOD=topN
APP_BID_N=1
```

```bash
docker compose up -d
docker compose run --rm app ./app
```

## Проверка

Unit-тесты:

```bash
make test
```

Проверка gRPC:

```bash
grpcurl -plaintext -d '{}' localhost:9001 rates.RateService/HealthCheck
grpcurl -plaintext -d '{}' localhost:9001 rates.RateService/GetRates
```

Проверка данных в PostgreSQL:

```bash
docker compose exec postgres psql -U postgres -d rates -c "SELECT ask, bid, timestamp FROM rates ORDER BY id DESC LIMIT 10;"
```
