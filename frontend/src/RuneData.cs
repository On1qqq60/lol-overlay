using System;
using System.Collections;
using System.Collections.Generic;
using System.ComponentModel;
using System.Globalization;
using System.IO;
using System.Net;
using System.Text;
using System.Threading;
using System.Web.Script.Serialization;
using System.Windows;
using System.Windows.Media;
using System.Windows.Media.Imaging;
using System.Windows.Threading;

namespace LolBuildOverlay
{
    public class ChampRow : INotifyPropertyChanged
    {
        public string Id { get; set; }
        public string Name { get; set; }
        ImageSource _icon;
        public ImageSource Icon
        {
            get { return _icon; }
            set
            {
                _icon = value;
                var handler = PropertyChanged;
                if (handler != null) handler(this, new PropertyChangedEventArgs("Icon"));
            }
        }
        public event PropertyChangedEventHandler PropertyChanged;
        public override string ToString() { return Name; }
    }

    public static class ChampionIcons
    {
        const string Cdn = "https://ddragon.leagueoflegends.com/cdn/16.18.1/img/champion/";
        const int Workers = 6;
        static readonly Dictionary<string, BitmapImage> Map = new Dictionary<string, BitmapImage>();
        static readonly HashSet<string> Waiting = new HashSet<string>();
        static readonly Queue<ChampRow> Queue = new Queue<ChampRow>();
        static readonly List<ChampRow> Ready = new List<ChampRow>();
        static readonly object Gate = new object();
        static int _workers;
        static bool _flushPosted;
        static bool _dirReady;

        static ChampionIcons()
        {
            try
            {
                ServicePointManager.SecurityProtocol =
                    (SecurityProtocolType)3072 | (SecurityProtocolType)768 | (SecurityProtocolType)192;
                ServicePointManager.DefaultConnectionLimit = 8;
            }
            catch { }
        }

        public static void Preload(List<ChampRow> rows)
        {
            if (rows == null) return;
            foreach (var row in rows) Request(row);
        }

        public static void Request(ChampRow row)
        {
            if (row == null || string.IsNullOrEmpty(row.Id)) return;
            BitmapImage cached = null;
            lock (Gate)
            {
                if (Map.TryGetValue(row.Id, out cached) && cached != null)
                {
                }
                else
                {
                    cached = null;
                    if (!Waiting.Add(row.Id)) return;
                    Queue.Enqueue(row);
                    Kick();
                }
            }
            if (cached != null && row.Icon != cached) row.Icon = cached;
        }

        static void Kick()
        {
            while (_workers < Workers && Queue.Count > 0)
            {
                _workers++;
                ThreadPool.QueueUserWorkItem(delegate { Pump(); });
            }
        }

        static void Pump()
        {
            try
            {
                while (true)
                {
                    ChampRow row;
                    lock (Gate)
                    {
                        if (Queue.Count == 0) return;
                        row = Queue.Dequeue();
                    }
                    var bmp = Load(row.Id);
                    lock (Gate) { Waiting.Remove(row.Id); }
                    if (bmp == null) continue;
                    lock (Gate) { Map[row.Id] = bmp; }
                    Publish(row);
                }
            }
            finally
            {
                lock (Gate)
                {
                    _workers--;
                    if (_workers < 0) _workers = 0;
                    Kick();
                }
            }
        }

        static void Publish(ChampRow row)
        {
            var app = Application.Current;
            if (app == null)
            {
                Apply(row);
                return;
            }
            lock (Ready)
            {
                Ready.Add(row);
                if (_flushPosted) return;
                _flushPosted = true;
            }
            app.Dispatcher.BeginInvoke(DispatcherPriority.Background, new Action(Flush));
        }

        static void Flush()
        {
            ChampRow[] batch;
            lock (Ready)
            {
                batch = Ready.ToArray();
                Ready.Clear();
                _flushPosted = false;
            }
            foreach (var row in batch) Apply(row);
        }

