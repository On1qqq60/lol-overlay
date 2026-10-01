using System;
using System.Diagnostics;
using System.IO;
using System.Runtime.InteropServices;
using System.Runtime.InteropServices.ComTypes;
using System.Text;
using WinForms = System.Windows.Forms;

namespace LolBuildOverlay
{
    static class ToastNote
    {
        public const string AppId = "LolBuild.Overlay";

        public static void Show(string body, WinForms.NotifyIcon tray)
        {
            var text = body ?? "";
            var ok = false;
            try { ok = TryToast(text); }
            catch { ok = false; }
            if (!ok) Balloon(tray, text);
        }

        static bool TryToast(string body)
        {
            InstallShortcut();
            var xml = "<toast><visual><binding template=\"ToastGeneric\"><text>Запрос сейчас недоступен</text><text>"
                + Xml(body) + "</text></binding></visual><audio silent=\"true\"/></toast>";
            var script = "$ErrorActionPreference='Stop';"
                + "[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null;"
                + "[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null;"
                + "$doc = New-Object Windows.Data.Xml.Dom.XmlDocument;"
                + "$doc.LoadXml($env:LOLBUILD_TOAST_XML);"
                + "$toast = [Windows.UI.Notifications.ToastNotification]::new($doc);"
                + "[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('" + AppId + "').Show($toast);"
                + "Start-Sleep -Seconds 2";
            var encoded = Convert.ToBase64String(Encoding.Unicode.GetBytes(script));
            var psi = new ProcessStartInfo();
            psi.FileName = Path.Combine(Environment.SystemDirectory, "WindowsPowerShell\\v1.0\\powershell.exe");
            psi.Arguments = "-NoProfile -ExecutionPolicy Bypass -EncodedCommand " + encoded;
            psi.CreateNoWindow = true;
            psi.UseShellExecute = false;
            psi.WindowStyle = ProcessWindowStyle.Hidden;
            psi.EnvironmentVariables["LOLBUILD_TOAST_XML"] = xml;
            var proc = Process.Start(psi);
            if (proc == null) return false;
            if (!proc.WaitForExit(8000)) return true;
            return proc.ExitCode == 0;
        }

        static void Balloon(WinForms.NotifyIcon tray, string body)
        {
            if (tray == null) return;
            var text = body;
            if (text.Length > 240) text = text.Substring(0, 237) + "...";
            try { tray.ShowBalloonTip(8000, "Запрос сейчас недоступен", text, WinForms.ToolTipIcon.Info); }
            catch { }
        }

        static string Xml(string text)
        {
            if (string.IsNullOrEmpty(text)) return "";
            return text.Replace("&", "&amp;").Replace("<", "&lt;").Replace(">", "&gt;");
        }

        static void InstallShortcut()
        {
            var dir = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.ApplicationData),
                "Microsoft", "Windows", "Start Menu", "Programs");
            Directory.CreateDirectory(dir);
            var path = Path.Combine(dir, "lol.build.lnk");
            var exe = System.Reflection.Assembly.GetExecutingAssembly().Location;
            var link = (IShellLinkW)new ShellLink();
            link.SetPath(exe);
            link.SetWorkingDirectory(Path.GetDirectoryName(exe));
            link.SetDescription("lol.build");
            link.SetIconLocation(exe, 0);
            var store = (IPropertyStore)link;
            var key = AppUserModelId;
            var variant = new PropVariant();
            variant.vt = 31;
            variant.value = Marshal.StringToCoTaskMemUni(AppId);
            try
            {
                store.SetValue(ref key, ref variant);
                store.Commit();
            }
            finally
            {
                if (variant.value != IntPtr.Zero) Marshal.FreeCoTaskMem(variant.value);
            }
            ((IPersistFile)link).Save(path, true);
        }

        static PropertyKey AppUserModelId
        {
            get
            {
                return new PropertyKey
                {
                    fmtid = new Guid("9F4C2855-9F79-4B39-A8D0-E1D42DE1D5F3"),
                    pid = 5
                };
            }
        }

        [StructLayout(LayoutKind.Sequential)]
        struct PropertyKey
        {
            public Guid fmtid;
            public uint pid;
        }

        [StructLayout(LayoutKind.Sequential)]
        struct PropVariant
        {
            public ushort vt;
            public ushort r1;
            public ushort r2;
            public ushort r3;
            public IntPtr value;
        }

        [ComImport, Guid("886D8EEB-8CF2-4446-8D02-CDBA1DBDCF99"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
        interface IPropertyStore
        {
            void GetCount(out uint count);
            void GetAt(uint index, out PropertyKey key);
            void GetValue(ref PropertyKey key, out PropVariant value);
            void SetValue(ref PropertyKey key, ref PropVariant value);
            void Commit();
        }

        [ComImport, Guid("000214F9-0000-0000-C000-000000000046"), InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
        interface IShellLinkW
        {
            void GetPath();
            void GetIDList();
            void SetIDList();
            void GetDescription();
            void SetDescription([MarshalAs(UnmanagedType.LPWStr)] string name);
            void GetWorkingDirectory();
            void SetWorkingDirectory([MarshalAs(UnmanagedType.LPWStr)] string dir);
            void GetArguments();
            void SetArguments();
            void GetHotkey();
            void SetHotkey();
            void GetShowCmd();
            void SetShowCmd();
            void GetIconLocation();
            void SetIconLocation([MarshalAs(UnmanagedType.LPWStr)] string path, int icon);
            void SetRelativePath();
            void Resolve();
            void SetPath([MarshalAs(UnmanagedType.LPWStr)] string path);
        }

        [ComImport, Guid("00021401-0000-0000-C000-000000000046")]
        class ShellLink { }
    }
}
