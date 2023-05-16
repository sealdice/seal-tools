package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
)

func Unzip(f *zip.File, destin string) error {
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
}
