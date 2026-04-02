# NMS LTE

Система управления LTE-сетью.

## Структура проекта

```text
cmd/nms-lte/                 запуск приложения
internal/app/                сборка зависимостей и HTTP сервера
internal/httpapi/            HTTP обработчики
internal/model/              модели данных
internal/store/memory/       хранилище в памяти
internal/service/ne/         сервис узлов
internal/service/inventory/  сервис инвентаря
internal/service/cm/         сервис конфигурации
internal/service/fault/      сервис событий и heartbeat
internal/service/pm/         сервис метрик
internal/id/                 генератор идентификаторов
```

## Полезные документы

- `docs/DESIGN_GUIDE.md` — правила проектирования и разработки.
- `docs/ARCHITECTURE.md` — архитектура и границы MVP.
- `docs/weeks/README.md` — план работ по неделям.
- `docs/specs/NOTES_PLAN.md` — обзор спецификаций с приоритетами.

## Разделение по сервисам и функциям

- `ne`: регистрация узла, чтение списка узлов, статус узла.
- `inventory`: синхронизация и чтение последнего снимка инвентаря.
- `cm`: создание запроса изменения и пошаговое выполнение (`lock/edit-config/validate/commit/unlock`).
- `fault`: регистрация событий, heartbeat проверка.
- `pm`: сбор и чтение метрик.

## Запуск

```bash
go run ./cmd/nms-lte
```

По умолчанию сервер стартует на `:8080`. Можно задать порт:

```bash
PORT=9090 go run ./cmd/nms-lte
```

Для `NETCONF` подключения сервис читает переменные окружения:

```bash
NETCONF_USERNAME=admin
NETCONF_PASSWORD=admin
NETCONF_PORT=830
NETCONF_RECONNECT_INTERVAL=5s
NETCONF_SCHEMA_PATH=/path/to/yang/modules
```

Статусы узла в API:

- `connecting` — идет первичное подключение;
- `connected` — сессия активна, возможности узла считаны;
- `disconnected` — подключение сейчас недоступно;
- `reconnecting` — сервис восстанавливает сессию после обрыва.

Поле `address` в API можно передавать как `host` или как `host:port`.

## Зависимости

```bash
sudo apt install -y \
  libssl-dev \
  libssh-dev \
  libcurl4-openssl-dev \
  pkg-config
```

## Сборка библиотек

```bash
make bootstrap 
```

## Сборка и запуск netconf клиента

```bash
make build-netconf-cli 
```

Для проверки netconf клиента необходимо запустить netconf сервер

```bash
cd .build/libnetconf2/examples
./server
```

Запуск netconf клиента в другом терминале

```bash
./netconf-client
```

## Быстрые примеры API

Создать узел:

```bash
curl -X POST http://localhost:8080/api/v1/ne \
  -H 'Content-Type: application/json' \
  -d '{"name":"enb-1","address":"10.0.0.1","vendor":"vendor-a"}'
```

Синхронизировать инвентарь:

```bash
curl -X POST http://localhost:8080/api/v1/ne/<NE_ID>/inventory/sync
```

Запустить изменение конфигурации:

```bash
curl -X POST http://localhost:8080/api/v1/cm/requests \
  -H 'Content-Type: application/json' \
  -d '{"ne_id":"<NE_ID>","parameter":"cell.pci","value":"101"}'
```

Проверка heartbeat:

```bash
curl -X POST http://localhost:8080/api/v1/ne/<NE_ID>/heartbeat/check \
  -H 'Content-Type: application/json' \
  -d '{"healthy":true}'
```

Собрать метрику:

```bash
curl -X POST http://localhost:8080/api/v1/ne/<NE_ID>/pm/collect \
  -H 'Content-Type: application/json' \
  -d '{"metric":"availability"}'
```

## Тесты

```bash
go test ./...
```
