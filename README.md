# Metrics

Сервер и агент сбора метрик (Clean Architecture). Код в `app/`, миграции в `migrations/metrics/`.

**Требования:** Go 1.24+, [Make](https://www.gnu.org/software/make/). Для integration-тестов и бенчмарков Postgres — **Docker**. Для профилирования нагрузки — [hey](https://github.com/rakyll/hey). Для шифрования — [OpenSSL](https://www.openssl.org/).

Все команды ниже — **из корня репозитория** (каталог с `Makefile`).

---

## 1. Сборка

```bash
make build
```

Бинарники: `app/bin/server`, `app/bin/agent` (тег `go_json` для Gin).

---

## 2. Запуск

### 2.1. Сервер

Терминал 1 — in-memory по умолчанию (Postgres не нужен):

```bash
CONFIG=configs/server.yaml make run-server
```

HTTP: `http://127.0.0.1:8080`, pprof: `http://127.0.0.1:8080/debug/pprof/`.

### 2.2. Агент

Терминал 2:

```bash
CONFIG=configs/agent.yaml make run-agent
```

Агент шлёт метрики на адрес из `app/configs/agent.yaml` (`http://127.0.0.1:8080`).

### 2.3. Быстрая проверка API

Пока сервер запущен:

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://127.0.0.1:8080/ping
curl -s http://127.0.0.1:8080/ | head
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"id":"test","type":"gauge","value":1}' \
  http://127.0.0.1:8080/update/
```

Ожидается `200` на `/ping` и ответ HTML на `/`.

### 2.4. Запуск бинарниками

```bash
cd app
./bin/server -c configs/server.yaml
./bin/agent -c configs/agent.yaml
```

---

## 3. Конфигурация

| Файл | Назначение |
|------|------------|
| `app/configs/server.yaml` | сервер |
| `app/configs/agent.yaml` | агент |

**Приоритет:** env → флаги → YAML → значения по умолчанию.

`make run-*` выполняет `cd app`, поэтому для YAML задай `CONFIG=configs/....yaml` (см. шаги выше) или флаг `-c`.

### 3.1. Флаги сервера

```bash
cd app
go run -tags=go_json ./cmd/server -c configs/server.yaml \
  -a 127.0.0.1:8080 \
  -f metrics.json \
  -r \
  -i 300s \
  -k secret \
  --crypto-key ../rsa_private.pem \
  -d 'postgres://user:pass@localhost:5432/metrics?sslmode=disable' \
  --retry
```

| Флаг | Env | Смысл |
|------|-----|--------|
| `-a` | `ADDRESS` | адрес прослушивания |
| `-c` | `CONFIG` | путь к YAML |
| `-l` | `LOG_LEVEL` | уровень логов |
| `-i` | `STORE_INTERVAL` | интервал сброса на диск |
| `-f` | `FILE_STORAGE_PATH` | файл с метриками |
| `-r` | `RESTORE` | восстановить метрики из файла при старте |
| `-d` | `DATABASE_DSN` | Postgres (при отсутствии in-memory repo) |
| `-k` | `KEY` | ключ HMAC |
| `--crypto-key` | `CRYPTO_KEY` | PEM приватного RSA |
| `--retry` | `ENABLE_RETRY` | retry запросов |

### 3.2. Флаги агента

```bash
cd app
go run -tags=go_json ./cmd/agent -c configs/agent.yaml \
  -a http://127.0.0.1:8080 \
  -p 2s -r 10s \
  -k secret \
  --crypto-key ../rsa_public.pem \
  -l 2
