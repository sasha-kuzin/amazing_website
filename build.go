package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	// Очищаем папку bin перед сборкой
	clearBin()

	// Сборка проекта
	buildProject()

	// Копирование данных weatherdata
	copyWeatherData()

	fmt.Println("Сборка и копирование завершены успешно.")
}

// clearBin удаляет все файлы из папки bin
func clearBin() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "del", "/Q", "bin\\*")
	} else {
		cmd = exec.Command("rm", "-rf", "bin/*")
	}
	runCommand(cmd, "Ошибка при очистке папки bin")
}

// buildProject выполняет сборку проекта в зависимости от ОС
func buildProject() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("go", "build", "-o", "bin/amazing_website.exe", "./cmd/amazing_website")
	} else {
		cmd = exec.Command("go", "build", "-o", "bin/amazing_website", "./cmd/amazing_website")
	}
	runCommand(cmd, "Ошибка при сборке проекта")
}

// copyWeatherData копирует папку weatherdata в bin
func copyWeatherData() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", "xcopy", "/E", "/I", "/Y", ".\\internal\\weather\\weatherdata\\*", ".\\bin\\weatherdata\\")
	} else {
		cmd = exec.Command("cp", "-r", "./internal/weather/weatherdata", "./bin/")
	}
	runCommand(cmd, "Ошибка при копировании папки")
}

// runCommand выполняет команду и обрабатывает ошибки
func runCommand(cmd *exec.Cmd, errorMessage string) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(errorMessage+":", err)
		os.Exit(1)
	}
}
