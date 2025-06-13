package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var targets = []string{
	"~/Library/Preferences",
	"~/Library/Caches",
	"~/Library/Logs",
	"~/Library/Application Support/JetBrains",
	"~/Applications",
	"~/Applications/JetBrains Toolbox",
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

func getPhpStormDirs() ([]string, error) {
	var found []string

	for _, base := range targets {
		basePath := expandPath(base)
		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			name := entry.Name()
			lower := strings.ToLower(name)
			if strings.Contains(lower, "phpstorm") {
				fullPath := filepath.Join(basePath, name)
				found = append(found, fullPath)
			}
		}
	}
	return found, nil
}

func getNewestPhpStormApp() string {
	appsPath := expandPath("~/Applications")
	entries, err := os.ReadDir(appsPath)
	if err != nil {
		return ""
	}

	var phpstormApps []struct {
		path string
		time time.Time
	}

	for _, entry := range entries {
		if strings.Contains(strings.ToLower(entry.Name()), "phpstorm") {
			fullPath := filepath.Join(appsPath, entry.Name())
			info, err := os.Stat(fullPath)
			if err == nil && info.IsDir() {
				phpstormApps = append(phpstormApps, struct {
					path string
					time time.Time
				}{path: fullPath, time: info.ModTime()})
			}
		}
	}

	if len(phpstormApps) == 0 {
		return ""
	}

	sort.Slice(phpstormApps, func(i, j int) bool {
		return phpstormApps[i].time.After(phpstormApps[j].time)
	})

	return phpstormApps[0].path
}

func confirm(prompt string) bool {
	fmt.Print(prompt + " [y/N]: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return input == "y" || input == "yes"
	}
	return false
}

func deletePaths(paths []string) {
	for _, path := range paths {
		err := os.RemoveAll(path)
		if err != nil {
			fmt.Printf("❌ Ошибка при удалении %s: %v\n", path, err)
		} else {
			fmt.Printf("✅ Удалено: %s\n", path)
		}
	}
}

func main() {
	fmt.Println("🔍 Поиск папок PhpStorm...")

	allDirs, err := getPhpStormDirs()
	if err != nil {
		fmt.Println("Ошибка при сканировании директорий.")
		return
	}

	if len(allDirs) == 0 {
		fmt.Println("🙈 Ничего не найдено.")
		return
	}

	currentApp := getNewestPhpStormApp()
	fmt.Printf("📌 Предположительно актуальная версия PhpStorm: %s\n", currentApp)

	var toDelete []string
	for _, path := range allDirs {
		if path == currentApp {
			continue
		}
		toDelete = append(toDelete, path)
	}

	if len(toDelete) == 0 {
		fmt.Println("✅ Нечего удалять — осталась только актуальная версия.")
		return
	}

	fmt.Println("\nБудут удалены:")
	for _, p := range toDelete {
		fmt.Println(" -", p)
	}

	if confirm("\nУдалить эти папки?") {
		deletePaths(toDelete)
	} else {
		fmt.Println("❎ Отмена.")
	}
}
