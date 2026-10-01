using System;
using System.Collections;
using System.Collections.Generic;
using System.IO;
using System.Net;
using System.Text;
using System.Web.Script.Serialization;

namespace LolBuildOverlay
{
    public static class SpeShuClient
    {
        const string Endpoint = "https://speshu.ai/api/v1/chat/completions";

        public static EngineResult Recommend(string file, string styleLabel, string plan, MatchState match, bool explainItems, string wish, string damage)
        {
            if (string.IsNullOrEmpty(AppSettings.SpeShuKey))
                return Fail("в settings.ini нет speshu_key");
            if (string.IsNullOrEmpty(AppSettings.SpeShuModel))
                return Fail("в settings.ini нет speshu_model");

            string system;
            string user;
            try
            {
                system = ReadPrompt("system.txt");
                user = ReadPrompt(file + ".txt");
            }
            catch (Exception ex)
            {
                return Fail(ex.Message);
            }

            user = user.Replace("{{STYLE}}", styleLabel ?? "")
                       .Replace("{{PLAN}}", string.IsNullOrEmpty(plan) ? "ещё нет" : plan)
                       .Replace("{{LIVE}}", LiveText(match));
            user += explainItems
                ? "\n\nРежим: предметы. Эта строка главнее абзацев про текст. play оставь пустым массивом. У каждого из 6 предметов why: имя, зачем он в этой сборке и против кого из снимка."
                : "\n\nРежим: игра. Эта строка главнее абзацев про why. В play дай 4–6 коротких строк, как играть эту катку до 7-й минуты: роль, с кем размениваться и чего не делать. why у каждого предмета оставь пустой строкой. Шесть предметов, ботинки и early всё равно заполни.";
            if (damage == "ad")
                user += "\n\nТип урона: физический. Все 6 предметов и путь до них собираются от силы атаки этого чемпиона.";
            else if (damage == "ap")
                user += "\n\nТип урона: магический. Все 6 предметов и путь до них собираются от силы умений этого чемпиона.";
            if (!string.IsNullOrEmpty(wish))
                user += "\n\nПожелание игрока, его надо выполнить: " + wish + ". Чемпион и роль остаются из снимка. Сборка должна быть сильной на этом герое и при этом делать то, что он просит.";

            try
            {
                var content = Complete(file, system, user);
                var planJson = ExtractObject(content);
                var build = ParseBuild(planJson, match);
                if (build == null) return Fail("модель вернула не сборку");
                return new EngineResult
                {
                    Build = build,
                    Reasons = explainItems ? (build.SlotReasons ?? new string[0]) : new string[0],
                    Play = explainItems ? new string[0] : (build.Play ?? new string[0]),
                    SeedName = styleLabel,
                    NextName = NextName(build),
                    FromLive = match != null && match.InGame
                };
            }
            catch (Exception ex)
            {
                return Fail(ex.Message);
            }
        }

        public static string PlanText(RecommendedBuild build)
        {
            if (build == null) return "";
            var parts = new List<string>();
            if (build.Items != null)
            {
                for (var i = 0; i < build.Items.Length; i++)
                {
                    var name = ItemNames.Get(build.Items[i]);
                    if (!string.IsNullOrEmpty(name)) parts.Add((i + 1) + ". " + name);
                }
            }
            if (build.Early != null && build.Early.Length > 0)
            {
                var early = new List<string>();
                foreach (var id in build.Early)
                {
                    var name = ItemNames.Get(id);
                    if (!string.IsNullOrEmpty(name)) early.Add(name);
                }
                if (early.Count > 0) parts.Add("до легендарки: " + string.Join(" → ", early.ToArray()));
            }
            if (!string.IsNullOrEmpty(build.Boots))
                parts.Add("ботинки: " + ItemNames.Get(build.Boots));
            if (build.NextIndex >= 0 && build.NextIndex < 6 && build.Items != null && build.NextIndex < build.Items.Length)
                parts.Add("следующий: " + ItemNames.Get(build.Items[build.NextIndex]));
            else if (build.NextIndex == 6)
                parts.Add("следующий: " + ItemNames.Get(build.Boots));
            return string.Join("; ", parts.ToArray());
        }

