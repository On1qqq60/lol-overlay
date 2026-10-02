using System;
using System.Collections.Generic;
using System.ComponentModel;
using System.Globalization;
using System.Threading;
using System.Windows;
using System.Windows.Controls;
using System.Windows.Data;
using System.Windows.Controls.Primitives;
using System.Windows.Input;
using System.Windows.Media;
using System.Windows.Media.Animation;
using System.Windows.Threading;

namespace LolBuildOverlay
{
    public class RunesPanel : Border
    {
        readonly TextBox _search;
        readonly TextBlock _picked;
        readonly Image _pickedIcon;
        ChampRow _faceChamp;
        readonly ListBox _list;
        readonly Popup _popup;
        readonly Border _face;
        readonly Border _drop;
        readonly TextBlock _status;
        readonly StackPanel _body;
        readonly List<ChampRow> _champs;
        string _liveName = "";
        string _livePos = "";
        string _loadedId = "";
        bool _userPicked;
        bool _filling;
        int _gen;
        int _pageToken;
        bool _pageBusy;
        RunePageView _shown;
        DateTime _popupClosed;

        public RunesPanel()
        {
            Width = 420;
            ClipToBounds = true;
            VerticalAlignment = VerticalAlignment.Top;
            Background = new SolidColorBrush(Color.FromArgb(230, 12, 14, 18));
            BorderBrush = new SolidColorBrush(Color.FromArgb(90, 232, 196, 110));
            BorderThickness = new Thickness(1);
            CornerRadius = new CornerRadius(8);
            Padding = new Thickness(10, 8, 10, 10);
            Margin = new Thickness(0, 0, 10, 0);

            _champs = ChampionBook.All();
            ChampionIcons.Preload(_champs);
            var ink = new SolidColorBrush(Color.FromRgb(236, 232, 226));
            var gold = new SolidColorBrush(Color.FromRgb(200, 170, 110));
            var field = new SolidColorBrush(Color.FromArgb(230, 14, 16, 20));
            var panel = new SolidColorBrush(Color.FromArgb(245, 12, 14, 18));

            _pickedIcon = new Image
            {
                Width = 24,
                Height = 24,
                Margin = new Thickness(0, 0, 8, 0),
                Stretch = Stretch.UniformToFill,
                VerticalAlignment = VerticalAlignment.Center,
                Visibility = Visibility.Collapsed
            };
            _picked = new TextBlock
            {
                Text = "Чемпион",
                FontSize = 15,
                VerticalAlignment = VerticalAlignment.Center,
                Foreground = new SolidColorBrush(Color.FromRgb(170, 166, 156))
            };
            var chevron = new TextBlock
            {
                Text = "▾",
                FontSize = 12,
                Margin = new Thickness(8, 0, 0, 0),
                VerticalAlignment = VerticalAlignment.Center,
                Foreground = new SolidColorBrush(Color.FromRgb(232, 196, 110))
            };
            var faceRow = new DockPanel();
            DockPanel.SetDock(chevron, Dock.Right);
            DockPanel.SetDock(_pickedIcon, Dock.Left);
            faceRow.Children.Add(chevron);
            faceRow.Children.Add(_pickedIcon);
            faceRow.Children.Add(_picked);
            _face = new Border
            {
                Background = field,
                BorderBrush = gold,
                BorderThickness = new Thickness(1),
                CornerRadius = new CornerRadius(4),
                Padding = new Thickness(8, 5, 8, 5),
                Cursor = Cursors.Hand,
                Child = faceRow
            };
            _popupClosed = DateTime.MinValue;
            _face.MouseLeftButtonUp += (s, e) =>
            {
                e.Handled = true;
                if ((DateTime.UtcNow - _popupClosed).TotalMilliseconds < 250) return;
                if (!_popup.IsOpen) OpenDrop();
            };

            _search = new TextBox
            {
                MinHeight = 34,
                FontSize = 16,
                Margin = new Thickness(0, 0, 0, 6),
                Padding = new Thickness(8, 4, 8, 4),
                Background = field,
                Foreground = ink,
                BorderBrush = gold,
                CaretBrush = new SolidColorBrush(Color.FromRgb(232, 196, 110)),
                ToolTip = "Поиск чемпиона"
            };
            _search.TextChanged += (s, e) =>
            {
                if (!_filling) Refill();
            };
            _search.KeyDown += (s, e) =>
            {
                if (e.Key == Key.Escape)
                {
                    e.Handled = true;
                    _popup.IsOpen = false;
                    return;
                }
                if (e.Key == Key.Down)
                {
                    e.Handled = true;
                    if (_list.Items.Count == 0) return;
                    _list.SelectedIndex = 0;
                    var first = _list.ItemContainerGenerator.ContainerFromIndex(0) as ListBoxItem;
                    if (first != null) first.Focus();
                    return;
                }
                if (e.Key != Key.Enter) return;
                e.Handled = true;
                if (_list.Items.Count == 0) return;
                var row = _list.Items[0] as ChampRow;
                if (row != null) Select(row, true);
            };

            _list = new ListBox
            {
                Background = panel,
                Foreground = ink,
                BorderThickness = new Thickness(0),
                MaxHeight = 280,
                FontSize = 15,
                ItemContainerStyle = ChampItemStyle(),
                ItemTemplate = ChampTemplate(),
                ItemsPanel = new ItemsPanelTemplate
                {
                    VisualTree = new FrameworkElementFactory(typeof(VirtualizingStackPanel))
                }
            };
            ScrollViewer.SetCanContentScroll(_list, true);
            VirtualizingPanel.SetIsVirtualizing(_list, true);
            VirtualizingPanel.SetVirtualizationMode(_list, VirtualizationMode.Recycling);
            _list.Template = DarkListTemplate(panel);
            _list.PreviewMouseLeftButtonUp += (s, e) =>
            {
                var item = ItemsControl.ContainerFromElement(_list, e.OriginalSource as DependencyObject) as ListBoxItem;
                if (item == null) return;
                var row = item.Content as ChampRow;
                if (row != null) Select(row, true);
            };
            _list.KeyDown += (s, e) =>
            {
                if (e.Key == Key.Escape)
                {
                    e.Handled = true;
                    _popup.IsOpen = false;
                    return;
                }
                if (e.Key != Key.Enter) return;
                e.Handled = true;
                var row = _list.SelectedItem as ChampRow;
                if (row != null) Select(row, true);
            };

            var dropCol = new StackPanel();
            dropCol.Children.Add(_search);
            dropCol.Children.Add(_list);
            _drop = new Border
            {
                Background = panel,
                BorderBrush = gold,
                BorderThickness = new Thickness(1),
                CornerRadius = new CornerRadius(4),
                Padding = new Thickness(6),
                Margin = new Thickness(0, 4, 0, 0),
                Child = dropCol
            };
            _popup = new Popup
            {
                PlacementTarget = _face,
                Placement = PlacementMode.Bottom,
                StaysOpen = false,
                AllowsTransparency = true,
                Child = _drop
            };
            _popup.Closed += (s, e) => { _popupClosed = DateTime.UtcNow; };

            _status = new TextBlock
            {
                FontSize = 12,
                Foreground = new SolidColorBrush(Color.FromRgb(170, 166, 156)),
                Margin = new Thickness(0, 8, 0, 0),
                TextWrapping = TextWrapping.Wrap,
                Text = _champs.Count == 0 ? "Список чемпионов не найден" : "Выбери чемпиона"
            };

            _body = new StackPanel { Margin = new Thickness(0, 8, 0, 0) };
            var scroll = new ScrollViewer
            {
                MaxHeight = 520,
                VerticalScrollBarVisibility = ScrollBarVisibility.Auto,
                HorizontalScrollBarVisibility = ScrollBarVisibility.Disabled,
                Content = _body
            };

            var title = Ui.GlowText("Руны", 15, Color.FromRgb(232, 196, 110));
            title.Margin = new Thickness(0, 0, 0, 8);
            var col = new StackPanel();
            col.Children.Add(title);
            col.Children.Add(_face);
            col.Children.Add(_popup);
            col.Children.Add(_status);
            col.Children.Add(scroll);
            Child = col;
        }

