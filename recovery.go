package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func recoveryWithGui() error {
	clearScreen()

	var archivedFiles []string
	var targetExt string
	if runtime.GOOS == "windows" {
		targetExt = ".zip"
	} else {
		targetExt = ".gz"
	}

	if err := filepath.Walk(workingDirectory, func(p string, i fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(p) == targetExt {
			archivedFiles = append(archivedFiles, p)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("搜索更新目标时出现错误\n%w", err)
	}

	if len(archivedFiles) == 0 {
		return fmt.Errorf("错误：没有发现任何更新")
	}

	clearScreen()

	var choice int
	fmt.Println("发现以下备份：")
	for i, b := range archivedFiles {
		fmt.Printf("[%d] %s\n", i+1, filepath.Base(b))
	}
	if len(archivedFiles) == 1 {
		fmt.Print("\n请输入对应数字进行选择（1）")
	} else {
		fmt.Printf("\n请输入对应数字进行选择（1–%d）", len(archivedFiles))
	}

	_, err := fmt.Scanln(&choice)
	if err != nil {
		return fmt.Errorf("出现错误，请确认输入的是数字\n%w", err)
	}

	if choice > len(archivedFiles) || choice <= 0 {
		return fmt.Errorf("错误：无效序号")
	}

	updatePath := filepath.Join(workingDirectory, archivedFiles[choice-1])
	err = CheckUpdateAndInstall(targetExt, updatePath)
	if err != nil {
		return err
	}

	fmt.Println("安装成功！请先实验骰子能否启动，然后（如果可能的话）恢复备份。")

	return nil
}

func CheckUpdateAndInstall(targetExt string, updatePath string) error {
	if targetExt == ".zip" {
		z, err := zip.OpenReader(updatePath)
		if err != nil {
			return fmt.Errorf("打开压缩文件时出现错误\n%w", err)
		}
		defer z.Close()

		m, err := checkUpdateValidZip(z)
		if err != nil {
			return fmt.Errorf("安装文件不完整，缺失以下文件或文件夹\n%s", strings.Join(m, " "))
		}

		err = trWithHandle(z, workingDirectory, true)
		if err != nil {
			return fmt.Errorf("解压缩时发生错误\n%w", err)
		}
	} else {
		osReader, err := os.Open(updatePath)
		if err != nil {
			return fmt.Errorf("打开压缩文件时出现错误\n%w", err)
		}
		defer osReader.Close()

		gzReader, err := gzip.NewReader(osReader)
		if err != nil {
			return fmt.Errorf("打开压缩文件时出现错误\n%w", err)
		}
		defer gzReader.Close()

		m, err := checkUpdateValidTarGz(gzReader)
		if err != nil {
			if m != nil {
				return fmt.Errorf("安装文件不完整，缺失以下文件或文件夹\n%s", strings.Join(m, " "))
			} else {
				return fmt.Errorf("遍历压缩文件时出现以下错误\n%w", err)
			}
		}

		// 上个循环中迭代器已经到头了，所以要重新来过
		_, err = osReader.Seek(0, io.SeekStart)
		if err != nil {
			return fmt.Errorf("重置迭代器时出现问题\n%w", err)
		}
		err = gzReader.Reset(osReader)
		if err != nil {
			return fmt.Errorf("重置迭代器时出现问题\n%w", err)
		}

		err = trWithHandleGz(osReader, gzReader, workingDirectory, true)
		if err != nil {
			return fmt.Errorf("解压缩时发生错误\n%w", err)
		}
	}
	return nil
}

func checkUpdateValidZip(z *zip.ReadCloser) ([]string, error) {
	var essentialFiles = []string{"sealdice-core", "go-cqhttp/", "data/", "frontend/"}
	var missingFiles []string

	for _, f := range z.File {
		if strings.Contains(f.Name, "..") {
			continue
		}

		for _, fn := range essentialFiles {
			found := false
			if filepath.Join(filepath.Base(f.Name), fn) == f.Name {
				found = true
			}

			if !found {
				missingFiles = append(missingFiles, fn)
			}
		}
	}

	if len(missingFiles) > 0 {
		return missingFiles, fmt.Errorf("缺少必要文件")
	}

	return nil, nil
}

func checkUpdateValidTarGz(g *gzip.Reader) ([]string, error) {
	var essentialFiles = []string{"sealdice-core", "go-cqhttp/", "data/", "frontend/"}
	var missingFiles []string

	var markers = map[string]bool{}
	for _, fn := range essentialFiles {
		markers[fn] = false
	}

	tr := tar.NewReader(g)
	for {
		f, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		} else if strings.Contains(f.Name, "..") {
			continue
		}

		for _, fn := range essentialFiles {
			if strings.Contains(f.Name, fn) {
				markers[fn] = true
			}
		}
	}

	for fn, exist := range markers {
		if !exist {
			missingFiles = append(missingFiles, fn)
		}
	}

	if len(missingFiles) > 0 {
		return missingFiles, fmt.Errorf("缺少必要文件")
	}

	return nil, nil
}
