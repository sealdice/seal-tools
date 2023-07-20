use crate::unarchive::unarchive;
use crate::SELECTION_HELP;
use inquire::{Select, Text};
use std::error::Error;
use std::path::Path;

pub fn restore_gui() -> Result<(), Box<dyn Error>> {
    let mut target = String::new();
    let backup_dir = Path::new("./backups");
    let options = if backup_dir.exists() {
        vec!["从备份目录中选择", "指定其他目录的文件"]
    } else {
        vec!["指定其他目录的文件"]
    };
    let selection = Select::new("您希望如何恢复备份？", options)
        .with_help_message(SELECTION_HELP)
        .prompt()?;
    match selection {
        "从备份目录中选择" => {
            let entries = backup_dir.read_dir()?;
            let mut files = vec![];
            for entry in entries {
                let entry = entry?;
                let path = entry.path();
                if !path.is_file() {
                    continue;
                }
                let name = path
                    .file_name()
                    .ok_or("unintelligible file name")?
                    .to_str()
                    .ok_or("cannot convert file name to standard string")?;
                if match Path::new(name).extension() {
                    Some(ext) => ext
                        .to_str()
                        .ok_or("cannot convert file extension to standard string")?,
                    None => continue,
                } != "zip"
                {
                    continue;
                }
                files.push(name.to_string());
            }
            target = format!(
                "backups/{}",
                Select::new("请选择要恢复的备份", files)
                    .with_help_message(SELECTION_HELP)
                    .prompt()?
                    .trim()
            );
        }
        "指定其他目录的文件" => {
            target = Text::new("请输入文件路径").prompt()?;
        }
        _ => {}
    }

    let rep_cq = matches!(
        Select::new(
            "您要替换 go-cqhttp 数据吗? (如果要，将会使用该备份的登录数据)",
            vec!["是", "否"]
        )
        .with_starting_cursor(1)
        .prompt()?,
        "是"
    );
    println!("{target}");
    restore_raw(Path::new(&target), rep_cq)
}

pub fn restore_raw(path: &Path, rep_cq: bool) -> Result<(), Box<dyn Error>> {
    let except: Vec<&Path> = if !rep_cq {
        vec![Path::new("data/default/extra")]
    } else {
        vec![]
    };
    unarchive(path, Path::new("./"), except)
}
