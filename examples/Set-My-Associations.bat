@echo off
setlocal EnableExtensions DisableDelayedExpansion

rem ============================== CONFIG ==============================
set "TEXT_EXE=C:\PortableApps\Notepad3\Notepad3.exe"
set "TEXT_EXTS=.txt,.log,.md,.markdown,.nfo,.ini,.cfg,.conf,.csv,.json,.xml,.yaml,.yml,.toml,.ps1,.psm1,.bat,.cmd,.reg,.sql,.c,.h,.cpp,.hpp,.cs,.java,.py,.js,.jsx,.tsx,.css"

set "IMAGE_EXE=C:\PortableApps\Honeyview\Honeyview.exe"
set "IMAGE_EXTS=.jpg,.jpeg,.jpe,.jfif,.png,.gif,.bmp,.webp,.tif,.tiff,.tga,.ico,.avif,.heic,.psd,.dds"

set "MEDIA_EXE=C:\PortableApps\PotPlayer\PotPlayerMini64.exe"
set "MEDIA_EXTS=.mp3,.flac,.wav,.aac,.m4a,.ogg,.opus,.wma,.ape,.ac3,.mid,.midi,.mp4,.mkv,.avi,.mov,.wmv,.flv,.webm,.m4v,.mpg,.mpeg,.ts,.m2ts,.mts,.vob,.3gp,.rm,.rmvb,.m3u,.m3u8,.pls"

rem 7zFM.exe is the file opener; 7z.exe is the console compression tool.
set "ARCHIVE_EXE=C:\PortableApps\7-Zip\7zFM.exe"
set "ARCHIVE_EXTS=.7z,.zip,.rar,.001,.tar,.gz,.gzip,.tgz,.bz2,.bzip2,.tbz,.tbz2,.xz,.txz,.zst,.cab,.iso,.wim,.arj,.cpio,.lzh,.lha,.rpm,.deb"
rem ============================ END CONFIG ============================

set "SETFTA=%~dp0..\SetFTA.exe"
if not exist "%SETFTA%" (
    echo [ERROR] SetFTA.exe was not found in the parent directory.
    pause
    exit /b 2
)

fltmc.exe >nul 2>&1
if errorlevel 1 (
    set "SETFTA_CALLER=%~f0"
    powershell.exe -NoProfile -Command "Start-Process -FilePath $env:SETFTA_CALLER -Verb RunAs"
    exit /b
)

"%SETFTA%" ^
  --app "PortableAssoc.Notepad3|Text documents and source files|%TEXT_EXE%|%TEXT_EXTS%" ^
  --app "PortableAssoc.Honeyview|Images|%IMAGE_EXE%|%IMAGE_EXTS%" ^
  --app "PortableAssoc.PotPlayer|Audio and video|%MEDIA_EXE%|%MEDIA_EXTS%" ^
  --app "PortableAssoc.7Zip|Archives|%ARCHIVE_EXE%|%ARCHIVE_EXTS%"

exit /b %errorlevel%
