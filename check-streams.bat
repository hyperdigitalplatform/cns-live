@echo off
echo ==========================================
echo    Stream Counter Quick Check
echo ==========================================
echo.

echo [API Stats]
curl -s http://localhost:8000/api/v1/stream/stats | python -m json.tool | findstr "active_streams total_viewers camera_id camera_name source"
echo.

echo [Valkey Counts]
docker exec cctv-valkey valkey-cli GET "stream:count:DUBAI_POLICE" > temp_dubai.txt
docker exec cctv-valkey valkey-cli GET "stream:count:METRO" > temp_metro.txt
set /p DUBAI=<temp_dubai.txt
set /p METRO=<temp_metro.txt
echo   DUBAI_POLICE: %DUBAI%
echo   METRO:        %METRO%
del temp_dubai.txt temp_metro.txt 2>nul
echo.
echo ==========================================