```

| Флаг | Env | Смысл |
|------|-----|--------|
| `-a` | `ADDRESS` | URL сервера |
| `-c` | `CONFIG` | путь к YAML |
| `-p` | `POLL_INTERVAL` | опрос runtime/gopsutil |
| `-r` | `REPORT_INTERVAL` | отправка метрик на сервер |
| `-k` | `KEY` | ключ HMAC |
| `--crypto-key` | `CRYPTO_KEY` | PEM публичного RSA |
| `-l` | `RATE_LIMIT` | параллельные воркеры (0 = последовательно) |

---

## 4. Postgres (опционально)

1. Подними Postgres и создай БД.
2. Миграции:

```bash
export DATABASE_DSN='postgres://user:pass@localhost:5432/metrics?sslmode=disable'
make migrate
```

3. Сервер с DSN:

```bash
cd app
DATABASE_DSN="$DATABASE_DSN" go run -tags=go_json ./cmd/server -c configs/server.yaml
```

---

## 5. RSA-ключи (опционально)

В корне репозитория:

```bash
openssl genrsa -out rsa_private.pem 2048
openssl rsa -in rsa_private.pem -pubout -out rsa_public.pem
```

| Файл | Кто | Параметр |
|------|-----|----------|
| `rsa_private.pem` | сервер | `crypto_key` / `CRYPTO_KEY` |
| `rsa_public.pem` | агент | `crypto_key` / `CRYPTO_KEY` |

---

## 6. Тесты

```bash
make test              # unit
make test-contract     # контракт агент - сервер
make test-component    # HTTP-сценарии сервера
make test-integration  # Postgres (Docker, тег integration)
make test-e2e          # e2e
make test-all          # все тесты выше
```

Покрытие тестов (~70%):

```bash
make cover       # unit + integration + contract + component + e2e, -coverpkg=./...
make cover-unit  # только unit
```

Профиль: `app/coverage.out`. HTML: `cd app && go tool cover -html=coverage.out`.

### 6.1. Integration и бенчмарки репозитория

Нужны **Docker** и `-tags=integration`.

```bash
make test-integration
```

Вручную (из `app/`):

```bash
cd app
go test -tags=integration -count=1 -v -run TestMetricRepository_ \
  ./internal/pkg/adapters/repository/postgres/

go test -tags=integration -count=1 -v \
  -run TestMetricRepository_CreateAndGetByID \
  ./internal/pkg/adapters/repository/postgres/

go test -tags=integration -bench='BenchmarkCreateBatch|BenchmarkUpdateBatch' \
  -benchmem -count=5 -benchtime=2s -run='^$' \
  ./internal/pkg/adapters/repository/postgres/
