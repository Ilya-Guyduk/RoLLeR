package main

import (
	"fmt"
	"os/exec"
)

type Step struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"desc"`
	Dependence  interface{} `yaml:"dependence"`
	Location    Location    `yaml:"location"`
	Helm        HelmAction  `yaml:"helm"`
	Yum         YumAction   `yaml:"yum"`
	Atomic      bool        `yaml:"atomic"`
	PreCheck    Check       `yaml:"pre_check"`
	PreScript   Script      `yaml:"pre_script"`
	PostCheck   Check       `yaml:"post_check"`
	PostScript  Script      `yaml:"post_script"`
	Rollback    bool        `yaml:"rollback"`
}

type HelmAction struct {
	Actions     string `yaml:"actions"`
	HelmDir     string `yaml:"helm_dir"`
	ReleaseName string `yaml:"release_name"`
	Version     string `yaml:"version"`
}

type YumAction struct {
	Actions    string `yaml:"actions"`
	PacketName string `yaml:"packet_name"`
	Version    string `yaml:"version"`
}

func runCommand(command string, args []string) error {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	logMessage("DEBUG", fmt.Sprintf("Command output: %s", output))
	return err
}

// Функция для выполнения действия Helm
func executeHelmAction(helm HelmAction) error {
	logMessage("INFO", fmt.Sprintf("Executing Helm action: %s on release %s, version %s", helm.Actions, helm.ReleaseName, helm.Version))
	args := []string{helm.Actions, helm.ReleaseName, "--namespace", helm.HelmDir, "--version", helm.Version}
	return runCommand("helm", args)
}

// Функция для выполнения действия Yum
func executeYumAction(yum YumAction) error {
	logMessage("INFO", fmt.Sprintf("Executing Yum action: %s on package %s, version %s", yum.Actions, yum.PacketName, yum.Version))
	args := []string{yum.Actions, yum.PacketName + "-" + yum.Version}
	return runCommand("yum", args)
}

func processStep(step Step) error {
	fmt.Printf("__________________________________________________________________\n")
	logMessage("INFO", fmt.Sprintf("Processing step: %s", step.Name))
	if step.Description != "" {
		logMessage("INFO", fmt.Sprintf("Desc: %s", step.Description))
	}

	// Проверка наличия атомарного этапа
	if step.Atomic {
		ATOMIC_STAGE = step.Atomic
		logMessage("DEBUG", fmt.Sprintf("ATOMIC_STEP: %v", ATOMIC_STAGE))
	}

	// Проверка наличия пре-чека
	if (Check{}) != step.PreCheck {
		logMessage("INFO", fmt.Sprintf("Running pre-check for step: %s", step.Name))
		if err := executeCheck(step.PreCheck); err != nil {
			return fmt.Errorf("pre-check failed for step %s: %v", step.Name, err)
		}
	} else {
		logMessage("DEBUG", fmt.Sprintf("pre_check for step %s is missing", step.Name))
	}

	// Проверка наличия пре-скрипта
	if (Script{}) != step.PreScript {
		logMessage("INFO", fmt.Sprintf("Running pre-script for step: %s", step.Name))
		if err := executeCheck(step.PreCheck); err != nil {
			return fmt.Errorf("pre-script failed for step %s: %v", step.Name, err)
		}
	} else {
		logMessage("DEBUG", fmt.Sprintf("pre-script for step %s is missing", step.Name))
	}

	if step.Helm.ReleaseName != "" {
		if err := executeHelmAction(step.Helm); err != nil {
			return fmt.Errorf("helm action failed for step %s: %v", step.Name, err)
		}
	}

	if step.Yum.PacketName != "" {
		if err := executeYumAction(step.Yum); err != nil {
			return fmt.Errorf("yum action failed for step %s: %v", step.Name, err)
		}
	}

	logMessage("INFO", fmt.Sprintf("Running post-check for step: %s", step.Name))
	if err := executeCheck(step.PostCheck); err != nil {
		return fmt.Errorf("post-check failed for step %s: %v", step.Name, err)
	}
	return nil
}
