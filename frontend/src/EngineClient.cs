using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Text;
using System.Web.Script.Serialization;

namespace LolBuildOverlay
{
    public class EngineSlot
    {
        public int itemId { get; set; }
        public string name { get; set; }
        public double priority { get; set; }
        public string role { get; set; }
    }

    public class EngineRec
    {
        public EngineSlot[] build { get; set; }
        public EngineSlot nextItem { get; set; }
        public bool hasNext { get; set; }
        public string[] reasons { get; set; }
        public string seedName { get; set; }
        public bool offrole { get; set; }
        public int[] ownedItems { get; set; }
        public EngineSlot[] early { get; set; }
    }

    public class EngineResult
    {
        public RecommendedBuild Build;
        public string[] Reasons = new string[0];
        public string[] Play = new string[0];
        public string SeedName = "";
        public string NextName = "";
        public string Error;
        public bool FromLive;

        public bool Ok { get { return Build != null && string.IsNullOrEmpty(Error); } }
    }

    public static class EngineClient
    {
        public static string ExePath()
        {
            return Path.Combine(AppPaths.Root, "recommend.exe");
        }

        public static string DataRoot()
        {
            var dir = AppPaths.Root;
            for (var i = 0; i < 8 && !string.IsNullOrEmpty(dir); i++)
            {
                if (IsBackend(dir)) return dir;
                var nested = Path.Combine(dir, "backend");
                if (IsBackend(nested)) return nested;
                var parent = Path.GetDirectoryName(dir.TrimEnd(Path.DirectorySeparatorChar, Path.AltDirectorySeparatorChar));
                if (string.IsNullOrEmpty(parent) || parent == dir) break;
                dir = parent;
            }
            return AppPaths.Root;
        }

        static bool IsBackend(string dir)
        {
            return File.Exists(Path.Combine(dir, "go.mod")) &&
                   File.Exists(Path.Combine(dir, "data", "items.json"));
        }

        public static string DemoFixture()
        {
            return Path.Combine(DataRoot(), "testdata", "fixtures", "draft_start.json");
        }

        public static EngineResult Recommend(bool preferLive)
        {
            var exe = ExePath();
            if (!File.Exists(exe))
                return Fail("нет recommend.exe — запусти frontend\\Start.bat");

            if (preferLive)
            {
                var live = Run(exe, true, null, true);
                if (live.Ok) return live;
                var snap = LiveClient.SnapPath();
                if (File.Exists(snap))
                {
                    var fromSnap = Run(exe, false, snap, true);
                    if (fromSnap.Ok) return fromSnap;
                }
                return Fail(string.IsNullOrEmpty(live.Error) ? "клиент игры не отвечает :2999" : live.Error);
            }

            var fixture = DemoFixture();
            if (File.Exists(fixture))
            {
                var demo = Run(exe, false, fixture, false);
                if (demo.Ok) return demo;
            }

            return Fail("движок не вернул сборку");
        }

        private static EngineResult Run(string exe, bool liveFlag, string fixture, bool fromLive)
        {
            var root = DataRoot();
            var args = new StringBuilder();
            args.Append("-json -root ").Append(Quote(root));
            if (liveFlag) args.Append(" -live");
            else args.Append(" -fixture ").Append(Quote(fixture));

            var psi = new ProcessStartInfo
            {
                FileName = exe,
                Arguments = args.ToString(),
                WorkingDirectory = root,
                UseShellExecute = false,
                RedirectStandardOutput = true,
                RedirectStandardError = true,
                CreateNoWindow = true
            };

            try
            {
                using (var p = Process.Start(psi))
                {
                    if (p == null) return Fail("не удалось запустить recommend.exe");
                    var stdout = p.StandardOutput.ReadToEnd();
                    var stderr = p.StandardError.ReadToEnd();
                    if (!p.WaitForExit(12000))
                    {
                        try { p.Kill(); } catch { }
                        return Fail("движок не ответил за 12с");
                    }
                    if (p.ExitCode != 0)
                    {
                        var msg = (stderr ?? "").Trim();
                        if (string.IsNullOrEmpty(msg)) msg = (stdout ?? "").Trim();
                        if (string.IsNullOrEmpty(msg)) msg = "recommend exit " + p.ExitCode;
                        if (msg.Length > 180) msg = msg.Substring(0, 180);
                        return Fail(msg);
                    }

                    var rec = Parse(stdout);
                    if (rec == null) return Fail("движок вернул пустой JSON");
                    var mapped = Map(rec);
                    if (mapped == null) return Fail("в ответе нет предметов");
                    return new EngineResult
                    {
                        Build = mapped,
                        Reasons = rec.reasons ?? new string[0],
                        SeedName = rec.seedName ?? "",
                        NextName = rec.hasNext && rec.nextItem != null ? rec.nextItem.name : "",
                        FromLive = fromLive
                    };
                }
            }
            catch (Exception ex)
            {
                return Fail(ex.Message);
            }
        }