        static void Apply(ChampRow row)
        {
            BitmapImage bmp;
            lock (Gate)
            {
                if (!Map.TryGetValue(row.Id, out bmp) || bmp == null) return;
            }
            if (row.Icon != bmp) row.Icon = bmp;
        }

        static BitmapImage Load(string id)
        {
            try
            {
                if (!_dirReady)
                {
                    Directory.CreateDirectory(AppPaths.ChampionIconsDir);
                    _dirReady = true;
                }
                var path = Path.Combine(AppPaths.ChampionIconsDir, id + ".png");
                if (!File.Exists(path))
                {
                    using (var wc = new WebClient())
                        wc.DownloadFile(Cdn + id + ".png", path);
                }
                return FromFile(path);
            }
            catch
            {
                return null;
            }
        }

        static BitmapImage FromFile(string path)
        {
            if (!File.Exists(path)) return null;
            try
            {
                var bmp = new BitmapImage();
                bmp.BeginInit();
                bmp.CacheOption = BitmapCacheOption.OnLoad;
                bmp.DecodePixelWidth = 36;
                bmp.UriSource = new Uri(path, UriKind.Absolute);
                bmp.EndInit();
                bmp.Freeze();
                return bmp;
            }
            catch { return null; }
        }
    }

    public class RuneInfo
    {
        public int Id;
        public string Name;
        public string Icon;
    }

    public class RuneSlot
    {
        public int Id;
        public int Play;
        public int Win;
    }

    public class RunePageView
    {
        public int PrimaryStyle;
        public int SecondaryStyle;
        public int Play;
        public int Win;
        public double PickRate;
        public List<RuneSlot> Primary = new List<RuneSlot>();
        public List<RuneSlot> Secondary = new List<RuneSlot>();
        public List<RuneSlot> Shards = new List<RuneSlot>();
    }

    public static class ChampionBook
    {
        static List<ChampRow> _all;

        public static List<ChampRow> All()
        {
            if (_all != null) return _all;
            var list = new List<ChampRow>();
            if (File.Exists(AppPaths.ChampionsFile))
            {
                string[] lines;
                try { lines = File.ReadAllLines(AppPaths.ChampionsFile, Encoding.UTF8); }
                catch { lines = new string[0]; }
                foreach (var raw in lines)
                {
                    var line = (raw ?? "").Trim();
                    if (line.Length == 0 || line[0] == '#') continue;
                    var tab = line.IndexOf('\t');
                    if (tab <= 0) continue;
                    list.Add(new ChampRow
                    {
                        Id = line.Substring(0, tab).Trim(),
                        Name = line.Substring(tab + 1).Trim()
                    });
                }
            }
            _all = list;
            return list;
        }

        public static string Slug(string name)
        {
            if (string.IsNullOrEmpty(name)) return "";
            foreach (var champ in All())
            {
                if (string.Equals(champ.Name, name, StringComparison.OrdinalIgnoreCase)
                    || string.Equals(champ.Id, name, StringComparison.OrdinalIgnoreCase))
                    return (champ.Id ?? "").ToLowerInvariant();
            }
            var key = Letters(name);
            if (key.Length == 0) return "";
            foreach (var champ in All())
            {
                if (Letters(champ.Name) == key || Letters(champ.Id) == key)
                    return (champ.Id ?? "").ToLowerInvariant();
            }
            return "";
        }

        static string Letters(string value)
        {
            var sb = new StringBuilder();
            foreach (var ch in (value ?? "").ToLowerInvariant())
            {
                if (char.IsLetterOrDigit(ch)) sb.Append(ch);
            }
            return sb.ToString();
        }
    }

    public static class RuneBook
    {
        static Dictionary<int, RuneInfo> _map;

        public static RuneInfo Get(int id)
        {
            Ensure();
            RuneInfo info;
            if (_map.TryGetValue(id, out info)) return info;
            return null;
        }