        void OpenDrop()
        {
            _filling = true;
            _search.Text = "";
            _filling = false;
            Refill();
            _drop.Width = _face.ActualWidth > 1 ? _face.ActualWidth : 380;
            _popup.IsOpen = true;
            Dispatcher.BeginInvoke(DispatcherPriority.Input, new Action(delegate
            {
                Keyboard.Focus(_search);
            }));
        }

        static Style ChampItemStyle()
        {
            var style = new Style(typeof(ListBoxItem));
            var template = new ControlTemplate(typeof(ListBoxItem));
            var border = new FrameworkElementFactory(typeof(Border));
            border.SetValue(Border.BackgroundProperty, new TemplateBindingExtension(Control.BackgroundProperty));
            border.SetValue(Border.PaddingProperty, new TemplateBindingExtension(Control.PaddingProperty));
            var presenter = new FrameworkElementFactory(typeof(ContentPresenter));
            presenter.SetValue(ContentPresenter.ContentProperty, new TemplateBindingExtension(ContentControl.ContentProperty));
            presenter.SetValue(ContentPresenter.ContentTemplateProperty, new TemplateBindingExtension(ContentControl.ContentTemplateProperty));
            presenter.SetValue(ContentPresenter.VerticalAlignmentProperty, VerticalAlignment.Center);
            border.AppendChild(presenter);
            template.VisualTree = border;
            style.Setters.Add(new Setter(Control.TemplateProperty, template));
            style.Setters.Add(new Setter(Control.BackgroundProperty, Brushes.Transparent));
            style.Setters.Add(new Setter(Control.ForegroundProperty, new SolidColorBrush(Color.FromRgb(236, 232, 226))));
            style.Setters.Add(new Setter(Control.PaddingProperty, new Thickness(8, 4, 8, 4)));
            style.Setters.Add(new Setter(Control.CursorProperty, Cursors.Hand));
            style.Setters.Add(new Setter(Control.HorizontalContentAlignmentProperty, HorizontalAlignment.Left));
            var over = new Trigger { Property = UIElement.IsMouseOverProperty, Value = true };
            over.Setters.Add(new Setter(Control.BackgroundProperty, new SolidColorBrush(Color.FromArgb(50, 232, 196, 110))));
            style.Triggers.Add(over);
            var on = new Trigger { Property = ListBoxItem.IsSelectedProperty, Value = true };
            on.Setters.Add(new Setter(Control.BackgroundProperty, new SolidColorBrush(Color.FromArgb(70, 232, 196, 110))));
            style.Triggers.Add(on);
            return style;
        }

