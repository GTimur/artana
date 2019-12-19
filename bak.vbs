'���� - ����� ��� ��������� ����� ��� ���������:
arcsource = WScript.Arguments.Item(0)

'���� - ��� ����� ����������� ������ �������:
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

'�������� ���� ���� �� ������� ���� (��� ������� ��� �������� ����
'����������� �� ��������� ���� ����� ������������ ��.

t = ""

'if (dtmDay-1) < 10 then 
't = "0"
'end if 

VDATE = dtmYear & "" & dtmMonth & "" & t & dtmDay-1   '���� ������������

'������� ����� ��� ��������� ���������
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

'�������� *.* � �����
set filesys=CreateObject("Scripting.FileSystemObject")
filesys.CopyFile arcsource & "\*.*", arcpath & "\" & dtmYear & "\" & dtmMonth & "\" & t & dtmDay,TRUE
'������� ����� � �������� ����������
filesys.DeleteFile arcsource & "\*.*", True 