        static string ReadPrompt(string name)
        {
            var path = Path.Combine(AppPaths.PromptsDir, name);
            if (!File.Exists(path))
                throw new FileNotFoundException("нет файла " + path);
            return File.ReadAllText(path, Encoding.UTF8);
        }

        static string Complete(string file, string system, string user)
        {
            var ser = new JavaScriptSerializer { MaxJsonLength = int.MaxValue };
            var body = ser.Serialize(new Dictionary<string, object>
            {
                { "model", AppSettings.SpeShuModel },
                { "temperature", 0.2 },
                { "stream", true },
                { "reasoning", new Dictionary<string, object> { { "effort", "low" } } },
                { "response_format", new Dictionary<string, object> { { "type", "json_object" } } },
                { "messages", new object[]
                    {
                        new Dictionary<string, object> { { "role", "system" }, { "content", system } },
                        new Dictionary<string, object> { { "role", "user" }, { "content", user } }
                    }
                }
            });

            var req = (HttpWebRequest)WebRequest.Create(Endpoint);
            req.Method = "POST";
            req.ContentType = "application/json";
            req.Headers["Authorization"] = "Bearer " + AppSettings.SpeShuKey;
            req.Timeout = 90000;
            req.ReadWriteTimeout = 90000;
            req.AllowReadStreamBuffering = false;
            var bytes = Encoding.UTF8.GetBytes(body);
            req.ContentLength = bytes.Length;
            using (var stream = req.GetRequestStream())
                stream.Write(bytes, 0, bytes.Length);

            try
            {
                using (var resp = (HttpWebResponse)req.GetResponse())
                using (var stream = resp.GetResponseStream())
                {
                    var content = ReadAnswer(ser, stream);
                    Log(file, content);
                    return content;
                }
            }
            catch (WebException ex)
            {
                var detail = "";
                if (ex.Response != null)
                {
                    using (var reader = new StreamReader(ex.Response.GetResponseStream(), Encoding.UTF8))
                        detail = reader.ReadToEnd();
                }
                Log(file, string.IsNullOrEmpty(detail) ? ex.Message : detail);
                if (detail.Length > 180) detail = detail.Substring(0, 180);
                throw new InvalidOperationException(string.IsNullOrEmpty(detail) ? ex.Message : detail);
            }
        }

        static void Log(string file, string text)
        {
            try
            {
                var path = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "model.log");
                var block = DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss") + " " + file
                    + Environment.NewLine + (text ?? "") + Environment.NewLine + Environment.NewLine;
                File.AppendAllText(path, block, Encoding.UTF8);
            }
            catch { }
        }

        static string ReadAnswer(JavaScriptSerializer ser, Stream stream)
        {
            var content = new StringBuilder();
            try
            {
                using (var reader = new StreamReader(stream, Encoding.UTF8))
                {
                    string line;
                    while ((line = reader.ReadLine()) != null)
                    {
                        line = line.Trim();
                        if (line.Length == 0) continue;
                        if (line[0] == '{')
                        {
                            var raw = line.EndsWith("}") ? line : line + reader.ReadToEnd();
                            return ContentOf(ser.DeserializeObject(raw));
                        }
                        if (!line.StartsWith("data:")) continue;
                        var data = line.Substring(5).Trim();
                        if (data == "[DONE]") break;
                        if (data.Length == 0 || data[0] != '{') continue;
                        try { content.Append(PieceOf(ser.DeserializeObject(data))); }
                        catch { }
                    }
                }
            }
            catch (Exception ex)
            {
                if (content.Length == 0 || content.ToString().IndexOf('{') < 0) throw;
                if (!(ex is IOException) && !(ex is WebException)) throw;
            }
            if (content.Length == 0) throw new InvalidOperationException("пустой ответ SpeShu");
            return content.ToString();
        }

