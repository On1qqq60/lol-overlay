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
    }

    public class EngineResult
    {
        public RecommendedBuild Build;
        public string[] Reasons = new string[0];
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
            var local = Path.Combine(AppPaths.Root, "recommend.exe");
            if (File.Exists(local)) return local;
            var repo = Path.GetFullPath(Path.Combine(AppPaths.Root, ".."));
            var fromRepo = Path.Combine(repo, "recommend.exe");
            if (File.Exists(fromRepo)) return fromRepo;
            var nested = Path.Combine(repo, "engine", "recommend.exe");
            if (File.Exists(nested)) return nested;
            return local;
        }

        public static string DataRoot()
        {
            var repo = Path.GetFullPath(Path.Combine(AppPaths.Root, ".."));
            if (File.Exists(Path.Combine(repo, "data", "items.json"))) return repo;
            var nested = Path.Combine(repo, "engine");
            if (File.Exists(Path.Combine(nested, "data", "items.json"))) return nested;
            var nextToExe = Path.GetDirectoryName(ExePath());
            if (!string.IsNullOrEmpty(nextToExe) && File.Exists(Path.Combine(nextToExe, "data", "items.json")))
                return nextToExe;
            return AppPaths.Root;
        }

        public static string DemoFixture()
        {
            return Path.Combine(DataRoot(), "testdata", "fixtures", "draft_start.json");
        }

        public static EngineResult Recommend(bool preferLive)
        {
            var exe = ExePath();
            if (!File.Exists(exe))
                return Fail("нет recommend.exe — запусти overlay\\Start.bat");

            if (preferLive)
            {
                var live = Run(exe, true, null);
                if (live.Ok) return live;
            }

            var fixture = DemoFixture();
            if (File.Exists(fixture))
            {
                var demo = Run(exe, false, fixture);
                if (demo.Ok)
                {
                    if (preferLive)
                    {
                        var extra = new List<string> { "live API недоступен — демо-сборка из фикстуры" };
                        if (demo.Reasons != null) extra.AddRange(demo.Reasons);
                        demo.Reasons = extra.ToArray();
                    }
                    return demo;
                }
                if (!preferLive) return demo;
            }

            return Fail(preferLive ? "клиент игры не отвечает :2999" : "движок не вернул сборку");
        }

        private static EngineResult Run(string exe, bool live, string fixture)
        {
            var root = DataRoot();
            var args = new StringBuilder();
            args.Append("-json -root ").Append(Quote(root));
            if (live) args.Append(" -live");
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
                        FromLive = live
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

            string nextId = null;
            if (rec.hasNext && rec.nextItem != null && rec.nextItem.itemId > 0)
                nextId = rec.nextItem.itemId.ToString();

            var items = new List<string>();
            string boots = null;

            foreach (var s in rec.build)
            {
                if (s == null || s.itemId <= 0) continue;
                var id = s.itemId.ToString();
                var role = s.role ?? "";
                if (role == "boots")
                {
                    if (boots == null || id == nextId) boots = id;
                    continue;
                }
                if (role == "component" && id != nextId) continue;
                if (role == "start" && id != nextId) continue;
                if (items.Count < 5) items.Add(id);
            }

            if (nextId != null && nextId != boots && !items.Contains(nextId))
            {
                items.Insert(0, nextId);
                while (items.Count > 5) items.RemoveAt(items.Count - 1);
            }

            while (items.Count < 5) items.Add(items.Count > 0 ? items[items.Count - 1] : "1055");
            if (string.IsNullOrEmpty(boots)) boots = "3006";

            var nextIndex = 0;
            if (nextId == boots) nextIndex = 5;
            else
            {
                var i = items.IndexOf(nextId);
                if (i >= 0) nextIndex = i;
            }

            return new RecommendedBuild
            {
                Items = items.ToArray(),
                Boots = boots,
                NextIndex = nextIndex
            };
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
