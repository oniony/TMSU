use std::path::PathBuf;
use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(about, version, long_about = None)]
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
    #[command(about = "Show database information")]
    Info,
}
