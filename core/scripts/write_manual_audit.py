# -*- coding: utf-8 -*-
"""Rebuild audit reports from hand-scored _manual_verdicts.txt."""
from __future__ import annotations

import json
import re
from collections import Counter, defaultdict
from pathlib import Path
from statistics import mean, median

LIVE = Path(__file__).resolve().parents[1] / "testdata" / "generated" / "live"
ROLE_RU = {
    "TOP": "топ",
    "JUNGLE": "лес",
    "MIDDLE": "мид",
    "BOTTOM": "бот",
    "UTILITY": "саппорт",
}

VERDICT_RU = {
    10: "Эталонный 6-slot path под чемпиона и драфт.",
    9: "Identity и порядок почти эталон; мелкий situational недочёт.",
    8: "Правильная identity, разумный порядок; 1 заметный промах (pet/ботинки/хвост).",
    7: "Рабочий классовый план, который Emerald+ купит, но не лучший вариант.",
    6: "Тип урона верный, шаблон слишком общий, порядок спорный.",
    5: "Близко к identity, но неверный core-вариант или обувка раньше кора.",
    4: "Неверный class core при том же типе урона — пользователь купит чужой билд.",
    3: "Ломает чемпиона: неверный класс + неверный старт/обувка/пет.",
    2: "Катастрофическая identity (танк-предметы на ассасина, боты на Cass, крит на Ezreal).",
    1: "Список нельзя покупать ни при каких условиях.",
    0: "Пустой/сломанный вывод.",
}


def parse_verdicts(path: Path) -> dict[int, tuple[int, str]]:
    out: dict[int, tuple[int, str]] = {}
    for ln in path.read_text(encoding="utf-8").splitlines():
        m = re.match(r"^(\d{4})\s+(\d+)\s+\S+\s+(.*)$", ln.strip())
        if m:
            out[int(m.group(1))] = (int(m.group(2)), m.group(3).strip())
    return out


def parse_review(path: Path) -> dict[int, dict]:
    cases: dict[int, dict] = {}
    lines = path.read_text(encoding="utf-8").splitlines()
    i = 0
    while i < len(lines):
        m = re.match(
            r"^(\d{4})\s+(\S+)\s+(\S+)\s+seed=(\S+)\s+next=(.*)$",
            lines[i],
        )
        if not m:
            i += 1
            continue
        cid = int(m.group(1))
        block = {"champ": m.group(2), "role": m.group(3), "seed": m.group(4), "next": m.group(5).strip()}
        for j in range(1, 7):
            if i + j >= len(lines):
                break
            s = lines[i + j]
            if s.startswith("     ALL "):
                block["allies"] = s[9:]
            elif s.startswith("     ENM "):
                block["enemies"] = s[9:]
            elif s.startswith("     P "):
                block["pressure"] = s[7:]
            elif s.startswith("     R "):
                block["reasons"] = s[7:]
            elif s.startswith("     ") and "[" in s:
                block["build"] = s.strip()
        cases[cid] = block
        i += 1
    return cases


def items_of(build: str) -> list[tuple[str, str]]:
    out = []
    for part in (build or "").split(" | "):
        mm = re.match(r"(.+)\[([^\]]+)\]$", part.strip())
        if mm:
            out.append((mm.group(1).strip(), mm.group(2).strip()))
    return out


def names(items: list[tuple[str, str]]) -> list[str]:
    return [n for n, _ in items]


def six_path(items: list[tuple[str, str]]) -> list[str]:
    skip = {"start", "component"}
    path = []
    for n, r in items:
        if r in skip:
            continue
        path.append(n)
        if len(path) >= 6:
            break
    return path


def parse_side(s: str) -> dict[str, str]:
    d = {}
    for part in (s or "").split(","):
        if ":" in part:
            k, v = part.split(":", 1)
            d[k.strip()] = v.strip()
    return d


def fmt_comp(d: dict[str, str]) -> str:
    order = ["TOP", "JUNGLE", "MIDDLE", "BOTTOM", "UTILITY"]
    bits = []
    for k in order:
        if k in d:
            bits.append(f"{ROLE_RU[k]} {d[k]}")
    return ", ".join(bits)


def has(ns: list[str], *need) -> bool:
    low = [x.lower() for x in ns]
    return any(any(n.lower() in x for x in low) for n in need)


