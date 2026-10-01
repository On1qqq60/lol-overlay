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
    public static class AppSettings
    {
        public static double Opacity = 1.0;
        public static int IconSize = 22;
        public static uint Vk = 0x77;
        public static string KeyName = "F8";
        public static bool ShowBuild = true;
        public static string SpeShuKey = "";
        public static string SpeShuModel = "qwen/qwen3.8-omni-flash";

        public static void Load()
        {
            if (!File.Exists(AppPaths.SettingsFile)) return;
            try
            {
                foreach (var line in File.ReadAllLines(AppPaths.SettingsFile))
                {
                    var i = line.IndexOf('=');
                    if (i <= 0) continue;
                    var k = line.Substring(0, i).Trim();
                    var v = line.Substring(i + 1).Trim();
                    if (k == "opacity")
                    {
                        double d;
                        if (double.TryParse(v, NumberStyles.Any, CultureInfo.InvariantCulture, out d))
                            Opacity = Math.Max(0.25, Math.Min(1.0, d));
                    }
                    else if (k == "size")
                    {
                        int n;
                        if (int.TryParse(v, out n)) IconSize = Math.Max(16, Math.Min(40, n));
                    }
                    else if (k == "vk")
                    {
                        uint n;
                        if (uint.TryParse(v, out n)) Vk = n;
                    }
                    else if (k == "key") KeyName = v;
                    else if (k == "build") ShowBuild = v != "0" && v.ToLowerInvariant() != "false";
                    else if (k == "speshu_key") SpeShuKey = v;
                    else if (k == "speshu_model" && v.Length > 0) SpeShuModel = v;
                }
            }
            catch { }
        }

        public static void Save()
        {
            try
            {
                var lines = new List<string>();
                if (File.Exists(AppPaths.SettingsFile))
                    lines.AddRange(File.ReadAllLines(AppPaths.SettingsFile));
                Upsert(lines, "opacity", Opacity.ToString("0.##", CultureInfo.InvariantCulture));
                Upsert(lines, "size", IconSize.ToString(CultureInfo.InvariantCulture));
                Upsert(lines, "vk", Vk.ToString(CultureInfo.InvariantCulture));
                Upsert(lines, "key", KeyName);
                Upsert(lines, "build", ShowBuild ? "1" : "0");
                if (!string.IsNullOrEmpty(SpeShuKey)) Upsert(lines, "speshu_key", SpeShuKey);
                if (!string.IsNullOrEmpty(SpeShuModel)) Upsert(lines, "speshu_model", SpeShuModel);
                File.WriteAllText(AppPaths.SettingsFile, string.Join("\r\n", lines.ToArray()) + "\r\n");
            }
            catch { }
        }

        static void Upsert(List<string> lines, string key, string value)
        {
            var prefix = key + "=";
            for (var i = 0; i < lines.Count; i++)
            {
                var t = lines[i].Trim();
                if (t.StartsWith(prefix, StringComparison.OrdinalIgnoreCase))
                {
                    lines[i] = prefix + value;
                    return;
                }
            }
            lines.Add(prefix + value);
        }
    }
}