        public static string NameOf(int id)
        {
            var info = Get(id);
            if (info == null || string.IsNullOrEmpty(info.Name)) return "#" + id.ToString(CultureInfo.InvariantCulture);
            return info.Name;
        }

        static void Ensure()
        {
            if (_map != null) return;
            _map = new Dictionary<int, RuneInfo>();
            if (!File.Exists(AppPaths.RunesFile)) return;
            string[] lines;
            try { lines = File.ReadAllLines(AppPaths.RunesFile, Encoding.UTF8); }
            catch { return; }
            foreach (var raw in lines)
            {
                var line = (raw ?? "").Trim();
                if (line.Length == 0 || line[0] == '#') continue;
                var parts = line.Split('|');
                if (parts.Length < 4) continue;
                int id;
                if (!int.TryParse(parts[1], NumberStyles.Integer, CultureInfo.InvariantCulture, out id)) continue;
                _map[id] = new RuneInfo { Id = id, Name = parts[2], Icon = parts[3] };
            }
        }
    }

    public static class RuneIcons
    {
        const string Cdn = "https://ddragon.leagueoflegends.com/cdn/img/";
        static readonly Dictionary<int, BitmapImage> Map = new Dictionary<int, BitmapImage>();
        static readonly object Gate = new object();

        public static ImageSource Get(int id)
        {
            var info = RuneBook.Get(id);
            if (info == null || string.IsNullOrEmpty(info.Icon)) return null;
            BitmapImage bmp;
            if (Map.TryGetValue(id, out bmp) && bmp != null) return bmp;
            lock (Gate)
            {
                if (Map.TryGetValue(id, out bmp) && bmp != null) return bmp;
                var path = Path.Combine(AppPaths.RuneIconsDir, id.ToString(CultureInfo.InvariantCulture) + ".png");
                bmp = FromFile(path);
                if (bmp == null) bmp = Download(info.Icon, path);
                if (bmp != null) Map[id] = bmp;
                return bmp;
            }
        }

        static BitmapImage Download(string icon, string path)
        {
            try
            {
                Directory.CreateDirectory(AppPaths.RuneIconsDir);
                using (var wc = new WebClient())
                    wc.DownloadFile(Cdn + icon, path);
                return FromFile(path);
            }
            catch
            {
                try { if (File.Exists(path)) File.Delete(path); } catch { }
                return null;
            }
        }

        static BitmapImage FromFile(string path)
        {
            if (!File.Exists(path)) return null;
            try
            {
                var bmp = new BitmapImage();
                bmp.BeginInit();
                bmp.CacheOption = BitmapCacheOption.OnLoad;
                bmp.UriSource = new Uri(path, UriKind.Absolute);
                bmp.EndInit();
                bmp.Freeze();
                return bmp;
            }
            catch { return null; }
        }
    }

    public static class OpggRunes
    {
        const string Base = "https://lol-api-champion.op.gg/api/KR/champions/ranked/";
        static readonly Dictionary<string, List<RunePageView>> Cache = new Dictionary<string, List<RunePageView>>();
        static readonly object Gate = new object();
        static readonly string[] Positions = new string[] { "top", "jungle", "mid", "adc", "support" };

        public static string PositionOf(string code)
        {
            var pos = (code ?? "").Trim().ToUpperInvariant();
            if (pos == "MID") pos = "MIDDLE";
            if (pos == "TOP") return "top";
            if (pos == "JUNGLE") return "jungle";
            if (pos == "MIDDLE") return "mid";
            if (pos == "BOTTOM") return "adc";
            if (pos == "UTILITY") return "support";
            return "";
        }

        public static List<RunePageView> Load(string slug, string prefer)
        {
            slug = (slug ?? "").Trim().ToLowerInvariant();
            prefer = prefer ?? "";
            var key = slug + "|" + prefer;
            lock (Gate)
            {
                List<RunePageView> hit;
                if (Cache.TryGetValue(key, out hit)) return hit;
            }
            var pages = Fetch(slug, prefer);
            lock (Gate) { Cache[key] = pages; }
            return pages;
        }

