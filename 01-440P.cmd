@echo off
SET SCRIPTPATH=C:\FOIV\440p\OUT
SET BASEPATH=C:\FOIV\440p\OUT
SET FILESPATH=%BASEPATH%\FILES
SET BAK=%BASEPATH%\DATA\ARCHIVE
SET LOG=%BASEPATH%\DATA\LOG

C:
cd %SRCPATH%

for /F %%i in ('dir /b /a "%FILESPATH%\*"') do (
    echo В папке есть файлы -- обработка!
    goto launch_app
    
)
echo Папка обработки пуста. Завершено!
exit
echo byebye

:launch_app
set PATH=%PATH%;C:\Program Files\MDPREI\spki
set SRCPATH=%BASEPATH%\FILES
set RESPATH=%BASEPATH%\OUT
set TMPPATH=%BASEPATH%\TMP

rename %BASEPATH%\CLEAN\PB*.xml PB*.bak
rename %BASEPATH%\CLEAN\*.xml *.vrb
rename %BASEPATH%\CLEAN\PB*.bak PB*.xml


REM ***** SETTING SIGN
C:
cd %SRCPATH%
@dir *.xml /b > list
@dir *.vrb /b >> list

for /F %%i in (list) do (
spki1utl.exe -registry -profile FOIV2019 -silent %LOG%\foiv-440p-sign.log  -sign -data %%i -out %TMPPATH%\%%i
SET erFile = %%i
IF NOT EXIST %%i GOTO ERR
rem @del %%i
)

cscript.exe %SCRIPTPATH%\bak.vbs "%FILESPATH%" "%BAK%"

REM ***** COMPRESSING GZIP
cd %TMPPATH%
@dir *.vrb /b > list

for /F %%i in (list) do (
7z a -tgzip -mx9 %%i.gz %%i -sdel
)

rename *.gz *.
move /y *.xml %RESPATH%

REM ***** ENCRYPTION VRB

cd %TMPPATH%
@dir *.vrb /b > list

for /F %%i in (list) do (
spki1utl.exe -registry -profile FOIV2019 -reclist %SCRIPTPATH%\abon.lst -silent %LOG%\foiv-440p.log -encrypt -in %%i -out %RESPATH%\%%i
SET erFile = %%i
IF NOT EXIST %%i GOTO ERR
)

del /q *.vrb
del /q list

GOTO OK

:ERR
echo ERROR NO FILES FOUND
exit

:OK
echo SCRIPT DONE

cd %SCRIPTPATH%
start /w 02-run-artana.cmd
del /s/q %SCRIPTPATH%\OUT
del /s/q %SCRIPTPATH%\TMP

exit