using System;
using System.Collections.Generic;
using System.IO;
using System.Text;

namespace LolBuildOverlay
{
    public static class ChampionCards
    {
        public static string Find(string championName, string position)
        {
            var cards = Load();
            if (cards.Count == 0 || string.IsNullOrEmpty(championName)) return "";
            var key = Norm(championName);
            var pos = (position ?? "").Trim().ToUpperInvariant();
            string any = "";
            string exact = "";
            foreach (var card in cards)
            {
                if (!card.Keys.Contains(key)) continue;
                if (card.Role == "*") any = card.Text;
                else if (pos.Length > 0 && card.Role == pos) exact = card.Text;
            }
            if (exact.Length > 0) return exact;
            if (any.Length > 0) return any;
            return "";
        }

        static List<Card> Load()
        {
            var list = new List<Card>();
            var path = Path.Combine(AppPaths.PromptsDir, "cards.txt");
            if (!File.Exists(path)) return list;
            string[] lines;
            try { lines = File.ReadAllLines(path, Encoding.UTF8); }
            catch { return list; }
            foreach (var raw in lines)
            {
                var line = (raw ?? "").Trim();
                if (line.Length == 0 || line[0] == '#') continue;
                var parts = line.Split(new[] { '|' }, 3);
                if (parts.Length < 3) continue;
                var keys = new HashSet<string>();
                foreach (var name in parts[0].Split(','))
                {
                    var n = Norm(name);
                    if (n.Length > 0) keys.Add(n);
                }
                if (keys.Count == 0) continue;
                var role = parts[1].Trim().ToUpperInvariant();
                if (role.Length == 0) role = "*";
                var text = parts[2].Trim();
                if (text.Length == 0) continue;
                list.Add(new Card { Keys = keys, Role = role, Text = text });
            }
            return list;
        }

        static string Norm(string s)
        {
            var sb = new StringBuilder();
            foreach (var c in (s ?? "").ToLowerInvariant())
            {
                if (char.IsLetterOrDigit(c)) sb.Append(c);
            }
            return sb.ToString();
        }

        sealed class Card
        {
            public HashSet<string> Keys;
            public string Role;
            public string Text;
        }
    }
}
