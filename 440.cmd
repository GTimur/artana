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

SET SCRIPT_DIR=E:\temp\fns\ARTANA
E:
CD %SCRIPT_DIR%
del /s/q %SCRIPT_DIR%\ARJ
cls
cd %SCRIPT_DIR%\ZDATA
artana.exe -src="%SCRIPT_DIR%\files" -dst="%SCRIPT_DIR%\OUT" -verba="C:\temp" -keyfsr="7020942009" -findex=1
%SCRIPT_DIR%\ZDATA\mv.cmd
pause

