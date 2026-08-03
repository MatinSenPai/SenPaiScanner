@echo off
setlocal

cd /d "%~dp0"

echo Building Go mobile library for Android...

where gomobile >nul 2>nul
if %errorlevel% neq 0 (
    echo gomobile not found. Installing...
    go install golang.org/x/mobile/cmd/gomobile@latest
    go install golang.org/x/mobile/cmd/gobind@latest
    for /f "delims=" %%i in ('go env GOPATH') do set GOPATH=%%i
    set "PATH=%PATH%;%GOPATH%\bin"
    gomobile init
    if errorlevel 1 exit /b %errorlevel%
)

if not exist app\libs mkdir app\libs

echo Running gomobile bind...
gomobile bind -v -target=android/arm64,android/arm -androidapi 21 -javapkg=com.matinsenpai.senpaiscanner -o app\libs\senpaiscanner.aar ..\mobile
if errorlevel 1 (
    echo gomobile bind failed.
    exit /b %errorlevel%
)

echo Successfully built senpaiscanner.aar!
endlocal
