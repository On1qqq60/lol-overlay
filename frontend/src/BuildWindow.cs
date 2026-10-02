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
    public class BuildWindow : Window
    {
        private readonly StackPanel _row;
        private readonly StackPanel _early;
        private Grid _styles;
        private Button _damageAd;
        private Button _damageAp;
        private Button _wishBtn;
        private TextBox _wishBox;
        private string _activeDamage;
        private readonly StackPanel _changes;
        private readonly Border _chrome;
        private readonly Border _body;
        private readonly Border _details;
        private readonly Button _scan;
        private readonly Button _gear;
        private readonly Button _infoBtn;
        private readonly Button _arrow;
        private readonly RunesPanel _runes;
        private bool _runesOpen;
        private readonly Button _fold;
        private bool _infoOpen;
        private Point _infoDown;
        private ItemChange[] _last;
        private string _lastTitle;
        private string[] _lastReasons = new string[0];
        private bool _waiting;
        private bool _playMode;
        private bool _chooseMode;

        public event Action AnalyzeClicked;
        public event Action SettingsClicked;
        public event Action<string> StyleClicked;
        public event Action<string> DamagePicked;
        public event Action<string> WishSubmitted;

        private readonly Button _styleStandard;
        private readonly Button _styleDamage;
        private readonly Button _styleDefense;
        private string _activeStyle;
        private double _gameTime;
        private bool _inGame;

        public BuildWindow()
        {
            Title = "lol.build";
            Icon = AppIcon.Wpf();
            WindowStyle = WindowStyle.None;
            AllowsTransparency = true;
            Background = Brushes.Transparent;
            ResizeMode = ResizeMode.NoResize;
            Topmost = true;
            ShowInTaskbar = true;
            SizeToContent = SizeToContent.WidthAndHeight;

            _row = new StackPanel { Orientation = Orientation.Horizontal };
            _early = new StackPanel { Orientation = Orientation.Horizontal };
            _scan = Ui.MiniButton("", "Обновить сборку");
            _scan.Content = Ui.RefreshGlyph();
            _scan.Click += (s, e) => { e.Handled = true; if (AnalyzeClicked != null) AnalyzeClicked(); };
            _gear = Ui.MiniButton("⚙", "Настройки");
            _gear.Click += (s, e) => { e.Handled = true; if (SettingsClicked != null) SettingsClicked(); };
            _fold = Ui.MiniButton("▼", "Показать объяснения");
            _fold.Click += (s, e) => { e.Handled = true; ToggleInfo(); };

            _infoBtn = new Button
            {
                Cursor = Cursors.SizeAll,
                Template = Ui.GhostCircleTemplate(),
                Content = Ui.StarGlyph(),
                ToolTip = "Переместить оверлей"
            };
            _infoBtn.Click += (s, e) =>
            {
                e.Handled = true;
            };
            _infoBtn.PreviewMouseLeftButtonDown += (s, e) =>
            {
                _infoDown = e.GetPosition(this);
            };
            _infoBtn.PreviewMouseMove += (s, e) =>
            {
                if (e.LeftButton != MouseButtonState.Pressed) return;
                var now = e.GetPosition(this);
                if (Math.Abs(now.X - _infoDown.X) + Math.Abs(now.Y - _infoDown.Y) <= 4) return;
                DragMove();
            };

            _arrow = new Button
            {
                Cursor = Cursors.Hand,
                Template = Ui.GhostCircleTemplate(),
                Content = UggMark(),
                ToolTip = "Руны",
                VerticalAlignment = VerticalAlignment.Center
            };
            _arrow.Click += (s, e) =>
            {
                e.Handled = true;
                ToggleRunes();
            };

            _infoBtn.VerticalAlignment = VerticalAlignment.Center;
            var earlyRow = new StackPanel
            {
                Orientation = Orientation.Horizontal,
                Margin = new Thickness(0, 0, 0, 2)
            };
            earlyRow.Children.Add(_infoBtn);
            earlyRow.Children.Add(new FrameworkElement { Width = 6 });
            earlyRow.Children.Add(_early);

            var top = new StackPanel { Orientation = Orientation.Horizontal };
            top.Children.Add(_arrow);
            top.Children.Add(new FrameworkElement { Width = 6 });
            top.Children.Add(_row);
            top.Children.Add(new FrameworkElement { Width = 8 });
            top.Children.Add(_scan);
            top.Children.Add(_gear);

            _changes = new StackPanel { HorizontalAlignment = HorizontalAlignment.Stretch };
            _details = new Border
            {
                HorizontalAlignment = HorizontalAlignment.Stretch,
                Background = new SolidColorBrush(Color.FromArgb(214, 10, 12, 16)),
                BorderBrush = new SolidColorBrush(Color.FromArgb(70, 232, 196, 110)),
                BorderThickness = new Thickness(1),
                CornerRadius = new CornerRadius(6),
                Padding = new Thickness(10, 8, 10, 8),
                Margin = new Thickness(0, 6, 0, 0),
                Visibility = Visibility.Collapsed,
                Child = _changes
            };

            _styleStandard = StyleButton("standard", "Стандарт");
            _styleDamage = StyleButton("damage", "Урон");
            _styleDefense = StyleButton("defensive", "Защита");
            _fold.VerticalAlignment = VerticalAlignment.Center;
            _damageAd = SubStyle("ad", "Физический (ад)");
            _damageAp = SubStyle("ap", "Магический (ап)");
            _wishBtn = StyleButton("wish", "Я хочу ...");
            _wishBox = new TextBox
            {
                Margin = new Thickness(0, 4, 0, 0),
                MinHeight = 26,
                FontSize = 13,
                Padding = new Thickness(8, 4, 8, 4),
                Background = new SolidColorBrush(Color.FromArgb(230, 14, 16, 20)),
                Foreground = new SolidColorBrush(Color.FromRgb(236, 232, 226)),
                BorderBrush = new SolidColorBrush(Color.FromRgb(200, 170, 110)),
                CaretBrush = new SolidColorBrush(Color.FromRgb(232, 196, 110)),
                Visibility = Visibility.Collapsed,
                ToolTip = "Напиши пожелание и нажми Enter"
            };
            _wishBox.KeyDown += (s, e) =>
            {
                if (e.Key != Key.Enter) return;
                e.Handled = true;
                SubmitWish();
            };
            _styles = new Grid();
            _styles.ColumnDefinitions.Add(new ColumnDefinition { Width = GridLength.Auto });
            _styles.ColumnDefinitions.Add(new ColumnDefinition { Width = GridLength.Auto });
            _styles.ColumnDefinitions.Add(new ColumnDefinition { Width = GridLength.Auto });
            _styles.ColumnDefinitions.Add(new ColumnDefinition { Width = GridLength.Auto });
            _styles.ColumnDefinitions.Add(new ColumnDefinition { Width = GridLength.Auto });
            _styles.RowDefinitions.Add(new RowDefinition { Height = GridLength.Auto });
            _styles.RowDefinitions.Add(new RowDefinition { Height = GridLength.Auto });
            _styles.RowDefinitions.Add(new RowDefinition { Height = GridLength.Auto });
            Put(_fold, 0, 0);
            Put(_styleStandard, 1, 0);
            Put(_styleDamage, 2, 0);
            Put(_styleDefense, 3, 0);
            Put(_wishBtn, 4, 0);
            _styles.Children.Add(_fold);
            _styles.Children.Add(_styleStandard);
            _styles.Children.Add(_styleDamage);
            _styles.Children.Add(_styleDefense);
            _styles.Children.Add(_wishBtn);
            _damageAd.Visibility = Visibility.Collapsed;
            _damageAp.Visibility = Visibility.Collapsed;
            Put(_damageAd, 2, 1);
            Put(_damageAp, 3, 1);
            _styles.Children.Add(_damageAd);
            _styles.Children.Add(_damageAp);
            Grid.SetColumn(_wishBox, 0);
            Grid.SetColumnSpan(_wishBox, 5);
            Grid.SetRow(_wishBox, 2);
            _styles.Children.Add(_wishBox);

            var col = new StackPanel { VerticalAlignment = VerticalAlignment.Top };
            col.Children.Add(earlyRow);
            col.Children.Add(top);
            col.Children.Add(_styles);
            col.Children.Add(_details);

            _runes = new RunesPanel { Visibility = Visibility.Collapsed };
            var shell = new Grid { VerticalAlignment = VerticalAlignment.Top };
            shell.ColumnDefinitions.Add(new ColumnDefinition { Width = GridLength.Auto });
            shell.ColumnDefinitions.Add(new ColumnDefinition { Width = GridLength.Auto });
            Grid.SetColumn(_runes, 0);
            Grid.SetColumn(col, 1);
            shell.Children.Add(_runes);
            shell.Children.Add(col);

            _body = new Border
            {
                Background = Brushes.Transparent,
                Padding = new Thickness(0),
                Child = shell,
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

            ShowWaiting();
            SyncFold();
            PaintStyles();
        }

        static void Put(UIElement el, int col, int row)
        {
            Grid.SetColumn(el, col);
            Grid.SetRow(el, row);
        }

        static Image UggMark()
        {
            var img = new Image
            {
                Stretch = Stretch.Uniform,
                HorizontalAlignment = HorizontalAlignment.Stretch,
                VerticalAlignment = VerticalAlignment.Stretch,
                Margin = new Thickness(1)
            };
            if (File.Exists(AppPaths.UggLogo))
            {
                var bmp = new BitmapImage();
                bmp.BeginInit();
                bmp.CacheOption = BitmapCacheOption.OnLoad;
                bmp.UriSource = new Uri(AppPaths.UggLogo, UriKind.Absolute);
                bmp.EndInit();
                bmp.Freeze();
                img.Source = bmp;
            }
            return img;
        }

        Button StyleButton(string id, string label)
        {
            var button = new Button
            {
                Cursor = Cursors.Hand,
                Template = Ui.GhostButtonTemplate(),
                Margin = new Thickness(0, 0, 10, 0),
                VerticalAlignment = VerticalAlignment.Center,
                ToolTip = id == "wish" ? "Своё пожелание к сборке" : "Стартовый промпт: " + label,
                Tag = id
            };
            button.Click += (s, e) =>
            {
                e.Handled = true;
                if (id == "damage") { ToggleDamage(); return; }
                if (id == "wish") { ToggleWish(); return; }
                HideDamageChoices();
                if (StyleClicked != null) StyleClicked(id);
            };
            return button;
        }

        Button SubStyle(string id, string label)
        {
            var button = new Button
            {
                Cursor = Cursors.Hand,
                Template = Ui.GhostButtonTemplate(),
                Margin = new Thickness(0, 2, 8, 0),
                VerticalAlignment = VerticalAlignment.Center,
                HorizontalAlignment = HorizontalAlignment.Left,
                ToolTip = label,
                Tag = id
            };
            button.Click += (s, e) =>
            {
                e.Handled = true;
                _activeDamage = id;
                PaintDamage();
                if (DamagePicked != null) DamagePicked(id);
            };
            return button;
        }

        void ShowDamage(bool open)
        {
            var vis = open ? Visibility.Visible : Visibility.Collapsed;
            if (_damageAd != null) _damageAd.Visibility = vis;
            if (_damageAp != null) _damageAp.Visibility = vis;
        }

        void ToggleDamage()
        {
            var open = _damageAd == null || _damageAd.Visibility != Visibility.Visible;
            ShowDamage(open);
            PaintDamage();
        }

        public void HideDamageChoices()
        {
            ShowDamage(false);
        }

        public void SetActiveDamage(string id)
        {
            _activeDamage = id ?? "";
            if (_activeDamage.Length > 0) ShowDamage(true);
            PaintDamage();
        }

        void PaintDamage()
        {
            PaintSub(_damageAd, "Физический (ад)");
            PaintSub(_damageAp, "Магический (ап)");
        }

        void PaintSub(Button button, string label)
        {
            if (button == null) return;
            var on = (string)button.Tag == _activeDamage;
            var text = Ui.GlowText(label, 11, on
                ? Color.FromRgb(255, 244, 210)
                : Color.FromRgb(210, 170, 90));
            text.TextDecorations = TextDecorations.Underline;
            button.Content = text;
        }

        void ToggleWish()
        {
            var open = _wishBox.Visibility != Visibility.Visible;
            _wishBox.Visibility = open ? Visibility.Visible : Visibility.Collapsed;
            if (open)
            {
                _wishBox.Focus();
                return;
            }
            SubmitWish();
        }

        void SubmitWish()
        {
            var text = _wishBox.Text == null ? "" : _wishBox.Text.Trim();
            if (text.Length == 0 || WishSubmitted == null) return;
            WishSubmitted(text);
        }

        public void SetActiveStyle(string id)
        {
            _activeStyle = id;
            PaintStyles();
        }

        void PaintStyles()
        {
            PaintStyle(_styleStandard, "Стандарт");
            PaintStyle(_styleDamage, "Урон");
            PaintStyle(_styleDefense, "Защита");
            if (_wishBtn != null)
                _wishBtn.Content = Ui.GlowText("Я хочу ...", 12, Color.FromRgb(232, 196, 110));
            PaintDamage();
        }

        void PaintStyle(Button button, string label)
        {
            if (button == null) return;
            var on = (string)button.Tag == _activeStyle;
            button.Content = Ui.GlowText(label, 12, on
                ? Color.FromRgb(255, 244, 210)
                : Color.FromRgb(140, 132, 118));
        }

        public void ShowDetails()
        {
            _infoOpen = true;
            _details.Visibility = Visibility.Visible;
            SyncFold();
        }

        public void SetBusy(bool busy) { Opacity = busy ? 0.5 : AppSettings.Opacity; }

        public void SetClock(double gameTime, bool inGame)
        {
            _gameTime = gameTime;
            _inGame = inGame;
            ApplyEarlyVisibility();
            if (_playMode && HideEarly()) HoldForItems();
        }

        public void SetLiveChampion(string name, string position)
        {
            if (_runes != null) _runes.SetLive(name, position);
        }

        void ToggleRunes()
        {
            double right = 0;
            var placed = !double.IsNaN(Left) && !double.IsNaN(ActualWidth) && ActualWidth > 0;
            if (placed) right = Left + ActualWidth;
            var top = Top;
            _runesOpen = !_runesOpen;
            _runes.Visibility = _runesOpen ? Visibility.Visible : Visibility.Collapsed;
            _arrow.ToolTip = _runesOpen ? "Скрыть руны" : "Руны";
            UpdateLayout();
            if (placed && ActualWidth > 0) Left = right - ActualWidth;
            if (!double.IsNaN(top)) Top = top;
        }

        void AlignToBuild()
        {
            var left = _arrow.Width + 6;
            if (left < 6) left = AppSettings.IconSize + 6;
            _styles.Margin = new Thickness(left + 3, 2, 0, 0);
            _early.Margin = new Thickness(0);
        }

        bool HideEarly()
        {
            return _inGame && _gameTime >= 7 * 60;
        }

        void ApplyEarlyVisibility()
        {
            var vis = _early.Children.Count == 0 || HideEarly()
                ? Visibility.Collapsed
                : Visibility.Visible;
            if (_early.Visibility != vis) _early.Visibility = vis;
        }

        public void SetBuild(RecommendedBuild build, bool live)
        {
            var size = AppSettings.IconSize;
            var star = size;
            var ctrl = Math.Max(12, size - 10);
            _infoBtn.Width = _infoBtn.Height = star;
            _arrow.Width = _arrow.Height = star;
            _scan.Width = _scan.Height = ctrl;
            _gear.Width = _gear.Height = ctrl;
            _fold.Width = _fold.Height = ctrl;
            _scan.Margin = new Thickness(0, 0, 4, 0);
            _fold.Margin = new Thickness(0, 0, 4, 0);
            _gear.Margin = new Thickness(0, 0, 0, 0);
            AlignToBuild();
            if (_chooseMode) { }
            else if (_waiting) ShowWaiting();
            else if (_last != null) PaintDetails();

            _chrome.BorderBrush = Brushes.Transparent;

            _row.Children.Clear();
            for (var i = 0; i < 6; i++)
            {
                var id = (build.Items != null && i < build.Items.Length) ? build.Items[i] : "";
                _row.Children.Add(Ui.ItemIcon(id, size, build.NextIndex == i));
            }
            var boot = Ui.ItemIcon(build.Boots, size, build.NextIndex == 6);
            boot.Margin = new Thickness(10, 2, 3, 2);
            boot.BorderBrush = build.NextIndex == 6
                ? new SolidColorBrush(Color.FromRgb(200, 170, 110))
                : new SolidColorBrush(Color.FromArgb(140, 10, 200, 185));
            _row.Children.Add(boot);

            _early.Children.Clear();
            if (build.Early != null)
            {
                for (var i = 0; i < build.Early.Length; i++)
                {
                    if (string.IsNullOrEmpty(build.Early[i])) continue;
                    _early.Children.Add(Ui.ItemIcon(build.Early[i], size, build.EarlyNext == i));
                }
            }
            ApplyEarlyVisibility();
        }

        public void ShowChoose()
        {
            if (_chooseMode && _infoOpen) return;
            _chooseMode = true;
            _waiting = false;
            _playMode = false;
            _changes.Children.Clear();
            _changes.Children.Add(DetailLine("Выберите билд", 15, Color.FromRgb(232, 196, 110)));
            var row = new StackPanel
            {
                Orientation = Orientation.Horizontal,
                HorizontalAlignment = HorizontalAlignment.Center,
                Margin = new Thickness(0, 4, 0, 0)
            };
            row.Children.Add(ChoiceLink("standard", "Стандарт"));
            row.Children.Add(ChoiceLink("damage", "Урон"));
            row.Children.Add(ChoiceLink("defensive", "Защита"));
            _changes.Children.Add(row);
            ShowDetails();
        }

        Button ChoiceLink(string id, string label)
        {
            var text = Ui.GlowText(label, 16, Color.FromRgb(255, 244, 214));
            text.TextDecorations = TextDecorations.Underline;
            var button = new Button
            {
                Content = text,
                Cursor = Cursors.Hand,
                Template = Ui.GhostButtonTemplate(),
                Margin = new Thickness(12, 0, 12, 0),
                ToolTip = "Выбрать билд: " + label,
                Background = Brushes.Transparent,
                BorderThickness = new Thickness(0),
                Focusable = false
            };
            button.Click += (s, e) =>
            {
                e.Handled = true;
                if (StyleClicked != null) StyleClicked(id);
            };
            return button;
        }

        public void ShowPlay(string[] lines)
        {
            _chooseMode = false;
            _waiting = false;
            _playMode = true;
            _last = new ItemChange[0];
            _lastTitle = "Рекомендации по игре";
            _lastReasons = lines != null && lines.Length > 0
                ? lines
                : new[] { "Модель не прислала рекомендации по игре." };
            PaintDetails();
            ShowDetails();
        }

        void HoldForItems()
        {
            _chooseMode = false;
            _playMode = false;
            _waiting = false;
            _last = new ItemChange[0];
            _lastTitle = "Объяснения предметов";
            _lastReasons = new[] { "7-я минута прошла. Следующий запрос распишет каждый из шести предметов: зачем он и против кого." };
            PaintDetails();
            ShowDetails();
        }

        public void ShowWaiting()
        {
            _chooseMode = false;
            _playMode = false;
            _waiting = true;
            _last = new ItemChange[0];
            _lastTitle = null;
            _lastReasons = new string[0];
            _changes.Children.Clear();
            _changes.Children.Add(DetailLine("Здесь будут рекомендации по игре", 15, Color.FromRgb(232, 196, 110)));
            _infoOpen = true;
            _details.Visibility = Visibility.Visible;
            SyncFold();
        }

        public void SetChanges(ItemChange[] changes, string title)
        {
            _chooseMode = false;
            _playMode = false;
            _waiting = false;
            _last = changes;
            _lastTitle = title;
            PaintDetails();
        }

        public void SetReasons(string[] reasons)
        {
            if (_chooseMode) return;
            _playMode = false;
            _lastReasons = reasons ?? new string[0];
            if (_waiting) return;
            PaintDetails();
        }

        void PaintDetails()
        {
            _changes.Children.Clear();

            if (!string.IsNullOrEmpty(_lastTitle))
                _changes.Children.Add(DetailLine(_lastTitle, 15, Color.FromRgb(232, 196, 110)));

            var changes = _last;
            if (changes != null)
            {
                var size = AppSettings.IconSize;
                for (var i = 0; i < changes.Length; i++)
                {
                    var c = changes[i];
                    var block = new StackPanel
                    {
                        Margin = new Thickness(0, 0, 0, 8),
                        HorizontalAlignment = HorizontalAlignment.Stretch
                    };
                    var row = new StackPanel
                    {
                        Orientation = Orientation.Horizontal,
                        HorizontalAlignment = HorizontalAlignment.Center
                    };
                    row.Children.Add(Ui.ItemIcon(c.from, size, false));
                    row.Children.Add(Ui.GlowText("  →  ", 15, Color.FromRgb(232, 166, 69)));
                    row.Children.Add(Ui.ItemIcon(c.to, size, true));
                    block.Children.Add(row);
                    block.Children.Add(DetailLine(
                        ItemNames.Get(c.from) + "  →  " + ItemNames.Get(c.to),
                        14, Color.FromRgb(255, 236, 200)));
                    block.Children.Add(DetailLine(c.why, 14, Color.FromRgb(236, 232, 226)));
                    _changes.Children.Add(block);
                }
            }

            if (_lastReasons == null) return;
            var n = 0;
            for (var i = 0; i < _lastReasons.Length && n < 8; i++)
            {
                var line = _lastReasons[i];
                if (string.IsNullOrEmpty(line)) continue;
                _changes.Children.Add(DetailLine("• " + line, 13, Color.FromRgb(226, 222, 214)));
                n++;
            }
        }

        static TextBlock DetailLine(string text, double size, Color color)
        {
            var line = Ui.GlowText(text, size, color);
            line.TextAlignment = TextAlignment.Center;
            line.HorizontalAlignment = HorizontalAlignment.Stretch;
            return line;
        }

        private void ToggleInfo()
        {
            _infoOpen = !_infoOpen;
            _details.Visibility = _infoOpen ? Visibility.Visible : Visibility.Collapsed;
            SyncFold();
        }

        public void HideDetails()
        {
            _infoOpen = false;
            _details.Visibility = Visibility.Collapsed;
            SyncFold();
        }

        private void SyncFold()
        {
            if (_fold == null) return;
            _fold.Content = Ui.ChevronGlyph(!_infoOpen);
            _fold.ToolTip = _infoOpen ? "Скрыть объяснения" : "Показать объяснения";
        }
    }
}
