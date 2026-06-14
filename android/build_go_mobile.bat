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
)

if not exist app\libs mkdir app\libs

echo Running gomobile bind...
gomobile bind -v -target=android/arm64,android/arm -androidapi 21 -javapkg=com.ehsatn.magik -o app\libs\magik.aar ..\mobile

echo Successfully built magik.aar!
endlocal
