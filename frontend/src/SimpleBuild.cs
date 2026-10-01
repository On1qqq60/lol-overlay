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
}
