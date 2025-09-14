mod args;
mod command;
mod database;
mod constants;

use std::process;
use args::{Args, Commands};
use crate::constants::APPLICATION_NAME;

fn main() {
    let args = Args::parse();

    let result = match args.command {
        Commands::Info => command::info::execute(args.database),
    };

    match result {
        Ok(_) => (),
        Err(e) => {
            eprintln!("{}: {}", APPLICATION_NAME, e);
            process::exit(1)
        },
    }
}
