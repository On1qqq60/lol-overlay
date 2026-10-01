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
            if (string.IsNullOrEmpty(id) || id == "0")
            {
                return new Border
                {
                    Width = size,
                    Height = size,
                    Margin = new Thickness(3, 2, 3, 2),
                    BorderThickness = new Thickness(1),
                    BorderBrush = new SolidColorBrush(Color.FromArgb(50, 255, 255, 255)),
                    Background = new SolidColorBrush(Color.FromArgb(80, 12, 10, 8)),
                    SnapsToDevicePixels = true
                };
            }
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
            return new System.Windows.Shapes.Path
            {
                Width = 14,
                Height = 14,
                Stretch = Stretch.Uniform,
                Fill = new SolidColorBrush(Color.FromRgb(232, 196, 110)),
                HorizontalAlignment = HorizontalAlignment.Center,
                VerticalAlignment = VerticalAlignment.Center,
                Effect = Glow(),
                Data = Geometry.Parse(
                    "M10,9 H14 V6 H17 L12,1 7,6 H10 Z " +
                    "M9,10 H6 V7 L1,12 6,17 V14 H9 Z " +
                    "M15,10 H18 V7 L23,12 18,17 V14 H15 Z " +
                    "M10,15 V18 H7 L12,23 17,18 H14 V15 Z")
            };
        }

        public static UIElement InfoGlyph()
        {
            return StarGlyph();
        }
    }
}
