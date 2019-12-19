/*
  	Archive katana (artana)
	Сортировщик файлов для 440-П

	Файлы формируются с помощью программы архиватора ARJ32.
	Каждый Архивный файл содержит не более 50 файлов и имеет размер не более 50 мб.

  	Наименование Архивного файла имеет следующую структуру:
	<AFN_SSSSSSS_RRRRRRR_ГГГГММДД_NNNNN.arj>
*/
package main

import (
	"artana/artanasub"
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type Config struct {
	SrcPath    string // Путь к файлам для обработки
	DstPath    string // Путь для выгрузки подготовленных файлов
	MaxSize    int64  // Максимальный размер файлов в архиве (50 МБ = 52428800 Байт)
	MaxCount   int    // Максимальное количество файлов в архиве
	SignPath   string // Путь к бинарным файлам Сигнатуры
	ScriptPath string // Путь к расположению скрипта постобработки
	KeyFSR     string // Ключ шифрования ФСР (номер получателя) для скрипта постобработки
	FileIndex  int    // Начальное значение счетчика файлов ARJ
	jsonFile   string
}

type File440 struct {
	File string // Полное имя файла (включает путь)
	Part string // Номер запроса (текстовый)
	Size int64  // Размер файла
	Grp  int    // Номер для внутренней групиировки по номеру запроса
	Done bool   // признак обработки файла
}

var GRPIDX = 0
var dirs []string

func main() {
	fmt.Println("Artana 1.10 (C) 2019 UMK BANK")

	fmt.Println("Утилита группировки файлов для архива 440-П. (Помощь: artana.exe -h)")
	fmt.Println("\nПример запуска:")
	fmt.Println("artana.exe -src=\"C:\\temp\\src\" -dst=\"C:\\temp\\out\" -signatura=\"C:\\Program Files\\MDPREI\\spki\"")

	//	var fPathSrc = flag.String("src", "E:\\temp\\fns\\test", "Путь к директории для исходных файлов. Пример: \"C:\\temp\\src\"")
	//	var fPathDst = flag.String("dst", "E:\\temp\\fns\\out", "Путь к директории для выгрузки файлов. Пример: \"C:\\temp\\dst\"")
	var fPathSrc = flag.String("src", "", "Путь к директории для исходных файлов. Пример: \"C:\\temp\\src\"")
	var fPathDst = flag.String("dst", ".\\out", "Путь к директории для выгрузки файлов. Пример: \"C:\\temp\\dst\"")
	var fPathSignatura = flag.String("signatura", "C:\\Program Files\\MDPREI\\spki", "Путь установки СКЗИ Сигнатура.")
	var fMaxSize = flag.Int64("maxsize", 52428800, "Максимальный размер файлов в архиве (в байтах). Минимум 10.")
	var fMaxCount = flag.Int("maxcount", 50, "Максимальное количество файлов в архиве. Минимум 1.")
	var fPathScript = flag.String("script", ".", "Путь расположения скрипта постобработки. Пример: \"C:\\temp\\script\"")
	var fFileIndex = flag.Int("findex", 1, "Начальный номер ARJ-архива. Минимум 1.")

	flag.Parse()

	if strings.Compare(*fPathSrc, "") == 0 || strings.Compare(*fPathDst, "") == 0 {
		fmt.Println("Ошибка указания исходной директории или директории выгрузки:")
		fmt.Println("SRC:", *fPathSrc)
		fmt.Println("DST:", *fPathDst)
		os.Exit(2)
	}

	if *fMaxSize <= 10 || *fMaxCount <= 1 || *fFileIndex < 1 {
		fmt.Println("Ошибка указания параметров обработки:")
		fmt.Println("MaxSize   (>10):", *fMaxSize)
		fmt.Println("MaxCount   (>1):", *fMaxCount)
		fmt.Println("FileIndex (>=1):", *fFileIndex)
		os.Exit(2)
	}

	// Если директории не существует
	if _, err := os.Stat(*fPathSrc); os.IsNotExist(err) {
		log.Fatal("Неверно указана директория исходных файлов! (", *fPathSrc, ")")
	}

        // Если директории не существует
	if _, err := os.Stat(*fPathSignatura); os.IsNotExist(err) {
		log.Fatal("Неверно указана директория установки Signatura! (", *fPathSignatura, ")")
	}


	var cfg = Config{
		SrcPath:    *fPathSrc,
		DstPath:    *fPathDst,
		MaxSize:    *fMaxSize,
		MaxCount:   *fMaxCount,
		SignPath:   *fPathSignatura,
		ScriptPath: *fPathScript,
		FileIndex:  *fFileIndex,
	}

	// Трансформируем путь вида . в реальный путь
	if strings.Compare(*fPathScript, ".") == 0 {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
		cfg.ScriptPath = dir
	}
	if strings.Compare(*fPathDst, ".") == 0 {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
		cfg.DstPath = dir
	}
	if strings.Compare(*fPathSrc, ".") == 0 {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
		cfg.SrcPath = dir
	}
	if strings.Compare(*fPathSignatura, ".") == 0 {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
		cfg.SignPath = dir
	}


	// Если директории не существует
	if _, err := os.Stat(cfg.ScriptPath); os.IsNotExist(err) {
		log.Fatal("Неверно указана директория скрипта постобработки! (", cfg.ScriptPath, ")")
	}

	// Удаляем все данные в выходной директории
	fmt.Print("\nОчищаем: ", cfg.ArcDirDstNow())
	if err := os.RemoveAll(cfg.ArcDirDstNow()); err != nil {
		log.Fatal("Ошибка при отчистке выходной директории! (", cfg.ArcDirDstNow(), ")")
	}
	fmt.Println(" - [Готово!]")

	var file440 File440
	var files440 []File440

	parts := make(map[string]string)

	var masks = []string{"*.xml","*.vrb"}
	fmt.Println("\nБудут обработаны файлы по следующей маске: ", masks, "\n")
	fmt.Println("Источник: \"", *fPathSrc, "\"")
	fmt.Println("Назначение: \"", *fPathDst, "\"")

	files := artanasub.FindFiles(cfg.SrcPath, masks)

	// Выбираем номера запросов из сообщений для группировки
	for k := range files {
		if strings.Contains(k, "BVS") {
			parts[filepath.Base(k)[8:36]] = k // только та часть файла где указано имя запроса ZSV
		}
		if strings.Contains(k, "BOS") {
			parts[filepath.Base(k)[8:36]] = k // только та часть файла где указано имя запроса
		}
		if strings.Contains(k, "PB") {
			parts[filepath.Base(k)[7:35]] = k // только та часть файла где указано имя запроса
		}

	}

	// Группируем файлы в массиве
	grpcnt := 0
	for k, _ := range parts {
		for r := range files {
			if strings.Contains(r, k) {
				file440 = File440{File: r, Part: k, Size: artanasub.GetFileSize(r), Grp: grpcnt, Done: false}
				files440 = append(files440, file440)
			}
		}
		grpcnt++
	}

	// Выгрузка групп
	mkdir := true
	var size int64 = 0
	var count int = 0

	fmt.Print("\nОБРАБОТКА: ")
	for g := 0; g < grpcnt; g++ {
		fmt.Print(".")
		if size >= cfg.MaxSize || count >= cfg.MaxCount {
			//fmt.Println("DEBUG:", size, count)
			// Елси предел достигнут - обнуляем счетчики
			// и выгружаем следующий файл в новый каталог
			size = 0
			count = 0
			s, c, err := Unload(files440, g, cfg, true, size, count)
			if err != nil {
				log.Fatal(err)
			}
			size = s
			count = c
			mkdir = false
			continue
		}
		s, c, err := Unload(files440, g, cfg, mkdir, size, count)
		if err != nil {
			log.Fatal(err)
		}
		if g == 0 {
			mkdir = false
		}
		size = s // собираем общий размер данных
		count = c
	}
	fmt.Println(" ГОТОВО! Файлов[", len(files440), "]")

	cfg.GenScript()
}

// Выгружает данные группы файлов в папки с учетом ограничений maxcount maxsize
func Unload(files440 []File440, group int, cfg Config, mkdir bool, sizesum int64, count int) (int64, int, error) {
	if mkdir {
		GRPIDX += 1
	}

	// Выгружаем в группе первым BVS файл
	for i := range files440 {
		if files440[i].Grp == group && !files440[i].Done && strings.Contains(path.Base(files440[i].File), "BVS") {

			if err := cfg.MakeCopy(files440[i].File, GRPIDX, mkdir); err != nil {
				log.Printf("Unload: MakeCopy: %v\n", err)
				return 0, 0, err
			}

			files440[i].Done = true
			mkdir = false
			sizesum += files440[i].Size
			count++
		}
	}

	// Выгружаем остальные файлы группы
	for i := range files440 {
		if files440[i].Grp == group && !files440[i].Done {

			//Если файлов слишком много или размер привышает лимит
			if sizesum >= cfg.MaxSize || count >= cfg.MaxCount {
				GRPIDX += 1
				sizesum = 0
				count = 0
				mkdir = true
			}

			if err := cfg.MakeCopy(files440[i].File, GRPIDX, mkdir); err != nil {
				log.Printf("Unload: MakeCopy: %v\n", err)
				return 0, 0, err
			}
			files440[i].Done = true
			mkdir = false

			sizesum += files440[i].Size
			count++
		}
	}

	return sizesum, count, nil
}

// Копирует указанный файл в архивную директорию
// file = полный путь включая имя файла
func (c *Config) MakeCopy(file string, grpidx int, mkdir bool) error {

	//Создаем директорию для выгрузки вида ГГГГ\ММ\ДД\0101
	if mkdir {
		if err := artanasub.MakeDir(c.ArcDirDstNow() + "\\" + fmt.Sprintf("%03d", grpidx)); err != nil {
			return err
		}
		// Сохраняем данные о созданных директориях для скрипта постобработки
		dirs = append(dirs, c.ArcDirDstNow()+"\\"+fmt.Sprintf("%03d", grpidx))
	}

	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		log.Printf("MakeCopy: Невозможно открыть файл: %v\n", err)
		return err
	}

	dstfilename := c.ArcDirDstNow() + "\\" + fmt.Sprintf("%03d", grpidx) + "\\" + filepath.Base(file)

	if _, err := os.Stat(dstfilename); os.IsExist(err) {
		// Файл существует и не будет перезаписан
		return err
	}

	df, err := os.Create(dstfilename)
	defer df.Close()

	if _, err = io.Copy(df, f); err != nil {
		return err
	}
	if err := df.Sync(); err != nil {
		return err
	}

	return err
}

