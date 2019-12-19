'путь - папка где наход€тс€ файлы дл€ архивации:
arcsource = WScript.Arguments.Item(0)

'путь - где будет создаватьс€ дерево архивов:
arcpath = WScript.Arguments.Item(1)

strComputer = "."

' Date and time

Set objWMIService = GetObject("winmgmts:\\" & strComputer & "\root\cimv2")
Set colItems = objWMIService.ExecQuery("Select * from Win32_OperatingSystem")

For Each objItem in colItems
    dtmLocalTime = objItem.LocalDateTime
    dtmMonth = Mid(dtmLocalTime, 5, 2)
    dtmDay = Mid(dtmLocalTime, 7, 2)
    dtmYear = Left(dtmLocalTime, 4)
Next

'¬ычитаем один день из текущей даты (при условии что архивный файл
'формируетс€ на следующий день после формировани€ Ё—.

t = ""

'if (dtmDay-1) < 10 then 
't = "0"
'end if 

VDATE = dtmYear & "" & dtmMonth & "" & t & dtmDay-1   'дата формировани€

'создаем папки дл€ архивации сообщений
dim filesys
set filesys=CreateObject("Scripting.FileSystemObject")
If Not filesys.FolderExists(arcpath) Then
filesys.CreateFolder(arcpath)
End If
If Not filesys.FolderExists(arcpath & "\" & dtmYear) Then
filesys.CreateFolder(arcpath & "\" & dtmYear)
End If
If Not filesys.FolderExists(arcpath & "\" & dtmYear & "\" & dtmMonth) Then
filesys.CreateFolder(arcpath & "\" & dtmYear & "\" & dtmMonth)
End If
If Not filesys.FolderExists(arcpath & "\" & dtmYear & "\" & dtmMonth & "\" & t & dtmDay) Then
filesys.CreateFolder(arcpath & "\" & dtmYear & "\" & dtmMonth & "\" & t & dtmDay)
End If

'копируем *.* в архив
set filesys=CreateObject("Scripting.FileSystemObject")
filesys.CopyFile arcsource & "\*.*", arcpath & "\" & dtmYear & "\" & dtmMonth & "\" & t & dtmDay,TRUE
'удал€ем файлы в исходной директории
filesys.DeleteFile arcsource & "\*.*", True 
