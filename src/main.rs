use crate::cli::{Cli, Commands};
use crate::cmd_patch::{patch_gui, patch_raw};
use crate::cmd_restore::{restore_gui, restore_raw};
use clap::Parser;
use inquire::Select;
use std::io::Read;
use std::path::Path;
use std::{io, process};

mod cli;
mod cmd_patch;
mod cmd_restore;
mod unarchive;

static SELECTION_HELP: &str = "↑↓ [上下方向键]移动选项，↩︎ [回车键]确认选择";

fn main() {
    let cli = Cli::parse();
    if let Some(cmd) = cli.command {
        if let Err(err) = match cmd {
            Commands::Patch { file, delete_data } => patch_raw(Path::new(&file), delete_data),
            Commands::Restore { file, replace_cq } => restore_raw(Path::new(&file), replace_cq),
        } {
            exit_with(err, 1);
        }
    } else {
        //TODO: Detect environment
        let selection = Select::new("请选择您需要的操作", vec!["恢复备份", "修补海豹"])
            .with_help_message(SELECTION_HELP)
            .prompt();
        match selection {
            Ok(choice) => match choice.trim() {
                "恢复备份" => {
                    if let Err(err) = restore_gui() {
                        exit_with(err, 1);
                    }
                }
                "修补海豹" => {
                    if let Err(err) = patch_gui() {
                        exit_with(err, 1);
                    }
                }
                _ => {}
            },
            Err(err) => exit_with(err, 1),
        }
    }
}

fn exit_with(err: impl std::fmt::Display, code: i32) {
    if code != 0 {
        eprintln!("{err}");
    } else {
        println!("{err}");
    }
    if cfg!(windows) {
        println!("按下回车键退出……");
        _ = io::stdin().read(&mut []);
    }
    process::exit(code);
}
