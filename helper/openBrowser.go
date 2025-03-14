package helper

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// Указываем путь к вашему HTML файлу
	http.ServeFile(w, r, "./index.html")
}

// Функция для открытия браузера
func OpenBrowser(addr string) {
	var cmd *exec.Cmd

	// В зависимости от операционной системы открываем браузер
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", "http://localhost"+addr)
	case "linux":
		cmd = exec.Command("xdg-open", "http://localhost"+addr)
	case "windows":
		cmd = exec.Command("cmd", "/C", "start", "http://localhost"+addr)
	default:
		fmt.Println("Неизвестная операционная система. Открытие браузера не поддерживается.")
		return
	}

	// Выполнение команды
	err := cmd.Start()
	if err != nil {
		fmt.Println("Ошибка при попытке открыть браузер:", err)
	}
}