        static List<RunePageView> Fetch(string slug, string prefer)
        {
            string best = null;
            var bestPlay = -1;
            foreach (var pos in Order(prefer))
            {
                string json = null;
                try { json = Download(Base + slug + "/build?position=" + pos + "&tier=emerald_plus&hl=ru_RU"); }
                catch { continue; }
                var play = PeekPlay(json);
                if (play > bestPlay)
                {
                    best = json;
                    bestPlay = play;
                }
                if (play >= 1500) break;
            }
            if (string.IsNullOrEmpty(best)) return new List<RunePageView>();
            return Parse(best);
        }

        static string[] Order(string prefer)
        {
            var list = new List<string>();
            if (!string.IsNullOrEmpty(prefer)) list.Add(prefer);
            foreach (var pos in Positions)
            {
                if (pos != prefer) list.Add(pos);
            }
            return list.ToArray();
        }

        static string Download(string url)
        {
            using (var wc = new TimedClient())
            {
                wc.Encoding = Encoding.UTF8;
                wc.Headers[HttpRequestHeader.UserAgent] = "Mozilla/5.0";
                wc.Headers[HttpRequestHeader.Accept] = "application/json";
                return wc.DownloadString(url);
            }
        }

        static int PeekPlay(string json)
        {
            try
            {
                var pages = PagesOf(json);
                if (pages.Count == 0) return 0;
                var first = pages[0] as Dictionary<string, object>;
                return IntOf(first, "play");
            }
            catch { return 0; }
        }

        static List<RunePageView> Parse(string json)
        {
            var list = new List<RunePageView>();
            foreach (var raw in PagesOf(json))
            {
                var page = raw as Dictionary<string, object>;
                if (page == null) continue;
                list.Add(ParsePage(page));
                if (list.Count >= 2) break;
            }
            return list;
        }

        static ArrayList PagesOf(string json)
        {
            var ser = new JavaScriptSerializer();
            var root = ser.DeserializeObject(json) as Dictionary<string, object>;
            if (root == null || !root.ContainsKey("data")) return new ArrayList();
            var data = root["data"] as Dictionary<string, object>;
            if (data == null || !data.ContainsKey("rune_pages")) return new ArrayList();
            return AsList(data["rune_pages"]);
        }

        static ArrayList AsList(object raw)
        {
            var list = raw as ArrayList;
            if (list != null) return list;
            var copy = new ArrayList();
            var arr = raw as object[];
            if (arr != null) copy.AddRange(arr);
            return copy;
        }

        static RunePageView ParsePage(Dictionary<string, object> page)
        {
            var view = new RunePageView();
            view.PrimaryStyle = IntOf(page, "primary_page_id");
            view.SecondaryStyle = IntOf(page, "secondary_page_id");
            view.Play = IntOf(page, "play");
            view.Win = IntOf(page, "win");
            view.PickRate = DblOf(page, "pick_rate");
            var primary = Slots(4);
            var shards = Slots(3);
            var secondary = new Dictionary<int, int[]>();
            if (page.ContainsKey("builds"))
            {
                var builds = AsList(page["builds"]);
                if (builds != null)
                {
                    foreach (var raw in builds)
                    {
                        var build = raw as Dictionary<string, object>;
                        if (build == null) continue;
                        var play = IntOf(build, "play");
                        var win = IntOf(build, "win");
                        AddSlots(primary, Ids(build, "primary_rune_ids"), play, win);
                        AddBag(secondary, Ids(build, "secondary_rune_ids"), play, win);
                        AddSlots(shards, Ids(build, "stat_mod_ids"), play, win);
                    }
                }
            }
            view.Primary = TopSlots(primary);
            view.Secondary = TopBag(secondary, 2);
            view.Shards = TopSlots(shards);
            return view;
        }