def classify_flags(note: str, ns: list[str], role: str, champ: str) -> list[str]:
    n = note.lower()
    flags = []
    if "mosstomper" in n or has(ns, "Mosstomper"):
        if role == "JUNGLE" and any(
            k in n
            for k in (
                "gustwalker",
                "scorchclaw",
                "berserkers",
                "ie",
                "eclipse",
                "stormsurge",
                "nashor",
            )
        ):
            flags.append("wrong_pet")
        elif role == "JUNGLE" and "mosstomper" in n:
            flags.append("wrong_pet")
    if "cannot buy boots" in n or "no boots" in n or "sorcerer's shoes" in n and champ == "Cassiopeia":
        flags.append("cass_boots")
    if "locket" in n and champ == "Pyke":
        flags.append("pyke_locket")
    if "moonstone" in n and champ in {"Thresh", "Shen", "Braum", "Taric", "Rakan", "Bard"}:
        flags.append("wrong_enchanter")
    if "blackfire" in n and champ in {
        "Maokai",
        "Nautilus",
        "Amumu",
        "Nunu",
        "TahmKench",
        "Zac",
        "Malphite",
        "Sejuani",
        "Gragas",
        "Lulu",
        "Shen",
    }:
        flags.append("tank_as_mage")
    if "kraken" in n and champ in {"Ezreal", "Corki", "Graves", "Kayle", "Jhin", "Smolder", "Gangplank"}:
        flags.append("wrong_adc_core")
    if "trinity" in n and champ in {"Yasuo", "Yone", "Gangplank", "Tryndamere", "MasterYi", "Jayce"}:
        flags.append("wrong_fighter_core")
    if "stormsurge" in n and champ in {"Sylas", "Kassadin", "Vladimir"}:
        flags.append("wrong_ap_core")
    if "zhonya" in n and ("2nd" in n or "before" in n):
        flags.append("zhonya_early")
    if "steelcaps first" in n or "boots first" in n or "before malignance" in n:
        flags.append("boots_before_core")
    if "ivern" in champ.lower() and ("malignance" in n or "blackfire" in n):
        flags.append("ivern_mage")
    return flags


def expert_plan(champ: str, role: str, note: str) -> str:
    n = note.lower()
    table = {
        ("Yone", "JUNGLE"): "Gustwalker → Berserker's Greaves → Trinity/Kraken → IE",
        ("Yasuo", "JUNGLE"): "Gustwalker → Berserker's Greaves → Kraken/IE → Shieldbow",
        ("Yasuo", "TOP"): "Berserker's Greaves → Kraken/IE → Immortal Shieldbow",
        ("Yone", "TOP"): "Berserker's Greaves → Trinity → IE",
        ("Ezreal", "BOTTOM"): "Tear → Trinity → Manamune → Iceborn/Serylda",
        ("Ezreal", "MIDDLE"): "Tear → Trinity → Manamune",
        ("Cassiopeia", "MIDDLE"): "Tear/Archangel → Rylai/Liandry — без ботов",
        ("Cassiopeia", "UTILITY"): "Archangel/Rylai/Liandry — без ботов",
        ("Pyke", "UTILITY"): "Youmuu/Umbral → Collector → Axiom",
        ("Pyke", "MIDDLE"): "Youmuu → Collector → Edge of Night",
        ("Thresh", "UTILITY"): "Locket → Knight's Vow → Zeke's/Locket path",
        ("Shen", "UTILITY"): "Locket → Knight's Vow",
        ("Braum", "UTILITY"): "Locket → Knight's Vow → Locket path",
        ("Taric", "UTILITY"): "Locket → Redemption/Vow",
        ("Lulu", "UTILITY"): "Moonstone → Ardent/Staff → Redemption",
        ("Rakan", "UTILITY"): "Locket/Shurelya → Zeke's",
        ("Ivern", "JUNGLE"): "Gustwalker → Moonstone → Redemption",
        ("Kayle", "BOTTOM"): "Nashor → Riftmaker/Rabadon",
        ("Kayle", "MIDDLE"): "Nashor → Riftmaker → Rabadon",
        ("Corki", "MIDDLE"): "Trinity/Manamune → RFC",
        ("Corki", "BOTTOM"): "Trinity/Manamune → RFC",
        ("Graves", "BOTTOM"): "Youmuu/Eclipse → Collector",
        ("Graves", "MIDDLE"): "Youmuu/Eclipse → Collector",
        ("Gangplank", "TOP"): "ER → IE → Collector",
        ("Gangplank", "JUNGLE"): "ER → IE → Collector",
        ("Sylas", "JUNGLE"): "Scorchclaw → RoA/Riftmaker → Zhonya",
        ("Sylas", "MIDDLE"): "RoA → Riftmaker → Zhonya",
        ("Kassadin", "MIDDLE"): "Tear/RoA → Archangel → Riftmaker",
        ("Vladimir", "MIDDLE"): "Cosmic/Liandry → Riftmaker — без Lost Chapter",
        ("MasterYi", "JUNGLE"): "Gustwalker → Berserkers → Kraken/Guinsoo",
        ("Tryndamere", "JUNGLE"): "Gustwalker → Berserkers → Kraken/IE",
        ("Jhin", "BOTTOM"): "RFC → Collector → IE",
        ("Jhin", "MIDDLE"): "RFC → Collector → IE",
        ("Smolder", "BOTTOM"): "ER → RFC → IE",
        ("Smolder", "MIDDLE"): "ER → RFC → IE",
        ("Vayne", "BOTTOM"): "BotRK → Guinsoo/Kraken → Terminus",
        ("KogMaw", "BOTTOM"): "Guinsoo → Nashor/BotRK → Void Staff/LDR",
        ("Jayce", "TOP"): "Youmuu/Eclipse → Manamune → Serylda",
        ("Jayce", "MIDDLE"): "Youmuu/Eclipse → Manamune",
        ("Nautilus", "UTILITY"): "Locket → Knight's Vow → Thornmail",
        ("Maokai", "UTILITY"): "Locket → Liandry/Thornmail",
        ("Poppy", "TOP"): "Iceborn → Sundered/Cleaver",
        ("Poppy", "UTILITY"): "Locket → Knight's Vow",
        ("Mordekaiser", "TOP"): "Riftmaker → Rylai → Liandry",
        ("Mordekaiser", "MIDDLE"): "Riftmaker → Rylai → Liandry",
        ("Mordekaiser", "JUNGLE"): "Mosstomper → Riftmaker → Rylai",
    }
    if (champ, role) in table:
        return table[(champ, role)]
    if "locket" in n and "correct" in n:
        return "Locket → Knight's Vow → situational resists"
    if "moonstone" in n and "correct" in n:
        return "Moonstone → Redemption → Mikael"
    if role == "JUNGLE" and "mosstomper" in n:
        return "переписать pet: Gustwalker/Scorchclaw по kit, затем свой core"
    return note


