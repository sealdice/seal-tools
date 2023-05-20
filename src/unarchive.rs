use crate::exit_with;
use clearscreen::clear;
use std::error::Error;
use std::{fs, io, path};

pub(crate) fn crau(dest: &str, wd: &str) -> Result<(), String> {
    match check_archive_replacement(dest, wd) {
        Ok(dup) => {
            if !dup.is_empty() {
                println!(
                    "解压后，以下文件将被覆盖：\n{}\n以上文件将被覆盖，您确认吗？(y/N)",
                    dup.join("\n")
                );
                let mut choice = String::new();
                if let Err(e) = io::stdin().read_line(&mut choice) {
                    exit_with(format!("意外错误：{e}"), 1);
                }
                if choice.trim() != "y" {
                    exit_with("操作已取消", 0);
                }
            }

            if let Err(e) = unarchive(dest, wd) {
                exit_with(format!("解压文件时发生错误：{e}"), 1);
            }
        }
        Err(e) => return Err(format!("试图检查覆盖项目时发生错误：{e}")),
    }

    Ok(())
}

pub(crate) fn list_files(dir: &str, ext: &str) -> Result<Vec<String>, Box<dyn Error>> {
    let mut files = vec![];

    let entries = fs::read_dir(dir)?;
    for entry in entries {
        let entry = entry?;
        let entry_path = entry.path();
        if !entry_path.is_file() {
            continue;
        }

        let file_name = entry_path
            .file_name()
            .ok_or("无法读取文件名称")?
            .to_str()
            .ok_or("非法文件名称")?;
        let file_ext = match path::Path::new(file_name).extension() {
            Some(ext) => ext.to_str().ok_or("非法文件扩展")?,
            None => continue,
        };
        if file_ext == ext {
            files.push(String::from(file_name));
        }
    }

    Ok(files)
}

pub(crate) fn check_archive_replacement(
    src: &str,
    dest: &str,
) -> Result<Vec<String>, Box<dyn Error>> {
    let mut duplicates = vec![];

    let file = fs::File::open(src)?;
    let src_ext = path::Path::new(src)
        .extension()
        .ok_or("无法读取文件名称")?
        .to_str()
        .ok_or("非法文件名称")?;

    if src_ext == "zip" {
        let mut archive = zip::ZipArchive::new(file)?;
        for i in 0..archive.len() {
            let file = archive.by_index(i)?;
            let file_name = file.enclosed_name().ok_or(
                "警告-操作已停止：正在尝试访问的文件包含不安全的目录，请检查文件来源是否合规",
            )?;
            let dest_path = path::Path::new(dest).join(file_name);
            if dest_path.exists() {
                duplicates.push(dest_path.to_string_lossy().into_owned());
            }
        }
    } else {
        let gz_reader = flate2::read::GzDecoder::new(file);
        let mut archive = tar::Archive::new(gz_reader);
        for entry in archive.entries()? {
            let file = entry?;
            let file_name = file.path()?;
            if !is_path_safe(file_name.components()) {
                Err("警告-操作已停止：正在尝试访问的文件包含不安全的目录，请检查文件来源是否合规")?;
            }
            let dest_path = path::Path::new(dest).join(file_name);
            if dest_path.exists() {
                duplicates.push(dest_path.to_string_lossy().into_owned());
            }
        }
    }

    Ok(duplicates)
}

pub(crate) fn unarchive(src: &str, dest: &str) -> Result<(), Box<dyn Error>> {
    let file = fs::File::open(src)?;
    let src_ext = path::Path::new(src)
        .extension()
        .ok_or("无法解析文件")?
        .to_str()
        .ok_or("无法解析文件")?;

    if src_ext == "zip" {
        let mut archive = zip::ZipArchive::new(file)?;
        for i in 0..archive.len() {
            let mut file = archive.by_index(i)?;
            let file_name = file.enclosed_name().ok_or(
                "警告-操作已停止：正在尝试访问的文件包含不安全的目录，请检查文件来源是否合规",
            )?;
            let dest_path = path::Path::new(dest).join(file_name);
            if let Some(parent) = dest_path.parent() {
                if !parent.exists() {
                    println!("creatring {:#?}", parent);
                    fs::create_dir_all(parent)?;
                }
            }
            let mut outfile = fs::File::create(&dest_path)?;
            println!("copying {:#?}", dest_path);
            io::copy(&mut file, &mut outfile)?;
        }
    } else {
        let gz_reader = flate2::read::GzDecoder::new(file);
        let mut archive = tar::Archive::new(gz_reader);
        for entry in archive.entries()? {
            let mut file = entry?;
            let file_name = file.path()?;
            if !is_path_safe(file_name.components()) {
                Err("警告-操作已停止：正在尝试访问的文件包含不安全的目录，请检查文件来源是否合规")?;
            }
            let dest_path = path::Path::new(dest).join(file_name);

            if let Some(parent) = dest_path.parent() {
                if !parent.exists() {
                    println!("creatring {:#?}", parent);
                    fs::create_dir_all(parent)?;
                }
            }
            let mut outfile = fs::File::create(&dest_path)?;
            println!("copying {:#?}", dest_path);
            io::copy(&mut file, &mut outfile)?;
        }
    }

    _ = clear();
    Ok(())
}

fn is_path_safe(components: path::Components) -> bool {
    let normals: Vec<path::Component> = components
        .into_iter()
        .filter(|c| matches!(c, path::Component::Normal(_)))
        .collect();

    !normals.is_empty()
}
