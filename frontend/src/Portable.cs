using System;
using System.IO;
using System.IO.Compression;
using System.Reflection;

namespace LolBuildOverlay
{
    static class Portable
    {
        public static void Ensure()
        {
            var dir = Path.GetDirectoryName(Assembly.GetExecutingAssembly().Location);
            if (string.IsNullOrEmpty(dir)) return;
            var asm = Assembly.GetExecutingAssembly();
            Stream packed = null;
            var names = asm.GetManifestResourceNames();
            for (var i = 0; i < names.Length; i++)
            {
                if (!names[i].EndsWith("payload.zip", StringComparison.OrdinalIgnoreCase)) continue;
                packed = asm.GetManifestResourceStream(names[i]);
                break;
            }
            if (packed == null) return;
            using (packed)
            using (var zip = new ZipArchive(packed, ZipArchiveMode.Read, true))
            {
                foreach (var entry in zip.Entries)
                {
                    if (string.IsNullOrEmpty(entry.Name)) continue;
                    var rel = entry.FullName.Replace('/', Path.DirectorySeparatorChar);
                    if (rel.Contains(".." + Path.DirectorySeparatorChar) || rel.StartsWith("..")) continue;
                    var dest = Path.GetFullPath(Path.Combine(dir, rel));
                    var root = Path.GetFullPath(dir + Path.DirectorySeparatorChar);
                    if (!dest.StartsWith(root, StringComparison.OrdinalIgnoreCase)) continue;
                    if (File.Exists(dest)) continue;
                    var folder = Path.GetDirectoryName(dest);
                    if (!string.IsNullOrEmpty(folder)) Directory.CreateDirectory(folder);
                    using (var input = entry.Open())
                    using (var output = File.Create(dest))
                        input.CopyTo(output);
                }
            }
        }
    }
}
