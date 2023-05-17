package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func TruncateRestoreGz(gzPath string, destin string, confirmation bool, exclude ...string) error {
	f, err := os.Open(gzPath)
	if err != nil {
		return fmt.Errorf("打开压缩文件时出现错误\n%w", err)
	}
	defer f.Close()

	g, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("打开压缩文件时出现错误\n%w", err)
	}
	defer g.Close()

	return trWithHandleGz(f, g, destin, confirmation, exclude...)
}

func trWithHandleGz(of *os.File, g *gzip.Reader, destin string, confirmation bool, exclude ...string) error {
	var excPaths []string
	excPaths = append(excPaths, exclude...)

	tr := tar.NewReader(g)

	if confirmation {
		var duplicate []string
		var excluded []string
		for {
			f, err := tr.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			} else if strings.Contains(f.Name, "..") {
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

		_, err := of.Seek(0, io.SeekStart)
		if err != nil {
			return fmt.Errorf("重置迭代器时出现问题\n%w", err)
		}
		err = g.Reset(of)
		if err != nil {
			return fmt.Errorf("重置迭代器时出现问题\n%w", err)
		}

		tr = tar.NewReader(g)
	}

	for {
		f, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		} else if strings.Contains(f.Name, "..") {
			continue
		}

		if len(excPaths) > 0 {
			for _, ep := range excPaths {
				if ep != filepath.Join(destin, f.Name) {
					fmt.Println(f.Name)
					if err := Ungz(tr, f, destin); err != nil {
						return err
					}
				}
			}
		} else {
			if err := Ungz(tr, f, destin); err != nil {
				return err
			}
		}
	}

	return nil
}

func Ungz(r *tar.Reader, f *tar.Header, destin string) error {
	destPath := filepath.Join(destin, f.Name)
	if f.FileInfo().IsDir() {
		//TODO: 这里和 100、106 行的权限再好好斟酌一下
		err := os.MkdirAll(destPath, 0755)
		if err != nil {
			return fmt.Errorf("创建目标目录`%s`时出错：%w", destPath, err)
		}
	} else {
		err := os.MkdirAll(filepath.Dir(destPath), 0755)
		if err != nil {
			return fmt.Errorf("创建目标目录`%s`时出错：%w", destPath, err)
		}

		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(f.Mode))
		if err != nil {
			return fmt.Errorf("准备复制文件`%s`时出错：%w", filepath.Base(destPath), err)
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, r)
		if err != nil {
			return fmt.Errorf("复制文件`%s`时出错：%w", destPath, err)
		}
	}

	return nil
}
