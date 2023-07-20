use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(disable_help_flag = true)]
pub struct Cli {
    #[command(subcommand)]
    pub command: Option<Commands>,
}

#[derive(Subcommand)]
#[command(disable_help_subcommand = true)]
pub enum Commands {
    Patch {
        #[arg(short, long)]
        file: String,
        #[arg(short = 'x')]
        delete_data: bool,
    },
    Restore {
        #[arg(short, long)]
        file: String,
        #[arg(short, long = "replace")]
        replace_cq: bool,
    },
}
