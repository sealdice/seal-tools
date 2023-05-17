package main

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"runtime"
)

func recoveryWithGui(wd string) error {
	clearScreen()

	var archivedFiles []string
	var targetExt string
	if runtime.GOOS == "windows" {
		targetExt = ".zip"
	} else {
		targetExt = ".gz"
	}

	if err := filepath.Walk(wd, func(p string, i fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(p) == targetExt {
			archivedFiles = append(archivedFiles, p)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("搜索更新目标时出现错误\n%s\n", err)
	}

	if len(archivedFiles) == 0 {
		return fmt.Errorf("错误：没有发现任何更新")
	}

	clearScreen()

	var choice int
	fmt.Println("发现以下备份：")
	for i, b := range archivedFiles {
		fmt.Printf("[%d] %s",
			i+1, path.Base(b))
	}
	fmt.Printf("\n请输入对应数字进行选择（0–%d）", len(archivedFiles))

	_, err := fmt.Scanln(&choice)
	if err != nil {
		return fmt.Errorf("出现错误，请确认输入的是数字\n%s\n", err)
	}

	if choice > len(archivedFiles) || choice <= 0 {
		return fmt.Errorf("错误：无效序号")
	}

	fmt.Println(path.Join(wd, archivedFiles[choice-1]))

	return nil
}
