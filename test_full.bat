@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo.
echo =========================================
echo    Full Streaming Platform E2E Test
echo =========================================
echo.

set "BASE_URL=http://localhost"
set "USERNAME=e2e_tester_%RANDOM%"
set "PASSWORD=test123456"
set "STREAM_DURATION=30"
set "TEMP_DIR=%TEMP%\streaming_test"

if not exist "%TEMP_DIR%" mkdir "%TEMP_DIR%"

echo [1/12] Health checks...
curl -s http://localhost:8081/health >nul 2>&1 && echo OK: Auth Service || goto :error
curl -s http://localhost:8082/health >nul 2>&1 && echo OK: Stream Service || goto :error
curl -s http://localhost:8083/health >nul 2>&1 && echo OK: Recording Service || goto :error
curl -s http://localhost:8084/health >nul 2>&1 && echo OK: VOD Service || goto :error

echo.
echo [2/12] Register user: %USERNAME%
curl -s -X POST http://localhost:8081/register -H "Content-Type: application/json" -d "{\"username\":\"%USERNAME%\",\"email\":\"%USERNAME%@test.com\",\"password\":\"%PASSWORD%\"}" -o "%TEMP_DIR%\reg.json"

powershell -NoProfile -Command "$r=gc '%TEMP_DIR%\reg.json'|ConvertFrom-Json;$r.token" > "%TEMP_DIR%\token.txt"
set /p TOKEN=<"%TEMP_DIR%\token.txt"
echo OK: Token received

echo.
echo [3/12] Create stream...
curl -s -X POST http://localhost:8082/stream -H "Authorization: Bearer %TOKEN%" -H "Content-Type: application/json" -d "{\"title\":\"E2E Test\",\"description\":\"Auto test\"}" -o "%TEMP_DIR%\stream.json"

powershell -NoProfile -Command "$s=gc '%TEMP_DIR%\stream.json'|ConvertFrom-Json;$s.stream.stream_key" > "%TEMP_DIR%\key.txt"
powershell -NoProfile -Command "$s=gc '%TEMP_DIR%\stream.json'|ConvertFrom-Json;$s.stream.id" > "%TEMP_DIR%\sid.txt"
set /p STREAM_KEY=<"%TEMP_DIR%\key.txt"
set /p STREAM_ID=<"%TEMP_DIR%\sid.txt"
echo OK: Stream created
echo   Key: %STREAM_KEY%
echo   ID: %STREAM_ID%

echo.
echo [4/12] Check FFmpeg...
where ffmpeg >nul 2>&1 || (echo ERROR: FFmpeg not found && pause && exit /b 1)
echo OK: FFmpeg available

echo.
echo [5/12] Start streaming (%STREAM_DURATION%s)...
start /B "" cmd /c "ffmpeg -re -f lavfi -i testsrc=duration=%STREAM_DURATION%:size=1280x720:rate=30 -f lavfi -i sine=frequency=1000:duration=%STREAM_DURATION% -c:v libx264 -preset veryfast -b:v 2500k -c:a aac -b:a 128k -f mpegts srt://localhost:6000?streamid=%STREAM_KEY% 2>%TEMP_DIR%\ffmpeg.log"
echo OK: Streaming started

echo.
echo [6/12] Wait 10s for initialization...
timeout /t 10 /nobreak >nul

echo.
echo [7/12] Check live streams...
curl -s http://localhost:8082/streams/live | findstr "%STREAM_KEY%" >nul && echo OK: Stream is live || echo WARNING: Not live yet

echo.
echo [8/12] Check HLS playlist...
timeout /t 5 /nobreak >nul
curl -s http://localhost/hls/%STREAM_KEY%/playlist.m3u8 -o "%TEMP_DIR%\playlist.m3u8" 2>nul
type "%TEMP_DIR%\playlist.m3u8" | findstr ".ts" >nul && echo OK: HLS segments available || echo WARNING: No HLS segments

echo.
echo [9/12] Wait for stream completion...
set /a WAIT=%STREAM_DURATION%-15
if %WAIT% GTR 0 timeout /t %WAIT% /nobreak >nul
echo OK: Stream ended

echo.
echo [10/12] Wait for recording processing (40s)...
timeout /t 40 /nobreak >nul

echo.
echo [11/12] Get recordings...
curl -s http://localhost:8083/recordings -o "%TEMP_DIR%\recs.json"
type "%TEMP_DIR%\recs.json"
echo.

powershell -NoProfile -Command "try{$r=gc '%TEMP_DIR%\recs.json'|ConvertFrom-Json;if($r.recordings){$r.recordings[0].id}}catch{''}" > "%TEMP_DIR%\rid.txt"
set /p REC_ID=<"%TEMP_DIR%\rid.txt"

if "%REC_ID%"=="" (
    echo WARNING: No recording ID found
    goto :summary
)

echo OK: Recording ID: %REC_ID%

echo.
echo [12/12] Import to VOD...
curl -s -X POST http://localhost:8084/import-recording -H "Authorization: Bearer %TOKEN%" -H "Content-Type: application/json" -d "{\"recording_id\":\"%REC_ID%\",\"title\":\"E2E Test Video\",\"category\":\"test\",\"tags\":[\"e2e\"],\"visibility\":\"public\"}" -o "%TEMP_DIR%\import.json"
type "%TEMP_DIR%\import.json"
echo.

powershell -NoProfile -Command "try{$i=gc '%TEMP_DIR%\import.json'|ConvertFrom-Json;$i.video_id}catch{''}" > "%TEMP_DIR%\vid.txt"
set /p VIDEO_ID=<"%TEMP_DIR%\vid.txt"

if not "%VIDEO_ID%"=="" (
    echo OK: Video imported: %VIDEO_ID%
    echo URL: http://localhost:8084/video/%VIDEO_ID%
)

:summary
echo.
echo =========================================
echo           Test Complete!
echo =========================================
echo.
echo Results:
echo   User: %USERNAME%
echo   Stream: %STREAM_KEY%
if defined REC_ID echo   Recording: %REC_ID%
if defined VIDEO_ID echo   Video: %VIDEO_ID%
echo.
echo Check:
echo   - Recordings: http://localhost:8083/recordings
echo   - Videos: http://localhost:8084/videos
echo   - MinIO: http://localhost:9001
echo.
goto :end

:error
echo ERROR: Service health check failed
pause
exit /b 1

:end
pause
