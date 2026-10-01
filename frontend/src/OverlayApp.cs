using System;
using System.Collections.Generic;
using System.Globalization;
using System.IO;
using System.Net;
using System.Runtime.InteropServices;
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
    public partial class OverlayApp : Application
    {
        [DllImport("user32.dll")]
        private static extern bool RegisterHotKey(IntPtr hWnd, int id, uint fsModifiers, uint vk);

        [DllImport("user32.dll")]
        private static extern bool UnregisterHotKey(IntPtr hWnd, int id);

        [DllImport("user32.dll", SetLastError = true)]
        private static extern bool SetWindowPos(IntPtr hWnd, IntPtr hWndInsertAfter, int x, int y, int cx, int cy, uint flags);

        [DllImport("shell32.dll", CharSet = CharSet.Unicode)]
        private static extern int SetCurrentProcessExplicitAppUserModelID(string appId);

        private static readonly IntPtr HwndTopmost = new IntPtr(-1);
        private const uint SwpNoMove = 0x0002;
        private const uint SwpNoSize = 0x0001;
        private const uint SwpNoActivate = 0x0010;

        private const int HotkeyId = 1;
        private const int HideHotkeyId = 2;
        private const int WmHotkey = 0x0312;

        private BuildWindow _build;
        private SettingsWindow _settings;
        private WinForms.NotifyIcon _tray;
        private HwndSource _hwnd;
        private DispatcherTimer _poll;
        private DispatcherTimer _live;
        private bool _connectedOnce;
        private bool _engineBusy;
        private bool _engineReady;
        private RecommendedBuild _original;
        private RecommendedBuild _current;
        private string _style;
        private string _styleLabel;
        private string _damage;
        private string _wish = "";
        private bool _concealed;
        private string _planText = "";
        private bool _liveLook;
        private MatchState _match;
        private int _starts;
        private double _lastPost7 = -1;
        private bool _fetching;
        private double _clockGame;
        private DateTime _clockUtc = DateTime.UtcNow;
        private bool _clockInGame;
        private static System.Threading.Mutex _instance;

        [STAThread]
        public static void Main()
        {
            var mine = false;
            _instance = new System.Threading.Mutex(false, "Local\\LolBuildOverlay");
            try { mine = _instance.WaitOne(0); }
            catch (AbandonedMutexException) { mine = true; }
            if (!mine) return;
            try { Portable.Ensure(); } catch { }
            var app = new OverlayApp();
            app.ShutdownMode = ShutdownMode.OnExplicitShutdown;
            app.Startup += (s, e) => app.Boot();
            app.Run();
        }

        private void Boot()
        {
            try { SetCurrentProcessExplicitAppUserModelID(ToastNote.AppId); } catch { }
            AppSettings.Load();
            ItemNames.Load();
            ImageCache.LoadAll();

            _original = SimpleBuild.Initial();
            _current = _original;
            _build = new BuildWindow();
            _build.AnalyzeClicked += () => Analyze(true);
            _build.StyleClicked += ChooseStyle;
            _build.DamagePicked += ChooseDamage;
            _build.WishSubmitted += SubmitWish;
            _build.SettingsClicked += OpenSettings;
            _build.Show();
            ApplyVisual();
            _build.UpdateLayout();
            PlaceTopRight(_build, 16, 16);

            _build.SourceInitialized += (s, e) => BindHotkey();
            _build.Closed += (s, e) => Quit();
            SetupTray();

            _poll = new DispatcherTimer { Interval = TimeSpan.FromSeconds(1) };
            _poll.Tick += (s, e) => { BumpTopmost(); TickClock(); };
            _poll.Start();
            _live = new DispatcherTimer { Interval = TimeSpan.FromSeconds(2) };
            _live.Tick += (s, e) => FetchAsync();
            _live.Start();
            FetchAsync();
            Analyze(true);
            SetBuildVisible(AppSettings.ShowBuild);
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
            _tray.Icon = AppIcon.WinForms();
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
            try { UnregisterHotKey(helper.Handle, HideHotkeyId); } catch { }
            RegisterHotKey(helper.Handle, HotkeyId, 0, AppSettings.Vk);
            if (AppSettings.HideVk != AppSettings.Vk)
                RegisterHotKey(helper.Handle, HideHotkeyId, 0, AppSettings.HideVk);
        }

        private IntPtr WndProc(IntPtr hwnd, int msg, IntPtr wParam, IntPtr lParam, ref bool handled)
        {
            if (msg == WmHotkey && wParam.ToInt32() == HotkeyId)
            {
                Analyze(true);
                handled = true;
            }
            else if (msg == WmHotkey && wParam.ToInt32() == HideHotkeyId)
            {
                ToggleConceal();
                handled = true;
            }
            return IntPtr.Zero;
        }

        private void ToggleConceal()
        {
            if (_build == null) return;
            if (_build.IsVisible)
            {
                _concealed = true;
                _build.Hide();
                return;
            }
            _concealed = false;
            AppSettings.ShowBuild = true;
            _build.Show();
            _build.Topmost = true;
        }

        private void ChooseDamage(string kind)
        {
            var same = _style == "damage" && _damage == kind;
            _damage = kind;
            if (_build != null) _build.SetActiveDamage(kind);
            if (same)
            {
                AskModel(false);
                return;
            }
            if (_style == "damage")
            {
                _styleLabel = StyleLabel("damage");
                AskModel(true);
                return;
            }
            ChooseStyle("damage");
        }

        private void SubmitWish(string text)
        {
            _wish = text ?? "";
            if (_wish.Length == 0) return;
            var had = !string.IsNullOrEmpty(_style);
            if (!had)
            {
                _style = "standard";
                _styleLabel = StyleLabel("standard");
                if (_build != null) _build.SetActiveStyle("standard");
            }
            var before7 = NowGame() < 7 * 60;
            AskModel(!had || (before7 && _starts < 3));
        }

        private void ChooseStyle(string style)
        {
            if (style != "damage")
            {
                _damage = "";
                if (_build != null)
                {
                    _build.SetActiveDamage("");
                    _build.HideDamageChoices();
                }
            }
            if (!string.IsNullOrEmpty(_style) && style == _style)
            {
                AskModel(false);
                return;
            }
            var deny = Deny(true);
            if (deny != null)
            {
                Notify(deny);
                return;
            }
            if (_engineBusy)
            {
                Notify("Запрос уже ушёл в модель и ответ ещё не пришёл. Пока он не вернётся, новый запрос не отправляется.");
                return;
            }
            _style = style;
            _styleLabel = StyleLabel(style);
            if (_build != null) _build.SetActiveStyle(style);
            AskModel(true);
        }

        string StyleLabel(string style)
        {
            if (style == "damage" && _damage == "ap") return "дамажный, магический";
            if (style == "damage" && _damage == "ad") return "дамажный, физический";
            if (style == "damage") return "дамажный";
            if (style == "defensive") return "защитный";
            return "стандартный";
        }

        private void AskModel(bool start)
        {
            if (string.IsNullOrEmpty(_style)) return;
            var deny = Deny(start);
            if (deny != null)
            {
                Notify(deny);
                return;
            }
            if (_engineBusy)
            {
                Notify("Запрос уже ушёл в модель и ответ ещё не пришёл. Пока он не вернётся, новый запрос не отправляется.");
                return;
            }
            var explain = NowGame() >= 7 * 60;
            var stamp = NowGame();
            if (stamp < 0) stamp = 0;
            var previousPost = _lastPost7;
            if (start) _starts++;
            if (explain) _lastPost7 = stamp;
            _engineBusy = true;
            if (_build != null) _build.SetBusy(true);
            var file = start ? _style : "update";
            var style = _styleLabel;
            var plan = _planText;
            var match = _match;
            ThreadPool.QueueUserWorkItem(_ =>
            {
                EngineResult result = null;
                var wish = _wish;
                var damage = _damage;
                try { result = SpeShuClient.Recommend(file, style, plan, match, explain, wish, damage); }
                catch (Exception ex) { result = new EngineResult { Error = ErrorText(ex) }; }
                Dispatcher.BeginInvoke(new Action(() => ApplyModel(result, explain, start, previousPost)));
            });
        }

        string Deny(bool start)
        {
            var now = NowGame();
            var before7 = now < 7 * 60;
            if (start)
            {
                if (!before7 && _starts > 0)
                {
                    var msg = "Стиль можно менять только до 7-й минуты, и за это время уходит не больше 3 запросов. Сейчас " + Clock(now) + ", время смены стиля вышло, выбранный билд зафиксирован.";
                    var later = Post7Wait(now);
                    if (later != null) return msg + " " + later;
                    return msg + " Обновление сборки сейчас можно отправить: оно распишет каждый предмет.";
                }
                if (before7 && _starts >= 3)
                {
                    var when = now < 0 ? "Время матча клиент ещё не отдал." : "Сейчас " + Clock(now) + ", до 7:00 осталось " + Clock(7 * 60 - now) + ".";
                    return "До 7-й минуты можно отправить только 3 запроса на выбор стиля: стандарт, урон или защита. Все три уже использованы. " + when + " В 7:00 откроется обновление этой сборки, дальше оно уходит раз в 5 минут и объясняет каждый предмет.";
                }
                return null;
            }

            if (_starts <= 0)
                return "Сначала выберите билд в панели: Стандарт, Урон или Защита. Первый запрос уходит только после этого нажатия.";
            if (before7)
            {
                var when = now < 0
                    ? "Время матча клиент ещё не отдал, поэтому 7:00 ещё не наступила."
                    : "Сейчас " + Clock(now) + ", до 7:00 осталось " + Clock(7 * 60 - now) + ".";
                var left = 3 - _starts;
                return "До 7-й минуты уходят только запросы на выбор стиля, максимум 3. " + when + " Смена стиля ещё доступна: осталось " + left + " из 3. Обновление сборки и объяснение каждого предмета начнутся с первого запроса после 7:00.";
            }
            return Post7Wait(now);
        }

        string Post7Wait(double now)
        {
            if (_lastPost7 < 0) return null;
            var ready = _lastPost7 + 5 * 60;
            if (now >= ready) return null;
            return "После 7-й минуты запросы уходят раз в 5 минут игрового времени. Прошлый ушёл в " + Clock(_lastPost7) + ", следующий откроется в " + Clock(ready) + ". Сейчас " + Clock(now) + ", осталось " + Clock(ready - now) + ".";
        }

        void Notify(string text)
        {
            ThreadPool.QueueUserWorkItem(_ => ToastNote.Show(text, _tray));
        }

        double NowGame()
        {
            if (!_clockInGame) return -1;
            var extra = (DateTime.UtcNow - _clockUtc).TotalSeconds;
            if (extra < 0) extra = 0;
            if (extra > 6) extra = 6;
            return _clockGame + extra;
        }

        static string Clock(double sec)
        {
            if (sec < 0) sec = 0;
            var s = (int)sec;
            return (s / 60) + ":" + (s % 60).ToString("00");
        }

        void TickClock()
        {
            if (!_clockInGame || _build == null) return;
            _build.SetClock(NowGame(), true);
        }

        static string ErrorText(Exception ex)
        {
            if (ex == null) return "";
            if (ex.InnerException == null) return ex.Message;
            return ex.Message + " " + ex.InnerException.Message;
        }

        private void ApplyModel(EngineResult result, bool explain, bool start, double previousAt)
        {
            _engineBusy = false;
            if (_build != null) _build.SetBusy(false);
            if (result == null || !result.Ok)
            {
                if (explain) _lastPost7 = previousAt;
                if (start && _starts > 0) _starts--;
                var err = result != null ? result.Error : "SpeShu не ответил";
                if (_build != null)
                {
                    _build.SetChanges(_current != null ? _current.Changes : new ItemChange[0], err);
                    _build.SetReasons(new string[0]);
                    _build.ShowDetails();
                }
                return;
            }
            var next = result.Build;
            next.Changes = new ItemChange[0];
            _current = next;
            _original = CloneBuild(next);
            _planText = SpeShuClient.PlanText(next);
            _liveLook = result.FromLive || LiveClient.Connected;
            _engineReady = true;
            ApplyVisual();
            if (!explain)
            {
                _build.ShowPlay(result.Play);
                if (NowGame() >= 7 * 60) _build.SetClock(NowGame(), true);
                return;
            }
            var title = (string.IsNullOrEmpty(result.SeedName) ? "сборка" : result.SeedName) + (result.FromLive ? " · live" : "");
            if (!string.IsNullOrEmpty(result.NextName)) title = title + " → " + result.NextName;
            _build.SetChanges(new ItemChange[0], title);
            var reasons = result.Reasons;
            if (reasons == null || reasons.Length == 0)
                reasons = new[] { "Модель не расписала предметы." };
            _build.SetReasons(reasons);
            _build.ShowDetails();
        }

        private void Analyze(bool preferLive)
        {
            if (!string.IsNullOrEmpty(_style))
            {
                AskModel(false);
                return;
            }
            if (_clockInGame && _engineReady)
            {
                Notify("Сначала выберите билд в панели: Стандарт, Урон или Защита. Первый запрос уходит только после этого нажатия.");
                if (_build != null) _build.ShowChoose();
                return;
            }
            if (_engineBusy) return;
            _engineBusy = true;
            if (_build != null) _build.SetBusy(true);
            ThreadPool.QueueUserWorkItem(_ =>
            {
                EngineResult result = null;
                try { result = EngineClient.Recommend(preferLive); }
                catch (Exception ex) { result = new EngineResult { Error = ErrorText(ex) }; }
                Dispatcher.BeginInvoke(new Action(() => ApplyEngine(result)));
            });
        }

        private void ApplyEngine(EngineResult result)
        {
            _engineBusy = false;
            if (_build != null) _build.SetBusy(false);
            if (result == null || !result.Ok)
            {
                if (_liveLook && _current != null)
                {
                    if (_build != null && result != null && !string.IsNullOrEmpty(result.Error))
                        _build.SetChanges(_current.Changes ?? new ItemChange[0], result.Error);
                        _build.SetReasons(new string[0]);
                    return;
                }
                var demo = EngineClient.Recommend(false);
                if (demo != null && demo.Ok)
                {
                    ApplyEngine(demo);
                    return;
                }
                var fallback = SimpleBuild.Adapt(_original, _match);
                var err = result != null ? result.Error : "движок недоступен";
                if (!SimpleBuild.Same(_current, fallback))
                {
                    fallback.Changes = new ItemChange[0];
                    _current = fallback;
                }
                ApplyVisual();
                if (string.IsNullOrEmpty(_style))
                {
                    if (_clockInGame) _build.ShowChoose();
                    else _build.ShowWaiting();
                }
                else
                {
                    _build.SetChanges(new ItemChange[0], err);
                    _build.SetReasons(new string[0]);
                }
                return;
            }

            if (!result.FromLive && _liveLook)
                return;

            var next = result.Build;
            var title = EngineTitle(result);
            if (!_engineReady || (result.FromLive && !_liveLook))
            {
                _original = CloneBuild(next);
                _engineReady = true;
            }

            next.Changes = new ItemChange[0];
            _current = next;
            _liveLook = result.FromLive || LiveClient.Connected;
            ApplyVisual();
            if (string.IsNullOrEmpty(_style))
            {
                if (_build == null) return;
                if (_clockInGame) _build.ShowChoose();
                else _build.ShowWaiting();
                return;
            }
            _build.SetChanges(new ItemChange[0], title);
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
                if (!AppSettings.ShowBuild)
                    SetBuildVisible(false);
                else if (!_concealed)
                    SetBuildVisible(true);
            };
            _settings.Closed += (s, e) => { _settings = null; };
            _settings.Show();
        }

        private void BumpTopmost()
        {
            Raise(_build);
        }

        private static void Raise(Window w)
        {
            if (w == null || !w.IsVisible) return;
            try
            {
                if (!w.Topmost) w.Topmost = true;
                var h = new WindowInteropHelper(w).Handle;
                if (h != IntPtr.Zero)
                    SetWindowPos(h, HwndTopmost, 0, 0, 0, 0, SwpNoMove | SwpNoSize | SwpNoActivate);
            }
            catch { }
        }

        private void FetchAsync()
        {
            if (_fetching) return;
            _fetching = true;
            ThreadPool.QueueUserWorkItem(_ =>
            {
                MatchState match = null;
                try { match = LiveClient.Fetch(); }
                catch { match = null; }
                Dispatcher.BeginInvoke(new Action(() =>
                {
                    _fetching = false;
                    if (match != null && match.InGame)
                    {
                        var wasIn = _clockInGame;
                        var fresh = wasIn && _clockGame > match.GameTime + 45;
                        if (fresh)
                        {
                            _starts = 0;
                            _lastPost7 = -1;
                            _style = null;
                            _styleLabel = null;
                            _planText = "";
                            _connectedOnce = false;
                            if (_build != null) _build.SetActiveStyle("");
                        }
                        _clockGame = match.GameTime;
                        _clockInGame = true;
                        _clockUtc = DateTime.UtcNow;
                        _match = match;
                        if (_build != null)
                        {
                            _build.SetClock(match.GameTime, true);
                            if ((fresh || !wasIn) && string.IsNullOrEmpty(_style))
                                _build.ShowChoose();
                        }
                    }
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
                if (LiveClient.Connected && match != null && match.Me != null && !_connectedOnce)
                {
                    _connectedOnce = true;
                    _tray.ShowBalloonTip(2000, "lol.build", "Игра: " + match.Me.championName, WinForms.ToolTipIcon.Info);
                }
                if (!LiveClient.Connected && !_clockInGame)
                    _connectedOnce = false;
            }
            catch { }
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
            if (_live != null) _live.Stop();
            if (_hwnd != null)
            {
                try { UnregisterHotKey(_hwnd.Handle, HotkeyId); } catch { }
            try { UnregisterHotKey(_hwnd.Handle, HideHotkeyId); } catch { }
            }
            if (_tray != null)
            {
                _tray.Visible = false;
                _tray.Dispose();
            }
            Shutdown();
        }
    }
}