        static DataTemplate ChampTemplate()
        {
            var row = new FrameworkElementFactory(typeof(StackPanel));
            row.SetValue(StackPanel.OrientationProperty, Orientation.Horizontal);
            var slot = new FrameworkElementFactory(typeof(Border));
            slot.SetValue(FrameworkElement.WidthProperty, 28.0);
            slot.SetValue(FrameworkElement.HeightProperty, 28.0);
            slot.SetValue(FrameworkElement.MarginProperty, new Thickness(0, 0, 10, 0));
            var icon = new FrameworkElementFactory(typeof(Image));
            icon.SetValue(FrameworkElement.WidthProperty, 28.0);
            icon.SetValue(FrameworkElement.HeightProperty, 28.0);
            icon.SetValue(Image.StretchProperty, Stretch.UniformToFill);
            icon.SetBinding(Image.SourceProperty, new Binding("Icon"));
            slot.AppendChild(icon);
            var name = new FrameworkElementFactory(typeof(TextBlock));
            name.SetValue(TextBlock.FontSizeProperty, 15.0);
            name.SetValue(FrameworkElement.VerticalAlignmentProperty, VerticalAlignment.Center);
            name.SetValue(TextBlock.ForegroundProperty, new SolidColorBrush(Color.FromRgb(236, 232, 226)));
            name.SetBinding(TextBlock.TextProperty, new Binding("Name"));
            row.AppendChild(slot);
            row.AppendChild(name);
            return new DataTemplate { VisualTree = row };
        }

