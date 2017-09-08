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
)

type Config struct {
	SrcPath  string // Путь к файлам для обработки
	DstPath  string // Путь для выгрузки подготовленных файлов
	MaxSize  int64  // Максимальный размер файлов в архиве (50 МБ = 52428800 Байт)
	MaxCount int    // Максимальное количество файлов в архиве
	jsonFile string
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
	fmt.Println("Artana v 1.0 (C) 2017 UMK BANK")
	fmt.Print("Утилита группировки файлов для архива 440-П. ")
	fmt.Println("Справка по параметрам: -? или -h")
	//	var fPathSrc = flag.String("src", "E:\\temp\\fns\\test", "Путь к директории для исходных файлов. Пример: \"C:\\temp\\src\"")
	//	var fPathDst = flag.String("dst", "E:\\temp\\fns\\out", "Путь к директории для выгрузки файлов. Пример: \"C:\\temp\\dst\"")
	var fPathSrc = flag.String("src", ".", "Путь к директории для исходных файлов. Пример: \"C:\\temp\\src\"")
	var fPathDst = flag.String("dst", ".\\out", "Путь к директории для выгрузки файлов. Пример: \"C:\\temp\\dst\"")

	var fMaxSize = flag.Int64("maxsize", 52428800, "Максимальный размер файлов в архиве")
	var fMaxCount = flag.Int("maxcount", 50, "Максимальное количество файлов в архиве")
	flag.Parse()

	if strings.Compare(*fPathSrc, "") == 0 || strings.Compare(*fPathDst, "") == 0 || *fMaxSize <= 10 || *fMaxCount <= 1 {
		fmt.Println("Ошибка при указании параметров:")
		fmt.Println("SRC:", *fPathSrc)
		fmt.Println("DST:", *fPathDst)
		fmt.Println("MaxSize:", *fMaxSize)
		fmt.Println("MaxCount:", *fMaxCount)
		os.Exit(2)
	}

	// Если директории не существует
	if _, err := os.Stat(*fPathSrc); os.IsNotExist(err) {
		log.Fatal("Ошибка директории исходных файлов! (", *fPathSrc, ")")
	}

	// Если директории не существует
	/*if _, err := os.Stat(*fPathDst); os.IsNotExist(err) {
		log.Fatal("Ошибка директории исходных файлов! (", *fPathDst,")")
	}*/

	var cfg = Config{
		SrcPath:  *fPathSrc,
		DstPath:  *fPathDst,
		MaxSize:  *fMaxSize,
		MaxCount: *fMaxCount,
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

	GenScript()
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
func GenScript() error {
	//start /wait FcolseOW.exe /@%vrb%\311-unsgn.scr

	cryptosc := `
; Установить получателей файла
To 2001941009

; Зашифровать все файлы по маске
Crypt {{.Path}}\BNP*.xml
Start

; Завершить работу программы
Exit`

	// Подготовим данные для шаблона
	type Folder struct {
		Path string
	}
	var folders []Folder
	for _, r := range dirs {
		folders = append(folders, Folder{Path: r})
	}

	// Шаблон для подстановки данных о пути к файлам в скрипт
	t := template.Must(template.New("CryptScript").Parse(cryptosc))

	for _, r := range folders {
		err := t.Execute(os.Stdout, r)
		if err != nil {
			log.Println("Ошибка обрабтки шаблона CryptScript:", err)
		}
	}

	fmt.Println(dirs)

	return nil
}
