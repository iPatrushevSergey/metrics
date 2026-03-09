# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Запуск

```bash
# Сервер
go build -o server ./cmd/server
./server

# Агент (в отдельном терминале)
go build -o agent ./cmd/agent
./agent
```

**Тесты и бенчмарки:**
```bash
go test ./internal/... -v
go test ./internal/... -run=^$ -bench=. -benchmem
```

## Профилирование памяти

Одна и та же нагрузка до и после оптимизаций (порядок сохранять).

**Нагрузка (4 команды по очереди, сервер запущен):**
```bash
hey -z 15s -c 8 -m POST -H "Content-Type: application/json" -d "{\"id\":\"cpu\",\"type\":\"gauge\",\"value\":1.5}" http://localhost:8080/update/
hey -z 15s -c 6 -m POST -H "Content-Type: application/json" -d "[{\"id\":\"a\",\"type\":\"gauge\",\"value\":1},{\"id\":\"b\",\"type\":\"counter\",\"delta\":1}]" http://localhost:8080/updates/
hey -z 20s -c 10 http://localhost:8080/
hey -z 15s -c 8 -m POST -H "Content-Type: application/json" -d "{\"id\":\"cpu\",\"type\":\"gauge\"}" http://localhost:8080/value/
```

**До оптимизаций:** собрать сервер → запустить → дать нагрузку (4 команды) → снять профиль:
```bash
curl -s "http://localhost:8080/debug/pprof/heap" -o profiles/base.pprof
```

**После оптимизаций:** пересобрать → запустить → та же нагрузка (4 команды) → снять профиль:
```bash
curl -s "http://localhost:8080/debug/pprof/heap" -o profiles/result.pprof
```

**Сравнение (diff по alloc_space):**
```bash
go tool pprof -http=:9096 -diff_base=profiles/base.pprof profiles/result.pprof
```
В браузере: Sample → **alloc_space**, затем Top. Отрицательные значения — уменьшение аллокаций.

**Что удалось снизить (по diff alloc_space):**
- **compress/flate.NewWriter** — Flat −209408 MB (−75%), Cum −254398 MB (−91.5%); пул gzip writer/reader.
- **compress/flate.initDeflate** — −43747 MB (−15.7%); 
- **huffmanEncoder.generate** — −2150 MB (−0.8%).

- Цепочка обработки запроса: **GzipGinMiddleware**, **LoggerMiddleware**, **HashMiddleware**, **net/http.serve** — Cum порядка −246600…−249100 MB (−88…−90%).
- **compressResponseWriter.Write/Close**, **responseBodyWriter.Write** — Cum −254144 MB, −252856 MB (−91%).
- **GetJSON** (gin.JSON, WriteJSON) — Cum −150914 MB (−54.3%).
- **GetAll** (handler + template) — Cum −97055 MB (−34.9%); **text/template.execute** — −98660 MB (−35.5%).