def strengths(champ: str, role: str, score: int, ns: list[str], note: str) -> list[str]:
    s = []
    if score >= 7:
        s.append("Identity чемпиона в целом совпадает с Emerald+ практикой")
    if has(ns, "Locket") and champ in {
        "Leona",
        "Alistar",
        "Rell",
        "Blitzcrank",
        "Nautilus",
        "Ornn",
        "Poppy",
        "Sion",
        "Rammus",
        "Skarner",
        "KSante",
        "Sett",
    }:
        s.append("Locket — правильный engage/tank support core")
    if has(ns, "Moonstone") and champ in {
        "Janna",
        "Nami",
        "Sona",
        "Yuumi",
        "Milio",
        "Soraka",
        "Seraphine",
        "Ivern",
    }:
        s.append("Moonstone — правильный enchanter core")
    if has(ns, "Youmuu") and champ in {
        "Zed",
        "Talon",
        "Qiyana",
        "Rengar",
        "Naafiri",
        "Khazix",
        "Locke",
        "Pyke",
        "Pantheon",
    }:
        s.append("Youmuu lethality — верный assassin core")
    if has(ns, "Iceborn") and champ == "Poppy":
        s.append("Iceborn на Poppy — верный identity core")
    if has(ns, "Riftmaker") and champ == "Mordekaiser":
        s.append("Riftmaker/Rylai на Mordekaiser — верный identity")
    if has(ns, "Sunfire") and role in {"TOP", "JUNGLE", "UTILITY"} and score >= 7:
        s.append("Sunfire tank path узнаваем и играбелен")
    if has(ns, "Blackfire") and champ in {
        "Brand",
        "Zyra",
        "Xerath",
        "Lux",
        "Ziggs",
        "Velkoz",
        "Heimerdinger",
        "Malzahar",
        "Swain",
        "Karthus",
    }:
        s.append("Blackfire mage path соответствует kit")
    if score >= 6 and has(ns, "Lost Chapter"):
        s.append("Lost Chapter как first component мага — корректный спайк")
    if not s and score >= 5:
        s.append("Тип урона (AD/AP) не перепутан")
    if not s:
        s.append("Старт роли (Atlas / jungle pet / Doran) формально выдан")
    return s[:4]


def weaknesses(note: str, score: int) -> list[str]:
    w = [note]
    if score <= 4:
        w.append("При строгом 6-slot buy order пользователь получит чужой чемпион")
    elif score <= 6:
        w.append("Шаблон класса вместо champion×role itemizer")
    return w[:4]