        void ShowFace(ChampRow row)
        {
            if (_faceChamp != null) _faceChamp.PropertyChanged -= FaceIconChanged;
            _faceChamp = row;
            if (row == null) return;
            row.PropertyChanged += FaceIconChanged;
            _pickedIcon.Source = row.Icon;
            _pickedIcon.Visibility = Visibility.Visible;
            ChampionIcons.Request(row);
        }

        void FaceIconChanged(object sender, PropertyChangedEventArgs e)
        {
            if (e.PropertyName != "Icon" || _faceChamp == null) return;
            _pickedIcon.Source = _faceChamp.Icon;
            _pickedIcon.Visibility = Visibility.Visible;
        }

        static ControlTemplate DarkListTemplate(Brush background)
        {
            var template = new ControlTemplate(typeof(ListBox));
            var border = new FrameworkElementFactory(typeof(Border));
            border.SetValue(Border.BackgroundProperty, background);
            var scroll = new FrameworkElementFactory(typeof(ScrollViewer));
            scroll.SetValue(ScrollViewer.CanContentScrollProperty, true);
            scroll.SetValue(ScrollViewer.HorizontalScrollBarVisibilityProperty, ScrollBarVisibility.Disabled);
            scroll.SetValue(ScrollViewer.VerticalScrollBarVisibilityProperty, ScrollBarVisibility.Auto);
            scroll.AppendChild(new FrameworkElementFactory(typeof(ItemsPresenter)));
            border.AppendChild(scroll);
            template.VisualTree = border;
            return template;
        }

        public void SetLive(string name, string position)
        {
            var pos = OpggRunes.PositionOf(position);
            var nameChanged = !Same(name, _liveName);
            var posArrived = pos.Length > 0 && pos != _livePos;
            _liveName = name ?? "";
            if (pos.Length > 0) _livePos = pos;
            if (nameChanged) _userPicked = false;
            if (_userPicked) return;
            var row = Find(_liveName);
            if (row == null) return;
            if (Same(row.Id, _loadedId) && !nameChanged && !posArrived) return;
            Select(row, false);
        }

        void Refill()
        {
            var q = (_search.Text ?? "").Trim();
            _filling = true;
            _list.Items.Clear();
            foreach (var champ in _champs)
            {
                if (q.Length > 0
                    && champ.Name.IndexOf(q, StringComparison.OrdinalIgnoreCase) < 0
                    && champ.Id.IndexOf(q, StringComparison.OrdinalIgnoreCase) < 0)
                    continue;
                _list.Items.Add(champ);
            }
            _filling = false;
        }

        void Select(ChampRow row, bool user)
        {
            if (row == null) return;
            if (user) _userPicked = true;
            _picked.Text = row.Name;
            _picked.Foreground = new SolidColorBrush(Color.FromRgb(236, 232, 226));
            ShowFace(row);
            _popup.IsOpen = false;
            Load(row);
        }

        ChampRow Find(string name)
        {
            if (string.IsNullOrEmpty(name)) return null;
            foreach (var champ in _champs)
            {
                if (Same(champ.Name, name)) return champ;
            }
            var key = Norm(name);
            foreach (var champ in _champs)
            {
                if (Norm(champ.Name) == key || Norm(champ.Id) == key) return champ;
            }
            return null;
        }