        static List<Dictionary<int, int[]>> Slots(int n)
        {
            var list = new List<Dictionary<int, int[]>>();
            for (var i = 0; i < n; i++) list.Add(new Dictionary<int, int[]>());
            return list;
        }

        static void AddSlots(List<Dictionary<int, int[]>> slots, List<int> ids, int play, int win)
        {
            var n = ids.Count;
            if (n > slots.Count) n = slots.Count;
            for (var i = 0; i < n; i++) AddBag(slots[i], new List<int> { ids[i] }, play, win);
        }

        static void AddBag(Dictionary<int, int[]> bag, List<int> ids, int play, int win)
        {
            foreach (var id in ids)
            {
                if (id <= 0) continue;
                int[] cur;
                if (!bag.TryGetValue(id, out cur))
                {
                    cur = new int[2];
                    bag[id] = cur;
                }
                cur[0] += play;
                cur[1] += win;
            }
        }

        static List<RuneSlot> TopSlots(List<Dictionary<int, int[]>> slots)
        {
            var list = new List<RuneSlot>();
            foreach (var bag in slots)
            {
                var top = Best(bag);
                if (top != null) list.Add(top);
            }
            return list;
        }

        static List<RuneSlot> TopBag(Dictionary<int, int[]> bag, int take)
        {
            var list = new List<RuneSlot>();
            var used = new Dictionary<int, bool>();
            for (var n = 0; n < take; n++)
            {
                RuneSlot best = null;
                foreach (var pair in bag)
                {
                    if (used.ContainsKey(pair.Key)) continue;
                    if (best == null || pair.Value[0] > best.Play)
                        best = new RuneSlot { Id = pair.Key, Play = pair.Value[0], Win = pair.Value[1] };
                }
                if (best == null) break;
                used[best.Id] = true;
                list.Add(best);
            }
            return list;
        }

        static RuneSlot Best(Dictionary<int, int[]> bag)
        {
            RuneSlot best = null;
            foreach (var pair in bag)
            {
                if (best == null || pair.Value[0] > best.Play)
                    best = new RuneSlot { Id = pair.Key, Play = pair.Value[0], Win = pair.Value[1] };
            }
            return best;
        }

        static List<int> Ids(Dictionary<string, object> d, string key)
        {
            var list = new List<int>();
            if (d == null || !d.ContainsKey(key)) return list;
            var raw = AsList(d[key]);
            if (raw == null) return list;
            foreach (var item in raw)
            {
                try { list.Add(Convert.ToInt32(Convert.ToDouble(item, CultureInfo.InvariantCulture))); }
                catch { }
            }
            return list;
        }

        static int IntOf(Dictionary<string, object> d, string key)
        {
            if (d == null || !d.ContainsKey(key) || d[key] == null) return 0;
            try { return Convert.ToInt32(Convert.ToDouble(d[key], CultureInfo.InvariantCulture)); }
            catch { return 0; }
        }

        static double DblOf(Dictionary<string, object> d, string key)
        {
            if (d == null || !d.ContainsKey(key) || d[key] == null) return 0;
            try { return Convert.ToDouble(d[key], CultureInfo.InvariantCulture); }
            catch { return 0; }
        }

        class TimedClient : WebClient
        {
            protected override WebRequest GetWebRequest(Uri address)
            {
                var request = base.GetWebRequest(address);
                if (request != null) request.Timeout = 8000;
                return request;
            }
        }
    }

    public static class OpggBuilds
    {
        const string Base = "https://lol-api-champion.op.gg/api/GLOBAL/champions/ranked/";
        static readonly Dictionary<string, string> Cache = new Dictionary<string, string>();
        static readonly object Gate = new object();
        static readonly string[] Positions = new string[] { "top", "jungle", "mid", "adc", "support" };
        static readonly HashSet<string> Skip = new HashSet<string>(new[]
        {
            "1001", "2422", "3005", "3047", "3008", "3006", "3009", "3010", "3020", "3111", "3117", "3158",
            "1082", "2003", "2031", "2033", "1056", "1054", "1055", "3340", "3363", "3364"
        });

