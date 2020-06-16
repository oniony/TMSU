use structopt::StructOpt;

use std::fmt;

use termion::color;

use crate::api;
use crate::errors::*;

/// Shows database information
#[derive(Debug, StructOpt)]
pub struct InfoOptions {
    /// Shows statistics
    #[structopt(short, long)]
    stats: bool,

    /// Shows tag usage breakdown
    #[structopt(short, long)]
    usage: bool,
}

impl InfoOptions {
    pub fn execute(&self, global_opts: &super::GlobalOptions) -> Result<()> {
        let db_path = super::locate_db(&global_opts.database)?;
        info!("Database path: {}", db_path.display());

        let use_colors = super::should_use_colour(&global_opts.color);

        let info_output = api::info::run_info(&db_path, self.stats, self.usage)?;
        print_output(&info_output, use_colors);

        Ok(())
    }
}

fn print_output(info_data: &api::info::InfoOutput, use_colors: bool) {
    print_entry("Database", &info_data.db_path.display(), use_colors);
    print_entry("Root path", &info_data.root_path.display(), use_colors);
    print_entry("Size", &info_data.size, use_colors);

    if let Some(stats_data) = &info_data.stats_info {
        print_stats(stats_data, use_colors);
    }

    if let Some(usage_data) = &info_data.usage_info {
        print_usage(usage_data, use_colors);
    }
}

fn print_stats(stats_data: &api::info::StatsOutput, use_colors: bool) {
    println!();

    print_entry("Tags", &stats_data.tag_count, use_colors);
    print_entry("Values", &stats_data.value_count, use_colors);
    print_entry("Files", &stats_data.file_count, use_colors);
    let taggings = stats_data.file_tag_count;
    print_entry("Taggings", &taggings, use_colors);

    let average_tags_per_file = taggings as f64 / stats_data.file_count as f64;
    print_entry(
        "Mean tags per file",
        &format!("{:.2}", average_tags_per_file),
        use_colors,
    );
    let average_files_per_tag = taggings as f64 / stats_data.tag_count as f64;
    print_entry(
        "Mean files per tag",
        &format!("{:.2}", average_files_per_tag),
        use_colors,
    );
}

fn print_usage(usage_data: &api::info::UsageOutput, use_colors: bool) {
    let max_tag_width = usage_data
        .tag_data
        .iter()
        .map(|tag_info| tag_info.name.len())
        .max()
        .unwrap_or_default();

    println!();
    for tag_info in &usage_data.tag_data {
        print_tag_info(tag_info, max_tag_width, use_colors);
    }
}

fn print_entry<T: fmt::Display>(key: &str, value: &T, use_colors: bool) {
    if use_colors {
        println!(
            "{}: {}{}{}",
            key,
            color::Fg(color::Green),
            value,
            color::Fg(color::Reset)
        );
    } else {
        println!("{}: {}", key, value);
    }
}

fn print_tag_info(tag_info: &api::info::TagInfo, max_tag_width: usize, use_colors: bool) {
    if use_colors {
        println!(
            "  {:width$} {}{}{}",
            &tag_info.name,
            color::Fg(color::Yellow),
            tag_info.count,
            color::Fg(color::Reset),
            width = max_tag_width
        );
    } else {
        println!(
            "  {:width$} {}",
            &tag_info.name,
            tag_info.count,
            width = max_tag_width
        );
    }
}
