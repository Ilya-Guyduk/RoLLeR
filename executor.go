package main

import (
	"fmt"
	"strings"
)

func executeCheck(check Check) error {

	// Поиск локации выполнения чека
	host, err := findLocation(check)
	if err != nil {
		return fmt.Errorf("Check failed: %v", err)
	} else {
		// Дальнейшая работа с host, если проверки прошли
		logMessage("DEBUG", fmt.Sprintf("HostConfig найден: %+v", *host))
	}

	if DRY_RUN_FLAG {

		if check.Bash.User_script != "" {
			logMessage("INFO", fmt.Sprintf("Executing check script: %s", check.Bash.User_script))
		} else if check.Run != "" {
			logMessage("INFO", fmt.Sprintf("Executing command: %s", check.Run))
		}
		return nil // Не выполняем команду, только логируем
	}

	// Если флаг не установлен, выполняем команду как обычно
	if check.Bash.User_script != "" {
		logMessage("INFO", fmt.Sprintf("Executing check script: %s", check.Bash.User_script))
		return runCommand("bash", []string{check.Bash.User_script})
	} else if check.Run != "" {
		logMessage("INFO", fmt.Sprintf("Executing command: %s", check.Run))
		cmdParts := strings.Fields(check.Run)
		return runCommand(cmdParts[0], cmdParts[1:])
	}

	return nil
}
