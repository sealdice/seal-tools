use crate::exit_with;
use crate::unarchive::{crau, list_files};
use std::{io, path};

pub(crate) fn restore_with_gui(wd: &str) -> Result<(), String> {
    println!("您希望如何恢复备份？");
    println!("[0] 退出程序\n[1] 从工作目录的文件中选择\n[2] 指定一个本地文件");
    println!("请输入序号(0–2)");
    let mut choice = String::new();
    if let Err(e) = io::stdin().read_line(&mut choice) {
        return Err(format!("意外错误：{e}"));
    }
    let choice: i32 = choice.trim().parse().unwrap_or(-1);

    match choice {
        0 => exit_with("", 0),
        1 => {
            let mut backup_dir = wd;
            let bdir = path::Path::new(wd).join("backups");
            if bdir.exists() {
                backup_dir = bdir.to_str().ok_or("无法读取备份文件夹")?;
            }

            return match list_files(backup_dir, "zip") {
                Ok(files) => {
                    if files.is_empty() {
                        exit_with("没有在工作目录下发现任何可能的备份文件", 1);
                    }

                    println!("发现以下可能的文件：");
                    for (pos, file) in files.iter().enumerate() {
                        println!("[{}] {}", pos + 1, file);
                    }
                    if files.len() == 1 {
                        println!("请输入序号选择(1)");
                    } else {
                        println!("请输入序号选择(1-{})", files.len());
                    }

                    let mut choice = String::new();
                    if let Err(e) = io::stdin().read_line(&mut choice) {
                        return Err(format!("意外错误：{e}"));
                    }
                    let choice: usize = choice.trim().parse().unwrap_or(0);
                    if choice == 0 || choice > files.len() {
                        return Err(String::from("无效选择"));
                    }

                    let dir = path::Path::new(backup_dir).join(files[choice - 1].clone());
                    let dir_str = dir.to_str().ok_or("无法导航至备份文件")?;

                    restore_backup(wd, Some(String::from(dir_str)))
                }
                Err(e) => Err(format!("获取文件列表时发生错误：{e}")),
            };
        }
        2 => {
            println!("请输入文件绝对路径：");
            let mut input = String::new();
            if let Err(e) = io::stdin().read_line(&mut input) {
                return Err(format!("意外错误：{e}"));
            }
            return restore_backup(wd, Some(String::from(input.trim())));
        }
        _ => return Err(String::from("无效选择")),
    }
    Ok(())
}

pub(crate) fn restore_backup(wd: &str, backup: Option<String>) -> Result<(), String> {
    let dest = backup.unwrap();
    crau(&dest, wd)
}
