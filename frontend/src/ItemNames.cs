using System;
using System.Collections.Generic;
using System.Globalization;
using System.IO;
using System.Net;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Controls.Primitives;
using System.Windows.Input;
using System.Windows.Interop;
using System.Windows.Media;
using System.Windows.Media.Effects;
using System.Windows.Media.Imaging;
using System.Windows.Threading;
using WinForms = System.Windows.Forms;

namespace LolBuildOverlay
{
    public static class ItemNames
    {
        private static readonly Dictionary<string, string> Map = new Dictionary<string, string>();

        public static void Load()
        {
            Map.Clear();
            if (!File.Exists(AppPaths.NamesFile)) return;
            try
            {
                foreach (var line in File.ReadAllLines(AppPaths.NamesFile, Encoding.UTF8))
                {
                    var i = line.IndexOf('=');
                    if (i <= 0) continue;
                    Map[line.Substring(0, i).Trim()] = line.Substring(i + 1).Trim();
                }
            }
            catch { }
        }

        public static string Get(string id)
        {
            string name;
            if (id != null && Map.TryGetValue(id, out name) && !string.IsNullOrEmpty(name))
                return name;
            return id ?? "";
        }

        public static bool Knows(string id)
        {
            return id != null && Map.ContainsKey(id);
        }

        public static string FindByHead(string head)
        {
            if (string.IsNullOrEmpty(head)) return "";
            head = head.Trim();
            if (head.Length < 4) return "";
            var found = "";
            foreach (var pair in Map)
            {
                var name = pair.Value;
                if (string.IsNullOrEmpty(name)) continue;
                var hit = name.IndexOf(head, StringComparison.OrdinalIgnoreCase) >= 0
                    || head.IndexOf(name, StringComparison.OrdinalIgnoreCase) >= 0;
                if (!hit) continue;
                if (found.Length > 0 && found != pair.Key) return "";
                found = pair.Key;
            }
            return found;
        }

        public static string FindId(string name)
        {
            if (string.IsNullOrEmpty(name)) return "";
            var want = name.Trim();
            if (Knows(want)) return want;
            foreach (var pair in Map)
            {
                if (string.Equals(pair.Value, want, StringComparison.OrdinalIgnoreCase))
                    return pair.Key;
            }
            return "";
        }
    }
}