        static string PieceOf(object parsed)
        {
            var doc = parsed as Dictionary<string, object>;
            if (doc == null || !doc.ContainsKey("choices")) return "";
            var choices = AsItems(doc["choices"]);
            if (choices.Count == 0) return "";
            var choice = choices[0] as Dictionary<string, object>;
            if (choice == null) return "";
            var delta = choice.ContainsKey("delta") ? choice["delta"] as Dictionary<string, object> : null;
            var piece = TextField(delta, "content");
            if (piece.Length > 0) return piece;
            var message = choice.ContainsKey("message") ? choice["message"] as Dictionary<string, object> : null;
            return TextField(message, "content");
        }

        static string TextField(Dictionary<string, object> obj, string key)
        {
            if (obj == null || !obj.ContainsKey(key) || obj[key] == null) return "";
            var text = obj[key].ToString();
            return text ?? "";
        }

        static string ContentOf(object root)
        {
            var doc = root as Dictionary<string, object>;
            if (doc == null || !doc.ContainsKey("choices")) throw new InvalidOperationException("пустой ответ SpeShu");
            var choices = AsItems(doc["choices"]);
            if (choices.Count == 0) throw new InvalidOperationException("пустой ответ SpeShu");
            var choice = choices[0] as Dictionary<string, object>;
            if (choice == null || !choice.ContainsKey("message")) throw new InvalidOperationException("пустой ответ SpeShu");
            var message = choice["message"] as Dictionary<string, object>;
            if (message == null || !message.ContainsKey("content") || message["content"] == null)
                throw new InvalidOperationException("пустой ответ SpeShu");
            return message["content"].ToString();
        }

        static Dictionary<string, object> ExtractObject(string content)
        {
            if (string.IsNullOrEmpty(content)) throw new InvalidOperationException("модель ничего не ответила");
            var start = content.IndexOf('{');
            var end = content.LastIndexOf('}');
            if (start < 0 || end <= start) throw new InvalidOperationException("в ответе нет JSON");
            var ser = new JavaScriptSerializer { MaxJsonLength = int.MaxValue };
            var obj = ser.DeserializeObject(content.Substring(start, end - start + 1)) as Dictionary<string, object>;
            if (obj == null) throw new InvalidOperationException("в ответе нет JSON");
            return obj;
        }

        static RecommendedBuild ParseBuild(Dictionary<string, object> obj, MatchState match)
        {
            var items = new List<string>();
            var whys = new List<string>();
            if (obj.ContainsKey("items") && obj["items"] != null)
            {
                foreach (var raw in AsItems(obj["items"]))
                {
                    var id = "";
                    var why = "";
                    var row = raw as Dictionary<string, object>;
                    if (row != null)
                    {
                        if (row.ContainsKey("id") && row["id"] != null) id = Resolve(row["id"].ToString());
                        else if (row.ContainsKey("item") && row["item"] != null) id = Resolve(row["item"].ToString());
                        if (row.ContainsKey("why") && row["why"] != null) why = row["why"].ToString().Trim();
                    }
                    else
                    {
                        id = Resolve(raw == null ? "" : raw.ToString());
                    }
                    if (id.Length == 0 || items.Contains(id)) continue;
                    items.Add(id);
                    whys.Add(why);
                }
            }
            var bootsEarly = PullBoots(items, whys);
            DropExclusive(items, whys);
            if (items.Count > 6)
            {
                items.RemoveRange(6, items.Count - 6);
                if (whys.Count > 6) whys.RemoveRange(6, whys.Count - 6);
            }
            var reasons = AlignReasons(items, whys, ReasonsOf(obj));
            while (items.Count < 6) items.Add("");
            var boots = Resolve(obj.ContainsKey("boots") && obj["boots"] != null ? obj["boots"].ToString() : "");
            if (boots.Length == 0) boots = bootsEarly;
            var footwear = match != null && match.MagicalFootwear;
            var early = new List<string>();
            foreach (var raw in AsList(obj, "early"))
            {
                var id = Resolve(raw);
                if (id.Length == 0) continue;
                if (footwear && id == "1001") id = "2422";
                if (items.Contains(id) || id == boots) continue;
                early.Add(id);
            }
            var next = 0;
            if (obj.ContainsKey("nextIndex") && obj["nextIndex"] != null)
                int.TryParse(Convert.ToString(obj["nextIndex"]), out next);
            if (next < 0 || next > 6) next = 0;
            if (items[0].Length == 0 && boots.Length == 0 && early.Count == 0) return null;
            var earlyNext = -1;
            var owned = Owned(match);
            for (var i = 0; i < early.Count; i++)
            {
                if (!owned.Contains(early[i])) { earlyNext = i; break; }
            }
            return new RecommendedBuild
            {
                Items = items.ToArray(),
                SlotReasons = reasons.ToArray(),
                Play = PlayOf(obj),
                Boots = boots,
                NextIndex = next,
                Early = early.ToArray(),
                EarlyNext = earlyNext
            };
        }

