# seal-tools (Rust)
 一些提供给 SealDice 的小工具。与 main 分支不同，该分支用 Rust 语言重构。

## 发行版本
目前无法在 macOS 上进行跨平台编译，造成此错误的原因仍然在排查中。此外，为其他系统编译通常涉及到 OpenSSL 不兼容的问题，必须启用 `rustls`：
```toml
curl = { version = "0.4.44", default-features = false, features = ["rustls"] }
```

## 开发指南
文档会在后续版本完善。