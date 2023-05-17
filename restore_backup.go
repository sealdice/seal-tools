package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
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
			return fmt.Errorf("解析备份目录时出现错误\n%s\n", err)
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
		fmt.Printf("\n请输入对应数字进行选择（0–%d）", len(backups))

		_, err := fmt.Scanln(&choice)
		if err != nil {
			return fmt.Errorf("出现错误，请确认输入的是数字\n%s\n", err)
		}

		if choice > len(backups) || choice <= 0 {
			return fmt.Errorf("错误：无效序号")
		}

		backupTarget = backups[choice-1]
	}

	err := TruncateRestore(backupTarget, wd, true)
	if err != nil {
		return fmt.Errorf("出现错误\n%s\n", err)
	} else {
		fmt.Println("备份覆盖成功！")
	}

	return nil
}

func TruncateRestore(zipPath string, destin string, confirmation bool, exclude ...string) error {
	// 打开压缩包
	z, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("不能打开压缩文件：%w", err)
	}
	defer z.Close()

	return trWithHandle(z, destin, confirmation, exclude...)
}

func trWithHandle(z *zip.ReadCloser, destin string, confirm bool, exclude ...string) error {
	var excPaths []string
	excPaths = append(excPaths, exclude...)

	if confirm {
		var duplicate []string
		var excluded []string
		for _, f := range z.File {
			// 根据 GitHub 安全检查，这边要防止一下 Zip Slip
			if strings.Contains(f.Name, "..") {
				continue
			}
			
			d := path.Join(destin, f.Name)
			if _, err := os.Stat(d); err == nil || !errors.Is(err, os.ErrNotExist) {
				if len(excPaths) > 0 {
					for _, ep := range excPaths {
						if ep == d {
							excluded = append(excluded, d)
						} else {
							duplicate = append(duplicate, d)
						}
					}
				} else {
					duplicate = append(duplicate, d)
				}
			}
		}

		if len(duplicate) > 0 {
			var conf string
			fmt.Printf("以下文件将被覆盖：\n%s\n", strings.Join(duplicate, "\n"))

			if len(excluded) > 0 {
				fmt.Printf("以下文件将被跳过：\n%s\n", strings.Join(excluded, "\n"))
			}

			fmt.Print("您确认吗？(y/N)")
			_, err := fmt.Scanln(&conf)
			if err != nil {
				return err
			}

			if conf != "y" {
				return fmt.Errorf("操作已取消")
			}
		}
	}

	for _, f := range z.File {
		if strings.Contains(f.Name, "..") {
			continue
		}

		for _, ep := range excPaths {
			if ep != path.Join(destin, f.Name) {
				// 不要在 for 循环里直接使用 defer
				// 出错提前 return 了没关系，但如果没出错，每一次循环的文件都会开着直到循环全部结束
				// 所以，最好包装在函数里
				fmt.Println(f.Name)
				if err := Unzip(f, destin); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
