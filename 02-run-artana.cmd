@echo off
REM ========================================================================================
REM �奬� ࠡ��� �ਯ�
REM 1. ��頥� ����� ARJ
REM 2. ����᪠�� artana ��� ��㯯�஢�� ᮣ��᭮ 440-� 䠩��� �� ����� FILES � ��ࠢ��
REM 3. �믮��塞 ����� ᣥ���஢������ � artana �ਯ� mv.cmd ��� ��⮬���᪮� 
REM    ���⠭���� ��, ��஢����, ��娢�஢����, � ���⠭���� �� �� ��娢�
REM 4. ��⮢� � ���쭥�襩 ��ࠡ�⪥ � ��� 䠩�� ��室���� � ����� ARJ
REM ========================================================================================
REM artana.exe -findex=N ������ ��砫쭮� ���祭�� ����� ��娢�
SET SCRIPT_DIR=C:\FOIV\440p\OUT
C:
CD %SCRIPT_DIR%
del /s/q %SCRIPT_DIR%\ARJ
cls
REM -findex=N ������ ��砫쭮� ���祭�� ����� ��娢�
cd %SCRIPT_DIR%\ZDATA
artana.exe -src="%SCRIPT_DIR%\OUT" -dst="%SCRIPT_DIR%\DONE" -signatura="C:\Program Files\MDPREI\spki" -findex=1
rem pause
%SCRIPT_DIR%\ZDATA\mv.cmd
rem echo ����� ���������!
pause