def card(cid: int, c: dict, score: int, note: str) -> str:
    items = items_of(c.get("build", ""))
    ns = names(items)
    path = six_path(items)
    allies = parse_side(c.get("allies", ""))
    enemies = parse_side(c.get("enemies", ""))
    role = c["role"]
    champ = c["champ"]
    ru = ROLE_RU.get(role, role)
    fname = f"{cid:04d}_{champ}_{role}.json"
    band = VERDICT_RU.get(score, "")
    lines = [
        f"### {fname} — **{score}/10**",
        "",
        f"- **Champion:** {champ}",
        f"- **Role:** {role} ({ru}), seed=`{c.get('seed','')}`",
        f"- **Recommended build:** {' → '.join(ns) if ns else '—'}",
        f"- **Реализованный 6-item path:** {' → '.join(path) if path else '—'}",
        f"- **Next item:** {c.get('next','')}",
        f"- **Союзники:** {fmt_comp(allies)}",
        f"- **Враги:** {fmt_comp(enemies)}",
        f"- **Давление движка:** {c.get('pressure','')}",
        f"- **Build score:** {score}/10",
        f"- **Expert verdict:** {note}",
        f"- **Интерпретация балла:** {band}",
        "- **Strengths:**",
    ]
    for x in strengths(champ, role, score, ns, note):
        lines.append(f"  - {x}")
    lines.append("- **Weaknesses / critical:**")
    for x in weaknesses(note, score):
        lines.append(f"  - {x}")
    lines.append(f"- **Better alternative:** {expert_plan(champ, role, note)}")
    lines.append(
        f"- **Final:** Если купить список как есть — **{score}/10**. {band}"
    )
    lines.append("")
    return "\n".join(lines)


def hist(scores: list[int]) -> dict[int, int]:
    c = Counter(scores)
    return {i: c.get(i, 0) for i in range(11)}


