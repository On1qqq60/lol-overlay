using System;
using System.Collections;
using System.Collections.Generic;
using System.IO;
using System.Net;
using System.Net.Security;
using System.Security.Cryptography.X509Certificates;
using System.Text;
using System.Web.Script.Serialization;

namespace LolBuildOverlay
{
    public static class LiveClient
    {
        private static readonly string[] Hosts =
        {
            "https://127.0.0.1:2999/liveclientdata",
            "https://localhost:2999/liveclientdata"
        };

        public static bool Connected;
        public static string Status = "поиск клиента";
        public static string LastError = "";

        private static readonly object Gate = new object();
        private static readonly JavaScriptSerializer Ser = new JavaScriptSerializer { MaxJsonLength = int.MaxValue };
        private static string _lastLog;

        static LiveClient()
        {
            try
            {
                ServicePointManager.SecurityProtocol =
                    (SecurityProtocolType)3072 | (SecurityProtocolType)768 | (SecurityProtocolType)192;
            }
            catch
            {
                ServicePointManager.SecurityProtocol = SecurityProtocolType.Tls;
            }
            ServicePointManager.ServerCertificateValidationCallback = TrustAll;
            ServicePointManager.Expect100Continue = false;
            ServicePointManager.DefaultConnectionLimit = 16;
            ServicePointManager.CheckCertificateRevocationList = false;
        }

        private static bool TrustAll(object sender, X509Certificate cert, X509Chain chain, SslPolicyErrors errors)
        {
            return true;
        }

        public static MatchState Fetch()
        {
            lock (Gate)
            {
                try
                {
                    foreach (var host in Hosts)
                    {
                        var state = FetchFrom(host);
                        if (state != null && state.Me != null)
                        {
                            Ok(state);
                            return state;
                        }
                    }
                    Fail(string.IsNullOrEmpty(LastError) ? "клиент игры не отвечает :2999" : LastError);
                    return null;
                }
                catch (Exception ex)
                {
                    Fail(ex.Message);
                    Log("ex " + ex);
                    return null;
                }
            }
        }

        private static MatchState FetchFrom(string host)
        {
            var listRaw = GetJson(host + "/playerlist");
            if (!string.IsNullOrEmpty(listRaw))
            {
                var players = ParsePlayers(Deserialize(listRaw));
                if (players.Count > 0)
                {
                    var meName = Unquote(GetJson(host + "/activeplayername"));
                    var active = AsDict(Deserialize(GetJson(host + "/activeplayer")));
                    if (string.IsNullOrEmpty(meName) && active != null)
                    {
                        meName = FirstNonEmpty(
                            Str(active, "summonerName"),
                            Str(active, "riotIdGameName"),
                            Str(active, "riotId"));
                    }

                    double gameTime = 0;
                    var stats = AsDict(Deserialize(GetJson(host + "/gamestats")));
                    if (stats != null) gameTime = Num(stats, "gameTime");

                    var me = FindMe(players, meName);
                    if (me == null)
                    {
                        Log("name miss '" + meName + "', players=" + PlayersDump(players));
                        me = FirstHuman(players) ?? players[0];
                    }

                    return BuildState(me, players, gameTime, active);
                }
            }

            var allRaw = GetJson(host + "/allgamedata");
            if (string.IsNullOrEmpty(allRaw)) return null;

            var root = AsDict(Deserialize(allRaw));
            if (root == null) return null;

            var allPlayers = ParsePlayers(Get(root, "allPlayers"));
            if (allPlayers.Count == 0) return null;

            var active2 = AsDict(Get(root, "activePlayer"));
            var game = AsDict(Get(root, "gameData"));
            var meName2 = "";
            if (active2 != null)
            {
                meName2 = FirstNonEmpty(
                    Str(active2, "summonerName"),
                    Str(active2, "riotIdGameName"),
                    Str(active2, "riotId"));
            }

            var me2 = FindMe(allPlayers, meName2) ?? FirstHuman(allPlayers) ?? allPlayers[0];
            var gameTime2 = game != null ? Num(game, "gameTime") : 0;
            return BuildState(me2, allPlayers, gameTime2, active2);
        }

