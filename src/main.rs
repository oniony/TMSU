extern crate clap;

use clap::{App, AppSettings, Arg, SubCommand};

fn main() {
    let matches = App::new("TMSU").version("1.0.0")
                                  .template("{bin}\n\n{subcommands}\n\nGlobal options:\n\n{flags}\n\n{after-help}")
                                  .after_help("Specify subcommand name for detailed help on a particular subcommand, e.g. tmsu help files")
                                  .arg(Arg::with_name("verbose").short("v")
                                                                .long("verbose")
                                                                .help("Show verbose messages"))
                                  .subcommand(SubCommand::with_name("config").about("Views or amends database sesttings"))
                                  .subcommand(SubCommand::with_name("files").about("Lists files with particular tags"))
                                  .get_matches();
}
