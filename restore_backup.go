package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

func execBackupRestore(wd string) error {
	var backups []string
	backupDir := path.Join(wd, "backups")

	if err := filepath.Walk(backupDir, func(p string, i fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(p) == ".zip" {
			backups = append(backups, p)
		}

		return nil
	}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"解析备份目录时出现错误\n%s\n", err)
		exitGracefully(1)
	}

	if len(backups) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "错误：没有发现任何备份")
		exitGracefully(1)
	}

	clearScreen()

	var choice int
	fmt.Println("发现以下备份：")
	for i, b := range backups {
		fmt.Printf("[%d] %s",
			i+1, path.Base(b))
	}
	fmt.Printf("\n请输入对应数字进行选择（0–%d）", len(backups))

	_, err := fmt.Scanln(&choice)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"出现错误\n%s\n", err)
		exitGracefully(1)
	}

	if choice > len(backups) || choice <= 0 {
		_, _ = fmt.Fprintln(os.Stderr, "错误：无效序号")
		exitGracefully(1)
	} else {
		err = RestoreBackup(backups[choice-1], wd)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr,
				"出现错误\n%s\n", err)
			exitGracefully(1)
		} else {
			fmt.Println("备份覆盖成功！")
			exitGracefully(0)
		}
	}

	return nil
}

func RestoreBackup(backup string, destin string) error {
	// 打开压缩包
	z, err := zip.OpenReader(backup)
	if err != nil {
		return fmt.Errorf("不能打开压缩文件：%w", err)
	}
	defer z.Close()

	for _, f := range z.File {
		// 不要在 for 循环里直接使用 defer
		// 出错提前 return 了没关系，但如果没出错，每一次循环的文件都会开着直到循环全部结束
		// 所以，最好包装在函数里
		if err = func(f *zip.File) error {
			sf, err := f.Open()
			if err != nil {
				return fmt.Errorf("解压缩`%s`时出错：%w", f.Name, err)
			}
			defer sf.Close()

			destPath := path.Join(destin, f.Name)
			if f.FileInfo().IsDir() {
				//TODO: 这里和 100、106 行的权限再好好斟酌一下
				err = os.MkdirAll(destPath, 0750)
				if err != nil {
					return fmt.Errorf("创建目标目录`%s`时出错：%w", destPath, err)
				}
			} else {
				err = os.MkdirAll(path.Dir(destPath), 0750)
				if err != nil {
					return fmt.Errorf("创建目标目录`%s`时出错：%w", destPath, err)
				}
			}

			destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fmt.Errorf("准备复制文件`%s`时出错：%w", path.Base(destPath), err)
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, sf)
			if err != nil {
				return fmt.Errorf("复制文件`%s`时出错：%w", destPath, err)
			}

			return nil
		}(f); err != nil {
			return err
		}
	}
	return nil
}
