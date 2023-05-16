package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

const (
	AppName string = "SealTools"
	Version string = "0.0.1"
)

var (
	isIntegrated bool
	backupTarget string
)

func main() {
	flag.BoolVar(&isIntegrated, "i", false, "Is invoked from SealDice")
	flag.StringVar(&backupTarget, "t", "", "The backup to restore, in full path")
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"无法获取当前目录，请检查：\n%s\n%s\n%s\n",
			"1. 程序是否权限足够\n", "2. 根据错误提示排查问题", err)
		exitGracefully(1)
	}

	if isIntegrated {
		if backupTarget == "" {
			_, _ = fmt.Fprintln(os.Stderr, "错误：没有提供要恢复的备份")
			exitGracefully(1)
		}

		err = restoreBackup(backupTarget, wd)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "发生错误\n%s\n", err)
			exitGracefully(1)
		}
		exitGracefully(0)
	}

	fmt.Printf("%s%s (%s) by 檀轶步棋%s\n",
		strings.Repeat("=", 8), AppName, Version, strings.Repeat("=", 8))

	f, ok := checkSealValid(wd)
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr,
			"当前目录不完整，缺失以下文件或文件夹：\n%s\n%s\n",
			strings.Join(f, " "),
			"您是否已经将本程序放在了海豹核心的安装目录下（和 sealdice-core 等文件同个目录）？")
		exitGracefully(1)
	}

	var choice int
	fmt.Printf("请选择功能：\n%s\n%s\n",
		"[0] 退出程序\n[1] 恢复备份", "请输入对应数字进行选择（0–1）")

	_, err = fmt.Scanln(&choice)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"出现错误，请检查您输入的是否是数字\n%s\n", err)
		exitGracefully(1)
	}

	switch choice {
	case 1:
		err = ExecBackupRestore(wd)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			exitGracefully(1)
		}
	case 114514:
		// 呃，彩蛋？
		clearScreen()
		fmt.Println("无奈的请伸出手　在梦里等斑驳轻舟")
		exitGracefully(25)
	default:
		exitGracefully(0)
	}

	exitGracefully(0)
}

// Windows 命令行程序执行完后会直接关掉，改成按任意键退出
// 又：由 SealDice 调起时不需要手动退出
func exitGracefully(code int) {
	if runtime.GOOS == "windows" && !isIntegrated {
		fmt.Println("按任意键退出程序…")
		r := bufio.NewReader(os.Stdin)
		_, _, _ = r.ReadRune()
	}

	os.Exit(code)
}

func checkSealValid(wd string) ([]string, bool) {
	var essentialFiles = []string{"sealdice-core", "go-cqhttp", "data", "backups", "frontend"}
	var missingFiles []string

	for _, f := range essentialFiles {
		fp := path.Join(wd, f)
		if _, err := os.Stat(fp); err != nil && errors.Is(err, os.ErrNotExist) {
			missingFiles = append(missingFiles, f)
		}
	}

	if len(missingFiles) > 0 {
		return missingFiles, false
	} else {
		return nil, true
	}
}

func clearScreen() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}
