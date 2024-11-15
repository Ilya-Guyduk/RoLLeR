package main

import (
	"fmt"
)

// printDescription выводит описание этапа, если оно есть.
func printDescription(description string) {
	if description != "" {
		logMessage("INFO", fmt.Sprintf("========== %s", description))
	}
}

// handleAtomicStage устанавливает флаг атомарного обновления, если он присутствует.
func handleAtomicStage(isAtomic bool) {
	if isAtomic {
		ATOMIC_STAGE = isAtomic
		logMessage("DEBUG", fmt.Sprintf("ATOMIC_STAGE: %v", ATOMIC_STAGE))
	}
}
