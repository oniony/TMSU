use std::fmt::Display;

pub enum Separator {
    Nul,
    Newline,
}

impl Display for Separator {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Separator::Nul => write!(f, "\0"),
            Separator::Newline => write!(f, "\n"),
        }
    }
}