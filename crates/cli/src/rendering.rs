use std::fmt::Display;

#[derive(Debug, Clone, Copy, Eq, PartialEq)]
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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn display() {
        assert_eq!("\n", format!("{}", Separator::Newline));
        assert_eq!("\0", format!("{}", Separator::Nul));
    }
}