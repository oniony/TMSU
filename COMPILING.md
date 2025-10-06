# Compiling from Source Code

These instructions are for compiling TMSU from the source code.

You may wish to compile from source if you want to make changes in your version or
if no binary or package is available for your system.

If you are just a regular user and a package is available for your operating system,
you will most likely be better off using that, as the package manager will make it
easier to install updates and put the files in the right places for things like
shell completion to work.

## Steps

1. Install Rust

    TMSU is written in Rust. To compile from source you must install the Rust toolchain:

    * <https://rust-lang.org/tools/install/>

2. Clone the TMSU Repository:

    Most people will probably want to clone the reository with Git:

    ```
    git clone -b main https://github.com/oniony/TMSU.git
    ```

    For other Jujutsu fans:

    ```
    jj git clone --colocate https://github.com/oniony/TMSU.git
    jj bookmark track main@origin
    jj new main
    ```

3. Build

    From the project directory, use Cargo to build the project:

    ```
    cargo build --release
    ```

    The resultant binary will be available at `target/release/tmsu`.

- - -

Copyright 2011-2025 Paul Ruane

Copying and distribution of this file, with or without modification,
are permitted in any medium without royalty provided the copyright
notice and this notice are preserved.  This file is offered as-is,
without any warranty.
