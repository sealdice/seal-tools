# seal-tools (Rust)
 一些提供给 SealDice 的小工具。与 main 分支不同，该分支用 Rust 语言重构。

## 发行版本
目前无法在 macOS 上进行跨平台编译，造成此错误的原因仍然在排查中。此外，要进行跨平台编译，必须使用 `reqwest` 包的 `rustls-tls` feature：
```toml
reqwest = reqwest = { version = "0.11.18", default-features = false, features = ["rustls-tls", "blocking", "json"] }
```
目测开启此选项会使得编译后的文件体积增加 0.9MB 左右。

## 开发指南
文档会在后续版本完善。