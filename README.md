# sealTools
 一些提供给 SealDice 的小工具

## 开发指南
现在仅支持恢复备份，修补模式（recovery）开发中。

1. 程序需要放在海豹核心根目录下（`checkSealValid()` 会检查工作目录中是否有海豹的文件）。
2. 为方便海豹集成，提供命令行参数 `-i` 和 `-t <target>`，前者用于验证是否是其他程序调用，后者是要恢复的备份的*绝对*路径。如果没有 `-i` 则会使用命令行界面与用户互动。
3. 提供另一个命令行参数 `-w <directory>`，用来指定程序的工作路径，缺省为程序所在路径。
4. 所有错误都输出到 `os.Stderr`，可以在主程序中用一些方法（`cmd.StderrPipe` 之类）捕获。
5. 在 macOS 上开发，未验证对 Windows 的兼容性（尤其是复制/创建文件那里）。**`recovery.go` 中的代码还未被全面测试过。**