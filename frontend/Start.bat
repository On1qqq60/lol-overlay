@echo off
cd /d "%~dp0"
set "PATH=C:\Program Files\Go\bin;%PATH%"
set "CSC=%WINDIR%\Microsoft.NET\Framework64\v4.0.30319\csc.exe"
set "FX=%WINDIR%\Microsoft.NET\Framework64\v4.0.30319"
set "WPF=%FX%\WPF"
set "BACK=%~dp0..\backend"

where go >nul 2>&1
if errorlevel 1 (
  echo Go not found. Install Go 1.22+ from https://go.dev/dl/ and re-run frontend\Start.bat
  pause
  exit /b 1
)
if not exist "%CSC%" (
  echo csc.exe not found. Install .NET Framework 4.x Developer Pack / targeting pack.
  pause
  exit /b 1
)
if not exist "%BACK%\go.mod" (
  echo backend\go.mod not found.
  pause
  exit /b 1
)

if not exist bin mkdir bin
if not exist bin\data mkdir bin\data

echo Building recommend engine...
pushd "%BACK%"
go build -o "%~dp0bin\recommend.exe" ./cmd/recommend
if errorlevel 1 (
  echo Go build failed.
  popd
  pause
  exit /b 1
)
popd

echo Compiling overlay...
del /q bin\sources.rsp 2>nul
for %%f in (src\*.cs) do echo %%f>> bin\sources.rsp
"%CSC%" /nologo /target:winexe /out:bin\LolBuildOverlay.exe /win32icon:app.ico ^
  /r:"%WPF%\PresentationCore.dll" ^
  /r:"%WPF%\PresentationFramework.dll" ^
  /r:"%WPF%\WindowsBase.dll" ^
  /r:"%FX%\System.Xaml.dll" ^
  /r:"%FX%\System.Windows.Forms.dll" ^
  /r:"%FX%\System.Drawing.dll" ^
  /r:"%FX%\System.Web.Extensions.dll" ^
  /r:"%FX%\System.Core.dll" ^
  /r:"%FX%\System.IO.Compression.dll" ^
  /r:"%FX%\System.IO.Compression.FileSystem.dll" ^
  @bin\sources.rsp
if errorlevel 1 (
  echo Compile failed.
  pause
  exit /b 1
)

copy /y app.ico bin\app.ico >nul
copy /y data\names.txt bin\data\names.txt >nul
start "" "%~dp0bin\LolBuildOverlay.exe"