        static List<string> AsList(Dictionary<string, object> obj, string key)
        {
            var list = new List<string>();
            if (!obj.ContainsKey(key) || obj[key] == null) return list;
            foreach (var item in AsItems(obj[key]))
                list.Add(item == null ? "" : item.ToString());
            return list;
        }

        static ArrayList AsItems(object raw)
        {
            var list = raw as ArrayList;
            if (list != null) return list;
            var copy = new ArrayList();
            var arr = raw as object[];
            if (arr != null) copy.AddRange(arr);
            return copy;
        }

        static HashSet<string> Owned(MatchState match)
        {
            var owned = new HashSet<string>();
            if (match == null || match.Me == null || match.Me.items == null) return owned;
            foreach (var item in match.Me.items)
            {
                if (item != null && item.itemID > 0) owned.Add(item.itemID.ToString());
            }
            return owned;
        }

        static readonly string[] BootIds = new[]
        {
            "1001", "2422", "3005", "3047", "3008", "3006", "3009", "3010", "3020", "3111", "3117", "3158"
        };

        static readonly string[][] ExclusiveGroups = new[]
        {
            new[] { "3033", "3036", "6694" },
            new[] { "3003", "3004", "3040", "3042", "2526", "2530", "3119", "3121" },
            new[] { "3078", "3100", "2510", "3508", "6662" },
            new[] { "3074", "3748", "6698", "6631" },
            new[] { "3068", "6664" },
            new[] { "3139", "6035" },
            new[] { "3135", "3137" },
        };

        static string PullBoots(List<string> items, List<string> whys)
        {
            var found = "";
            var keep = new List<string>();
            var keepWhy = new List<string>();
            for (var i = 0; i < items.Count; i++)
            {
                if (Array.IndexOf(BootIds, items[i]) >= 0)
                {
                    if (found.Length == 0 && items[i] != "1001" && items[i] != "2422") found = items[i];
                    continue;
                }
                keep.Add(items[i]);
                keepWhy.Add(i < whys.Count ? whys[i] : "");
            }
            items.Clear();
            items.AddRange(keep);
            whys.Clear();
            whys.AddRange(keepWhy);
            return found;
        }

        static void DropExclusive(List<string> items, List<string> whys)
        {
            var keep = new List<string>();
            var keepWhy = new List<string>();
            for (var i = 0; i < items.Count; i++)
            {
                if (Clashes(items[i], keep)) continue;
                keep.Add(items[i]);
                keepWhy.Add(i < whys.Count ? whys[i] : "");
            }
            items.Clear();
            items.AddRange(keep);
            whys.Clear();
            whys.AddRange(keepWhy);
        }

