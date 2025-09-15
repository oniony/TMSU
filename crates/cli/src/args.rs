use std::path::PathBuf;
use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(about = "TMSU", version, long_about = None)]
#[command(disable_colored_help = true)]
pub struct Args {
    #[clap(short = 'D', long)]
    pub database: Option<PathBuf>,

    #[command(subcommand)]
    pub command: Commands,
}

impl Args {
    pub fn parse() -> Self {
        Parser::parse()
    }
}

#[derive(Subcommand, Debug)]
pub enum Commands {
    #[command(
        about = "Show database information",
        long_about = "Show database paths and metrics.")]
    Info,

    #[command(
        about = "Initialize a new database",
        long_about = "Initializes a new local database.

Creates a .tmsu directory under PATH and initializes a new empty database within it.
If no PATH is specified then the current working directory is assumed.
")]
    Init { path: Option<PathBuf> },
}
