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

此外，为其他系统编译通常涉及到 OpenSSL 不兼容的问题，必须启用 `rustls`：
```toml
curl = { version = "0.4.44", default-features = false, features = ["rustls"] }
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

Options:
  -w <WORKING_DIRECTORY>      The program's working directory [default: ./]
  -n, --nogui                 Run without GUI
  -h, --help                  Print help
  -V, --version               Print version
```

### `./seal-tools restore --help`
```text
Usage: seal-tools restore [OPTIONS]

Options:
  -b, --backup <BACKUP_PATH>      Specify a backup archive. Mandatory if run without gui
  -e, --except <PATH1, PATH2>...  Specify paths to be skipped
  -h, --help                      Print help
```

### `./seal-tools patch --help`
```text
Usage: seal-tools patch [OPTIONS]

Options:
  -p, --package <UPDATE_PATH>      Specify an update package. Mandatory if run without gui nor `-d`
  -d, --download                   Download the latest update package
  -n, --noinstall                  Cancel auto-installation after download. Only works when `-d` exists
  -e, --except <PATHS1, PATH2>...  Specify paths to be skipped
  -h, --help                       Print help
```

## 特别感谢
[熊砾](https://github.com/Lightinglight)：进行了大量的测试