#[macro_use]
extern crate log;

mod api;
mod cli;
mod errors;
mod storage;

fn main() {
    // Initialize the logging system
    pretty_env_logger::init();

    // Parse CLI args and dispatch to the right subcommand
    let result = cli::run();

    // If there is an error, print it and exit with a non-zero error code
    cli::print_error(result);
}
