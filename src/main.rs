use crate::cli::{Cli, Commands};
use crate::command_patch::{patch_seal, patch_with_gui};
use clap::Parser;
use clearscreen::clear;
use std::io::Read;
use std::{io, path, process};
use crate::command_restore::{restore_backup, restore_with_gui};

mod cli;
mod command_patch;
mod command_restore;
mod unarchive;

#[cfg(windows)]
static PACKAGE_EXT: &str = "zip";
#[cfg(not(windows))]
static PACKAGE_EXT: &str = "tar.gz";

fn main() {
    let mut cli = Cli::parse();

    if cli.nogui {
        todo!()
    }

    if cli.command.is_none() {
        let missing_files = check_self_integrity(&cli.dir);
        if !missing_files.is_empty() {
            println!(
                "工作路径下缺少以下文件：\n{}\n请检查工作路径`{}`是否正确，或者进入修复模式。\n您要进入修复模式吗？(y/N)",
                missing_files.join(" "), cli.dir
            );

            let mut choice = String::new();
            if let Err(e) = io::stdin().read_line(&mut choice) {
                exit_with(format!("意外错误：{e}"), 1);
            }

            if choice.trim() != "y" {
                exit_with("操作已取消", 0);
            }

            if let Err(e) = patch_with_gui(&cli.dir) {
                exit_with(e, 1);
            } else {
                exit_with(
                    "安装成功！请先实验骰子能否启动，然后（如果可能的话）恢复备份。",
                    0,
                );
            }
        }

        println!("========SealTools by 檀轶步棋=========");
        println!("请选择要使用的功能：");
        println!("[0] 退出程序\n[1] 恢复备份\n[2] 修补海豹");
        println!("请输入序号(0–2)");
        let mut choice = String::new();
        if let Err(e) = io::stdin().read_line(&mut choice) {
            exit_with(format!("意外错误：{e}"), 1);
        }
        let choice: i32 = choice.trim().parse().unwrap_or(-1);
        match choice {
            1 => cli.command = Some(Commands::Restore { backup: None }),
            2 => {
                cli.command = Some(Commands::Patch {
                    package: None,
                    download: false,
                    noinstall: false,
                })
            }
            114514 => {
                println!("等酒香醇　等你弹一曲古筝");
                exit_with("", 25);
            }
            0 => exit_with("操作已取消", 0),
            _ => exit_with("无效输入", 1),
        }
    }

    _ = clear();

    match cli.command.unwrap() {
        Commands::Restore { backup } => {
            if let Err(e) = if backup.is_none() {
                restore_with_gui(&cli.dir)
            } else {
                restore_backup(&cli.dir, backup)
            } {
                exit_with(e, 1);
            } else {
                exit_with("成功恢复备份！", 0)
            }
        },
        Commands::Patch {
            package,
            download,
            noinstall,
        } => {
            if let Err(e) = if package.is_none() && !download && !noinstall {
                patch_with_gui(&cli.dir)
            } else {
                patch_seal(&cli.dir, package, download, noinstall)
            } {
                exit_with(e, 1);
            } else {
                exit_with(
                    "安装成功！请先实验骰子能否启动，然后（如果可能的话）恢复备份。",
                    0,
                );
            }
        }
    }
}

fn check_self_integrity(wd: &str) -> Vec<String> {
    let mut essential_files = vec!["go-cqhttp/", "data/", "frontend/"];
    essential_files.push(if cfg!(windows) {
        "sealdice-core.exe"
    } else {
        "sealdice-core"
    });

    let mut missing_files: Vec<String> = vec![];
    for file in essential_files {
        let file_path = path::Path::new(wd).join(file);
        if !file_path.exists() {
            missing_files.push(String::from(file));
        }
    }

    missing_files
}

fn exit_with(err: impl std::fmt::Display, code: i32) {
    if cfg!(windows) {
        println!("按下回车键退出……");
        _ = io::stdin().read(&mut []);
    }

    if code != 0 {
        eprintln!("{err}");
    } else {
        println!("{err}");
    }
    process::exit(code);
}
