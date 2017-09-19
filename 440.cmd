@echo off

SET SCRIPT_DIR=E:\temp\fns\ARTANA
E:
CD %SCRIPT_DIR%
del /s/q %SCRIPT_DIR%\ARJ

cls

REM -findex=N задает начальное значение номера архива
cd %SCRIPT_DIR%\ZDATA
artana.exe -src="%SCRIPT_DIR%\files" -dst="%SCRIPT_DIR%\OUT" -verba="C:\temp" -keyfsr="7020942009" -findex=1