        private static MatchState BuildState(LivePlayer me, List<LivePlayer> players, double gameTime, Dictionary<string, object> active)
        {
            var state = new MatchState
            {
                Me = me,
                InGame = true,
                GameTime = gameTime,
                Gold = active != null ? Num(active, "currentGold") : 0
            };
            foreach (var p in players)
            {
                if (string.Equals(p.team, me.team, StringComparison.OrdinalIgnoreCase))
                    state.Allies.Add(p);
                else
                    state.Enemies.Add(p);
            }
            return state;
        }

        private static LivePlayer FindMe(List<LivePlayer> players, string meName)
        {
            if (string.IsNullOrEmpty(meName)) return null;
            foreach (var p in players)
            {
                if (SameName(p, meName)) return p;
            }
            return null;
        }

        private static LivePlayer FirstHuman(List<LivePlayer> players)
        {
            foreach (var p in players)
            {
                if (!p.isBot) return p;
            }
            return null;
        }

        private static bool SameName(LivePlayer p, string meName)
        {
            var a = NormName(meName);
            if (a.Length == 0) return false;
            var aShort = NormName(StripTag(meName));
            return Eq(p.summonerName, a) || Eq(p.riotIdGameName, a) || Eq(p.riotId, a)
                || Eq(StripTag(p.riotId), a) || Eq(p.summonerName, aShort)
                || Eq(p.riotIdGameName, aShort) || Eq(StripTag(p.riotId), aShort);
        }

        private static bool Eq(string x, string y)
        {
            return NormName(x) == y && y.Length > 0;
        }

        private static string StripTag(string s)
        {
            if (string.IsNullOrEmpty(s)) return "";
            var i = s.IndexOf('#');
            return i > 0 ? s.Substring(0, i) : s;
        }

        private static string NormName(string s)
        {
            return (s ?? "").Trim().ToLowerInvariant();
        }

        private static List<LivePlayer> ParsePlayers(object raw)
        {
            var list = new List<LivePlayer>();
            foreach (var row in Enumerate(raw))
            {
                var d = AsDict(row);
                if (d != null) list.Add(ParsePlayer(d));
            }
            return list;
        }

        private static LivePlayer ParsePlayer(Dictionary<string, object> d)
        {
            var champ = Str(d, "championName");
            if (string.IsNullOrEmpty(champ))
            {
                var raw = Str(d, "rawChampionName");
                var idx = raw.LastIndexOf('_');
                if (idx >= 0 && idx < raw.Length - 1) champ = raw.Substring(idx + 1);
            }

            var p = new LivePlayer
            {
                championName = champ,
                isBot = Bool(d, "isBot"),
                level = (int)Num(d, "level"),
                position = Str(d, "position"),
                role = Str(d, "role"),
                summonerName = Str(d, "summonerName"),
                riotIdGameName = Str(d, "riotIdGameName"),
                riotId = Str(d, "riotId"),
                team = Str(d, "team"),
                items = new List<LiveItem>(),
                scores = new LiveScores()
            };

            var scores = AsDict(Get(d, "scores"));
            if (scores != null)
            {
                p.scores.assists = (int)Num(scores, "assists");
                p.scores.creepScore = (int)Num(scores, "creepScore");
                p.scores.deaths = (int)Num(scores, "deaths");
                p.scores.kills = (int)Num(scores, "kills");
            }

            foreach (var it in Enumerate(Get(d, "items")))
            {
                var id = AsDict(it);
                if (id == null) continue;
                var itemId = (int)Num(id, "itemID");
                if (itemId <= 0) itemId = (int)Num(id, "itemId");
                p.items.Add(new LiveItem
                {
                    itemID = itemId,
                    slot = (int)Num(id, "slot"),
                    count = (int)Num(id, "count"),
                    displayName = Str(id, "displayName"),
                    price = (int)Num(id, "price")
                });
            }
            return p;
        }

        private static IEnumerable Enumerate(object raw)
        {
            var arr = raw as ArrayList;
            if (arr != null) return arr;
            var obj = raw as object[];
            if (obj != null) return obj;
            return new object[0];
        }