        static bool Clashes(string id, List<string> kept)
        {
            foreach (var group in ExclusiveGroups)
            {
                if (Array.IndexOf(group, id) < 0) continue;
                foreach (var other in kept)
                {
                    if (Array.IndexOf(group, other) >= 0) return true;
                }
            }
            return false;
        }

        static List<string> AlignReasons(List<string> ids, List<string> whys, string[] loose)
        {
            var used = new bool[loose == null ? 0 : loose.Length];
            var lines = new List<string>();
            for (var i = 0; i < ids.Count; i++)
            {
                if (ids[i].Length == 0) continue;
                var line = "";
                if (i < whys.Count && Fits(whys[i], ids[i], true)) line = whys[i];
                if (line.Length == 0 && loose != null)
                {
                    for (var r = 0; r < loose.Length; r++)
                    {
                        if (used[r] || !Fits(loose[r], ids[i], false)) continue;
                        line = loose[r];
                        used[r] = true;
                        break;
                    }
                }
                if (line.Length > 0) lines.Add(line);
            }
            return lines;
        }

        static bool Fits(string line, string id, bool allowUnnamed)
        {
            if (string.IsNullOrEmpty(line) || string.IsNullOrEmpty(id)) return false;
            var head = line.Trim();
            var cut = head.IndexOf(':');
            if (cut < 0) cut = head.IndexOf('—');
            if (cut > 0) head = head.Substring(0, cut).Trim();
            var named = ItemNames.FindByHead(head);
            if (named.Length > 0) return named == id;
            return allowUnnamed;
        }

        static string[] PlayOf(Dictionary<string, object> obj)
        {
            if (!obj.ContainsKey("play") || obj["play"] == null) return new string[0];
            var text = obj["play"] as string;
            if (text != null)
            {
                var parts = text.Split(new[] { '\n' }, StringSplitOptions.RemoveEmptyEntries);
                var lines = new List<string>();
                foreach (var part in parts)
                {
                    var line = part.Trim();
                    if (line.Length > 0) lines.Add(line);
                }
                return lines.ToArray();
            }
            var list = new List<string>();
            foreach (var item in AsItems(obj["play"]))
            {
                if (item == null) continue;
                var line = item.ToString().Trim();
                if (line.Length > 0) list.Add(line);
            }
            return list.ToArray();
        }

        static string[] ReasonsOf(Dictionary<string, object> obj)
        {
            if (!obj.ContainsKey("reasons")) return new string[0];
            var list = new List<string>();
            foreach (var item in AsItems(obj["reasons"]))
            {
                if (item == null) continue;
                var line = item.ToString().Trim();
                if (line.Length > 0) list.Add(line);
            }
            return list.ToArray();
        }

        static string Resolve(string token)
        {
            if (string.IsNullOrEmpty(token)) return "";
            token = token.Trim();
            if (ItemNames.Knows(token)) return token;
            return ItemNames.FindId(token);
        }

        static string NextName(RecommendedBuild build)
        {
            if (build == null) return "";
            if (build.NextIndex == 6) return ItemNames.Get(build.Boots);
            if (build.NextIndex >= 0 && build.Items != null && build.NextIndex < build.Items.Length)
                return ItemNames.Get(build.Items[build.NextIndex]);
            return "";
        }

        public static string LiveText(MatchState match)
        {
            if (match == null || match.Me == null || !match.InGame)
                return "Матч не идёт. Live Client недоступен.";
            var sb = new StringBuilder();
            var sec = (int)match.GameTime;
            sb.Append("Время ").Append(sec / 60).Append(":").Append((sec % 60).ToString("00"));
            var role = RoleName(match.Me);
            if (role.Length > 0)
                sb.Append("\nРоль: ").Append(role).Append(". Сборка и early с первого ответа под эту роль, не под привычную линию чемпиона.");
            else
                sb.Append("\nРоль клиент ещё не отдал. Не ставь чемпиона на мид по привычке: смотри заклинания и предметы.");
            sb.Append("\nЯ: ").Append(Line(match.Me)).Append(" золото ").Append((int)match.Gold);
            if (match.MagicalFootwear)
                sb.Append("\nРуна: Магическая обувь. Бесплатные ботинки на 12:00, в early ставь 2422, не 1001.");
            else
                sb.Append("\nРуны Магической обуви нет. В early обычные ботинки 1001 за 300 золота.");
            sb.Append("\nСоюзники:");
            AppendPlayers(sb, match.Allies);
            sb.Append("\nВраги:");
            AppendPlayers(sb, match.Enemies);
            return sb.ToString();
        }

