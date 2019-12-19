@echo off
REM ========================================================================================
REM Схема работы скрипта
REM 1. Очищаем папку ARJ
REM 2. Запускаем artana для группировки согласно 440-П файлов из папки FILES к отправке
REM 3. Выполняем запуск сгенерированного в artana скрипта mv.cmd для автоматической 
REM    простановки КА, шифрования, архивирования, и простановки КА на архивы
REM 4. Готовые к дальнейшей обработке в ПТК файлы находятся в папке ARJ
REM ========================================================================================
REM artana.exe -findex=N задает начальное значение номера архива
SET SCRIPT_DIR=C:\FOIV\440p\OUT
C:
CD %SCRIPT_DIR%
del /s/q %SCRIPT_DIR%\ARJ
cls
REM -findex=N задает начальное значение номера архива
cd %SCRIPT_DIR%\ZDATA
artana.exe -src="%SCRIPT_DIR%\OUT" -dst="%SCRIPT_DIR%\DONE" -signatura="C:\Program Files\MDPREI\spki" -findex=1
rem pause
%SCRIPT_DIR%\ZDATA\mv.cmd
rem echo КОНЕЦ ОБРАБОТКИ!
pause
