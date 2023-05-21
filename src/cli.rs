use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(author, version, about, long_about = None)]
pub(crate) struct Cli {
    /// The program's working directory
    #[arg(short='w', value_name="WORKING_DIRECTORY", default_value_t=String::from("./"))]
    pub(crate) dir: String,
    /// Run without GUI
    #[arg(short, long)]
    pub(crate) nogui: bool,
    #[command(subcommand)]
    pub(crate) command: Option<Commands>,
}

#[derive(Subcommand)]
pub(crate) enum Commands {
    /// Restore backup
    Restore {
        /// Specify a backup archive. Mandatory if run without gui
        #[arg(short, long, value_name = "BACKUP_PATH")]
        backup: Option<String>,
        /// Specify paths to be skipped
        #[arg(short, long, value_name = "PATH1, PATH2", value_parser, num_args = 1.., value_delimiter = ',')]
        except: Option<Vec<String>>,
    },
    /// Patch up SealDice
    Patch {
        /// Specify an update package. Mandatory if run without gui nor `-d`
        #[arg(short, long, value_name = "UPDATE_PATH")]
        package: Option<String>,
        /// Download the latest update package
        #[arg(short, long)]
        download: bool,
        /// Cancel auto-installation after download. Only works when `-d` exists
        #[arg(short, long)]
        noinstall: bool,
        /// Specify paths to be skipped
        #[arg(short, long, value_name = "PATHS1, PATH2", value_parser, num_args = 1.., value_delimiter = ',')]
        except: Option<Vec<String>>,
    },
}