        private static EngineRec Parse(string json)
        {
            if (string.IsNullOrEmpty(json)) return null;
            try
            {
                var ser = new JavaScriptSerializer { MaxJsonLength = int.MaxValue };
                return ser.Deserialize<EngineRec>(json);
            }
            catch
            {
                return null;
            }
        }

        private static RecommendedBuild Map(EngineRec rec)
        {
            if (rec == null || rec.build == null || rec.build.Length == 0) return null;

            var owned = new HashSet<string>();
            if (rec.ownedItems != null)
            {
                for (var i = 0; i < rec.ownedItems.Length; i++)
                {
                    if (rec.ownedItems[i] > 0)
                        owned.Add(rec.ownedItems[i].ToString());
                }
            }

            string nextId = null;
            string nextRole = "";
            if (rec.hasNext && rec.nextItem != null && rec.nextItem.itemId > 0)
            {
                nextId = rec.nextItem.itemId.ToString();
                nextRole = rec.nextItem.role ?? "";
            }

            var have = new List<string>();
            var rest = new List<string>();
            string boots = null;

            foreach (var s in rec.build)
            {
                if (s == null || s.itemId <= 0) continue;
                var id = s.itemId.ToString();
                var role = s.role ?? "";
                if (role == "boots")
                {
                    if (owned.Contains(id) || boots == null || id == nextId) boots = id;
                    continue;
                }
                if (role == "component") continue;
                if (role == "start" || role == "consumable") continue;
                if (have.Contains(id) || rest.Contains(id)) continue;
                if (owned.Contains(id)) have.Add(id);
                else rest.Add(id);
            }

            var items = new List<string>();
            foreach (var id in have)
            {
                if (items.Count >= 5) break;
                items.Add(id);
            }
            foreach (var id in rest)
            {
                if (items.Count >= 5) break;
                items.Add(id);
            }
            while (items.Count < 5) items.Add("");
            if (string.IsNullOrEmpty(boots)) boots = "3006";

            var highlight = nextId;
            if (nextRole == "component" || nextRole == "start" || nextRole == "consumable")
            {
                var parent = ParentLegendary(nextId, items);
                if (parent != null) highlight = parent;
            }

            var early = new List<string>();
            var earlyNext = -1;
            if (rec.early != null)
            {
                foreach (var s in rec.early)
                {
                    if (s == null || s.itemId <= 0) continue;
                    var id = s.itemId.ToString();
                    if (id == nextId && earlyNext < 0) earlyNext = early.Count;
                    early.Add(id);
                }
            }

            var nextIndex = -1;
            if (earlyNext < 0 && (nextId == boots || highlight == boots)) nextIndex = 6;
            else if (earlyNext < 0)
            {
                var i = -1;
                if (!string.IsNullOrEmpty(highlight)) i = items.IndexOf(highlight);
                if (i < 0 && nextId != null) i = items.IndexOf(nextId);
                if (i < 0)
                {
                    for (var n = 0; n < items.Count; n++)
                    {
                        if (!owned.Contains(items[n])) { i = n; break; }
                    }
                }
                if (i >= 0) nextIndex = i;
            }

            return new RecommendedBuild
            {
                Items = items.ToArray(),
                Early = early.ToArray(),
                Boots = boots,
                NextIndex = nextIndex,
                EarlyNext = earlyNext
            };
        }

        // component id → legendaries it builds into (first match in the row wins).
        private static readonly Dictionary<string, string[]> CraftsInto = new Dictionary<string, string[]>
        {
            {"1043", new[]{"3115","3153"}},
            {"1053", new[]{"3153","3072"}},
            {"1038", new[]{"3031","3032","3072"}},
            {"3144", new[]{"3153","3032"}},
            {"1052", new[]{"3115"}},
            {"1026", new[]{"3100","3165","3089","4645","3135"}},
            {"1058", new[]{"3089","4645","3157","3102"}},
            {"3802", new[]{"3118","6655","2503","6657"}},
            {"3108", new[]{"3118"}},
            {"3147", new[]{"6653"}},
            {"2508", new[]{"6653","2503"}},
            {"3134", new[]{"3142","6701","6692"}},
            {"3057", new[]{"3100","3078","6662"}},
            {"6670", new[]{"6672"}},
            {"3145", new[]{"4646","3152"}},
            {"3113", new[]{"3100","4646"}},
            {"2420", new[]{"3157"}},
            {"3916", new[]{"3165"}},
            {"1033", new[]{"3102"}},
            {"3133", new[]{"6610","3071"}}
        };

        private static string ParentLegendary(string componentId, List<string> row)
        {
            string[] parents;
            if (string.IsNullOrEmpty(componentId) || !CraftsInto.TryGetValue(componentId, out parents))
                return null;
            for (var i = 0; i < row.Count; i++)
            {
                for (var j = 0; j < parents.Length; j++)
                {
                    if (row[i] == parents[j]) return row[i];
                }
            }
            return null;
        }

        private static EngineResult Fail(string error)
        {
            return new EngineResult { Error = error };
        }

        private static string Quote(string path)
        {
            if (string.IsNullOrEmpty(path)) return "\"\"";
            return "\"" + path.Replace("\"", "\\\"") + "\"";
        }
    }
}