        public static string Text(string championName, string positionCode)
        {
            var slug = ChampionBook.Slug(championName);
            if (slug.Length == 0) return "";
            var prefer = OpggRunes.PositionOf(positionCode);
            var key = slug + "|" + prefer;
            lock (Gate)
            {
                string hit;
                if (Cache.TryGetValue(key, out hit)) return hit;
            }
            var text = Fetch(slug, prefer);
            if (text.Length == 0) return "";
            lock (Gate) { Cache[key] = text; }
            return text;
        }

        static string Fetch(string slug, string prefer)
        {
            string best = null;
            var bestPlay = -1;
            var bestPos = "";
            foreach (var pos in Order(prefer))
            {
                string json = null;
                try { json = Download(Base + slug + "/build?position=" + pos + "&tier=emerald_plus&hl=ru_RU"); }
                catch { continue; }
                var play = FirstPlay(json);
                if (play > bestPlay)
                {
                    best = json;
                    bestPlay = play;
                    bestPos = pos;
                }
                if (play >= 1500) break;
            }
            if (string.IsNullOrEmpty(best) || bestPos.Length == 0) return "";
            var cores = CoreLines(best);
            if (cores.Count == 0) return "";
            var later = "";
            try { later = LaterLine(Download(Base + slug + "/items/builds?position=" + bestPos + "&tier=emerald_plus&hl=ru_RU")); }
            catch { }
            var sb = new StringBuilder();
            sb.Append("Коры op.gg на этой роли. Это вариации кора: в стандарте выбери одну строку целиком.");
            for (var i = 0; i < cores.Count; i++)
                sb.Append("\n").Append(i + 1).Append(". ").Append(cores[i]);
            if (later.Length > 0) sb.Append("\n").Append(later);
            var boots = BootLine(best);
            if (boots.Length > 0) sb.Append("\n").Append(boots);
            return sb.ToString();
        }

        static string[] Order(string prefer)
        {
            var list = new List<string>();
            if (!string.IsNullOrEmpty(prefer)) list.Add(prefer);
            foreach (var pos in Positions)
            {
                if (pos != prefer) list.Add(pos);
            }
            return list.ToArray();
        }

        static string Download(string url)
        {
            using (var wc = new TimedClient())
            {
                wc.Encoding = Encoding.UTF8;
                wc.Headers[HttpRequestHeader.UserAgent] = "Mozilla/5.0";
                wc.Headers[HttpRequestHeader.Accept] = "application/json";
                return wc.DownloadString(url);
            }
        }

        class TimedClient : WebClient
        {
            protected override WebRequest GetWebRequest(Uri address)
            {
                var request = base.GetWebRequest(address);
                if (request != null) request.Timeout = 8000;
                return request;
            }
        }

        static int FirstPlay(string json)
        {
            var rows = Rows(json, "core_items");
            if (rows.Count == 0) return 0;
            return IntOf(rows[0] as Dictionary<string, object>, "play");
        }

        static List<string> CoreLines(string json)
        {
            var lines = new List<string>();
            foreach (var raw in Rows(json, "core_items"))
            {
                var row = raw as Dictionary<string, object>;
                if (row == null) continue;
                var ids = Legendaries(IdsOf(row));
                if (ids.Count < 3) continue;
                var parts = new List<string>();
                for (var i = 0; i < 3; i++)
                    parts.Add(ItemNames.Get(ids[i]) + " " + ids[i]);
                var pct = (DblOf(row, "pick_rate") * 100).ToString("0.0", CultureInfo.InvariantCulture).Replace('.', ',');
                lines.Add(string.Join(" → ", parts.ToArray()) + ", " + pct + "% игр");
                if (lines.Count == 3) break;
            }
            return lines;
        }