// Возвращает директорию вида ГГГГ\ММ\ДД с учетом даты.
func (c *Config) ArcDirDstNow() string {
	date := time.Now()
	res := strconv.Itoa(date.Year())                     //ГГГГ
	res += "\\" + fmt.Sprintf("%02d", int(date.Month())) //ММ
	res += "\\" + fmt.Sprintf("%02d", date.Day())        //ДД
	return c.DstPath + "\\" + res
}

// Генерирует скрипт постобработки
func (c *Config) GenScript() error {
	renamesc := `@echo off
SET SIGNATURA_PATH="{{.SignPath}}"
SET PATH=%SIGNATURA_PATH%;%PATH%
set NOW=%date:~6,4%%date:~3,2%%date:~0,2%
SET SCRIPT_DIR="{{.ScriptPath}}"
SET ARJPATH=%SCRIPT_DIR%\..\ARJ
{{.ScriptDrive}}:

REM ADD TO ARJ ARCHIVE
CD %SCRIPT_DIR%\..\TMP

{{range $paths := .Paths}}arj32.exe a -e AFN_0349830_MIFNS00_%NOW%_{{ $paths.Number }}.ARJ {{ $paths.Path }}\*.*
{{end}}

REM ***** SETTING SIGN IN TMP FOLDER

@dir *.arj /b > list

for /F %%i in (list) do (
spki1utl.exe -registry -profile FOIV2019 -silent %LOG%\foiv-440p-sign.log  -sign -data %%i -out %ARJPATH%\%%i
SET erFile = %%i
IF NOT EXIST %%i GOTO ERR
@del %%i
)

ECHO АРХИВЫ БЫЛИ СОЗДАНЫ
ECHO АРХИВЫ БЫЛИ ПОДПИСАНЫ КА
ECHO ДО НОВЫХ ВСТРЕЧ!

pause
exit`

	type Path struct {
		Path   string
		Number string // номер для формирования ARJ архива
	}

	// Подготовим данные для шаблона
	type Folders struct {
		Paths       []Path
		SignPath    string
		ScriptPath  string
		ScriptDrive string
	}

	var folders Folders
	var path Path
	var strbuffer bytes.Buffer // Используем буфер для конвертации шаблона в CP866 перед выводом в файл
	for i, r := range dirs {
		path = Path{Path: r, Number: fmt.Sprintf("%05d", i+c.FileIndex)}
		folders.Paths = append(folders.Paths, path)
	}
	folders.ScriptPath = c.ScriptPath
	folders.ScriptDrive = c.ScriptPath[:1]
	folders.SignPath = c.SignPath

	// Шаблон для подстановки данных о пути к файлам в скрипт
	trename := template.Must(template.New("CryptScript").Parse(renamesc))

	/* frename */
	frename, err := os.Create("mv.cmd")
	defer frename.Close()
	if err != nil {
		log.Printf("Ошибка создания файла скрипта: %v\n", err)
		return err
	}

	if err := trename.Execute(&strbuffer, folders); err != nil {
		log.Println("Ошибка обрабтки шаблона rename:", err)
	}

	buf, err := ToCP866(strbuffer.String())
	if err != nil {
		log.Printf("Ошибка конвертации CP866: %v\n", err)
		return err
	}
	frename.WriteString(buf)
	strbuffer.Reset()
	/**/

	//fmt.Println(dirs)

	return nil
}

func ToCP866(str string) (string, error) {
	sr := strings.NewReader(str)
	tr := transform.NewReader(sr, charmap.CodePage866.NewEncoder())
	buf, err := ioutil.ReadAll(tr)
	if err != nil {
		return "", err
	}

	return string(buf), err
}
