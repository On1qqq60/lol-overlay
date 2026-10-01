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
    public static class ImageCache
    {
        private const string Cdn = "https://ddragon.leagueoflegends.com/cdn/16.18.1/img/item/";
        private static readonly Dictionary<string, BitmapImage> Map = new Dictionary<string, BitmapImage>();
        private static readonly object Gate = new object();

        static ImageCache()
        {
            try
            {
                ServicePointManager.SecurityProtocol =
                    (SecurityProtocolType)3072 | (SecurityProtocolType)768 | (SecurityProtocolType)192;
            }
            catch { }
        }

        public static void LoadAll()
        {
            Map.Clear();
            if (!Directory.Exists(AppPaths.ItemsDir))
            {
                try { Directory.CreateDirectory(AppPaths.ItemsDir); } catch { return; }
            }
            foreach (var file in Directory.GetFiles(AppPaths.ItemsDir, "*.png"))
            {
                var bmp = FromFile(file);
                if (bmp != null)
                    Map[Path.GetFileNameWithoutExtension(file)] = bmp;
            }
        }

        public static ImageSource Get(string id)
        {
            if (string.IsNullOrEmpty(id) || id == "0") return null;
            BitmapImage bmp;
            if (Map.TryGetValue(id, out bmp) && bmp != null) return bmp;
            lock (Gate)
            {
                if (Map.TryGetValue(id, out bmp) && bmp != null) return bmp;
                var path = Path.Combine(AppPaths.ItemsDir, id + ".png");
                bmp = FromFile(path);
                if (bmp == null) bmp = Download(id, path);
                if (bmp != null) Map[id] = bmp;
                return bmp;
            }
        }

        private static BitmapImage Download(string id, string path)
        {
            try
            {
                Directory.CreateDirectory(AppPaths.ItemsDir);
                using (var wc = new WebClient())
                    wc.DownloadFile(Cdn + id + ".png", path);
                return FromFile(path);
            }
            catch
            {
                return null;
            }
        }

        private static BitmapImage FromFile(string path)
        {
            if (string.IsNullOrEmpty(path) || !File.Exists(path) || new FileInfo(path).Length < 32)
                return null;
            try
            {
                var bmp = new BitmapImage();
                bmp.BeginInit();
                bmp.UriSource = new Uri(path, UriKind.Absolute);
                bmp.CacheOption = BitmapCacheOption.OnLoad;
                bmp.CreateOptions = BitmapCreateOptions.IgnoreImageCache;
                bmp.DecodePixelWidth = 64;
                bmp.EndInit();
                bmp.Freeze();
                return bmp;
            }
            catch
            {
                return null;
            }
        }
    }
}
