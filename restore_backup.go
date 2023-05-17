package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

func backupRestoreWithGui(wd string) error {
	var backups []string
	var backupDir string

	if backupTarget == "" {
		backupDir = path.Join(wd, "backups")
		if _, err := os.Stat(backupDir); err != nil && errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("当前工作目录下没有 backups 文件夹")
		}

		if err := filepath.Walk(backupDir, func(p string, i fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if filepath.Ext(p) == ".zip" {
				backups = append(backups, p)
			}

			return nil
		}); err != nil {
			return fmt.Errorf("解析备份目录时出现错误\n%w", err)
		}

		if len(backups) == 0 {
			return fmt.Errorf("错误：没有发现任何备份")
		}

		clearScreen()

		var choice int
		fmt.Println("发现以下备份：")
		for i, b := range backups {
			fmt.Printf("[%d] %s\n",
				i+1, path.Base(b))
		}
		if len(backups) == 1 {
			fmt.Print("\n请输入对应数字进行选择（1）")
		} else {
			fmt.Printf("\n请输入对应数字进行选择（1–%d）", len(backups))
		}

		_, err := fmt.Scanln(&choice)
		if err != nil {
			return fmt.Errorf("出现错误，请确认输入的是数字\n%w", err)
		}

		if choice > len(backups) || choice <= 0 {
			return fmt.Errorf("错误：无效序号")
		}

		backupTarget = backups[choice-1]
	}

	err := TruncateRestore(backupTarget, wd, true)
	if err != nil {
		return fmt.Errorf("出现错误\n%w", err)
	} else {
		fmt.Println("备份覆盖成功！")
	}

	return nil
}
