use crate::unarchive::{check_archive_replacement, crau, list_files, unarchive};
use crate::{exit_with, PACKAGE_EXT};
use clearscreen::clear;
use serde::Deserialize;
use std::error::Error;
use std::{env, fs, io, path};

#[derive(Deserialize)]
struct VersionInfo {
    #[serde(rename = "versionLatest")]
    ver: String,
    #[serde(rename = "versionLatestDetail")]
    ver_detail: String,
    #[serde(rename = "versionLatestCode")]
    ver_code: i64,
    #[serde(rename = "versionLatestNote")]
    ver_note: Option<String>,
    #[serde(rename = "minUpdateSupportVersion")]
    min_version: i64,
    #[serde(rename = "newVersionUrlPrefix")]
    new_url: Option<String>,
}

pub(crate) fn patch_with_gui(wd: &str) -> Result<(), String> {
    println!("您希望如何修补？");
    println!("[0] 退出程序\n[1] 从工作目录的文件中选择\n[2] 指定一个本地文件\n[3] 从网络上下载");
    println!("请输入序号(0–3)");
    let mut choice = String::new();
    if let Err(e) = io::stdin().read_line(&mut choice) {
        return Err(format!("意外错误：{e}"));
    }
    let choice: i32 = choice.trim().parse().unwrap_or(-1);
    match choice {
        0 => exit_with("", 0),
        1 => {
            _ = clear();
            let ext = if cfg!(windows) { "zip" } else { "gz" };
            return match list_files(wd, ext) {
                Ok(files) => {
                    if files.is_empty() {
                        exit_with("没有在工作目录下发现任何可能的更新文件", 1);
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

                    patch_seal(wd, Some(files[choice - 1].clone()), false, false)
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
            return patch_seal(wd, Some(String::from(input.trim())), false, false);
        }
        3 => {
            _ = clear();
            return patch_seal(wd, None, true, false);
        }
        _ => return Err(String::from("无效选择")),
    }

    Ok(())
}

pub(crate) fn patch_seal(
    wd: &str,
    package: Option<String>,
    download: bool,
    noinstall: bool,
) -> Result<(), String> {
    let mut dest = String::new();

    if let Some(package_path) = package {
        dest = package_path;
    } else if download {
        let ver_info = match get_version() {
            Ok(ver) => ver,
            Err(e) => return Err(format!("从网络获取最新版本时出现错误：{e}")),
        };

        println!("获取到最新版本：{}", ver_info.ver);

        let mut os = env::consts::OS;
        let mut arch = env::consts::ARCH;

        os = match os {
            "macos" => "darwin",
            "linux" => "linux",
            "windows" => "windows",
            _ => return Err(format!("不支持的系统`{os}`")),
        };

        arch = match arch {
            "x86_64" => "amd64",
            "aarch64" => "arm64",
            "i686" => "i386",
            _ => return Err(format!("不支持的架构`{arch}`")),
        };

        if os == "darwin" && arch == "arm64" {
            arch = "amd64";
        }

        let file_name = format!(
            "sealdice-core_{}_{}_{}.{}",
            ver_info.ver, os, arch, PACKAGE_EXT
        );

        let mut target_url = String::new();
        if let Some(url) = ver_info.new_url {
            target_url = url;
        } else {
            target_url = String::from(
                "https://sealdice.coding.net/p/sealdice/d/sealdice-binaries/git/raw/master/",
            );
        }

        println!("从{}获取`{}`...", target_url, file_name);
        target_url = format!("{}/{}", target_url, file_name);

        let try_dest = path::Path::new(wd).join(file_name);
        dest = match try_dest.to_str() {
            Some(d) => d.to_string(),
            None => return Err(String::from("准备下载文件时发生意外错误")),
        };

        if let Err(e) = download(&dest, &target_url) {
            return Err(format!("下载更新时发生错误：{e}"));
        }

        if noinstall {
            exit_with("因为 `--noinstall`，现在退出……", 0);
        }

        fn get_version() -> Result<VersionInfo, Box<dyn Error>> {
            let url = "https://dice.weizaima.com/dice/api/version?versionCode=0";
            let resp = reqwest::blocking::get(url)?;
            let ver_info: VersionInfo = resp.json()?;
            Ok(ver_info)
        }

        fn download(dest: &str, url: &str) -> Result<(), Box<dyn Error>> {
            let mut resp = reqwest::blocking::get(url)?;
            let mut file = fs::File::create(dest)?;
            io::copy(&mut resp, &mut file)?;
            Ok(())
        }
    }

    crau(&dest, wd)
}
