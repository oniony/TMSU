mod args;
mod command;
mod database;
mod constants;

use std::error::Error;
use std::process;
use args::{Args, Commands};

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let db_path = match database::resolve(args.database) {
        Ok(db_path) => db_path,
        Err(error) => return fatal(error),
    };

    let result = match args.command {
        Commands::Info => command::info::execute(db_path),
        Commands::Init { path } => command::init::execute(db_path, path),
    };

    match result {
        Ok(_) => (),
        Err(error) => {
            fatal(error)
        }
    }
}

fn fatal(error: Box<dyn Error>) {
    eprintln!("{}: {}", constants::APPLICATION_NAME, error.to_string());
    process::exit(1)
}