        void Load(ChampRow row)
        {
            if (row == null) return;
            var gen = ++_gen;
            _pageToken++;
            _pageBusy = false;
            _shown = null;
            _loadedId = row.Id;
            _status.Text = "Загрузка " + row.Name + "…";
            _body.Children.Clear();
            var slug = row.Id.ToLowerInvariant();
            var prefer = _livePos;
            ThreadPool.QueueUserWorkItem(delegate
            {
                List<RunePageView> pages = null;
                string error = null;
                try { pages = OpggRunes.Load(slug, prefer); }
                catch (Exception ex) { error = ex.Message; }
                Dispatcher.BeginInvoke(new Action(delegate
                {
                    if (gen != _gen) return;
                    Paint(row, pages, error);
                }));
            });
        }

        void Paint(ChampRow row, List<RunePageView> pages, string error)
        {
            _body.Children.Clear();
            if (!string.IsNullOrEmpty(error) || pages == null || pages.Count == 0)
            {
                _status.Text = "Нет данных по рунам для " + row.Name;
                return;
            }
            _status.Text = row.Name;
            var chips = new Grid();
            for (var i = 0; i < pages.Count; i++)
            {
                chips.ColumnDefinitions.Add(new ColumnDefinition { Width = new GridLength(1, GridUnitType.Star) });
                var chip = PageChip(pages, i);
                chip.HorizontalAlignment = HorizontalAlignment.Stretch;
                chip.Margin = new Thickness(0, 0, i < pages.Count - 1 ? 6 : 0, 0);
                Grid.SetColumn(chip, i);
                chips.Children.Add(chip);
            }
            var head = new StackPanel { Margin = new Thickness(0, 0, 0, 8) };
            head.Children.Add(chips);
            _body.Children.Add(head);
            ShowPage(pages[0], false);
        }

        Border PageChip(List<RunePageView> pages, int index)
        {
            var page = pages[index];
            var selected = index == 0;
            var title = Ui.GlowText(RuneBook.NameOf(page.PrimaryStyle), 15, Color.FromRgb(236, 232, 226));
            title.TextTrimming = TextTrimming.CharacterEllipsis;
            title.MaxWidth = 128;
            var text = new StackPanel { VerticalAlignment = VerticalAlignment.Center, MaxWidth = 128 };
            text.Children.Add(title);
            text.Children.Add(new TextBlock
            {
                FontSize = 14,
                Margin = new Thickness(0, 2, 0, 0),
                Foreground = new SolidColorBrush(Color.FromRgb(232, 196, 110)),
                Text = Pct(page.PickRate)
            });
            text.Children.Add(new TextBlock
            {
                FontSize = 14,
                Foreground = new SolidColorBrush(Color.FromRgb(150, 196, 160)),
                Text = Win(page.Win, page.Play)
            });
            var keyId = page.Primary.Count > 0 ? page.Primary[0].Id : 0;
            var mark = new Image
            {
                Width = 36,
                Height = 36,
                Margin = new Thickness(0, 0, 8, 0),
                VerticalAlignment = VerticalAlignment.Center,
                Stretch = Stretch.Uniform,
                Source = keyId > 0 ? RuneIcons.Get(keyId) : null,
                ToolTip = keyId > 0 ? RuneBook.NameOf(keyId) : null
            };
            var box = new Grid { Margin = new Thickness(10, 8, 10, 8) };
            box.ColumnDefinitions.Add(new ColumnDefinition { Width = GridLength.Auto });
            box.ColumnDefinitions.Add(new ColumnDefinition { Width = new GridLength(1, GridUnitType.Star) });
            Grid.SetColumn(mark, 0);
            Grid.SetColumn(text, 1);
            box.Children.Add(mark);
            box.Children.Add(text);
            var button = new Button
            {
                Cursor = Cursors.Hand,
                Template = Ui.GhostButtonTemplate(),
                Content = box,
                HorizontalContentAlignment = HorizontalAlignment.Stretch,
                Background = Brushes.Transparent
            };
            var frame = new Border
            {
                CornerRadius = new CornerRadius(6),
                HorizontalAlignment = HorizontalAlignment.Stretch,
                Background = ChipFill(selected),
                BorderBrush = ChipEdge(selected),
                BorderThickness = new Thickness(1),
                Child = button
            };
            button.Click += (s, e) =>
            {
                e.Handled = true;
                var head = _body.Children.Count > 0 ? _body.Children[0] as StackPanel : null;
                var parent = head != null && head.Children.Count > 0 ? head.Children[0] as Grid : null;
                if (parent != null)
                {
                    foreach (var child in parent.Children)
                    {
                        var border = child as Border;
                        if (border == null) continue;
                        var on = border == frame;
                        border.Background = ChipFill(on);
                        border.BorderBrush = ChipEdge(on);
                    }
                }
                ShowPage(page, true);
            };
            return frame;
        }

