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
    public static class AppPaths
    {
        public static readonly string Root = AppDomain.CurrentDomain.BaseDirectory;
        public static readonly string ItemsDir = Path.Combine(Root, "assets", "items");
        public static readonly string NamesFile = FirstExisting(
            Path.Combine(Root, "data", "names.txt"),
            Path.Combine(Root, "..", "data", "names.txt"));
        public static readonly string SettingsFile = FirstExisting(
            Path.Combine(Root, "settings.ini"),
            Path.Combine(Root, "..", "settings.ini"));
        public static readonly string AppIconFile = FirstExisting(
            Path.Combine(Root, "app.ico"),
            Path.Combine(Root, "..", "app.ico"));
        public static readonly string PromptsDir = FirstDir(
            Path.Combine(Root, "prompts"),
            Path.GetFullPath(Path.Combine(Root, "..", "prompts")));

        static string FirstExisting(string besideExe, string besideProject)
        {
            if (File.Exists(besideExe)) return besideExe;
            var project = Path.GetFullPath(besideProject);
            if (File.Exists(project)) return project;
            return besideExe;
        }

        static string FirstDir(string besideExe, string besideProject)
        {
            if (Directory.Exists(besideExe)) return besideExe;
            if (Directory.Exists(besideProject)) return besideProject;
            return besideProject;
        }
    }
}
