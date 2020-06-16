use std::env;
use std::path::PathBuf;

use structopt::StructOpt;

use crate::api;
use crate::errors::*;

/// Initializes a new database
#[derive(Debug, StructOpt)]
pub struct InitOptions {
    /// Path where to initialize a new empty database
    #[structopt()]
    path: Option<PathBuf>,
}

impl InitOptions {
    pub fn execute(self) -> Result<()> {
        let curr_path = self.path.unwrap_or(env::current_dir()?);
        api::init::run_init(&curr_path)
    }
}
