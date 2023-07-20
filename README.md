# seal-tools
 一些提供给 SealDice 的小工具。

## 跨平台编译（以 aarch64-unknown-linux-gnu 为例）
1. 您需要安装其他系统的 Rust 标准库：
  ```bash
  rustup target add aarch64-unknown-linux-gnu
  ```
2. 安装对应架构的 gcc：
  ```bash
  sudo apt-get update && apt-get install gcc-aarch64-linux-gnu
  ```
3. 打开 `~/.cargo/config`，如果没有则新建一个，然后添加
  ```toml
  [target.aarch64-unknown-linux-gnu]
  linker = "aarch64-linux-gnu-gcc"
  ```
4. 进行编译：
  ```bash
  cargo build --release --target=aarch64-unknown-linux-gnu
  ```
5. 进行 strip 以减少文件大小
  ```bash
  aarch64-linux-gnu-strip ./target/aarch64-unknown-linux-gnu/release/seal-tools
  ```

## 开发指南
文档会在后续版本完善。您现在可以在 `src/cli.rs` 查看所有命令和参数的定义。以下是内置的帮助：

### `./seal-tools --help`
```text
Usage: seal-tools [OPTIONS] [COMMAND]

Commands:
  restore  Restore backup
  patch    Patch up SealDice
  help     Print this message or the help of the given subcommand(s)
```

### `./seal-tools restore --help`
```text
Usage: seal-tools restore [OPTIONS]

Options:
  -f, --file <PATH>     Specify a backup archive
  -x                    Specify if to delete the old `data/default` folder after restoring
```

### `./seal-tools patch --help`
```text
Usage: seal-tools patch [OPTIONS]

Options:
  -f, --file <PATH>     Specify a backup archive
  -r, --replace         Specify if to replace the current `data/default/extra/` folder with the one in backup file
```

## 特别感谢
[熊砾](https://github.com/Lightinglight)：进行了大量的测试