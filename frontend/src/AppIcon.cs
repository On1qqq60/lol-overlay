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
    public static class AppIcon
    {
        public static ImageSource Wpf()
        {
            try
            {
                if (!File.Exists(AppPaths.AppIconFile)) return null;
                var bmp = new BitmapImage();
                bmp.BeginInit();
                bmp.UriSource = new Uri(AppPaths.AppIconFile, UriKind.Absolute);
                bmp.CacheOption = BitmapCacheOption.OnLoad;
                bmp.EndInit();
                bmp.Freeze();
                return bmp;
            }
            catch { return null; }
        }

        public static System.Drawing.Icon WinForms()
        {
            try
            {
                if (File.Exists(AppPaths.AppIconFile))
                    return new System.Drawing.Icon(AppPaths.AppIconFile);
            }
            catch { }
            return System.Drawing.SystemIcons.Application;
        }
    }
}
