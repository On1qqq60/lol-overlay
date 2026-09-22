# LoL Build Overlay

Движок рекомендаций предметов для League of Legends по снапшоту **Live Client API** (`allgamedata`): класс чемпиона, роль, драфт врагов, давление по предметам и следующий слот к покупке.

## Требования

- [Go 1.22+](https://go.dev/dl/)
- Python 3.10+ (только для генерации фикстур и обновления данных)
- Запущенный клиент LoL **в матче** (для `-live`)
- .NET Framework 4.x (для WPF-оверлея, обычно уже есть на Windows)

## Оверлей

После `git clone` на Windows достаточно:

```powershell
overlay\Start.bat
```

Батник собирает Go-движок (`recommend.exe`) и C# HUD из исходников в этом репозитории — тот же алгоритм, что локально. Иконки предметов качаются с Data Dragon при первом показе. Нужны Go 1.22+ и .NET Framework 4.x (csc). Live Client отвечает только во время матча; без игры HUD покажет демо-фикстуру.

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
2. Берёт сид по **чемпиону×роли** (таблица семей). Класс (mage / assassin / marksman / fighter / tank / support) — запасной роутинг, не единственный шаблон. Танк с тегом fighter (Poppy и похожие) идёт в bruiser-сид, не в чистый Sunfire
3. Если в сумке уже лежит кор другой ветки — переключает семью (`CommitFamily`) и выкидывает скипнутый первый кор
4. Подстраивает приоритеты под драфт и live-pressure врагов
5. Учитывает lane-оппонента для ботинок, фронтлайн союзников
6. Синхронизирует старт с инвентарём **только** если стартер той же семьи (Blade ≠ Ring ≠ Shield ≠ Atlas ≠ pet)
7. На раннем бэке next — компонент к кору, не целая легендарка
8. Выдаёт порядок слотов и **Next item**

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

## Журнал алгоритма

### 21–22 сентября 2026

День — проход **всех пяти классов/ролей**, не правка одного сида Poppy. Сначала 1000 кейсов показали 5.3/10: система раздавала классовый шаблон, не билд чемпиона. Дальше — таблица семей champion×role, уважение сумки, компоненты на бэке, и по ~200 живых фикстур на роль: TOP → JUNGLE → MIDDLE → BOTTOM → UTILITY.

Оценки после прохода (ручной аудит):

| Роль | Старт | Мид / живой next | В катку |
|------|-------|------------------|---------|
| TOP | 6 → **7** после патча сумки | 4 → сумка бьёт сид | 5 → **7** |
| JUNGLE | 6 → **7.5** (петы) | 6.5 → **7–7.5** | 6 → **7** |
| MIDDLE | **7.5** | **7** (Corki ещё орёт) | **7** |
| BOTTOM | **7.5** | 6.5 → Kraken-commit | **7 → 8** после патча |
| UTILITY | **6.5** | **5.5** | **6** |

Саппорт слабее починенного бота и на ступень ниже первого леса. Класс угадан. Покупку «прямо сейчас» на саппах в катку рано.

#### Общие правила (это и есть перелопатка классов)

- **Champion×role identity.** Не «все танки = Sunfire / все маги = Blackfire / все бойцы = Trinity». Таблица семей по чемпиону и роли; класс — запасной роутинг.
- **`CommitFamily`.** Легендарка (и ключевой компонент) в сумке переключает семью. Скипнули первый кор — это валидная катка, оверлей идёт за сумкой. Иначе Ривен с Trinity слышит Eclipse, Kayle с Kraken+IE слышит Nashor.
- **`DropSkippedFirstCores`.** Trinity vs Eclipse, Sunfire vs Heartsteel, Locket vs Moonstone vs Blackfire, Kraken vs IE-ветка: незакрытая альтернатива выкидывается из плана.
- **Компонент, не легендарка на 500g.** `ComponentsToward` + `PickAffordable`: early next — Sheen / Bami’s / Lost Chapter / Noonquiver / Tear / Dirk / Kindlegem / Forbidden Idol / Tome. Полный Trinity / Sunfire / Locket / Moonstone / Blackfire в next при 500 золота — ложь.
- **Ботинки.** ADC и лес не переезжают на Steelcaps с AD-линии. Swifties в сиде — identity MS (не затираются lane Mercs/Steelcaps, пока в сиде нет танковых сапог). Cass / Yuumi без сапог. Команда после 8-й: большинство AD → Steelcaps, AP → Mercs, если сид это допускает.
- **Sheen.** На ADC / AP / саппорте Sheen не коммитит в Trinity. Essence Reaver = Sheen + Caulfield + Cloak. Смоулдер с Sheen остаётся на ER.
- **Emit / хвост.** 1 старт + 1 боты + до 5 легендарик. Кор и топ-оффенс пинятся; последние 2 слота — аукцион по давлению (хил / AP-берст / AD). Thornmail только если AD+хил сильнее AP-берста; иначе FoN / Banshee. `replaceTail` не съедает Dead Man’s / FoN.
- **Item gates.** Jhin без Kraken/Berserkers, Vladimir без Lost Chapter, Cass без сапог.

#### TOP

Имена, не «боецкий паштет»:

- Poppy и tank×fighter → `SeedTankFighter`: Iceborn, Sundered Sky, Cleaver. Не Sunfire-only.
- Garen / Darius / Sett — Stridebreaker. Urgot — Cleaver / Sterak / Titanic. Yorick — Trinity. Illaoi — Iceborn / Cleaver / Sterak. Shen — Sunfire + Titanic. Jayce — Trinity + Manamune. Nasus — Sheen-линия. Gwen — Nashor + Riftmaker. Teemo — Blackfire. Singed — Rylai после Blackfire. Gragas топ — танк. Volibear топ — Trinity (AP только если уже Nashor).
- Чистый танк (Malphite, Cho, Ornn, Sion, Rammus) — Sunfire / Heartsteel по kit. Cho без Riftmaker; грив танка — Thornmail, не Executioner’s.
- Next на бэке бойца — Sheen, танка — Bami’s.

#### JUNGLE

- **Питомец в buy, не Доран.** После фикса: не сто Mosstomper. Gustwalker / Scorchclaw / Mosstomper по kit (Lillia / Karthus — красный, Lee / Kha’Zix — серый, Rammus / Zac — зелёный). Ivern — Gustwalker + Moonstone, не маг с Lost Chapter.
- Yi / Bel’Veth — Kraken → BotRK → Guinsoo. Warwick — BotRK → Titanic. Olaf — Stridebreaker. Kha’Zix / Rengar — Youmuu / Collector. Kayn — Eclipse (форма — ещё дыра). Lillia / Udyr — Liandry + Rylai. Hecarim — Trinity. Volibear лес — Nashor + Riftmaker. Sylas лес — Rocketbelt, не Tear+RoA. Shaco AP больше не получает Collector.
- Танки на бэке: next = Bami’s, не целый Sunfire.

#### MIDDLE

- Первый бэк — компонент: Lost Chapter, Tear, Serrated Dirk, Hextech Alternator, Recurve, Sheen, Fiendish Codex, Bami’s, Noonquiver, B.F. Sword.
- Слеза на своих: Anivia, Kassadin, Ryze, Cass — Tear → RoA. Cass ещё Rylai.
- Ассасины: Talon / Zed / Qiyana / Naafiri / LeBlanc — lethality / AP burst, не боец.
- Имена магов: TF — Lich Bane (не Blackfire/Malignance). Orianna — Luden + Horizon. Viktor — Luden. Swain — Liandry + Rylai. Vlad — Cosmic + Riftmaker. Sylas — Rocketbelt + Riftmaker. Corki — Trinity + Manamune. Ahri / burst — не артиллерия.
- Corki с Kraken+IE, которому орёт Manamune — тот же баг, что Ezreal на боте; на боте починили, на миде ещё всплывает.

#### BOTTOM

Не один крит на всех:

- Jhin — Collector → Youmuu / IE / RFC, Ionians. Не Kraken, не Berserkers.
- Caitlyn — YunTal. MF — Youmuu. Lucian — Essence Reaver → Navori. Zeri — Stormrazor + Statikk. Smolder — ER + Shojin. Ezreal — Tear → Manamune → Trinity.
- Vayne / Kalista / Varus — BotRK → Guinsoo. Kog — Recurve → Guinsoo / Nashor / BotRK.
- Kaisa — гибрид, пока нет IE / YunTal.
- Обычный крит с Kraken+IE → RFC / Collector. **Kraken в сумке без IE** на Ezreal / Zeri / Vayne / Kalista / Kog / Varus не переписывает identity обратно в Manamune / «ещё один крит».
- Steelcaps с AD-линии на боте выключены. Next на старте — Noonquiver / Tear / Vamp / B.F. / Recurve.

#### UTILITY

Три коридора, не один паст. Класс (танк / энчантер / маг) почти всегда верный.

- **Танк-энгейдж** — Locket → Mercs → Knight's Vow → Thornmail: Leona, Naut, Rell, Alistar, Blitz, Maokai, Tahm, Braum, Thresh, Poppy.
- **Энчантер** — Moonstone → Ionians → Redemption → Mikael: Sona, Soraka, Lulu, Nami, Janna, Milio, Seraphine, Zilean. Юми без сапог.
- **Маг** — Atlas + Lost Chapter / Blackfire → Liandry → Zhonya: Brand, Zyra, Xerath, Vel’Koz, Lux, Morgana, Heimer. Старт **World Atlas (3865)**, не Доран и не кольцо с мида. Свипер — Oracle Lens, не Farsight ADC.

Имена поверх класса:

- Pyke — Dirk → Youmuu → Umbral. Вард-предмет ассасина, не Locket. Квест → Bloodsong.
- Pantheon — Eclipse + Umbral + Cleaver.
- Renata — Mandate.
- Rakan — Locket, дальше Shurelya.
- Bard — Locket + Dead Man’s + Shurelya.
- Karma — Moonstone + Mandate + Shurelya, не Blackfire с мида.
- Senna сапп — Eclipse / Mandate / lethality / Moonstone, **не** Essence Reaver (шина Люциана).

Когда сумка и сид одного класса — как на боте: Sona с Moonstone → Redemption; Soraka / Yuumi / Janna с Moonstone+Redemption → Mikael; Brand с Haunting Guise → Liandry; Leona с Locket → Knight's Vow.

Квест: Atlas → Compass → Bounty → Celestial / Dream Maker / Zaz’Zak / Sleigh / Bloodsong по семье.

Энчантеры без Void Staff / LDR. Brand / Zyra — Rylai. В магазине семей должны быть Ardent / Staff of Flowing Water / Dawncore, Frozen Heart / Abyssal / Trailblazer / Zeke’s, не только Locket / Vow / Thornmail.

##### Где на саппах ещё глупо (оценка 6.5 / 5.5 / 6)

- **Целый легендарный на 500g.** 39 стартов next = Locket, 26 = Moonstone (оба 2200). На топе/лесе/миде/боте уже Sheen / Bami’s / Lost Chapter / Noonquiver. Здесь откатились: при 500 золота нужен Kindlegem / Forbidden Idol / Tome, не Солари в лицо. Yuumi / Janna / Sona на 10-й с 350–450g, Compass + сапоги, next = Moonstone 2200. Heimer с 474g → Blackfire 2800.
- **Квест в next не существует.** Atlas в плане есть, апгрейд Compass → Bounty → финиш — нет. Оверлей пишет World Atlas, даже когда в сумке Zaz’Zak. Bloodsong и Sleigh в рекомендациях не видно. Пайк с Celestial Opposition в сумке — танковский вардовый, не Bloodsong.
- **Сид бьёт сумку.** На боте Ezreal с Kraken получает RFC, не Manamune. На сапах правило не включили / не держит:
  - Senna с Dream Maker + Moonstone + Redemption, 1200g+, next = Essence Reaver — худший кейс роли.
  - Pyke с Locket + Vow, next = Youmuu.
  - Pantheon с Locket, next = Eclipse.
  - Braum / Thresh / Taric / Rakan с Moonstone, next = Locket.
  - Bard / Galio / Naut / Maokai / Tahm с Blackfire+Liandry, next = Locket.
  - Lulu / Karma / Zilean с Blackfire, next = Moonstone.
- Тонкий магазин и чужой пенетрейшн (Void шестым на Sona/Yuumi/Janna, LDR на Galio/Renata).

Итог слабее бота и починенного топа. В катку рано, пока Locket не станет Kindlegem на бэке и пока Moonstone в сумке не перестанет проигрывать стартовому Locket.

#### Контракт / HUD

- JSON-теги на `Recommendation` / `Slot` / `Threat` / `Pressure`.
- `overlay/Start.bat` собирает `recommend.exe` из корня и зовёт `-json -live` с `-root` на этот tree.

## Замечания по Live Client

- В **Practice Tool** у человека часто `position: "NONE"` — jungle всё равно определяется по Smite / jungle pet, но `by_role` для `JUNGLE`/`MIDDLE` может не примениться.
- API отвечает только во время матча; в лобби / клиенте без игры будет connection refused.
- Сертификат локальный self-signed — клиент ходит с `InsecureSkipVerify` (как принято для Riot Live Client).

## Лицензия

Личные / учебные цели. League of Legends и связанные данные принадлежат Riot Games.
