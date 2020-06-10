use std::fs;
use std::path::{Path, PathBuf};

use crate::errors::*;
use crate::storage::{self, Storage, Transaction};

pub struct InfoOutput {
    pub db_path: PathBuf,
    pub root_path: PathBuf,
    pub size: u64,

    pub stats_info: Option<StatsOutput>,

    pub usage_info: Option<UsageOutput>,
}

pub struct StatsOutput {
    pub tag_count: u64,
    pub value_count: u64,
    pub file_count: u64,
    pub file_tag_count: u64,
}

pub struct UsageOutput {
    pub tag_data: Vec<TagInfo>,
}

pub struct TagInfo {
    pub name: String,
    pub count: u64,
}

pub fn run_info(db_path: &Path, with_stats: bool, with_usage: bool) -> Result<InfoOutput> {
    let mut store = Storage::open(&db_path)?;
    let mut tx = store.begin_transaction()?;

    let mut stats_data = None;
    if with_stats {
        stats_data = Some(compute_stats(&mut tx)?);
    }

    let mut usage_data = None;
    if with_usage {
        usage_data = Some(compute_usage(&mut tx)?);
    }

    tx.commit()?;

    Ok(InfoOutput {
        db_path: store.db_path.clone(),
        root_path: store.root_path.clone(),
        size: fs::metadata(db_path)?.len(),
        stats_info: stats_data,
        usage_info: usage_data,
    })
}

fn compute_stats(tx: &mut Transaction) -> Result<StatsOutput> {
    Ok(StatsOutput {
        tag_count: storage::tag_count(tx)?,
        value_count: storage::value_count(tx)?,
        file_count: storage::file_count(tx)?,
        file_tag_count: storage::tag_file_count(tx)?,
    })
}

fn compute_usage(tx: &mut Transaction) -> Result<UsageOutput> {
    let tag_usage = storage::tag_usage(tx)?;

    Ok(UsageOutput {
        tag_data: tag_usage
            .into_iter()
            .map(|tfc| TagInfo {
                name: tfc.name,
                count: tfc.file_count as u64,
            })
            .collect(),
    })
}