        static SolidColorBrush ChipFill(bool on)
        {
            return new SolidColorBrush(on
                ? Color.FromArgb(50, 232, 196, 110)
                : Color.FromArgb(30, 255, 255, 255));
        }

        static SolidColorBrush ChipEdge(bool on)
        {
            return new SolidColorBrush(on
                ? Color.FromRgb(232, 196, 110)
                : Color.FromArgb(40, 255, 255, 255));
        }

        void ShowPage(RunePageView page, bool animate)
        {
            if (page == null) return;
            if (animate && SamePage(_shown, page))
            {
                if (!_pageBusy) return;
                _pageToken++;
                _pageBusy = false;
                while (_body.Children.Count > 1) _body.Children.RemoveAt(1);
                PaintPage(_shown);
                return;
            }
            var token = ++_pageToken;
            while (_body.Children.Count > 1) _body.Children.RemoveAt(1);
            if (!animate)
            {
                _pageBusy = false;
                _shown = page;
                PaintPage(page);
                return;
            }
            _pageBusy = true;
            _body.Children.Add(LoadingRow());
            var timer = new DispatcherTimer { Interval = TimeSpan.FromMilliseconds(480) };
            timer.Tick += delegate
            {
                timer.Stop();
                if (token != _pageToken) return;
                _pageBusy = false;
                while (_body.Children.Count > 1) _body.Children.RemoveAt(1);
                _shown = page;
                PaintPage(page);
            };
            timer.Start();
        }

        static bool SamePage(RunePageView a, RunePageView b)
        {
            if (a == null || b == null) return false;
            return a.PrimaryStyle == b.PrimaryStyle && a.SecondaryStyle == b.SecondaryStyle;
        }

        static StackPanel LoadingRow()
        {
            var arc = new System.Windows.Shapes.Path
            {
                Width = 18,
                Height = 18,
                Stretch = Stretch.Uniform,
                Stroke = new SolidColorBrush(Color.FromRgb(232, 196, 110)),
                StrokeThickness = 2.2,
                StrokeStartLineCap = PenLineCap.Round,
                StrokeEndLineCap = PenLineCap.Round,
                Data = Geometry.Parse("M 16,9 A 7,7 0 1 1 9,2"),
                RenderTransformOrigin = new Point(0.5, 0.5),
                VerticalAlignment = VerticalAlignment.Center
            };
            var spin = new RotateTransform();
            arc.RenderTransform = spin;
            spin.BeginAnimation(RotateTransform.AngleProperty, new DoubleAnimation(0, 360, TimeSpan.FromMilliseconds(650))
            {
                RepeatBehavior = RepeatBehavior.Forever
            });
            var row = new StackPanel
            {
                Orientation = Orientation.Horizontal,
                Margin = new Thickness(0, 18, 0, 8)
            };
            row.Children.Add(arc);
            row.Children.Add(new TextBlock
            {
                Text = "Загрузка…",
                FontSize = 13,
                Margin = new Thickness(10, 0, 0, 0),
                VerticalAlignment = VerticalAlignment.Center,
                Foreground = new SolidColorBrush(Color.FromRgb(232, 196, 110))
            });
            return row;
        }

