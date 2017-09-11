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
	"flag"
	"fmt"
	"artana/artanasub"
	"log"
	"os"
	"strings"
	"path/filepath"
	"time"
	"strconv"
	"io"
	"path"
	"html/template"
	"io/ioutil"
	"golang.org/x/text/transform"
	"golang.org/x/text/encoding/charmap"
	"bytes"
)

type Config struct {
	SrcPath    string // Путь к файлам для обработки
	DstPath    string // Путь для выгрузки подготовленных файлов
	MaxSize    int64  // Максимальный размер файлов в архиве (50 МБ = 52428800 Байт)
	MaxCount   int    // Максимальное количество файлов в архиве
	VerbaPath  string // Путь к бинарным файлам verba-ow
	ScriptPath string // Путь к расположению скрипта постобработки
	jsonFile   string
}

type File440 struct {
	File string // Полное имя файла (включает путь)
	Part string // Номер запроса (текстовый)
	Size int64  // Размер файла
	Grp  int    // Номер для внутренней групиировки по номеру запроса
	Done bool   // признак обработки файла
}

var grpidx = 1
var dirs []string

func main() {
	fmt.Println("Artana v 1.0a (C) 2017 UMK BANK")
	fmt.Println("Утилита группировки файлов для архива 440-П. (Помощь: artana.exe -h)")
	fmt.Println("\nПример запуска:")
	fmt.Println("artana.exe -src=\"C:\\temp\\src\" -dst=\"C:\\temp\\out\" -verba=\"C:\\Program files\\MDPREI\\Verba-OW\"")

	//	var fPathSrc = flag.String("src", "E:\\temp\\fns\\test", "Путь к директории для исходных файлов. Пример: \"C:\\temp\\src\"")
	//	var fPathDst = flag.String("dst", "E:\\temp\\fns\\out", "Путь к директории для выгрузки файлов. Пример: \"C:\\temp\\dst\"")
	var fPathSrc = flag.String("src", "", "Путь к директории для исходных файлов. Пример: \"C:\\temp\\src\"")
	var fPathDst = flag.String("dst", ".\\out", "Путь к директории для выгрузки файлов. Пример: \"C:\\temp\\dst\"")

	var fMaxSize = flag.Int64("maxsize", 52428800, "Максимальный размер файлов в архиве (в байтах). Минимум 10.")
	var fMaxCount = flag.Int("maxcount", 50, "Максимальное количество файлов в архиве. Минимум 1.")
	var fPathVerba = flag.String("verba", "C:\\Program files\\MDPREI\\Verba-OW", "Путь установки Verba-OW. Пример: \"C:\\Program files\\MDPREI\\Verba-OW")
	var fPathScript = flag.String("script", ".", "Путь расположения скрипта постобработки. Пример: \"C:\\temp\\script\"")
	flag.Parse()

	if strings.Compare(*fPathSrc, "") == 0 || strings.Compare(*fPathDst, "") == 0 {
		fmt.Println("Ошибка указания исходной директории или директории выгрузки:")
		fmt.Println("SRC:", *fPathSrc)
		fmt.Println("DST:", *fPathDst)
		os.Exit(2)
	}

	if *fMaxSize <= 10 || *fMaxCount <= 1 {
		fmt.Println("Ошибка указания параметров обработки:")
		fmt.Println("MaxSize:", *fMaxSize)
		fmt.Println("MaxCount:", *fMaxCount)
		os.Exit(2)
	}

	if strings.Compare(*fPathSrc, "") == 0 || strings.Compare(*fPathDst, "") == 0 || *fMaxSize <= 10 || *fMaxCount <= 1 {
		fmt.Println("Ошибка при указании параметров:")
		fmt.Println("SRC:", *fPathSrc)
		fmt.Println("DST:", *fPathDst)
		fmt.Println("MaxSize:", *fMaxSize)
		fmt.Println("MaxCount:", *fMaxCount)
		fmt.Println("Verba:", *fPathVerba)
		fmt.Println("Script:", *fPathScript)

		os.Exit(2)
	}

	// Если директории не существует
	if _, err := os.Stat(*fPathSrc); os.IsNotExist(err) {
		log.Fatal("Неверно указана директория исходных файлов! (", *fPathSrc, ")")
	}

	// Если директории не существует
	if _, err := os.Stat(*fPathVerba); os.IsNotExist(err) {
		log.Fatal("Неверно указана директория установки Verba-OW! (", *fPathVerba, ")")
	}

	var cfg = Config{
		SrcPath:    *fPathSrc,
		DstPath:    *fPathDst,
		MaxSize:    *fMaxSize,
		MaxCount:   *fMaxCount,
		VerbaPath:  *fPathVerba,
		ScriptPath: *fPathScript,
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
	if strings.Compare(*fPathScript, ".") == 0 {
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
		cfg.ScriptPath = dir
	}

	// Если директории не существует
	if _, err := os.Stat(cfg.ScriptPath); os.IsNotExist(err) {
		log.Fatal("Неверно указана директория скрипта постобработки! (", cfg.ScriptPath, ")")
	}


	var file440 File440
	var files440 []File440

	parts := make(map[string]string)

	var masks = []string{"*.xml"}
	fmt.Println("\nБудут обработаны файлы по следующей маске:", masks, "\n")
	fmt.Println("Источник: \"", *fPathSrc, "\"")
	fmt.Println("Назначение: \"", *fPathDst, "\"")

	files := artanasub.FindFiles(cfg.SrcPath, masks)

	// Выбираем номера запросов из имени сообщений для группировки
	for k := range files {
		if strings.Contains(k, "BVS") {
			parts[k] = filepath.Base(k)[5:36] // только та часть файла где указано имя запроса ZSV
		}
	}

	// Группируем файлы в массиве
	grpcnt := 0
	for _, v := range parts {
		for k := range files {
			if strings.Contains(k, v) {
				file440 = File440{File: k, Part: v, Size: artanasub.GetFileSize(k), Grp: grpcnt, Done: false,}
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
		if size > cfg.MaxSize || count > cfg.MaxCount {
			//fmt.Println("DEBUG:", size, count)
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
	fmt.Println(" ГОТОВО!")

	cfg.GenScript()
}

// Выгружает данные группы файлов в папки с учетом ограничений maxcount maxsize
func Unload(files440 []File440, group int, cfg Config, mkdir bool, sizesum int64, count int) (int64, int, error) {
	if mkdir {
		grpidx += 1
	}

	// Выгружаем сначала BVS файл
	for i := range files440 {
		if files440[i].Grp == group && !files440[i].Done && strings.Contains(path.Base(files440[i].File), "BVS") {

			if err := cfg.MakeCopy(files440[i].File, grpidx, mkdir); err != nil {
				log.Printf("Unload: MakeCopy: %v\n", err)
				return 0, 0, err
			}

			files440[i].Done = true
			mkdir = false
			sizesum += files440[i].Size
			count++
		}
	}

	// Выгружаем остальные файлы
	for i := range files440 {
		if files440[i].Grp == group && !files440[i].Done {

			//Если файлов слишком много или размер привышает лимит
			if sizesum >= cfg.MaxSize || count >= cfg.MaxCount {
				grpidx += 1
				sizesum = 0
				count = 0
				mkdir = true
			}

			if err := cfg.MakeCopy(files440[i].File, grpidx, mkdir); err != nil {
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
	//start /wait FcolseOW.exe /@%vrb%\311-unsgn.scr

	renamesc := `@echo off
SET VERBA_PATH="{{.VerbaPath}}"
SET SCRIPT_DIR="{{.ScriptPath}}"
{{.ScriptDrive}}:
CD %SCRIPT_DIR%\FILES

	{{range .Paths}}
rename {{ . }}\PB*.xml PB*.bak
rename {{ . }}\*.xml *.vrb
rename {{ . }}\PB*.bak PB*.xml
{{end}}

start /wait FcolseOW.exe /@%VERBA_PATH%\sign440.scr
start /wait FcolseOW.exe /@%VERBA_PATH%\crypt440.scr

exit`

	signsc := `; Подписать все файлы по маске
{{range .Paths}}
	Sign {{ . }}\*.*
{{end}}
Start

; Завершить работу программы
Exit`

	cryptosc := `; Установить получателей файла
To 2001941009

; Зашифровать все файлы по маске
{{range .Paths}}
	Crypt {{ . }}\BNP*.vrb
{{end}}
Start

; Завершить работу программы
Exit`

	// Подготовим данные для шаблона
	type Folders struct {
		Paths       []string
		VerbaPath   string
		ScriptPath  string
		ScriptDrive string
	}

	var folders Folders
	var strbuffer bytes.Buffer // Используем буфер для конвертации шаблона в CP866 перед выводом в файл
	for _, r := range dirs {
		folders.Paths = append(folders.Paths, r)
	}
	folders.ScriptPath = c.ScriptPath
	folders.VerbaPath = c.VerbaPath
	folders.ScriptDrive = c.ScriptPath[:1]

	// Шаблон для подстановки данных о пути к файлам в скрипт
	trename := template.Must(template.New("CryptScript").Parse(renamesc))

	// Шаблон для подстановки данных о пути к файлам в скрипт
	tsign := template.Must(template.New("CryptScript").Parse(signsc))

	// Шаблон для подстановки данных о пути к файлам в скрипт
	tcrypt := template.Must(template.New("CryptScript").Parse(cryptosc))

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

	buf, err := ToCP866(strbuffer.String());
	if err != nil {
		log.Printf("Ошибка конвертации CP866: %v\n", err)
		return err
	}
	frename.WriteString(buf)
	strbuffer.Reset()
	/**/

	/* fsign */
	fsign, err := os.Create("sign440.sc")
	defer fsign.Close()
	if err != nil {
		log.Printf("Ошибка создания файла скрипта: %v\n", err)
		return err
	}

	if err := tsign.Execute(&strbuffer, folders); err != nil {
		log.Println("Ошибка обрабтки шаблона rename:", err)
	}

	buf, err = ToCP866(strbuffer.String());
	if err != nil {
		log.Printf("Ошибка конвертации CP866: %v\n", err)
		return err
	}
	fsign.WriteString(buf)
	strbuffer.Reset()
	/**/

	/* fcrypt */
	fcrypt, err := os.Create("crypt440.sc")
	defer fcrypt.Close()
	if err != nil {
		log.Printf("Ошибка создания файла скрипта: %v\n", err)
		return err
	}

	if err := tcrypt.Execute(&strbuffer, folders); err != nil {
		log.Println("Ошибка обрабтки шаблона rename:", err)
	}

	buf, err = ToCP866(strbuffer.String());
	if err != nil {
		log.Printf("Ошибка конвертации CP866: %v\n", err)
		return err
	}
	fcrypt.WriteString(buf)
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
