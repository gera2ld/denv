# denv

denv is a command-line tool for managing environment variables in a secure and efficient manner. It allows users to set, get, delete, and import/export environment variables, while also supporting encryption for sensitive data.

## Features

- Manage environment variables with ease
- Support for encryption of sensitive data
- Import and export environment variables from/to files
- Command-line interface for easy usage

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

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.
