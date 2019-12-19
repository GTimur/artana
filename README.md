# artana 1.10 (2019)
Утилита группировки файлов для архива 440-П

Решает следующую задачу согласно сказанного в 440-П:

	Файлы формируются с помощью программы архиватора ARJ32.
	Каждый Архивный файл содержит не более 50 файлов и имеет размер не более 50 мб.

	Наименование Архивного файла имеет следующую структуру:
	<AFN_SSSSSSS_RRRRRRR_ГГГГММДД_NNNNN.arj>

Ключи:
  	
	-dst string
		Путь к директории для выгрузки файлов. Пример: "C:\temp\dst" (default ".\\out")
	-findex int
	        Начальный номер ARJ-архива. Минимум 1. (default 1)
	-maxcount int
	        Максимальное количество файлов в архиве. Минимум 1. (default 50)
	-maxsize int
        	Максимальный размер файлов в архиве (в байтах). Минимум 10. (default 52428800)
	-script string
	        Путь расположения скрипта постобработки. Пример: "C:\temp\script" (default ".")
	-src string
	        Путь к директории для исходных файлов. Пример: "C:\temp\src"
	-signatura string
	        Путь установки СКЗИ Сигнатура. (default "C:\\Program Files\\MDPREI\\spki")

Пример использования:

	SET SCRIPT_DIR=E:\temp\fns\ARTANA
	E:
	CD %SCRIPT_DIR%
	del /s/q %SCRIPT_DIR%\ARJ
	cls
	artana.exe -src="%SCRIPT_DIR%\files" -dst="%SCRIPT_DIR%\OUT" -findex=1
	%SCRIPT_DIR%\ZDATA\mv.cmd