```

### 6.2. Моки (после изменения port-интерфейсов)

```bash
make generate-mocks
```

---

## 7. Профилирование памяти

Сервер с pprof: `CONFIG=configs/server.yaml make run-server`.

Каталог для снимков:

```bash
mkdir -p profiles
```

**Нагрузка** (4 команды подряд, сервер уже запущен):

```bash
hey -z 15s -c 8 -m POST -H "Content-Type: application/json" -d "{\"id\":\"cpu\",\"type\":\"gauge\",\"value\":1.5}" http://localhost:8080/update/
hey -z 15s -c 6 -m POST -H "Content-Type: application/json" -d "[{\"id\":\"a\",\"type\":\"gauge\",\"value\":1},{\"id\":\"b\",\"type\":\"counter\",\"delta\":1}]" http://localhost:8080/updates/
hey -z 20s -c 10 http://localhost:8080/
hey -z 15s -c 8 -m POST -H "Content-Type: application/json" -d "{\"id\":\"cpu\",\"type\":\"gauge\"}" http://localhost:8080/value/
```

**До оптимизаций:** сборка → запуск → нагрузка → heap:

```bash
curl -s "http://localhost:8080/debug/pprof/heap" -o profiles/base.pprof
```

**После оптимизаций:** пересборка → запуск → та же нагрузка → heap:

```bash
curl -s "http://localhost:8080/debug/pprof/heap" -o profiles/result.pprof
```

**Сравнение (diff, alloc_space):**

```bash
go tool pprof -http=:9096 -diff_base=profiles/base.pprof profiles/result.pprof
```

В UI: Sample → **alloc_space** → Top. Отрицательные значения — меньше аллокаций.

**Что удалось снизить (по diff alloc_space):**

- **compress/flate.NewWriter** — Flat −209408 MB (−75%), Cum −254398 MB (−91.5%); пул gzip writer/reader.
- **compress/flate.initDeflate** — −43747 MB (−15.7%);
- **huffmanEncoder.generate** — −2150 MB (−0.8%).
- Цепочка обработки запроса: **GzipGinMiddleware**, **LoggerMiddleware**, **HashMiddleware**, **net/http.serve** — Cum порядка −246600…−249100 MB (−88…−90%).
- **compressResponseWriter.Write/Close**, **responseBodyWriter.Write** — Cum −254144 MB, −252856 MB (−91%).
- **GetJSON** (gin.JSON, WriteJSON) — Cum −150914 MB (−54.3%).
- **GetAll** (handler + template) — Cum −97055 MB (−34.9%); **text/template.execute** — −98660 MB (−35.5%).

---

## 8. Сравнение batch-стратегий Postgres

Замеры из `repository_integration_test.go`.  
Окружение: Postgres 16 (testcontainers), `go test -tags=integration -benchmem -count=5 -benchtime=2s`  
Размеры батча: **200**, **2000**, **12000** строк.  
Подходы: **VALUES** (multi-row `$n`), **UNNEST**, **COPY**, **PREPARE** (pipeline в транзакции).

Среднее по 5 прогонам. Память — `B/op` (`-benchmem`), для читаемости в KiB/MiB.

### Insert

| Подход | 200 | | | 2000 | | | 12000 | | |
|--------|-----:|-----:|-----:|-----:|-----:|-----:|-----:|-----:|-----:|
| | время | память | allocs | время | память | allocs | время | память | allocs |
| **VALUES** | 53 ms | 313 KiB | 1131 | 39 ms | 3.4 MiB | 8374 | 212 ms | 24.7 MiB | 48415 |
| **UNNEST** | 8.8 ms | 169 KiB | 1138 | 55 ms | 1.4 MiB | 8348 | 105 ms | 8.9 MiB | 48355 |
| **COPY** | 9.4 ms | 152 KiB | 1313 | 45 ms | 603 KiB | 10322 | 59 ms | 1.9 MiB | 60324 |
| **PREPARE** | 9.4 ms | 251 KiB | 2500 | 27 ms | 2.1 MiB | 22313 | 124 ms | 14.1 MiB | 132326 |

### Update

| Подход | 200 | | | 2000 | | | 12000 | | |
|--------|-----:|-----:|-----:|-----:|-----:|-----:|-----:|-----:|-----:|
| | время | память | allocs | время | память | allocs | время | память | allocs |
| **VALUES** | 15 ms | 369 KiB | 1121 | 104 ms | 3.9 MiB | 8368 | 320 ms | 28.5 MiB | 48411 |
| **UNNEST** | 12.5 ms | 166 KiB | 1132 | 26 ms | 1.4 MiB | 8344 | 127 ms | 7.6 MiB | 48352 |
| **COPY** | 16.5 ms | 139 KiB | 1312 | 56 ms | 629 KiB | 10326 | 122 ms | 1.9 MiB | 60328 |
| **PREPARE** | 18 ms | 209 KiB | 2490 | 90 ms | 1.8 MiB | 22305 | 528 ms | 11.9 MiB | 132327 |

### Выводы

- **Insert 200:** VALUES в ~6× медленнее остальных; UNNEST/COPY/PREPARE близки по времени.
- **Insert 2000+:** COPY лидирует по скорости и памяти; PREPARE быстрый на insert, но много allocs.
- **Update:** UNNEST и COPY стабильны на больших батчах; VALUES и PREPARE деградируют на 12k.
- **VALUES на 12k:** два chunk’а SQL (лимит 65535 параметров при 6 полях на строку) — пик памяти ~25–29 MiB/op.
- Количество аллокаций можно постараться сделать меньше, но срок сдачи проекта ограничен.

### Решение для проекта

В `UpsertMetricsBatch` для Postgres выбран **UNNEST** (`CreateBatchWithUnnest` / `UpdateBatchWithUnnest`): один round-trip, предсказуемая память на больших батчах, стабильный update без temp-таблицы (в отличие от COPY). Альтернативы (`WithParams`, `WithCopy`, `WithPrepare`) остаются в репозитории для сравнения и бенчмарков.

---

## Makefile targets:

| Цель | Действие |
|------|----------|
| `build` / `build-server` / `build-agent` | сборка |
| `run-server` / `run-agent` | `go run` |
| `migrate` | goose, нужен `DATABASE_DSN` |
| `test` / `test-unit` / `test-contract` / `test-component` / `test-integration` / `test-e2e` / `test-all` | тесты |
| `cover` / `cover-unit` | покрытие |
| `generate-mocks` | gomock для port |