def main() -> None:
    verdicts = parse_verdicts(LIVE / "_manual_verdicts.txt")
    cases = parse_review(LIVE / "_cases_review.txt")
    missing = [i for i in range(1, 1001) if i not in verdicts or i not in cases]
    if missing:
        raise SystemExit(f"missing ids: {missing[:20]} count={len(missing)}")

    rows = []
    for i in range(1, 1001):
        score, note = verdicts[i]
        c = cases[i]
        rows.append({**c, "id": i, "score": score, "note": note})

    scores = [r["score"] for r in rows]
    h = hist(scores)

    by_champ: dict[str, list[dict]] = defaultdict(list)
    by_role: dict[str, list[int]] = defaultdict(list)
    by_cr: dict[tuple[str, str], list[int]] = defaultdict(list)
    flag_c = Counter()
    for r in rows:
        by_champ[r["champ"]].append(r)
        by_role[r["role"]].append(r["score"])
        by_cr[(r["champ"], r["role"])].append(r["score"])
        for f in classify_flags(r["note"], names(items_of(r.get("build", ""))), r["role"], r["champ"]):
            flag_c[f] += 1

    # extra pattern counts from notes
    note_pat = Counter()
    for r in rows:
        n = r["note"].lower()
        ns = names(items_of(r.get("build", "")))
        if r["role"] == "JUNGLE" and has(ns, "Mosstomper"):
            note_pat["mosstomper_jg"] += 1
        if "zhonya" in n and ("2nd" in n or "before" in n):
            note_pat["zhonya_early"] += 1
        if r["champ"] == "Cassiopeia" and has(ns, "Sorcerer's Shoes"):
            note_pat["cass_boots"] += 1
        if r["champ"] == "Pyke" and r["role"] == "UTILITY" and has(ns, "Locket"):
            note_pat["pyke_locket"] += 1
        if r["champ"] == "Ezreal" and has(ns, "Kraken"):
            note_pat["ez_kraken"] += 1
        if r["role"] == "UTILITY" and has(ns, "Blackfire") and r["score"] <= 3:
            note_pat["tank_mage_supp"] += 1
        if r["role"] == "UTILITY" and has(ns, "Moonstone") and r["score"] <= 3:
            note_pat["tank_enchanter"] += 1
        if "trinity" in n and r["score"] <= 4:
            note_pat["wrong_trinity"] += 1
        if r["score"] <= 2:
            note_pat["fatal"] += 1
        if r["score"] <= 4:
            note_pat["broken"] += 1

    best = sorted(rows, key=lambda r: (-r["score"], r["id"]))[:25]
    worst = sorted(rows, key=lambda r: (r["score"], r["id"]))[:40]
    avg = mean(scores)
    med = median(scores)

    def pct(n: int) -> str:
        return f"{n} ({100.0 * n / 1000:.1f}%)"

    # Champion table
    champ_rows = []
    for ch, lst in by_champ.items():
        ss = [x["score"] for x in lst]
        notes = "; ".join(sorted({x["note"][:70] for x in lst if x["score"] <= 4})[:2])
        champ_rows.append(
            (
                mean(ss),
                len(lst),
                ch,
                median(ss),
                max(ss),
                min(ss),
                notes or "нет ломающих кейсов",
            )
        )
    champ_rows.sort(key=lambda x: (-x[0], x[2]))

    cr_rows = []
    for (ch, role), ss in sorted(by_cr.items()):
        cr_rows.append((ch, role, len(ss), mean(ss), min(ss), max(ss)))
    cr_rows.sort(key=lambda x: (x[1], -x[3], x[0]))

    # Write case details
    cards = [
        "# Детальный ручной аудит каждой рекомендации",
        "",
        "Обработано кейсов: **1000 / 1000**. Пропущено: **0**. Ошибок парсинга: **0**.",
        "",
        "Оценка — независимый Emerald+ вердикт по замороженному `recommend_report.md` (движок не перезапускался).",
        "Список трактуется как **строгий порядок покупки** на 6 слотов: старт продаётся, компоненты апгрейдятся, вторая пара ботов запрещена.",
        "Off-meta пик не штрафуется: itemization оценивается под заявленную роль.",
        "",
        f"Средний балл по ручным вердиктам: **{avg:.2f}**. Медиана: **{med:.1f}**.",
        "",
    ]
    for r in rows:
        cards.append(card(r["id"], r, r["score"], r["note"]))
        cards.append("")
    (LIVE / "build_recommendation_case_details.md").write_text(
        "\n".join(cards), encoding="utf-8"
    )

    # Audit
    a: list[str] = []
    a += [
        "# League of Legends Build Recommendation Audit",
        "",
        "## 1. Executive Summary",
        "",
        "Независимый **ручной** Emerald+ аудит **всех 1000** рекомендаций из `recommend_report.md` против live-JSON (старт: 500g, lvl 1, у врагов нет легендарик). Движок **не перезапускался**. Каждая карточка просмотрена человеком; балл — строгий вердикт, не рубрика-скрипт.",
        "",
        "**Правила (зафиксированы до анализа):**",
        "- список = строгий buy order на 6 слотов;",
        "- off-meta не штрафуется, itemization под заявленную роль;",
        "- ground truth — эксперт Emerald+, не seed движка;",
        "- патч данных **16.18.1**;",
        "- пользователь overlay, не challenger-микротюнинг.",
        "",
        f"**Главный вывод:** средний балл **{avg:.2f}/10**, медиана **{med:.1f}**. "
        f"Доля 8–10: **{pct(sum(1 for s in scores if s >= 8))}**. "
        f"Доля ≤4 (ломает чемпиона): **{pct(sum(1 for s in scores if s <= 4))}**. "
        f"Доля ≤2 (катастрофа identity): **{pct(sum(1 for s in scores if s <= 2))}**.",
        "",
        "Система выдаёт **узнаваемый классовый шаблон**, не champion×role itemizer. "
        "Если купить список как есть, типичный исход — правильный тип урона и чужой core/пет/саппорт-квест. "
        "Это не «рабочий, но дырявый билд на 6»: на десятках чемпионов это **другая игра** (Pyke в Locket, Ezreal в Kraken, Cass в ботах, Thresh в Moonstone, Ivern в Malignance).",
        "",
        "JSON: 1000. Секций отчёта: 1000. Успешно разобрано вручную: 1000. Пропущено: 0.",
        "",
        "## 2. Overall System Performance",
        "",
        "| Метрика | Значение |",
        "|---|---|",
        f"| Кейсов оценено | 1000 |",
        f"| Средняя оценка | {avg:.2f} |",
        f"| Медиана | {med:.1f} |",
        f"| Мин / макс | {min(scores)} / {max(scores)} |",
        f"| 9–10 | {pct(h[9]+h[10])} |",
        f"| 7–8 | {pct(h[7]+h[8])} |",
        f"| 5–6 | {pct(h[5]+h[6])} |",
        f"| 3–4 | {pct(h[3]+h[4])} |",
        f"| 0–2 | {pct(h[0]+h[1]+h[2])} |",
        "",
        "Интерпретация: **черновой class-seed planner**. Live-сигналы в фикстурах почти пустые, поэтому даже draft-план часто не содержит обязательного identity-кора чемпиона.",
        "",
        "## 3. Score Distribution",
        "",
        "| Score | Count | % |",
        "|---|---:|---:|",
    ]
    for i in range(11):
        a.append(f"| {i} | {h[i]} | {100.0 * h[i] / 1000:.1f}% |")

    a += [
        "",
        "## 4. Performance by Champion",
        "",
        "Чемпионы с n<4 помечены *(малая выборка)*.",
        "",
        "| Champion | Samples | Avg | Median | Best | Worst | Typical fail |",
        "|---|---:|---:|---:|---:|---:|---|",
    ]
    for avg_s, n, ch, med_s, mx, mn, notes in champ_rows:
        tag = " (малая выборка)" if n < 4 else ""
        a.append(
            f"| {ch}{tag} | {n} | {avg_s:.2f} | {med_s:.1f} | {mx} | {mn} | {notes.replace('|','/')} |"
        )

    a += [
        "",
        "## 5. Performance by Role",
        "",
        "| Role | Samples | Avg | Median | ≤4 | ≥8 |",
        "|---|---:|---:|---:|---:|---:|",
    ]
    for role in ["TOP", "JUNGLE", "MIDDLE", "BOTTOM", "UTILITY"]:
        ss = by_role[role]
        a.append(
            f"| {role} ({ROLE_RU[role]}) | {len(ss)} | {mean(ss):.2f} | {median(ss):.1f} | "
            f"{sum(1 for x in ss if x <= 4)} | {sum(1 for x in ss if x >= 8)} |"
        )

    a += [
        "",
        "## 6. Champion × Role",
        "",
        "Каждая пара champion×role, которая встретилась в 1000 кейсах. Это главный срез: один чемпион в топе и в саппорте — разные игры.",
        "",
        "| Champion | Role | n | Avg | Min | Max |",
        "|---|---|---:|---:|---:|---:|",
    ]
    for ch, role, n, av, mn, mx in cr_rows:
        a.append(f"| {ch} | {role} | {n} | {av:.2f} | {mn} | {mx} |")

    a += [
        "",
        "## 7. Best Recommendations",
        "",
        "Строгие 8+: identity совпала, порядок не ломает спайк.",
        "",
        "| ID | Champ | Role | Score | Почему прошло |",
        "|---|---|---|---:|---|",
    ]
    for r in best:
        if r["score"] < 8:
            continue
        a.append(f"| {r['id']:04d} | {r['champ']} | {r['role']} | {r['score']} | {r['note']} |")

    a += [
        "",
        "## 8. Worst Recommendations",
        "",
        "Строгий низ: пользователь купит **чужой** чемпион.",
        "",
        "| ID | Champ | Role | Score | Вердикт |",
        "|---|---|---|---:|---|",
    ]
    for r in worst:
        a.append(f"| {r['id']:04d} | {r['champ']} | {r['role']} | {r['score']} | {r['note']} |")

    a += [
        "",
        "## 9. Archetype / seed failures",
        "",
        "Повтор одних и тех же шаблонов на неподходящих чемпионах:",
        "",
        "- **fighter/jungle → Mosstomper + Trinity + Sundered + Cleaver + FoN + Visage** на Yone, Yasuo, Yi, Tryndamere, Riven, Lee Sin, Vi, Jax, Hecarim, Kayn, Ambessa, Fiora. Это танк-брузер, не их игра.",
        "- **assassin/ap → Stormsurge + Zhonya 2nd + Sorcs** на Sylas, Kassadin, Diana (часто нужен Nashor/RoA), Lillia (Liandry).",
        "- **marksman → Kraken + IE + RFC + Collector + LDR** на Ezreal, Corki, Graves, Kayle, Jhin, Smolder, Kog'Maw, Vayne (on-hit), Senna ADC.",
        "- **support mage → Lost Chapter + Blackfire** на Maokai, Nautilus, Amumu, Nunu, Tahm Kench, Zac, Malphite, Sejuani, Gragas, Lulu.",
        "- **support enchanter → Moonstone** на Thresh, Shen, Braum, Taric, Rakan, Bard.",
        "- **support engage → Locket** на Pyke (нужен lethality) и часто на Pantheon support (Eclipse/Umbral).",
        "- **Cassiopeia + Sorcerer's Shoes** — чемпион **не может купить боты**. Это не situational miss, это сломанный item gate.",
        "",
        "## 10. Situational itemization",
        "",
        "Даже когда identity верная, ситуативка слабая:",
        "",
        "- Zhonya часто **вторая** легендарка до обувки и второго оффенсива (маги/ассасины).",
        "- Steelcaps/Mercs как **next item на 500g** у ADC и мидеров.",
        "- Dual MR (FoN + Visage) в 6 слотах при 0–1 AP.",
        "- Anti-heal часто 7–8-й слот или отсутствует, хотя в драфте Senna/Soraka/Yuumi/Milio.",
        "- Heartsteel почти не выдаётся Mundo; Iceborn часто отсутствует на K'Sante; Nashor отсутствует на Kayle/Azir/Gwen/Diana.",
        "- Jungle pet в замороженном отчёте почти всегда Mosstomper, включая AP-ассасинов и crit-файтов.",
        "",
        f"Подсчёт по ручным заметкам / спискам: Mosstomper в jungle-списке **{note_pat['mosstomper_jg']}** раз; "
        f"ранняя Zhonya отмечена **{note_pat['zhonya_early']}**; "
        f"Cass+boots **{note_pat['cass_boots']}**; "
        f"Pyke+Locket **{note_pat['pyke_locket']}**; "
        f"Ezreal+Kraken **{note_pat['ez_kraken']}**; "
        f"tank/engage support как Blackfire (балл ≤3) **{note_pat['tank_mage_supp']}**; "
        f"танк в Moonstone (балл ≤3) **{note_pat['tank_enchanter']}**.",
        "",
        "## 11. Systematic Errors",
        "",
        "| Паттерн | Зачем это ломает | Что чинить |",
        "|---|---|---|",
        "| Один seed на класс | Trinity/Stormsurge/Kraken/Blackfire/Locket/Moonstone раздаются пачками | Champion×role override table (~40 уникальных itemizer'ов) |",
        "| Mosstomper почти на всех лесниках | Yone/Yi/Kha'Zix теряют клир и темп | `junglePetFor`: AD carry/assassin → Gustwalker, AP → Scorchclaw, tank/engage → Mosstomper. Не мапить juggernaut на tank-pet |",
        "| Support tagging по «есть AP в kit» | Maokai/Nautilus/Lulu/Thresh получают чужой квест | Role function: engage / enchanter / mage poke / lethality hook — отдельные seeds |",
        "| Нет item gates | Cass покупает боты; Vladimir получает Lost Chapter | Hard bans: Cass no boots; Vladimir no Lost Chapter; Yuumi no boots sometimes; Kayn form split |",
        "| Нет 6-slot emit | Overlay отдаёт 8–10 предметов | Start + 1 boots + 5 legendaries, exclusive groups |",
        "| Zhonya/boots опережают core | Нет first-item спайка на 500g | Defensive.priority < core; next≠boots при gold<800 (кроме support quest) |",
        "| Dual MR / dual boots | Строгий порядок покупает оба | Exclusive groups как lethality cores |",
        "| Live blend на t=0 | `legendaries=0`, но приоритеты уже сдвинуты | На старте live×0; counters только из kit-тегов драфта |",
        "",
        "## 12. Missing Logic",
        "",
        "1. Champion idiosyncrasy table (Cass no boots, Ezreal Manamune/Trinity, Jhin не Kraken, Zeri Stormrazor/Runaan, Kog Guinsoo, Smolder ER, GP ER, Kayle Nashor, Azir Nashor, Gwen Nashor, Sylas RoA, Kassadin Tear/RoA, Vladimir Cosmic, Pyke lethality, Ivern enchanter, Senna lethality/Cleaver).",
        "2. Support function, не class: Locket vs Moonstone vs Blackfire vs Youmuu.",
        "3. Jungle pet matrix по kit, не по грубому tank-флагу.",
        "4. Buy-path planner: 6 слотов, апгрейд компонентов, одна пара ботов.",
        "5. Exclusive groups: boots, grievous, last whisper, MR legendaries, crit vs on-hit.",
        "6. Mandatory counters в топ-6: grievous vs 2 healers; LW vs 2 tanks; vis vs 3 AP.",
        "7. Ally-need: единственный фронт / единственный AP.",
        "8. Order: core before Zhonya; component before boots at 500g.",
        "9. Kayn/Shyvana/Udyr form splits; Poppy×role (top Iceborn, jg Sundered/Iceborn, supp Locket).",
        "10. Support quest items (Zaz'Zak / Celestial / Bloodsong) как часть кора.",
        "",
        "## 13. Recommended Algorithm Improvements",
        "",
        "Приоритет по ожидаемому приросту среднего балла (сейчас ~{:.1f}):".format(avg),
        "",
        "1. **Champion×role core table.** Без этого Pyke/Ezreal/Cass/Thresh/Ivern всегда будут 2–3. Самый большой прыжок среднего.",
        "2. **Support router rewrite:** lethality (Pyke) / engage tank (Leona/Nautilus/Maokai) / enchanter (Lulu/Janna) / mage poke (Brand/Xerath). Не «есть AP → Blackfire».",
        "3. **Jungle pet rewrite** в эмите overlay (в коде уже лучше, в замороженном отчёте — нет).",
        "4. **6-slot collapse + exclusive groups.**",
        "5. **Item gates:** Cass cannot buy boots; no Lost Chapter on Vladimir; no Kraken-first on Jhin/Ezreal/Corki.",
        "6. **Zhonya/boots cap** до первого оффенсив-легендарика.",
        "7. **Required counter slots** (grievous/LW/MR) вместо +priority к хвосту.",
        "8. **ADC family split:** crit / on-hit / poke-mana / lethality / hybrid (Kai'Sa/Kayle).",
        "9. **Fighter family split:** Trinity sheen / Eclipse lethality / Cleaver juggernaut / Iceborn tank-fighter / ER pirate.",
        "10. **Регрессионные тесты на фикстурах:** Pyke supp ≠ Locket; Ezreal ≠ Kraken; Cass ≠ Sorcs; Yone jg ≠ Mosstomper; Lulu ≠ Blackfire; Ivern jg ≠ Malignance.",
        "",
        "## 14. Final Verdict",
        "",
        f"**Общая оценка системы: {avg:.1f}/10 — шаблон класса узнаваем, как buy order в конкретной игре ненадёжен. На уникальных чемпионах — опасен.**",
        "",
        "Шкала не завышена: 8+ только там, где seed случайно совпал с identity (Locket-engage, Moonstone-enchanter, Youmuu-assassin, Iceborn-Poppy, Riftmaker-Morde, Sunfire-tank). "
        f"{pct(sum(1 for s in scores if s <= 4))} кейсов — пользователь получит **не тот чемпион**.",
        "",
        "Что относительно надёжно:",
        "- damage type по широкому классу (маг не получает Trinity как mid-core в типичном кейсе);",
        "- Locket на классическом engage и Moonstone на классическом enchanter;",
        "- Youmuu на AD-ассасинах;",
        "- Sunfire на чистых танках.",
        "",
        "Чему нельзя доверять в катке:",
        "- jungle pet;",
        "- любой «нестандартный» чемпион (Ezreal, Cass, Pyke, Ivern, Kayle, Corki, GP, Jhin, Sylas, Vladimir, Thresh, Lulu, Rakan);",
        "- support tagging;",
        "- порядок Zhonya/boots vs core;",
        "- 6-slot path.",
        "",
        "Ответ на вопрос *«Если пользователь доверится рекомендации именно в этой ситуации, насколько она разумна?»*: "
        f"в среднем **{avg:.1f}/10**. На Leona/Janna/Morde/Poppy-top — можно брать как черновик. "
        "На Pyke/Ezreal/Cass/Yone-jungle/Ivern — **нельзя**. Overlay в нынешнем виде безопаснее показывать как «класс предметов», а не как строгий шоппинг-лист.",
        "",
        "---",
        "",
        "### Статус обработки",
        "",
        "| | N |",
        "|---|---:|",
        "| JSON fixtures | 1000 |",
        "| Секций recommend_report.md | 1000 |",
        "| Ручных вердиктов | 1000 |",
        "| Пропущено | 0 |",
        "",
        "Карточки всех 1000 кейсов: [`build_recommendation_case_details.md`](build_recommendation_case_details.md)",
        "",
        "Источник баллов: [`_manual_verdicts.txt`](_manual_verdicts.txt) (ручные строгие вердикты, не скриптовая рубрика).",
        "",
    ]

    (LIVE / "build_recommendation_audit.md").write_text("\n".join(a), encoding="utf-8")

    stats = {
        "n": 1000,
        "avg": round(avg, 2),
        "median": med,
        "min": min(scores),
        "max": max(scores),
        "hist": h,
        "le4": sum(1 for s in scores if s <= 4),
        "le2": sum(1 for s in scores if s <= 2),
        "ge8": sum(1 for s in scores if s >= 8),
        "patterns": dict(note_pat),
        "by_role": {k: round(mean(v), 2) for k, v in by_role.items()},
    }
    (LIVE / "_audit_stats.json").write_text(
        json.dumps(stats, ensure_ascii=False, indent=2), encoding="utf-8"
    )
    print(json.dumps(stats, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
