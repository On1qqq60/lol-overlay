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
        public static readonly string NamesFile = Path.Combine(Root, "data", "names.txt");
        public static readonly string SettingsFile = Path.Combine(Root, "settings.ini");
    }

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
    }

    public static class AppSettings
    {
        public static double Opacity = 1.0;
        public static int IconSize = 22;
        public static uint Vk = 0x77;
        public static string KeyName = "F8";
        public static bool ShowBuild = true;

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
                }
            }
            catch { }
        }

        public static void Save()
        {
            try
            {
                File.WriteAllText(AppPaths.SettingsFile,
                    "opacity=" + Opacity.ToString("0.##", CultureInfo.InvariantCulture) + "\r\n" +
                    "size=" + IconSize + "\r\n" +
                    "vk=" + Vk + "\r\n" +
                    "key=" + KeyName + "\r\n" +
                    "build=" + (ShowBuild ? "1" : "0") + "\r\n");
            }
            catch { }
        }
    }

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

    public partial class OverlayApp : Application
    {
        [DllImport("user32.dll")]
        private static extern bool RegisterHotKey(IntPtr hWnd, int id, uint fsModifiers, uint vk);

        [DllImport("user32.dll")]
        private static extern bool UnregisterHotKey(IntPtr hWnd, int id);

        [DllImport("user32.dll", SetLastError = true)]
        private static extern bool SetWindowPos(IntPtr hWnd, IntPtr hWndInsertAfter, int x, int y, int cx, int cy, uint flags);

        private static readonly IntPtr HwndTopmost = new IntPtr(-1);
        private const uint SwpNoMove = 0x0002;
        private const uint SwpNoSize = 0x0001;
        private const uint SwpNoActivate = 0x0010;

        private const int HotkeyId = 1;
        private const int WmHotkey = 0x0312;

        private BuildWindow _build;
        private HudWindow _hud;
        private SettingsWindow _settings;
        private WinForms.NotifyIcon _tray;
        private HwndSource _hwnd;
        private DispatcherTimer _poll;
        private bool _connectedOnce;
        private bool _engineBusy;
        private bool _engineReady;
        private RecommendedBuild _original;
        private RecommendedBuild _current;
        private bool _liveLook;
        private MatchState _match;
        private string _invKey;
        private DateTime _lastAnalyzeUtc = DateTime.MinValue;

        [STAThread]
        public static void Main()
        {
            var app = new OverlayApp();
            app.ShutdownMode = ShutdownMode.OnExplicitShutdown;
            app.Startup += (s, e) => app.Boot();
            app.Run();
        }

        private void Boot()
        {
            AppSettings.Load();
            ItemNames.Load();
            ImageCache.LoadAll();

            _original = SimpleBuild.Initial();
            _current = _original;
            _build = new BuildWindow();
            _build.AnalyzeClicked += () => Analyze(true);
            _build.SettingsClicked += OpenSettings;
            _build.Show();
            ApplyVisual();
            _build.UpdateLayout();
            PlaceTopRight(_build, 16, 16);

            _hud = new HudWindow();
            _hud.ShowBuildChanged += SetBuildVisible;
            _hud.Show();
            _hud.UpdateLayout();
            PlaceTopRight(_hud, 10, 8);

            _build.SourceInitialized += (s, e) => BindHotkey();
            _build.Closed += (s, e) => Quit();
            SetupTray();

            _poll = new DispatcherTimer { Interval = TimeSpan.FromSeconds(1) };
            _poll.Tick += (s, e) =>
            {
                BumpTopmost();
                FetchAsync();
            };
            _poll.Start();
            FetchAsync();
            Analyze(false);
            SetBuildVisible(AppSettings.ShowBuild);
            if (_build.IsVisible)
            {
                _build.Left = SystemParameters.WorkArea.Right - _build.Width - 16;
                _build.Top = _hud.Top + _hud.Height + 8;
            }
        }

        private static void PlaceTopRight(Window w, double marginX, double marginY)
        {
            w.Left = SystemParameters.WorkArea.Right - w.Width - marginX;
            w.Top = SystemParameters.WorkArea.Top + marginY;
        }

        private void SetupTray()
        {
            _tray = new WinForms.NotifyIcon();
            _tray.Visible = true;
            _tray.Text = "lol.build";
            _tray.Icon = System.Drawing.SystemIcons.Application;
            var menu = new WinForms.ContextMenu();
            menu.MenuItems.Add("Настройки", (s, e) => OpenSettings());
            menu.MenuItems.Add("Обновить (" + AppSettings.KeyName + ")", (s, e) => Analyze(true));
            menu.MenuItems.Add("Сброс", (s, e) => Reset());
            menu.MenuItems.Add("Выход", (s, e) => Quit());
            _tray.ContextMenu = menu;
        }

        private void BindHotkey()
        {
            var helper = new WindowInteropHelper(_build);
            if (_hwnd == null)
            {
                _hwnd = HwndSource.FromHwnd(helper.Handle);
                if (_hwnd != null) _hwnd.AddHook(WndProc);
            }
            try { UnregisterHotKey(helper.Handle, HotkeyId); } catch { }
            RegisterHotKey(helper.Handle, HotkeyId, 0, AppSettings.Vk);
        }

        private IntPtr WndProc(IntPtr hwnd, int msg, IntPtr wParam, IntPtr lParam, ref bool handled)
        {
            if (msg == WmHotkey && wParam.ToInt32() == HotkeyId)
            {
                Analyze(true);
                handled = true;
            }
            return IntPtr.Zero;
        }

        private void Analyze(bool preferLive)
        {
            if (_engineBusy) return;
            _engineBusy = true;
            _lastAnalyzeUtc = DateTime.UtcNow;
            if (_build != null) _build.SetBusy(true);
            ThreadPool.QueueUserWorkItem(_ =>
            {
                EngineResult result = null;
                try { result = EngineClient.Recommend(preferLive); }
                catch (Exception ex) { result = new EngineResult { Error = ex.Message }; }
                Dispatcher.BeginInvoke(new Action(() => ApplyEngine(result)));
            });
        }

        private void ApplyEngine(EngineResult result)
        {
            _engineBusy = false;
            if (_build != null) _build.SetBusy(false);
            if (result == null || !result.Ok)
            {
                var fallback = SimpleBuild.Adapt(_original, _match);
                var err = result != null ? result.Error : "движок недоступен";
                if (!SimpleBuild.Same(_current, fallback))
                {
                    fallback.Changes = SimpleBuild.Diff(_original, fallback);
                    _current = fallback;
                }
                _liveLook = _match != null && LiveClient.Connected;
                ApplyVisual();
                _build.SetChanges(_current.Changes ?? new ItemChange[0], err);
                return;
            }

            var next = result.Build;
            var title = EngineTitle(result);
            if (!_engineReady || (result.FromLive && !_liveLook))
            {
                _original = CloneBuild(next);
                _engineReady = true;
            }

            var changes = SimpleBuild.Diff(_original, next);
            next.Changes = changes;
            _current = next;
            _liveLook = result.FromLive || LiveClient.Connected;
            ApplyVisual();
            _build.SetChanges(changes, title);
            _build.SetReasons(result.Reasons);
        }

        private static string EngineTitle(EngineResult result)
        {
            var seed = string.IsNullOrEmpty(result.SeedName) ? "сборка" : result.SeedName;
            if (result.FromLive) seed = seed + " · live";
            else seed = seed + " · демо";
            if (!string.IsNullOrEmpty(result.NextName))
                seed = seed + " → " + result.NextName;
            return seed;
        }

        private static RecommendedBuild CloneBuild(RecommendedBuild src)
        {
            if (src == null) return SimpleBuild.Initial();
            return new RecommendedBuild
            {
                Items = (string[])src.Items.Clone(),
                Boots = src.Boots,
                NextIndex = src.NextIndex,
                Changes = new ItemChange[0]
            };
        }

        private void ApplyVisual()
        {
            if (_build != null)
            {
                _build.Opacity = AppSettings.Opacity;
                _build.SetBuild(_current, _liveLook);
            }
            if (_hud != null) _hud.Opacity = AppSettings.Opacity;
        }

        private void SetBuildVisible(bool visible)
        {
            AppSettings.ShowBuild = visible;
            if (_build == null) return;
            if (visible)
            {
                _build.Show();
                _build.Topmost = true;
            }
            else
            {
                _build.HideDetails();
                _build.Hide();
            }
        }

        private void OpenSettings()
        {
            if (_settings != null)
            {
                _settings.Activate();
                return;
            }
            _settings = new SettingsWindow();
            _settings.Changed += () =>
            {
                AppSettings.Save();
                BindHotkey();
                ApplyVisual();
            };
            _settings.Closed += (s, e) => { _settings = null; };
            _settings.Show();
        }

        private void BumpTopmost()
        {
            Raise(_build);
            Raise(_hud);
        }

        private static void Raise(Window w)
        {
            if (w == null || !w.IsVisible) return;
            try
            {
                w.Topmost = false;
                w.Topmost = true;
                var h = new WindowInteropHelper(w).Handle;
                if (h != IntPtr.Zero)
                    SetWindowPos(h, HwndTopmost, 0, 0, 0, 0, SwpNoMove | SwpNoSize | SwpNoActivate);
            }
            catch { }
        }

        private void FetchAsync()
        {
            ThreadPool.QueueUserWorkItem(_ =>
            {
                MatchState match = null;
                try { match = LiveClient.Fetch(); }
                catch { match = null; }
                Dispatcher.BeginInvoke(new Action(() =>
                {
                    _match = match;
                    SetTray(match);
                }));
            });
        }

        private void SetTray(MatchState match)
        {
            if (_tray == null) return;
            try
            {
                var text = "lol.build · " + LiveClient.Status;
                if (text.Length > 63) text = text.Substring(0, 63);
                _tray.Text = text;
                if (LiveClient.Connected && match != null && match.Me != null)
                {
                    if (!_connectedOnce)
                    {
                        _connectedOnce = true;
                        _tray.ShowBalloonTip(2000, "lol.build", "Игра: " + match.Me.championName, WinForms.ToolTipIcon.Info);
                        Analyze(true);
                    }
                    else
                    {
                        var key = InventoryKey(match);
                        var stale = (DateTime.UtcNow - _lastAnalyzeUtc).TotalSeconds >= 4;
                        if (stale || key != _invKey)
                        {
                            _invKey = key;
                            Analyze(true);
                        }
                    }
                }
                if (!LiveClient.Connected)
                {
                    _connectedOnce = false;
                    _invKey = null;
                }
            }
            catch { }
        }

        private static string InventoryKey(MatchState match)
        {
            if (match == null || match.Me == null || match.Me.items == null)
                return "";
            var sb = new StringBuilder();
            for (var i = 0; i < match.Me.items.Count; i++)
            {
                var it = match.Me.items[i];
                if (it == null || it.itemID <= 0) continue;
                sb.Append(it.itemID).Append(',');
            }
            sb.Append('|').Append((int)(match.Gold / 50));
            return sb.ToString();
        }

        private void Reset()
        {
            _liveLook = false;
            _current = _original;
            ApplyVisual();
            _build.HideDetails();
        }

        private void Quit()
        {
            if (_poll != null) _poll.Stop();
            if (_hwnd != null)
            {
                try { UnregisterHotKey(_hwnd.Handle, HotkeyId); } catch { }
            }
            if (_tray != null)
            {
                _tray.Visible = false;
                _tray.Dispose();
            }
            Shutdown();
        }
    }

    public static class SimpleBuild
    {
        public static RecommendedBuild Initial()
        {
            return new RecommendedBuild
            {
                Items = new[] { "6672", "3031", "3072", "3094", "3026" },
                Boots = "3006",
                NextIndex = 0
            };
        }

        public static RecommendedBuild Adapt(RecommendedBuild original, MatchState match)
        {
            var items = (string[])original.Items.Clone();
            var boots = original.Boots;
            var armor = 0;
            var ap = 0;
            var heal = 0;

            if (match != null && match.Enemies != null)
            {
                foreach (var e in match.Enemies)
                {
                    if (e.items == null) continue;
                    foreach (var it in e.items)
                    {
                        if (it == null || it.slot >= 6) continue;
                        var id = it.itemID.ToString();
                        if (id == "3075" || id == "3076" || id == "3110" || id == "3068" || id == "3047" || id == "3143" || id == "3742")
                            armor++;
                        if (id == "3089" || id == "3118" || id == "3157" || id == "4645" || id == "3020")
                            ap++;
                        if (id == "3072" || id == "3153" || id == "3083" || id == "3107")
                            heal++;
                    }
                }
            }

            if (match == null)
            {
                armor = 2;
                ap = 1;
            }

            if (armor >= 1) items[1] = "3036";
            if (heal >= 1) items[1] = "3033";
            if (ap >= 1) items[2] = "3156";
            if (armor >= 2) boots = "3047";
            else if (ap >= 2) boots = "3111";

            return new RecommendedBuild { Items = items, Boots = boots, NextIndex = 1 };
        }

        public static bool Same(RecommendedBuild a, RecommendedBuild b)
        {
            if (a == null || b == null) return false;
            if (a.Boots != b.Boots) return false;
            if (a.Items == null || b.Items == null || a.Items.Length != b.Items.Length) return false;
            for (var i = 0; i < a.Items.Length; i++)
            {
                if (a.Items[i] != b.Items[i]) return false;
            }
            return true;
        }

        public static ItemChange[] Diff(RecommendedBuild original, RecommendedBuild now)
        {
            var list = new List<ItemChange>();
            for (var i = 0; i < 5; i++)
            {
                if (original.Items[i] == now.Items[i]) continue;
                list.Add(new ItemChange { from = original.Items[i], to = now.Items[i], why = Why(now.Items[i]) });
            }
            if (original.Boots != now.Boots)
                list.Add(new ItemChange { from = original.Boots, to = now.Boots, why = Why(now.Boots) });
            return list.ToArray();
        }

        private static string Why(string id)
        {
            switch (id)
            {
                case "3036": return "Враги на броне — LDR вместо предмета из начального списка.";
                case "3033": return "У врагов хил — Mortal Reminder.";
                case "3156": return "AP угроза — Maw вместо чистого урона.";
                case "3047": return "Много AD — Steelcaps вместо начальных ботинок.";
                case "3111": return "AP/CC — Mercury's вместо начальных ботинок.";
                default: return "Замена относительно изначального списка.";
            }
        }
    }

    public class BuildWindow : Window
    {
        private readonly StackPanel _row;
        private readonly StackPanel _changes;
        private readonly Border _chrome;
        private readonly Border _body;
        private readonly Border _details;
        private readonly Button _scan;
        private readonly Button _gear;
        private readonly Button _infoBtn;
        private bool _infoOpen;
        private bool _infoDragged;
        private Point _infoDown;
        private ItemChange[] _last;
        private string _lastTitle;

        public event Action AnalyzeClicked;
        public event Action SettingsClicked;

        public BuildWindow()
        {
            Title = "lol.build";
            WindowStyle = WindowStyle.None;
            AllowsTransparency = true;
            Background = Brushes.Transparent;
            ResizeMode = ResizeMode.NoResize;
            Topmost = true;
            ShowInTaskbar = false;
            SizeToContent = SizeToContent.WidthAndHeight;

            _row = new StackPanel { Orientation = Orientation.Horizontal };
            _scan = Ui.MiniButton("▶", "Обновить сборку");
            _scan.Click += (s, e) => { e.Handled = true; if (AnalyzeClicked != null) AnalyzeClicked(); };
            _gear = Ui.MiniButton("⚙", "Настройки");
            _gear.Click += (s, e) => { e.Handled = true; if (SettingsClicked != null) SettingsClicked(); };

            _infoBtn = new Button
            {
                Cursor = Cursors.SizeAll,
                Template = Ui.GhostCircleTemplate(),
                Content = Ui.StarGlyph(),
                ToolTip = "Показать замены"
            };
            _infoBtn.Click += (s, e) =>
            {
                e.Handled = true;
                if (_infoDragged) return;
                ToggleInfo();
            };
            _infoBtn.PreviewMouseLeftButtonDown += (s, e) =>
            {
                _infoDown = e.GetPosition(this);
                _infoDragged = false;
            };
            _infoBtn.PreviewMouseMove += (s, e) =>
            {
                if (e.LeftButton != MouseButtonState.Pressed) return;
                var now = e.GetPosition(this);
                if (Math.Abs(now.X - _infoDown.X) + Math.Abs(now.Y - _infoDown.Y) <= 4) return;
                _infoDragged = true;
                DragMove();
            };

            var top = new StackPanel { Orientation = Orientation.Horizontal };
            top.Children.Add(_infoBtn);
            top.Children.Add(new FrameworkElement { Width = 6 });
            top.Children.Add(_row);
            top.Children.Add(new FrameworkElement { Width = 8 });
            top.Children.Add(_scan);
            top.Children.Add(_gear);

            _changes = new StackPanel();
            _details = new Border
            {
                HorizontalAlignment = HorizontalAlignment.Left,
                Background = Brushes.Transparent,
                BorderThickness = new Thickness(0),
                Padding = new Thickness(0, 6, 0, 0),
                Visibility = Visibility.Collapsed,
                Child = _changes
            };

            var col = new StackPanel();
            col.Children.Add(top);
            col.Children.Add(_details);

            _body = new Border
            {
                Background = Brushes.Transparent,
                Padding = new Thickness(0),
                Child = col,
                Cursor = Cursors.Arrow
            };

            _chrome = new Border
            {
                Background = Brushes.Transparent,
                BorderBrush = Brushes.Transparent,
                BorderThickness = new Thickness(0),
                Padding = new Thickness(2),
                Child = _body,
                Cursor = Cursors.SizeAll
            };
            Content = _chrome;
            _chrome.MouseLeftButtonDown += (s, e) =>
            {
                if (e.ChangedButton != MouseButton.Left) return;
                var p = e.GetPosition(_chrome);
                const double grip = 4;
                var onEdge = p.X <= grip || p.Y <= grip
                    || p.X >= _chrome.ActualWidth - grip
                    || p.Y >= _chrome.ActualHeight - grip;
                if (!onEdge) return;
                DragMove();
                e.Handled = true;
            };

            SetChanges(new ItemChange[0], "Сборка");
        }

        public void SetBusy(bool busy) { Opacity = busy ? 0.5 : AppSettings.Opacity; }

        public void SetBuild(RecommendedBuild build, bool live)
        {
            var size = AppSettings.IconSize;
            var star = size;
            var ctrl = Math.Max(12, size - 10);
            _infoBtn.Width = _infoBtn.Height = star;
            _scan.Width = _scan.Height = ctrl;
            _gear.Width = _gear.Height = ctrl;
            _scan.Margin = new Thickness(0, 0, 4, 0);
            _gear.Margin = new Thickness(0, 0, 0, 0);
            if (_last != null) SetChanges(_last, _lastTitle);

            _chrome.BorderBrush = Brushes.Transparent;

            _row.Children.Clear();
            for (var i = 0; i < 5; i++)
                _row.Children.Add(Ui.ItemIcon(build.Items[i], size, build.NextIndex == i));
            var boot = Ui.ItemIcon(build.Boots, size, build.NextIndex == 5);
            boot.Margin = new Thickness(10, 2, 3, 2);
            boot.BorderBrush = build.NextIndex == 5
                ? new SolidColorBrush(Color.FromRgb(200, 170, 110))
                : new SolidColorBrush(Color.FromArgb(140, 10, 200, 185));
            _row.Children.Add(boot);
        }

        public void SetChanges(ItemChange[] changes, string title)
        {
            _last = changes;
            _lastTitle = title;
            _changes.Children.Clear();

            if (!string.IsNullOrEmpty(title))
                _changes.Children.Add(Ui.GlowText(title, 13, Color.FromRgb(232, 196, 110)));

            if (changes == null || changes.Length == 0)
            {
                _changes.Children.Add(Ui.GlowText("Со сборкой всё в порядке.", 13, Color.FromRgb(180, 230, 170)));
                return;
            }

            var size = AppSettings.IconSize;
            for (var i = 0; i < changes.Length; i++)
            {
                var c = changes[i];
                var block = new StackPanel { Margin = new Thickness(0, 0, 0, 8) };
                var row = new StackPanel { Orientation = Orientation.Horizontal };
                row.Children.Add(Ui.ItemIcon(c.from, size, false));
                row.Children.Add(Ui.GlowText("  →  ", 14, Color.FromRgb(232, 166, 69)));
                row.Children.Add(Ui.ItemIcon(c.to, size, true));
                block.Children.Add(row);
                block.Children.Add(Ui.GlowText(
                    ItemNames.Get(c.from) + "  →  " + ItemNames.Get(c.to),
                    12, Color.FromRgb(255, 236, 200)));
                block.Children.Add(Ui.GlowText(c.why, 12, Color.FromRgb(230, 226, 218)));
                _changes.Children.Add(block);
            }
        }

        public void SetReasons(string[] reasons)
        {
            if (reasons == null || reasons.Length == 0) return;
            var n = 0;
            for (var i = 0; i < reasons.Length && n < 8; i++)
            {
                var line = reasons[i];
                if (string.IsNullOrEmpty(line)) continue;
                _changes.Children.Add(Ui.GlowText("• " + line, 11, Color.FromRgb(180, 176, 168)));
                n++;
            }
        }

        private void ToggleInfo()
        {
            _infoOpen = !_infoOpen;
            _details.Visibility = _infoOpen ? Visibility.Visible : Visibility.Collapsed;
        }

        public void HideDetails()
        {
            _infoOpen = false;
            _details.Visibility = Visibility.Collapsed;
        }
    }

    public class HudWindow : Window
    {
        public event Action<bool> ShowBuildChanged;

        private readonly Button _lol;
        private readonly Popup _popup;
        private readonly Border _menu;
        private readonly ColorToggle _toggle;

        public HudWindow()
        {
            WindowStyle = WindowStyle.None;
            AllowsTransparency = true;
            Background = Brushes.Transparent;
            ResizeMode = ResizeMode.NoResize;
            Topmost = true;
            ShowInTaskbar = false;
            SizeToContent = SizeToContent.WidthAndHeight;

            _lol = new Button
            {
                Width = 28,
                Height = 16,
                Cursor = Cursors.Hand,
                ToolTip = "Меню оверлея",
                Template = Ui.GhostButtonTemplate(),
                Content = Ui.GlowText("LOL", 10, Color.FromRgb(232, 196, 110))
            };
            _lol.Click += (s, e) =>
            {
                e.Handled = true;
                ToggleMenu();
            };

            _toggle = new ColorToggle(AppSettings.ShowBuild);
            _toggle.Changed += OnToggle;

            var row = new StackPanel
            {
                Orientation = Orientation.Horizontal,
                VerticalAlignment = VerticalAlignment.Center
            };
            row.Children.Add(Ui.GlowText("Рекомендуемый билд", 12, Color.FromRgb(232, 228, 217)));
            var gap = new FrameworkElement { Width = 10 };
            row.Children.Add(gap);
            row.Children.Add(_toggle);

            _menu = new Border
            {
                Background = Brushes.Transparent,
                BorderThickness = new Thickness(0),
                Padding = new Thickness(0, 4, 0, 0),
                Child = row
            };

            _popup = new Popup
            {
                AllowsTransparency = true,
                Placement = PlacementMode.Bottom,
                PlacementTarget = _lol,
                StaysOpen = false,
                Child = _menu
            };

            var root = new Grid { Width = 28, Height = 16 };
            root.Children.Add(_lol);
            root.Children.Add(_popup);
            Content = root;
        }

        private void ToggleMenu()
        {
            if (_popup.IsOpen)
            {
                _popup.IsOpen = false;
                return;
            }
            _menu.Measure(new Size(double.PositiveInfinity, double.PositiveInfinity));
            _popup.HorizontalOffset = _lol.ActualWidth - _menu.DesiredSize.Width;
            _popup.VerticalOffset = 6;
            _popup.IsOpen = true;
        }

        private void OnToggle(bool on)
        {
            AppSettings.ShowBuild = on;
            AppSettings.Save();
            if (ShowBuildChanged != null) ShowBuildChanged(on);
        }
    }

    public class ColorToggle : Grid
    {
        public event Action<bool> Changed;

        private readonly Border _track;
        private readonly Border _knob;
        private bool _on;

        public ColorToggle(bool on)
        {
            Width = 42;
            Height = 22;
            Cursor = Cursors.Hand;
            VerticalAlignment = VerticalAlignment.Center;

            _track = new Border { CornerRadius = new CornerRadius(11) };
            _knob = new Border
            {
                Width = 16,
                Height = 16,
                CornerRadius = new CornerRadius(8),
                Background = Brushes.White,
                Margin = new Thickness(3, 0, 3, 0),
                VerticalAlignment = VerticalAlignment.Center
            };
            Children.Add(_track);
            Children.Add(_knob);
            MouseLeftButtonDown += (s, e) =>
            {
                e.Handled = true;
                Set(!_on, true);
            };
            Set(on, false);
        }

        public void Set(bool on, bool notify)
        {
            _on = on;
            _track.Background = new SolidColorBrush(on
                ? Color.FromRgb(46, 180, 80)
                : Color.FromRgb(210, 50, 50));
            _knob.HorizontalAlignment = on ? HorizontalAlignment.Right : HorizontalAlignment.Left;
            if (notify && Changed != null) Changed(on);
        }
    }

    public class SettingsWindow : Window
    {
        public event Action Changed;
        private bool _capture;
        private readonly Button _bind;

        public SettingsWindow()
        {
            Title = "Настройки";
            Width = 320;
            Height = 260;
            Topmost = true;
            ResizeMode = ResizeMode.NoResize;
            WindowStartupLocation = WindowStartupLocation.CenterScreen;
            Background = new SolidColorBrush(Color.FromRgb(14, 16, 20));
            Foreground = new SolidColorBrush(Color.FromRgb(232, 228, 217));

            var root = new StackPanel { Margin = new Thickness(16) };

            root.Children.Add(Label("Прозрачность"));
            var op = new Slider { Minimum = 0.25, Maximum = 1, Value = AppSettings.Opacity, TickFrequency = 0.05 };
            op.ValueChanged += (s, e) =>
            {
                AppSettings.Opacity = op.Value;
                if (Changed != null) Changed();
            };
            root.Children.Add(op);

            root.Children.Add(Label("Размер иконок"));
            var sz = new Slider { Minimum = 16, Maximum = 40, Value = AppSettings.IconSize, IsSnapToTickEnabled = true, TickFrequency = 1 };
            sz.ValueChanged += (s, e) =>
            {
                AppSettings.IconSize = (int)sz.Value;
                if (Changed != null) Changed();
            };
            root.Children.Add(sz);

            root.Children.Add(Label("Бинд обновления данных"));
            _bind = new Button
            {
                Content = AppSettings.KeyName,
                Height = 32,
                Margin = new Thickness(0, 4, 0, 0),
                Background = new SolidColorBrush(Color.FromRgb(28, 22, 12)),
                Foreground = new SolidColorBrush(Color.FromRgb(200, 170, 110)),
                BorderBrush = new SolidColorBrush(Color.FromRgb(200, 170, 110))
            };
            _bind.Click += (s, e) =>
            {
                _capture = true;
                _bind.Content = "нажмите клавишу…";
            };
            root.Children.Add(_bind);

            Content = root;
            PreviewKeyDown += OnKey;
        }

        private void OnKey(object sender, KeyEventArgs e)
        {
            if (!_capture) return;
            e.Handled = true;
            var key = e.Key == Key.System ? e.SystemKey : e.Key;
            var vk = (uint)KeyInterop.VirtualKeyFromKey(key);
            if (vk == 0) return;
            AppSettings.Vk = vk;
            AppSettings.KeyName = key.ToString();
            _bind.Content = AppSettings.KeyName;
            _capture = false;
            if (Changed != null) Changed();
        }

        private static TextBlock Label(string t)
        {
            return new TextBlock
            {
                Text = t,
                Margin = new Thickness(0, 10, 0, 4),
                Foreground = new SolidColorBrush(Color.FromRgb(200, 170, 110)),
                FontSize = 12
            };
        }
    }

    public static class Ui
    {
        public static Button MiniButton(string text, string tip)
        {
            return new Button
            {
                Width = 14,
                Height = 14,
                Margin = new Thickness(0, 0, 4, 0),
                Cursor = Cursors.Hand,
                ToolTip = tip,
                Template = GhostButtonTemplate(),
                Content = GlowText(text, 9, Color.FromRgb(232, 196, 110))
            };
        }

        public static Border ItemIcon(string id, int size, bool next)
        {
            var src = ImageCache.Get(id);
            UIElement child;
            if (src != null)
            {
                child = new Image { Stretch = Stretch.UniformToFill, Source = src };
            }
            else
            {
                var label = ItemNames.Get(id);
                if (string.IsNullOrEmpty(label)) label = id ?? "?";
                child = new TextBlock
                {
                    Text = label,
                    FontSize = Math.Max(7, size / 5.0),
                    Foreground = Brushes.White,
                    TextWrapping = TextWrapping.Wrap,
                    TextAlignment = TextAlignment.Center,
                    VerticalAlignment = VerticalAlignment.Center,
                    HorizontalAlignment = HorizontalAlignment.Center,
                    Margin = new Thickness(1)
                };
            }
            var border = new Border
            {
                Width = size,
                Height = size,
                Margin = new Thickness(3, 2, 3, 2),
                BorderThickness = new Thickness(1),
                BorderBrush = next
                    ? new SolidColorBrush(Color.FromRgb(232, 196, 110))
                    : new SolidColorBrush(Color.FromArgb(140, 255, 255, 255)),
                Background = src != null
                    ? Brushes.Transparent
                    : new SolidColorBrush(Color.FromArgb(180, 20, 18, 14)),
                Cursor = Cursors.Arrow,
                ClipToBounds = true,
                SnapsToDevicePixels = true,
                ToolTip = ItemNames.Get(id),
                Child = child
            };
            ToolTipService.SetInitialShowDelay(border, 120);
            ToolTipService.SetShowDuration(border, 12000);
            return border;
        }

        public static DropShadowEffect Glow()
        {
            return new DropShadowEffect
            {
                Color = Colors.Black,
                BlurRadius = 6,
                ShadowDepth = 0,
                Opacity = 0.95
            };
        }

        public static TextBlock GlowText(string text, double size, Color color)
        {
            return new TextBlock
            {
                Text = text,
                FontSize = size,
                FontWeight = FontWeights.SemiBold,
                Foreground = new SolidColorBrush(color),
                TextWrapping = TextWrapping.Wrap,
                VerticalAlignment = VerticalAlignment.Center,
                Effect = Glow()
            };
        }

        public static ControlTemplate GhostButtonTemplate()
        {
            var template = new ControlTemplate(typeof(Button));
            var border = new FrameworkElementFactory(typeof(Border));
            border.SetValue(Border.BackgroundProperty, Brushes.Transparent);
            border.SetValue(Border.BorderThicknessProperty, new Thickness(0));
            var presenter = new FrameworkElementFactory(typeof(ContentPresenter));
            presenter.SetValue(FrameworkElement.HorizontalAlignmentProperty, HorizontalAlignment.Center);
            presenter.SetValue(FrameworkElement.VerticalAlignmentProperty, VerticalAlignment.Center);
            border.AppendChild(presenter);
            template.VisualTree = border;
            return template;
        }

        public static ControlTemplate GhostCircleTemplate()
        {
            return GhostButtonTemplate();
        }

        public static ControlTemplate CircleButtonTemplate()
        {
            return GhostButtonTemplate();
        }

        public static ControlTemplate ScanButtonTemplate()
        {
            return GhostButtonTemplate();
        }

        public static UIElement StarGlyph()
        {
            return new TextBlock
            {
                Text = "★",
                FontSize = 16,
                FontWeight = FontWeights.Bold,
                Foreground = new SolidColorBrush(Color.FromRgb(232, 196, 110)),
                HorizontalAlignment = HorizontalAlignment.Center,
                VerticalAlignment = VerticalAlignment.Center,
                Effect = Glow()
            };
        }

        public static UIElement InfoGlyph()
        {
            return StarGlyph();
        }
    }
}
