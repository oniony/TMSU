extern crate clap;

use clap::{App, SubCommand};

fn main() {
    let matches = App::new("TMSU").version("1.0.0")
                                  .subcommand(SubCommand::with_name("files").about("Lists files with particular tags"))
                                  .get_matches();
}
