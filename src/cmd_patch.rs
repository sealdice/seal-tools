use crate::unarchive::unarchive;
use crate::SELECTION_HELP;
use inquire::{Select, Text};
use std::error::Error;
use std::path::Path;
use std::{fs, io};

pub fn patch_gui() -> Result<(), Box<dyn Error>> {
    let target = Text::new("请输入海豹安装文件路径").prompt()?;
    let selection = Select::new("您是否要删除 data 文件夹?", vec!["是", "否"])
        .with_starting_cursor(1)
        .with_help_message(SELECTION_HELP)
        .prompt()?;
    let del = selection == "是";
    patch_raw(Path::new(target.trim()), del)
}

pub fn patch_raw(path: &Path, del_data: bool) -> Result<(), Box<dyn Error>> {
    match unarchive(path, Path::new("./"), vec![]) {
        Ok(_) => {
            if del_data {
                println!("deleting data/default");
                return match fs::remove_dir_all("./data/default") {
                    Ok(()) => Ok(()),
                    Err(err) => {
                        if err.kind() == io::ErrorKind::NotFound {
                            //println!("encountered not found");
                            Ok(())
                        } else {
                            Err(err)?
                        }
                    }
                };
            }
            Ok(())
        }
        Err(err) => Err(err)?,
    }
}
