package path

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const gofileEnv = "GOFILE"

func GetInFilepath(fname string) string {
	if fname != "" {
		return fname
	}
	// default case
	sourceCodeFile := os.Getenv(gofileEnv)
	if sourceCodeFile == "" {
		return ""
	}
	dname, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(dname, sourceCodeFile)
}

func GetOutFilepath(fname string) string {
	dname, err := os.Getwd()
	if err != nil {
		return ""
	}

	if fname != "" {
		dir := filepath.Dir(fname)
		if dir == "." {
			return filepath.Join(dname, fname)
		}
		return fname
	}

	sourceCodeFile := os.Getenv(gofileEnv)

	const testSuffix = "_ai_test.go"
	const ext = ".go"

	sourceCodeFile = strings.TrimSuffix(sourceCodeFile, testSuffix) + testSuffix

	return filepath.Join(dname, sourceCodeFile)
}

// findRepoRoot ищет корень репозитория, двигаясь вверх от текущей директории
// и находя go.mod.
func GetRepoRoot() (string, error) {
	// Получаем текущую рабочую директорию, откуда был запущен go:generate
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Начинаем поиск с текущей директории
	dir := wd
	for {
		// Проверяем, есть ли go.mod в текущей директории
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Нашли! Возвращаем этот путь.
			return dir, nil
		}

		// Поднимаемся на один уровень вверх
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// Достигли корня файловой системы (/), но ничего не нашли
			return "", errors.New("не удалось найти корень репозитория (go.mod не найден)")
		}
		dir = parentDir
	}
}
