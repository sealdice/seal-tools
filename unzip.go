package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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

			d := filepath.Join(destin, f.Name)
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
			fmt.Printf("%s\n以上文件将被覆盖\n", strings.Join(duplicate, "\n"))

			if len(excluded) > 0 {
				fmt.Printf("%s\n以上文件将被跳过\n", strings.Join(excluded, "\n"))
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

		if len(excPaths) > 0 {
			for _, ep := range excPaths {
				if ep != filepath.Join(destin, f.Name) {
					// 不要在 for 循环里直接使用 defer
					// 出错提前 return 了没关系，但如果没出错，每一次循环的文件都会开着直到循环全部结束
					// 所以，最好包装在函数里
					fmt.Println(f.Name)
					if err := Unzip(f, destin); err != nil {
						return err
					}
				}
			}
		} else {
			if err := Unzip(f, destin); err != nil {
				return err
			}
		}
	}
	return nil
}

func Unzip(f *zip.File, destin string) error {
	sf, err := f.Open()
	if err != nil {
		return fmt.Errorf("解压缩`%s`时出错：%w", f.Name, err)
	}
	defer sf.Close()

	destPath := filepath.Join(destin, f.Name)
	if f.FileInfo().IsDir() {
		//TODO: 这里和 100、106 行的权限再好好斟酌一下
		err = os.MkdirAll(destPath, 0755)
		if err != nil {
			return fmt.Errorf("创建目标目录`%s`时出错：%w", destPath, err)
		}
	} else {
		err = os.MkdirAll(filepath.Dir(destPath), 0755)
		if err != nil {
			return fmt.Errorf("创建目标目录`%s`时出错：%w", destPath, err)
		}

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("准备复制文件`%s`时出错：%w", filepath.Base(destPath), err)
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, sf)
		if err != nil {
			return fmt.Errorf("复制文件`%s`时出错：%w", destPath, err)
		}
	}

	return nil
}
