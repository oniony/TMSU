mod info;
mod init;

use std::env;
use std::io;
use std::path::PathBuf;
use std::process;

use clap::arg_enum;
use structopt::StructOpt;

use crate::errors::*;

#[derive(Debug, StructOpt)]
#[structopt(
    name = "TMSU",
    about = "A tool for tagging your files and accessing them through a virtual filesystem"
)]
struct TmsuOptions {
    // Externalize global options to a separate struct for convenience
    #[structopt(flatten)]
    global_opts: GlobalOptions,

    #[structopt(subcommand)]
    cmd: SubCommands,
}

#[derive(Debug, StructOpt)]
pub struct GlobalOptions {
    /// Use the specified database
    #[structopt(short = "-D", long, env = "TMSU_DB", parse(from_os_str))]
    database: Option<PathBuf>,

    /// Colorize the output (auto/always/never)
    #[structopt(long, default_value = "auto")]
    color: ColorMode,
}

arg_enum! {
    #[derive(Debug)]
    enum ColorMode {
        Auto,
        Always,
        Never,
    }
}

#[derive(Debug, StructOpt)]
enum SubCommands {
    Info(info::InfoOptions),
    Init(init::InitOptions),
}

/// CLI entry point, dispatching to subcommands
pub fn run() -> Result<()> {
    let opt = TmsuOptions::from_args();
    println!("{:?}", opt);

    match opt.cmd {
        SubCommands::Info(info_opts) => info_opts.execute(&opt.global_opts),
        SubCommands::Init(init_opts) => init_opts.execute(),
    }
}

fn locate_db(db_path: &Option<PathBuf>) -> Result<PathBuf> {
    // Use the given path if available
    match db_path {
        Some(path) => Ok(path.clone()),
        // Fallback: look for the DB in parent directories
        None => match find_database_upwards()? {
            Some(path) => Ok(path),
            // Fallback: use the default database
            None => match get_user_default_db() {
                Some(path) => Ok(path),
                // OK, we finally give up...
                None => Err(ErrorKind::NoDatabaseFound(PathBuf::default()).into()),
            },
        },
    }
}

/// Look for .tmsu/db in the current directory and ancestors
fn find_database_upwards() -> Result<Option<PathBuf>> {
    let mut path = env::current_dir()?;

    loop {
        let mut db_path = path.clone();
        db_path.push(".tmsu");
        db_path.push("db");

        debug!("Looking for database at {:?}", &db_path);
        if db_path.is_file() {
            return Ok(Some(db_path));
        }

        match path.parent() {
            Some(parent) => {
                path = PathBuf::from(parent);
            }
            None => {
                return Ok(None);
            }
        }
    }
}

/// Return the path corresponding to $HOME/.tmsu/default.db,
/// or None if the home directory cannot be resolved
fn get_user_default_db() -> Option<PathBuf> {
    dirs::home_dir().map(|mut path| {
        path.push(".tmsu");
        path.push("default.db");
        path
    })
}

fn should_use_colour(color_mode: &ColorMode) -> bool {
    match color_mode {
        ColorMode::Always => true,
        ColorMode::Never => false,
        ColorMode::Auto => termion::is_tty(&io::stdout()),
    }
}

pub fn print_error(result: Result<()>) {
    if let Err(error) = result {
        eprintln!("Error: {}", error);

        if let Some(backtrace) = error.backtrace() {
            eprintln!("backtrace: {:?}", backtrace);
        }

        process::exit(1);
    }
}
