use rusqlite::Connection;
use std::error::Error;
use std::fmt::Display;

/// Application setting.
pub enum Setting {
    Root,
}

impl Display for Setting {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Setting::Root => write!(f, "database-root"),
        }
    }
}

/// The setting store.
pub struct Store<'s> {
    connection: &'s Connection,
}

impl Store<'_> {
    /// Creates a new setting store.
    pub fn new(connection: &Connection) -> Store {
        Store { connection }
    }

    pub fn read(&self, setting: Setting) -> Result<String, Box<dyn Error>> {
        let value = self.connection.query_one(
            "\
            SELECT value FROM setting WHERE name = ?;",
            [setting.to_string()],
            |r| r.get::<usize, String>(0),
        )?;

        Ok(value)
    }

    pub fn update(&self, setting: Setting, value: &str) -> Result<(), Box<dyn Error>> {
        let _ = self.connection.execute(
            "\
        INSERT INTO setting (name ,value)
        VALUES (?1, ?2)
        ON CONFLICT DO UPDATE
        SET value = ?2;
        ",
            (setting.to_string(), value),
        )?;

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use crate::database::setting::Setting;

    #[test]
    fn setting_display() {
        assert_eq!("database-root", format!("{}", Setting::Root));
    }
}
