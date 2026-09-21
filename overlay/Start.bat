@echo off
cd /d "%~dp0"
set "PATH=C:\Program Files\Go\bin;%PATH%"
set "CSC=%WINDIR%\Microsoft.NET\Framework64\v4.0.30319\csc.exe"
set "FX=%WINDIR%\Microsoft.NET\Framework64\v4.0.30319"
set "WPF=%FX%\WPF"

echo Building recommend engine...
pushd "%~dp0.."
if not exist "go.mod" (
  echo Engine go.mod not found next to overlay\
  popd
  pause
  exit /b 1
)
go build -o "%~dp0recommend.exe" ./cmd/recommend
if errorlevel 1 (
  echo Go build failed. Install Go from https://go.dev/dl/
  popd
  pause
  exit /b 1
)
popd

echo Compiling overlay...
"%CSC%" /nologo /target:winexe /out:LolBuildOverlay.exe ^
  /r:"%WPF%\PresentationCore.dll" ^
  /r:"%WPF%\PresentationFramework.dll" ^
  /r:"%WPF%\WindowsBase.dll" ^
  /r:"%FX%\System.Xaml.dll" ^
  /r:"%FX%\System.Windows.Forms.dll" ^
  /r:"%FX%\System.Drawing.dll" ^
  /r:"%FX%\System.Web.Extensions.dll" ^
  /r:"%FX%\System.Core.dll" ^
  Overlay.cs Models.cs LiveClient.cs EngineClient.cs
if errorlevel 1 (
  echo Compile failed.
  pause
  exit /b 1
)
start "" "%~dp0LolBuildOverlay.exe"
