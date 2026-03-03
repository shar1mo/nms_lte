# Заметки по спецификациям

Колонки:
- **Приоритет:** `P1`, `P2`, `P3` (этап 2+).
- **Кому читать:** `Студент 1` (сеть/NETCONF), `Студент 2` (API/данные), `Оба`.
- **Неделя:** ориентир по плану из `docs/weeks`.

| Файл | TS | Коротко о документе | Приоритет | Кому читать | Неделя | Для какого функционала |
|---|---|---|---|---|---|---|
| `docs/specs/28500-j00.pdf` | 28.500 | Общая архитектура и требования управления | P1 | Оба | 01 | Базовая архитектура NMS |
| `docs/specs/28530-j00.pdf` | 28.530 | Концепции и use-case orchestration | P1 | Оба | 01 | Границы MVP и сценарии управления |
| `docs/specs/28533-j40.pdf` | 28.533 | Архитектурный фреймворк управления | P1 | Оба | 01 | Разбиение на сервисы |
| `docs/specs/23003-j50.pdf` | 23.003 | Идентификаторы, адресация, naming | P1 | Студент 2 | 02/05 | DN/ID модель и API-ключи |
| `docs/specs/28658-j00.pdf` | 28.658 | LTE NRM IS (классы и связи E-UTRAN) | P1 | Оба | 02 | LTE профиль объектов для MVP |
| `docs/specs/28659-j00.pdf` | 28.659 | LTE NRM SS (CORBA/XML, DN mapping) | P1 | Оба | 02 | Сопоставление IS↔SS, подготовка bulk XML |
| `docs/specs/28622-j60.pdf` | 28.622 | Generic NRM IS | P1 | Оба | 02 | Базовые классы (ManagedElement и т.д.) |
| `docs/specs/28623-j60/28623-j60.pdf` | 28.623 | Generic NRM SS definitions | P1 | Оба | 02 | Практическое отображение модели |
| `docs/specs/32158-j10.pdf` | 32.158 | Правила REST Solution Set | P1 | Студент 2 | 03/11 | Правильный API стиль |
| `docs/specs/32160-j50.pdf` | 32.160 | Шаблон management-сервисов | P1 | Студент 2 | 03/11 | Структура API и сервисов |
| `docs/specs/28510-j00.pdf` | 28.510 | Требования к Configuration Management | P1 | Оба | 06 | CM pipeline и безопасность изменений |
| `docs/specs/28532-j30/28532-j30.pdf` | 28.532 | Generic management services | P1 | Оба | 06/11 | Базовые операции и сервисные шаблоны |
| `docs/specs/28545-h00.pdf` | 28.545 | Fault Supervision | P1 | Оба | 10 | Fault/heartbeat подсистема |
| `docs/specs/28550-j20/28550-j20.pdf` | 28.550 | Performance assurance | P1 | Оба | 09 | PM-lite и дальнейший PM-full |
| `docs/specs/32401-j00.pdf` | 32.401 | PM концепции и требования | P1 | Студент 2 | 09 | Модель хранения и чтения метрик |
| `docs/specs/28620-j20.pdf` | 28.620 | Umbrella/Federated информационная модель | P2 | Оба | 02/после MVP | Единая модель при росте системы |
| `docs/specs/28531-j20/28531-j20.pdf` | 28.531 | Provisioning | P2 | Студент 2 | 12 | Групповые и сервисные операции |
| `docs/specs/28515-j00.pdf` | 28.515 | Требования к Fault Management | P2 | Студент 1 | 10/после MVP | Углубление fault логики |
| `docs/specs/28525-j00.pdf` | 28.525 | Life Cycle Management требования | P2 | Оба | после MVP | Жизненный цикл сущностей и версий |
| `docs/specs/28537-j40.pdf` | 28.537 | Набор management capabilities | P2 | Оба | 14 | Чек-лист зрелости платформы |
| `docs/specs/28540-j30.pdf` | 28.540 | 5G NRM Stage 1 | P2 | Студент 1 | после MVP | Расширение модели ресурсов |
| `docs/specs/28541-j60/28541-j60.pdf` | 28.541 | 5G NRM Stage 2/3 | P2 | Студент 1 | после MVP | Детализация NRM и interworking |
| `docs/specs/22261-jd0.pdf` | 22.261 | Service requirements 5G (Stage 1) | P3 | Студент 2 | этап 2 | Связь NMS с бизнес-требованиями |
| `docs/specs/23288-j50.pdf` | 23.288 | 5GS analytics architecture | P3 | Студент 2 | этап 2 | Интеграция сетевой аналитики |
| `docs/specs/23501-j60.pdf` | 23.501 | Архитектура 5GS Stage 2 | P3 | Студент 1 | этап 2 | Контекст междоменных связей |
| `docs/specs/28104-j30/28104-j30.pdf` | 28.104 | Management Data Analytics | P3 | Студент 2 | 17/этап 2 | MDA и расширенная аналитика |
| `docs/specs/28314-j10.pdf` | 28.314 | Plug and Connect: концепции | P3 | Студент 1 | этап 2 | Автоматический onboarding |
| `docs/specs/28315-j00.pdf` | 28.315 | Plug and Connect: процедуры | P3 | Студент 1 | этап 2 | Процессы подключения узлов |
| `docs/specs/28316-j00.pdf` | 28.316 | Plug and Connect: форматы | P3 | Оба | этап 2 | Форматы данных onboarding |
| `docs/specs/28535-j00.pdf` | 28.535 | Assurance services requirements | P3 | Студент 2 | этап 2 | SLA/качество сервиса |
| `docs/specs/28536-j20/28536-j20.pdf` | 28.536 | Assurance services Stage 2/3 | P3 | Студент 2 | 17/этап 2 | Детализация assurance логики |
| `docs/specs/28538-j50/28538-j50.pdf` | 28.538 | Edge Computing Management | P3 | Студент 1 | этап 2 | Расширение в edge домен |
| `docs/specs/28552-j60.pdf` | 28.552 | 5G performance measurements | P3 | Студент 2 | этап 2 | Расширение PM модели |
| `docs/specs/28554-j60.pdf` | 28.554 | E2E KPI 5G | P3 | Студент 2 | этап 2 | Сквозные KPI и SLA отчеты |
| `docs/specs/28555-j00.pdf` | 28.555 | Network policy management Stage 1 | P3 | Оба | этап 2 | Policy-driven управление |

---

## Рекомендуемый минимум чтения для старта MVP

### Студент 1
1. `28658-j00.pdf`
2. `28659-j00.pdf`
3. `28622-j60.pdf`
4. `28510-j00.pdf`
5. `28545-h00.pdf`

### Студент 2
1. `28500-j00.pdf`
2. `32158-j10.pdf`
3. `32160-j50.pdf`
4. `28532-j30/28532-j30.pdf`
5. `32401-j00.pdf`
