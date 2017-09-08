package artanasub

import (
	"path/filepath"
	"log"
	"strings"
	"os"
)

//Выполняет поиск файлов в каталоге согласно списка масок
func FindFiles(dir string, mask []string) (files map[string]string) {
	var err error
	var list []string
	files = make(map[string]string)

	for i := range mask {
		list, err = filepath.Glob(dir + "\\" + strings.ToUpper(mask[i]))
		if err != nil {
			log.Println("findFiles error: ", err)
			return nil
		}
		//files = append(files, list...)
		for _, f := range list {
			files[f] = mask[i]
		}
	}
	for i := range mask {
		list, err = filepath.Glob(dir + "\\" + strings.ToLower(mask[i]))
		if err != nil {
			log.Println("findFiles error: ", err)
			return nil
		}
		//files = append(files, list...)
		for _, f := range list {
			files[f] = mask[i]
		}
	}
	// Удаляем дубликаты из результата
	// return Dedup(files)
	return files
}

func GetFileSize(path string) int64 {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	return info.Size()
}

// Создает директорию для архива
func MakeDir(path string) error {
	dir := path
	// Если директория существует то на выход
	if _, err := os.Stat(dir); os.IsExist(err) {
		return err
	}
	// Иначе - Создаем
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	return nil
}