        static string LaterLine(string json)
        {
            var data = Data(json);
            if (data == null || !data.ContainsKey("single_items")) return "";
            var parts = new List<string>();
            foreach (var raw in AsList(data["single_items"]))
            {
                var block = raw as Dictionary<string, object>;
                if (block == null || !block.ContainsKey("items")) continue;
                var depth = IntOf(block, "depth");
                if (depth < 4 || depth > 6) continue;
                var names = new List<string>();
                foreach (var item in AsList(block["items"]))
                {
                    var row = item as Dictionary<string, object>;
                    var ids = Legendaries(IdsOf(row));
                    if (ids.Count == 0) continue;
                    names.Add(ItemNames.Get(ids[0]) + " " + ids[0]);
                    if (names.Count == 5) break;
                }
                if (names.Count == 0) continue;
                parts.Add(depth + "-й: " + string.Join(", ", names.ToArray()));
            }
            if (parts.Count == 0) return "";
            return "После кора, слоты 4–6 только отсюда:\n" + string.Join("\n", parts.ToArray());
        }

        static string BootLine(string json)
        {
            var names = new List<string>();
            foreach (var raw in Rows(json, "boots"))
            {
                var row = raw as Dictionary<string, object>;
                var ids = IdsOf(row);
                if (ids.Count == 0) continue;
                var id = ids[0].ToString();
                if (!ItemNames.Knows(id)) continue;
                names.Add(ItemNames.Get(id) + " " + id);
                if (names.Count == 3) break;
            }
            if (names.Count == 0) return "";
            return "Ботинки: " + string.Join(", ", names.ToArray()) + ".";
        }

        static List<string> Legendaries(List<int> ids)
        {
            var list = new List<string>();
            if (ids == null) return list;
            foreach (var id in ids)
            {
                var text = id.ToString();
                if (Skip.Contains(text) || !ItemNames.Knows(text) || list.Contains(text)) continue;
                list.Add(text);
            }
            return list;
        }

        static List<int> IdsOf(Dictionary<string, object> row)
        {
            var list = new List<int>();
            if (row == null || !row.ContainsKey("ids") || row["ids"] == null) return list;
            var text = row["ids"] as string;
            if (text != null)
            {
                foreach (var part in text.Split(new[] { ' ' }, StringSplitOptions.RemoveEmptyEntries))
                {
                    int id;
                    if (int.TryParse(part, out id) && id > 0) list.Add(id);
                }
                return list;
            }
            foreach (var item in AsList(row["ids"]))
            {
                try
                {
                    var id = Convert.ToInt32(Convert.ToDouble(item, CultureInfo.InvariantCulture));
                    if (id > 0) list.Add(id);
                }
                catch { }
            }
            return list;
        }

        static ArrayList Rows(string json, string key)
        {
            var data = Data(json);
            if (data == null || !data.ContainsKey(key)) return new ArrayList();
            return AsList(data[key]);
        }

        static Dictionary<string, object> Data(string json)
        {
            var ser = new JavaScriptSerializer { MaxJsonLength = int.MaxValue };
            var root = ser.DeserializeObject(json) as Dictionary<string, object>;
            if (root == null || !root.ContainsKey("data")) return null;
            return root["data"] as Dictionary<string, object>;
        }

        static ArrayList AsList(object raw)
        {
            var list = raw as ArrayList;
            if (list != null) return list;
            var copy = new ArrayList();
            var arr = raw as object[];
            if (arr != null) copy.AddRange(arr);
            return copy;
        }

        static int IntOf(Dictionary<string, object> d, string key)
        {
            if (d == null || !d.ContainsKey(key) || d[key] == null) return 0;
            try { return Convert.ToInt32(Convert.ToDouble(d[key], CultureInfo.InvariantCulture)); }
            catch { return 0; }
        }

        static double DblOf(Dictionary<string, object> d, string key)
        {
            if (d == null || !d.ContainsKey(key) || d[key] == null) return 0;
            try { return Convert.ToDouble(d[key], CultureInfo.InvariantCulture); }
            catch { return 0; }
        }
    }
}