        private static string GetJson(string url)
        {
            if (string.IsNullOrEmpty(url)) return null;
            try
            {
                var req = (HttpWebRequest)WebRequest.Create(url);
                req.Method = "GET";
                req.Timeout = 4000;
                req.ReadWriteTimeout = 4000;
                req.KeepAlive = false;
                req.Proxy = null;
                req.AllowAutoRedirect = false;
                req.UserAgent = "LeagueOfLegendsClient/lol.build";
                req.Accept = "application/json";
                using (var resp = (HttpWebResponse)req.GetResponse())
                using (var stream = resp.GetResponseStream())
                {
                    if (stream == null) return null;
                    using (var reader = new StreamReader(stream, Encoding.UTF8))
                        return reader.ReadToEnd();
                }
            }
            catch (WebException ex)
            {
                LastError = WebError(ex);
                return null;
            }
            catch (Exception ex)
            {
                LastError = ex.Message;
                return null;
            }
        }

        private static string WebError(WebException ex)
        {
            var http = ex.Response as HttpWebResponse;
            if (http != null)
                return "HTTP " + (int)http.StatusCode;
            if (ex.Status == WebExceptionStatus.ConnectFailure)
                return "порт 2999 закрыт — зайди в саму игру, не в лобби";
            if (ex.Status == WebExceptionStatus.TrustFailure)
                return "SSL к 2999";
            if (ex.Status == WebExceptionStatus.Timeout)
                return "таймаут :2999";
            return ex.Status + " " + ex.Message;
        }

        private static void Ok(MatchState state)
        {
            Connected = true;
            var champ = state.Me != null ? state.Me.championName : "?";
            var bots = 0;
            foreach (var e in state.Enemies) if (e.isBot) bots++;
            Status = "онлайн · " + champ + (bots > 0 ? " vs bots" : "");
            LastError = "";
            Log("ok " + champ + " enemies=" + state.Enemies.Count + " bots=" + bots + " t=" + (int)state.GameTime);
        }

        private static void Fail(string err)
        {
            Connected = false;
            Status = "поиск :2999";
            LastError = err ?? "";
            Log("fail " + LastError);
        }

        private static void Log(string msg)
        {
            if (msg == _lastLog) return;
            _lastLog = msg;
            try
            {
                var path = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "live.log");
                File.AppendAllText(path, DateTime.Now.ToString("HH:mm:ss") + " " + msg + Environment.NewLine, Encoding.UTF8);
            }
            catch { }
        }

        private static string PlayersDump(List<LivePlayer> players)
        {
            var sb = new StringBuilder();
            foreach (var p in players)
            {
                if (sb.Length > 0) sb.Append("; ");
                sb.Append(p.championName).Append(p.isBot ? "(bot)" : "(you?)");
            }
            return sb.ToString();
        }

        private static object Deserialize(string json)
        {
            if (string.IsNullOrEmpty(json)) return null;
            try { return Ser.DeserializeObject(json); }
            catch (Exception ex)
            {
                LastError = "json: " + ex.Message;
                return null;
            }
        }

        private static Dictionary<string, object> AsDict(object o)
        {
            return o as Dictionary<string, object>;
        }

        private static object Get(Dictionary<string, object> d, string key)
        {
            object v;
            return d != null && d.TryGetValue(key, out v) ? v : null;
        }

        private static string Str(Dictionary<string, object> d, string key)
        {
            var v = Get(d, key);
            return v == null ? "" : Convert.ToString(v);
        }

        private static double Num(Dictionary<string, object> d, string key)
        {
            var v = Get(d, key);
            if (v == null || v is DBNull) return 0;
            try { return Convert.ToDouble(v); }
            catch { return 0; }
        }

        private static bool Bool(Dictionary<string, object> d, string key)
        {
            var v = Get(d, key);
            if (v == null) return false;
            try { return Convert.ToBoolean(v); }
            catch { return false; }
        }

        private static string Unquote(string s)
        {
            if (string.IsNullOrEmpty(s)) return "";
            s = s.Trim();
            if (s.Length >= 2 && s[0] == '"') s = s.Substring(1);
            if (s.Length >= 1 && s[s.Length - 1] == '"') s = s.Substring(0, s.Length - 1);
            return s.Replace("\\u0027", "'").Replace("\\/", "/");
        }

        private static string FirstNonEmpty(params string[] values)
        {
            foreach (var v in values)
                if (!string.IsNullOrEmpty(v)) return v;
            return "";
        }
    }
}