        static double PercentTail()
        {
            var text = new FormattedText(
                "88,8%",
                CultureInfo.GetCultureInfo("ru-RU"),
                FlowDirection.LeftToRight,
                new Typeface(new FontFamily("Segoe UI"), FontStyles.Normal, FontWeights.Normal, FontStretches.Normal),
                12,
                Brushes.White);
            var tail = 54 - text.Width;
            return tail > 0 ? tail : 0;
        }

        static StackPanel RateLegend()
        {
            var box = new StackPanel
            {
                HorizontalAlignment = HorizontalAlignment.Right,
                VerticalAlignment = VerticalAlignment.Center,
                Margin = new Thickness(8, 0, 0, 0)
            };
            box.Children.Add(new TextBlock
            {
                Text = "пикрейт — жёлтый",
                FontSize = 11,
                Foreground = new SolidColorBrush(Color.FromRgb(232, 196, 110))
            });
            box.Children.Add(new TextBlock
            {
                Text = "винрейт — зелёный",
                FontSize = 11,
                Margin = new Thickness(0, 2, 0, 0),
                Foreground = new SolidColorBrush(Color.FromRgb(150, 196, 160))
            });
            return box;
        }

        void PaintPage(RunePageView page)
        {
            _body.Children.Add(Section(RuneBook.NameOf(page.PrimaryStyle)));
            for (var i = 0; i < page.Primary.Count; i++)
                _body.Children.Add(RuneRow(page.Primary[i], page.Play, i == 0));
            _body.Children.Add(Section(RuneBook.NameOf(page.SecondaryStyle)));
            foreach (var slot in page.Secondary)
                _body.Children.Add(RuneRow(slot, page.Play, false));
            _body.Children.Add(Section("Осколки"));
            var shards = new StackPanel
            {
                Orientation = Orientation.Horizontal,
                VerticalAlignment = VerticalAlignment.Center
            };
            foreach (var slot in page.Shards)
                shards.Children.Add(Shard(slot));
            var foot = new Grid { Margin = new Thickness(0, 2, 1, 0) };
            foot.ColumnDefinitions.Add(new ColumnDefinition { Width = GridLength.Auto });
            foot.ColumnDefinitions.Add(new ColumnDefinition { Width = new GridLength(1, GridUnitType.Star) });
            Grid.SetColumn(shards, 0);
            var legend = RateLegend();
            legend.Margin = new Thickness(8, 0, PercentTail(), 0);
            legend.VerticalAlignment = VerticalAlignment.Center;
            Grid.SetColumn(legend, 1);
            foot.Children.Add(shards);
            foot.Children.Add(legend);
            _body.Children.Add(foot);
        }

        static TextBlock Section(string title)
        {
            return new TextBlock
            {
                Text = title,
                FontSize = 12,
                FontWeight = FontWeights.SemiBold,
                Margin = new Thickness(0, 8, 0, 2),
                Foreground = new SolidColorBrush(Color.FromRgb(200, 170, 110))
            };
        }

