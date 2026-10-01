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
    public class SettingsWindow : Window
    {
        public event Action Changed;
        private int _capture;
        private readonly Button _bind;
        private readonly Button _hideBind;

        public SettingsWindow()
        {
            Title = "lol.build";
            Icon = AppIcon.Wpf();
            Width = 340;
            Height = 390;
            Topmost = true;
            ResizeMode = ResizeMode.NoResize;
            WindowStartupLocation = WindowStartupLocation.CenterScreen;
            Background = new SolidColorBrush(Color.FromRgb(14, 16, 20));
            Foreground = new SolidColorBrush(Color.FromRgb(232, 228, 217));

            var root = new StackPanel { Margin = new Thickness(16) };

            var show = new CheckBox
            {
                Content = "Показывать сборку",
                IsChecked = AppSettings.ShowBuild,
                Margin = new Thickness(0, 0, 0, 8),
                Foreground = new SolidColorBrush(Color.FromRgb(232, 228, 217))
            };
            show.Checked += (s, e) => { AppSettings.ShowBuild = true; if (Changed != null) Changed(); };
            show.Unchecked += (s, e) => { AppSettings.ShowBuild = false; if (Changed != null) Changed(); };
            root.Children.Add(show);

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
            _bind = BindButton(AppSettings.KeyName, 1);
            root.Children.Add(_bind);

            root.Children.Add(Label("Бинд скрытия оверлея"));
            _hideBind = BindButton(AppSettings.HideKeyName, 2);
            root.Children.Add(_hideBind);

            Content = root;
            PreviewKeyDown += OnKey;
        }

        Button BindButton(string name, int slot)
        {
            var button = new Button
            {
                Content = name,
                Height = 32,
                Margin = new Thickness(0, 4, 0, 0),
                Background = new SolidColorBrush(Color.FromRgb(28, 22, 12)),
                Foreground = new SolidColorBrush(Color.FromRgb(200, 170, 110)),
                BorderBrush = new SolidColorBrush(Color.FromRgb(200, 170, 110))
            };
            button.Click += (s, e) =>
            {
                _capture = slot;
                button.Content = "нажмите клавишу…";
            };
            return button;
        }

        private void OnKey(object sender, KeyEventArgs e)
        {
            if (_capture == 0) return;
            e.Handled = true;
            var key = e.Key == Key.System ? e.SystemKey : e.Key;
            if (key == Key.Escape)
            {
                _bind.Content = AppSettings.KeyName;
                _hideBind.Content = AppSettings.HideKeyName;
                _capture = 0;
                return;
            }
            var vk = (uint)KeyInterop.VirtualKeyFromKey(key);
            if (vk == 0) return;
            if (_capture == 2)
            {
                AppSettings.HideVk = vk;
                AppSettings.HideKeyName = key.ToString();
                _hideBind.Content = AppSettings.HideKeyName;
            }
            else
            {
                AppSettings.Vk = vk;
                AppSettings.KeyName = key.ToString();
                _bind.Content = AppSettings.KeyName;
            }
            _capture = 0;
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
}