        static void AppendPlayers(StringBuilder sb, List<LivePlayer> players)
        {
            if (players == null || players.Count == 0)
            {
                sb.Append("\n- нет данных");
                return;
            }
            foreach (var p in players)
                sb.Append("\n- ").Append(Line(p));
        }

        static string Line(LivePlayer p)
        {
            if (p == null) return "";
            var sb = new StringBuilder();
            sb.Append(string.IsNullOrEmpty(p.championName) ? "?" : p.championName);
            var role = RoleName(p);
            if (role.Length > 0) sb.Append(" ").Append(role);
            sb.Append(" ур.").Append(p.level);
            var spells = SpellList(p);
            if (spells.Length > 0) sb.Append(" заклинания: ").Append(spells);
            if (p.scores != null)
                sb.Append(" ").Append(p.scores.kills).Append("/").Append(p.scores.deaths).Append("/").Append(p.scores.assists);
            sb.Append(" предметы: ").Append(ItemList(p));
            return sb.ToString();
        }

        static string RoleName(LivePlayer p)
        {
            if (p == null) return "";
            var pos = (p.position ?? "").Trim().ToUpperInvariant();
            if (pos == "JUNGLE") return "лес";
            if (pos == "TOP") return "верх";
            if (pos == "MIDDLE" || pos == "MID") return "мид";
            if (pos == "BOTTOM") return "бот";
            if (pos == "UTILITY") return "поддержка";
            if (HasSmite(p)) return "лес";
            return "";
        }

        static bool HasSmite(LivePlayer p)
        {
            if (p == null || p.summonerSpells == null) return false;
            return IsSmite(p.summonerSpells.summonerSpellOne) || IsSmite(p.summonerSpells.summonerSpellTwo);
        }

        static bool IsSmite(LiveSpell spell)
        {
            if (spell == null) return false;
            var name = ((spell.displayName ?? "") + " " + (spell.rawDisplayName ?? "")).ToLowerInvariant();
            return name.Contains("smite") || name.Contains("кара");
        }

        static string SpellList(LivePlayer p)
        {
            if (p == null || p.summonerSpells == null) return "";
            var names = new List<string>();
            AddSpell(names, p.summonerSpells.summonerSpellOne);
            AddSpell(names, p.summonerSpells.summonerSpellTwo);
            return string.Join(", ", names.ToArray());
        }

        static void AddSpell(List<string> names, LiveSpell spell)
        {
            if (spell == null) return;
            var name = spell.displayName;
            if (string.IsNullOrEmpty(name)) name = spell.rawDisplayName;
            if (!string.IsNullOrEmpty(name)) names.Add(name);
        }

        static string ItemList(LivePlayer p)
        {
            if (p.items == null || p.items.Count == 0) return "пусто";
            var names = new List<string>();
            foreach (var item in p.items)
            {
                if (item == null || item.itemID <= 0) continue;
                var name = item.displayName;
                if (string.IsNullOrEmpty(name)) name = ItemNames.Get(item.itemID.ToString());
                if (!string.IsNullOrEmpty(name)) names.Add(name);
            }
            if (names.Count == 0) return "пусто";
            return string.Join(", ", names.ToArray());
        }

        static EngineResult Fail(string error)
        {
            return new EngineResult { Error = error };
        }
    }
}
