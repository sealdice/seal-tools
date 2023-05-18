package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	AppName string = "SealTools"
	Version string = "0.1.0"
)

var (
	isIntegrated     bool
	backupTarget     string
	workingDirectory string
	installedUpdate  string
)

func main() {
	flag.BoolVar(&isIntegrated, "i", false, "Is invoked from SealDice")
	flag.StringVar(&backupTarget, "t", "", "The backup to restore, in absolute path")
	flag.StringVar(&workingDirectory, "w", "./", "The path where the program will run on")
	flag.StringVar(&installedUpdate, "u", "", "The absolute path for SealDice update, if available")
	flag.Parse()

	if stat, err := os.Stat(workingDirectory); err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"指定的工作路径不合法\n%s\n", err)
		exitGracefully(1)
	} else if !stat.IsDir() {
		_, _ = fmt.Fprintf(os.Stderr, "错误：指定的工作路径不是一个文件夹\n")
		exitGracefully(1)
	}

	if isIntegrated {
		if backupTarget != "" {
			err := TruncateRestore(backupTarget, workingDirectory, false)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "发生错误\n%s\n", err)
				exitGracefully(1)
			}
			exitGracefully(0)
		}

		if installedUpdate != "" {
			var targetExt string
			if runtime.GOOS == "windows" {
				targetExt = ".zip"
			} else {
				targetExt = ".gz"
			}
			err := CheckUpdateAndInstall(targetExt, installedUpdate)
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				exitGracefully(1)
			}
		}
	}

	fmt.Printf("%s%s (%s) by 檀轶步棋%s\n",
		strings.Repeat("=", 8), AppName, Version, strings.Repeat("=", 8))

	f, ok := checkSealValid()
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr,
			"当前目录不完整，缺失以下文件或文件夹：\n%s\n%s\n",
			strings.Join(f, " "),
			"您是否已经将本程序放在了海豹核心的安装目录下（和 sealdice-core 等文件同个目录）？")
		var conf string
		fmt.Println("您是否需要进入恢复模式？(y/N）")
		_, err := fmt.Scanln(&conf)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr,
				"出现错误\n%s\n", err)
			exitGracefully(1)
		}

		if conf != "y" {
			fmt.Println("操作已取消")
			exitGracefully(0)
		}

		err = recoveryWithGui()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			exitGracefully(1)
		}

		exitGracefully(1)
	}

	var choice int
	fmt.Printf("请选择功能：\n%s\n%s\n",
		"[0] 退出程序\n[1] 恢复备份\n[2] 修补 SealDice", "请输入对应数字进行选择（0–2）")

	_, err := fmt.Scanln(&choice)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr,
			"出现错误，请检查您输入的是否是数字\n%s\n", err)
		exitGracefully(1)
	}

	switch choice {
	case 1:
		err = backupRestoreWithGui()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			exitGracefully(1)
		}
	case 2:
		err = recoveryWithGui()
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

func checkSealValid() ([]string, bool) {
	var essentialFiles = []string{"sealdice-core", "go-cqhttp/", "data/", "frontend/"}
	var missingFiles []string

	for _, f := range essentialFiles {
		fp := filepath.Join(workingDirectory, f)
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