        static UIElement RuneRow(RuneSlot slot, int pagePlay, bool keystone)
        {
            var row = new Grid { Margin = new Thickness(keystone ? 0 : 1, keystone ? 0 : 2, keystone ? 0 : 1, keystone ? 0 : 2) };
            row.ColumnDefinitions.Add(new ColumnDefinition { Width = new GridLength(46) });
            row.ColumnDefinitions.Add(new ColumnDefinition { Width = new GridLength(1, GridUnitType.Star) });
            row.ColumnDefinitions.Add(new ColumnDefinition { Width = new GridLength(62) });
            row.ColumnDefinitions.Add(new ColumnDefinition { Width = new GridLength(54) });
            var icon = new Image
            {
                Width = keystone ? 40 : 26,
                Height = keystone ? 40 : 26,
                HorizontalAlignment = HorizontalAlignment.Left,
                VerticalAlignment = VerticalAlignment.Center,
                Source = RuneIcons.Get(slot.Id)
            };
            var name = new TextBlock
            {
                Text = RuneBook.NameOf(slot.Id),
                FontSize = keystone ? 16 : 13,
                FontWeight = keystone ? FontWeights.SemiBold : FontWeights.Normal,
                VerticalAlignment = VerticalAlignment.Center,
                Margin = new Thickness(0, 0, 8, 0),
                Foreground = new SolidColorBrush(keystone
                    ? Color.FromRgb(255, 236, 190)
                    : Color.FromRgb(236, 232, 226)),
                TextTrimming = TextTrimming.CharacterEllipsis
            };
            var share = pagePlay > 0 ? (double)slot.Play / pagePlay : 0;
            var pick = new TextBlock
            {
                Text = Pct(share),
                FontSize = 12,
                VerticalAlignment = VerticalAlignment.Center,
                Foreground = new SolidColorBrush(Color.FromRgb(232, 196, 110))
            };
            var win = new TextBlock
            {
                Text = Win(slot.Win, slot.Play),
                FontSize = 12,
                VerticalAlignment = VerticalAlignment.Center,
                Foreground = new SolidColorBrush(Color.FromRgb(150, 196, 160))
            };
            Grid.SetColumn(icon, 0);
            Grid.SetColumn(name, 1);
            Grid.SetColumn(pick, 2);
            Grid.SetColumn(win, 3);
            row.Children.Add(icon);
            row.Children.Add(name);
            row.Children.Add(pick);
            row.Children.Add(win);
            if (!keystone) return row;
            return new Border
            {
                Margin = new Thickness(0, 4, 0, 8),
                Padding = new Thickness(0, 6, 0, 6),
                CornerRadius = new CornerRadius(8),
                Background = new SolidColorBrush(Color.FromArgb(55, 232, 196, 110)),
                BorderBrush = new SolidColorBrush(Color.FromRgb(232, 196, 110)),
                BorderThickness = new Thickness(1),
                Child = row
            };
        }

        static Border Shard(RuneSlot slot)
        {
            var icon = new Image
            {
                Width = 20,
                Height = 20,
                Stretch = Stretch.Uniform,
                HorizontalAlignment = HorizontalAlignment.Center,
                VerticalAlignment = VerticalAlignment.Center,
                IsHitTestVisible = false,
                Source = RuneIcons.Get(slot.Id)
            };
            return new Border
            {
                Width = 34,
                Height = 34,
                Margin = new Thickness(0, 0, 8, 0),
                CornerRadius = new CornerRadius(17),
                Background = new SolidColorBrush(Color.FromArgb(220, 28, 26, 22)),
                BorderBrush = new SolidColorBrush(Color.FromRgb(200, 170, 110)),
                BorderThickness = new Thickness(1),
                ToolTip = RuneBook.NameOf(slot.Id),
                Child = icon
            };
        }

        static string Pct(double rate)
        {
            return (rate * 100).ToString("0.0", CultureInfo.GetCultureInfo("ru-RU")) + "%";
        }

        static string Win(int win, int play)
        {
            if (play <= 0) return "";
            return (100.0 * win / play).ToString("0.0", CultureInfo.GetCultureInfo("ru-RU")) + "%";
        }

        static bool Same(string a, string b)
        {
            return string.Equals(a ?? "", b ?? "", StringComparison.OrdinalIgnoreCase);
        }

        static string Norm(string value)
        {
            var chars = new List<char>();
            foreach (var ch in (value ?? "").ToLowerInvariant())
            {
                if (char.IsLetterOrDigit(ch)) chars.Add(ch);
            }
            return new string(chars.ToArray());
        }
    }
}
