# LoL Build Overlay

Движок рекомендаций предметов для League of Legends по снапшоту **Live Client API** (`allgamedata`): класс чемпиона, роль, драфт врагов, давление по предметам и следующий слот к покупке.

## Требования

- [Go 1.22+](https://go.dev/dl/)
- Python 3.10+ (только для генерации фикстур и обновления данных)
- Запущенный клиент LoL **в матче** (для `-live`)
- .NET Framework 4.x (для WPF-оверлея, обычно уже есть на Windows)

## Оверлей

C# HUD в `overlay/` собирает **это** Go-ядро и рисует следующий слот поверх клиента:

```powershell
overlay\Start.bat
```

Батник делает `go build ./cmd/recommend` в корне репозитория и кладёт `overlay\recommend.exe`. Оверлей вызывает его с `-json -root <корень>` (`-live`, иначе демо-фикстура).

## Быстрый старт (CLI)

```powershell
cd lol-overlay   # или lol-build-overlay
$env:Path = "C:\Program Files\Go\bin;" + $env:Path

# Рекомендация из живой игры
go run ./cmd/recommend -live

# JSON-вывод
go run ./cmd/recommend -live -json
```

Live Client доступен только пока идёт игра:

```text
https://127.0.0.1:2999/liveclientdata/allgamedata
```

Проверка / снимок на диск:

```powershell
curl.exe -k https://127.0.0.1:2999/liveclientdata/allgamedata -o snap.json
go run ./cmd/recommend -fixture snap.json
```

Удобный цикл без ранкеда: **Practice Tool** → боты → `-live` в другом терминале.

## CLI

| Команда | Назначение |
|--------|------------|
| `go run ./cmd/recommend -live` | Снимок с Live Client |
| `go run ./cmd/recommend -fixture PATH` | Локальный JSON `allgamedata` |
| `go run ./cmd/recommend -json` | Печать `Recommendation` как JSON |
| `go run ./cmd/bootstrap-tags` | Пересборка базовых тегов чемпионов |

Флаги `recommend`:

- `-root` — корень проекта с папкой `data/` (по умолчанию ищется от cwd)
- `-live` / `-fixture` — источник снапшота (нужен один из них)

## Что делает движок

1. Резолвит теги активного чемпиона (`data/champion_tags.json`, с учётом `by_role`)
2. Берёт seed-шаблон (mage / assassin / marksman / fighter / tank / support + jungle/support-роутинг)
3. Подстраивает приоритеты под драфт и live-pressure врагов
4. Учитывает lane-оппонента для ботинок, фронтлайн союзников
5. Синхронизирует старт с инвентарём **только** если стартер той же семьи (Blade ≠ Ring ≠ Shield ≠ Atlas ≠ pet)
6. Выдаёт порядок слотов и **Next item**

Поля вывода (текст):

- `seed` — имя шаблона (`mage/jungle`, `assassin`, `support`, …)
- `[offrole]` — чемпион не на типичной роли по тегам
- `Pressure` / `Top threats` — сигналы по врагам
- `Build order` — слоты по приоритету

## Данные

| Файл | Содержание |
|------|------------|
| `data/champions.json` | Чемпионы (Data Dragon) |
| `data/items.json` | Предметы |
| `data/champion_tags.json` | Primary-класс + теги + `by_role` |
| `data/champion_tag_fixes.json` / `overrides` | Ручные правки тегов |

Обновить предметы / чемпионов (нужен `pip install -r requirements.txt`):

```powershell
python scripts/fetch_items.py
python scripts/fetch_champions.py
```

## Фикстуры (оффлайн-тесты)

Генератор собирает полноценные `allgamedata` JSON: чемпионы на **натуральных** ролях, старты по классу×позиции.

```powershell
# 100 кейсов
python scripts/gen_live_fixtures.py --count 100

# + прогон recommend и отчёт
python scripts/gen_live_fixtures.py --count 50 --recommend --report out/report.md

# Свой каталог
python scripts/gen_live_fixtures.py --count 20 --out testdata/generated/live_onrole
```

Готовые ручные фикстуры: `testdata/fixtures/`.  
Сгенерированные (`testdata/generated/`) в git не коммитятся.

Один кейс вручную:

```powershell
go run ./cmd/recommend -fixture testdata/fixtures/draft_start.json
```

## Тесты

```powershell
go test ./internal/...
```

## Структура

```text
cmd/recommend/          CLI рекомендаций (JSON для оверлея)
cmd/bootstrap-tags/     бутстрап тегов
internal/engine/        Recommend, threat, pressure
internal/rules/         seeds, draft/live adjust, start sync
internal/tags/          таксономия чемпионов/предметов
internal/liveclient/    парсер Live Client API
internal/data/          загрузка data/
overlay/                WPF HUD, вызывает recommend.exe
scripts/                fetch + генератор фикстур
data/                   статический геймдата
testdata/fixtures/      эталонные снапшоты
```

## Замечания по Live Client

- В **Practice Tool** у человека часто `position: "NONE"` — jungle всё равно определяется по Smite / jungle pet, но `by_role` для `JUNGLE`/`MIDDLE` может не примениться.
- API отвечает только во время матча; в лобби / клиенте без игры будет connection refused.
- Сертификат локальный self-signed — клиент ходит с `InsecureSkipVerify` (как принято для Riot Live Client).

## Лицензия

Личные / учебные цели. League of Legends и связанные данные принадлежат Riot Games.
