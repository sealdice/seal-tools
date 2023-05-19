package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type VersionInfo struct {
	VersionLatest           string `json:"versionLatest"`
	VersionLatestDetail     string `json:"versionLatestDetail"`
	VersionLatestCode       int64  `json:"versionLatestCode"`
	VersionLatestNote       string `json:"versionLatestNote"`
	MinUpdateSupportVersion int64  `json:"minUpdateSupportVersion"`
	NewVersionUrlPrefix     string `json:"newVersionUrlPrefix"`
}

func recoveryWithGui() error {
	clearScreen()

	var archivedFiles []string

	if err := filepath.Walk(workingDirectory, func(p string, i fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(p) == UpdateExt {
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
	fmt.Println("发现以下可能的安装文件：")
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
	err = CheckUpdateAndInstall(UpdateExt, updatePath)
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
	var essentialFiles = []string{"sealdice-core.exe", "go-cqhttp/", "data/", "frontend/"}
	var missingFiles []string

	var markers = map[string]bool{}
	for _, fn := range essentialFiles {
		markers[fn] = false
	}

	for _, f := range z.File {
		if strings.Contains(f.Name, "..") {
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

func GetUpdateAndDownload() (string, error) {
	getVersionInfo := func() (*VersionInfo, error) {
		resp, err := http.Get("https://dice.weizaima.com/dice/api/version?versionCode=0")
		if err != nil {
			return nil, fmt.Errorf("与网站连接时出现错误\n%w", err)
		}
		defer resp.Body.Close()

		var ver VersionInfo
		err = json.NewDecoder(resp.Body).Decode(&ver)
		if err != nil {
			return nil, fmt.Errorf("解析响应时出现错误\n%w", err)
		}

		return &ver, nil
	}

	getFile := func(destin string, url string) error {
		resp, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("与网站连接时出现错误\n%w", err)
		}
		defer resp.Body.Close()

		op, err := os.Create(destin)
		if err != nil {
			return fmt.Errorf("创建目标文件时发生错误\n%w", err)
		}
		defer op.Close()

		_, err = io.Copy(op, resp.Body)
		if err != nil {
			return fmt.Errorf("复制文件时发生错误\n%w", err)
		}

		return nil
	}

	ver, err := getVersionInfo()
	if err != nil {
		return "", err
	}

	var arch string
	if runtime.GOARCH == "386" {
		arch = "i386"
	} else if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		arch = "amd64"
	} else {
		arch = runtime.GOARCH
	}

	fn := fmt.Sprintf("sealdice-core_%s_%s_%s.%s",
		ver.VersionLatest, runtime.GOOS, arch, UpdateExt)

	var fileUrl string
	if ver.NewVersionUrlPrefix != "" {
		fileUrl = ver.NewVersionUrlPrefix + "/" + fn
	} else {
		fileUrl = "https://sealdice.coding.net/p/sealdice/d/sealdice-binaries/git/raw/master/" + fn
	}

	final := filepath.Join(workingDirectory, fn)
	err = getFile(final, fileUrl)
	if err != nil {
		return "", err
	}

	return final, nil
}
