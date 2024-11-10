package main

import (
	"fmt"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

// Stage представляет этап обработки с его параметрами.
type Stage struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"desc"`
	Dependence  interface{} `yaml:"dependence"`
	Atomic      bool        `yaml:"atomic"`
	PreCheck    Check       `yaml:"pre_check"`
	PreScript   Script      `yaml:"pre_script"`
	PostCheck   Check       `yaml:"post_check"`
	PostScript  Script      `yaml:"post_scriprt"`
	Rollback    bool        `yaml:"rollback"`
	Steps       []Step      `yaml:"step"`
}

var ATOMIC_STAGE bool

// processStage обрабатывает этап (Stage) с различными проверками и скриптами.
func processStage(stage Stage) error {
	fmt.Println("######################################################")
	logMessage("INFO", fmt.Sprintf("Start stage: %s", stage.Name))

	// Выводим описание, если оно есть
	printDescription(stage.Description)

	// Запуск прогресс-бара в горутине
	bar := startProgressBar(len(stage.Steps), stage.Name)
	defer bar.Close()

	// Обрабатываем атомарный флаг
	handleAtomicStage(stage.Atomic)

	// Выполняем предварительные действия
	if err := runPreActions(stage); err != nil {
		return err
	}

	// Обрабатываем шаги
	for _, step := range stage.Steps {
		if err := processStep(step); err != nil {
			return err
		}
		// Обновляем прогресс-бар после каждого шага
		bar.Add(1)
	}

	// Выполняем пост-действия
	if err := runPostActions(stage); err != nil {
		return err
	}

	return nil
}

// startProgressBar создает и запускает прогресс-бар для этапа.
func startProgressBar(totalSteps int, stageName string) *progressbar.ProgressBar {
	bar := progressbar.NewOptions(
		totalSteps,
		progressbar.OptionSetDescription(fmt.Sprintf("Processing: %s", stageName)),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionSetRenderBlankState(true),
	)
	go func() {
		for i := 0; i < totalSteps; i++ {
			// Пауза, имитирующая выполнение каждого шага (можно настроить под вашу логику)
			time.Sleep(500 * time.Millisecond)
		}
	}()
	fmt.Println()
	return bar
}
