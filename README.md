# denv

denv is a command-line tool for managing environment variables in a secure and efficient manner. It allows users to set, get, delete, and import/export environment variables, while also supporting encryption for sensitive data.

## Features

- Manage environment variables with ease
- Support for encryption of sensitive data
- Import and export environment variables from/to files
- Command-line interface for easy usage

## Prerequisites

Before using `denv`, ensure the following tools are installed on your system:

- **[age](https://github.com/FiloSottile/age)**: A simple, modern, and secure encryption tool used by `denv` to encrypt environment variable data.

## Usage

### Running Commands

You can run commands using the following syntax:

```bash
./denv run -e key1 -e key2 -- command arg1 arg2
```

### Show Environment Variables

To display the environment variables, use:

```bash
./denv run -e key1 -e key2 --export
```

### Importing Environment Variables

To import environment variables from a directory:

```bash
./denv import <source>
```

### Exporting Environment Variables

To export all environment variables to a directory:

```bash
./denv export -o <outDir>
```

### Managing Recipients

You can manage encryption recipients with the following commands:

- List recipients: `./denv recipients`
- Add a recipient: `./denv recipientAdd <recipient>`
- Remove a recipient: `./denv recipientDel <recipient>`

## Data Storage

The `denv` tool organizes user data under the `DENV_ROOT` directory. Here's how the data is structured:

- **`DENV_ROOT`**: The root directory where all `denv` data is stored.
  - **`config.yml`**: A configuration file located at the root of `DENV_ROOT`. This file contains settings and metadata required for `denv` operations.
  - **`env/`**: A subdirectory where all environment variable data is securely stored. Each file in this directory is encrypted using `age` for enhanced security.
    - **`UNIQUIE_ID.age`**: Encrypted files representing individual environment variable sets. Each file is uniquely identified by a `UNIQUIE_ID`.

To customize the location of the root directory, you can set the `DENV_ROOT` environment variable. By default, it is set to `~/.config/denv`.

All data is encrypted and can be safely managed with version control tools like Git, ensuring both security and traceability.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
