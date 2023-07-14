# Chainstorage-cli

Chainstorage-cli is a powerful and versatile tool that allows you to interact with the rest api SDK of *web3 chain strorage*. This tool provides a set of commands to manage objects, buckets, and configurations.

## Features

- Create and manage buckets
- Upload files and folders
- Import CAR files
- List links from objects or buckets
- Remove buckets and objects
- Rename objects
- View and manage logs
- Get object information
- Manage configuration settings

## Installation

To install Chainstorage-cli, follow these steps:

1. Download the latest release from the [GitHub repository](https://github.com/solarfs/go-chainstorage-cli/releases).
2. Extract the downloaded package to a desired location on your system.
3. Add the tool's executable to your system's PATH variable to access it from anywhere in the terminal.

## Usage

To use Chainstorage-cli, open a terminal and run the tool's executable with the desired command and arguments.

```shell
gcscmd [command] [flags]
```

For example, to create a new bucket, use the following command:

```shell
gcscmd mb cs://my-bucket
```

Refer to the command's specific section in the manual for detailed usage instructions and available flags.

## Manual

For detailed information on how to use each command and its available options, refer to the [Manual](./MANUAL.md).

## Configuration

Chainstorage-cli allows you to configure various settings to customize its behavior. The configuration file can be found at `config.toml`. Update the configuration file as per your requirements.

## Credits

This project builds upon the work of the [go-carbites](https://github.com/alanshaw/go-carbites) created by [ Alan Shaw](https://github.com/alanshaw) and the [pget](https://github.com/Code-Hex/pget) created by [Kei Kamikawa](https://github.com/Code-Hex).

Please check out their repository for more information and to see their original work.

## License

Chainstorage-cli is open source and released under the [MIT License](./LICENSE).

## Contact

If you have any questions or need further assistance, feel free to contact our support team at https://www.solarfs.io/.

We hope you find Chainstorage-cli useful! Happy command-line operations!


