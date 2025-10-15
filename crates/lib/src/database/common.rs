#[derive(Clone, Debug, Eq, PartialEq)]
pub enum TagSpecificity {
    All,
    ExplicitOnly,
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum Casing {
    Insensitive,
    Sensitive,
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum FileTypeSpecificity {
    Any,
    FileOnly,
    DirectoryOnly,
}
