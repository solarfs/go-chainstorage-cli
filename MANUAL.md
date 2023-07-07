# Chainstorage-cli Manual

## Introduction
- chainstorage-cli is the command line application to interact with rest api SDK of *chain strorage*.
- CAR import support: import CAR-archived content with custom DAGs directly to the Cluster.

## Installation

```
go install github.com/paradeum-team/chainstorage-cli@latest
```

## Getting Started

The chainstorage-cli tool provides various commands to interact with the system. Use the following commands to perform specific operations:

- `config`: Manage configuration
- `get`: Get object
- `help`: Get help for any command
- `import`: Import CAR file
- `log`: Manage and display logs of the running daemon
- `ls`: List links from an object or bucket
- `mb`: Create a new bucket
- `put`: Upload a file or folder
- `rb`: Remove a bucket
- `rm`: Delete objects or clear bucket
- `rn`: Rename an object
- `version`: Show version information

### Usage

To execute a command, use the following format:

```
command [flags] [arguments]
```

For example, to create a new bucket, use the following command:

```
gcscmd mb [bucket-name]
```

Replace `[bucket-name]` with the desired name for the new bucket.



### Manage Configuration (config)

The `config` command allows you to manage the configuration of the application.

```shell
gcscmd config
```

#### Examples

To display the current configuration settings, use the following command:

```shell
gcscmd config
```



### Get Object (get)

The `get` command allows you to retrieve an object from the system. Use the following syntax:

```shell
gcscmd get cs://<BUCKET> [--name=<name>] [--cid=<cid>] [--downloadFolder=<downloadfolder>] [flags]
```

Replace `<BUCKET>` with the name of the bucket where the object is located. You can optionally specify the name or CID of the object using the `--name` or `--cid` flags, respectively.

#### Flags

- `-c, --cid <cid>`: The CID of the object you want to retrieve.
- `-n, --name <name>`: The name of the object you want to retrieve.
- `-d, --downloadFolder <downloadfolder>`: The folder of download.

#### Example

To get an object named "example.txt" from the bucket "my-bucket", use the following command:

```shell
gcscmd get cs://my-bucket --name=example.txt
```

This will retrieve the "example.txt" object from the "my-bucket" bucket.

Use the `get` command to retrieve objects from the system by specifying the bucket and optionally providing the name or CID of the object you want to retrieve.



### Help

The `help` command provides detailed help information for any command in the application.

```shell
gcscmd help [command]
```

Replace `[command]` with the name of the command for which you need help.



### Import CAR File (import)

The `import` command allows you to import a CAR file into the system. Use the following syntax:

```shell
gcscmd import <CARFILE> cs://<BUCKET>
```

#### Example

To import a CAR file into a bucket, use the following command:

```shell
gcscmd import CARFILE cs://BUCKET
```

Replace `CARFILE` with the path to the CAR file you want to import, and `BUCKET` with the name of the bucket where you want to import the file.

This command will import the specified CAR file into the specified bucket.



### Manage and Show Logs (log)

The `log` command allows you to manage and display logs. Use the following syntax:

```shell
gcscmd log <level>
```

#### Example

To manage and show logs of the running daemon, use the following command:

```shell
gcscmd log <level>
```

Replace `<level>` with the desired log level, such as `trace`, `debug`, `info`, `warn`, or `error`.


This command will manage and display logs of the specified log level.



### List Bucket or Object (ls)

The `ls` command allows you to list links from an object or bucket. Use the following syntax:

```shell
gcscmd ls [cs://<BUCKET>] [--name=<name>] [--cid=<cid>] [--Offset=<Offset>]
```

`cs://BUCKET`: The name of the bucket you want to get.

#### Flags

- `-c, --cid <cid>`: Specifies the CID of the object to list links from.
- `-n, --name <name>`: Specifies the name of the object to list links from.
- `-o, --offset <offset>`: Specifies the list offset, i.e., the number of links to skip before starting the listing. The default offset is 10.

#### Example

To list all buckets, use the following command:

```shell
gcscmd ls --offset <offset>
```

Replace `<offset>` with the desired list offset value.

To list links from an object or bucket, use the following command:

```shell
gcscmd ls cs://<BUCKET> --cid <cid> --offset <offset>
```

Replace `<BUCKET>` with the name of the bucket, `<cid>` with the CID of the object, and `<offset>` with the desired list offset value.

Alternatively, you can specify the object name using the `--name` flag:

```shell
gcscmd ls cs://<BUCKET> --name <name> --offset <offset>
```

Replace `<name>` with the name of the object and `<offset>` with the desired list offset value.



### Create Bucket (mb)

The `mb` command allows you to create a new bucket. Use the following syntax:

```shell
gcscmd mb cs://<BUCKET> [--storageNetworkCode=<storageNetworkCode>] [--bucketPrincipleCode=<bucketPrincipleCode>]
```

Replace `<BUCKET>` with the desired name for the new bucket.

#### Flags

- `-s, --storageNetworkCode`: Optional. The storage network code for the bucket. Default: 10001.
- `-b, --bucketPrincipleCode`: Optional. The bucket principle code. Default: 10001.

#### Examples

Create a bucket named "my-bucket" with default storage network and bucket principle codes:

```shell
gcscmd mb cs://my-bucket
```

Create a bucket named "my-bucket" with custom storage network and bucket principle codes:

```shell
gcscmd mb cs://my-bucket --storageNetworkCode=10001 --bucketPrincipleCode=10001
```



### Upload File or Folder (put)

The `put` command allows you to upload a file or folder to the system. Use the following syntax:

```shell
gcscmd put <FILE[/DIR...]> cs://<BUCKET> [flags]
```

- `FILE[/DIR...]`: The path to the file or folder you want to upload.
- `cs://BUCKET`: The target bucket where the file or folder should be uploaded.

#### Flags

- `-c, --carFile`: Optional. Specify a CAR file to upload.

#### Examples

Upload a single file:

```shell
gcscmd put /path/to/file.txt cs://my-bucket
```

Upload a folder:

```shell
gcscmd put /path/to/folder cs://my-bucket
```

Upload a file using a CAR file:

```shell
gcscmd put --carFile=/path/to/file.car cs://my-bucket
```



### Remove Bucket (rb)

The `rb` command allows you to remove a bucket from the system. Use the following syntax:

```shell
gcscmd rb cs://<BUCKET> [--force]
```

- `cs://BUCKET`: The name of the bucket you want to remove.

#### Flags

- `-f, --force`: Optional. If the bucket contains objects, adding this flag will prompt for confirmation before deletion.

#### Example

Remove a bucket:

```shell
gcscmd rb cs://my-bucket
```

Remove a bucket with confirmation for deletion if it contains objects:

```shell
gcscmd rb cs://my-bucket --force
```



### Delete Objects or Clear Bucket (rm)

The `rm` command allows you to delete objects, or empty a bucket by removing all objects within it. Use the following syntax:

```shell
gcscmd rm cs://<BUCKET> [--name=<name>] [--cid=<cid>] [--force]
```

- `cs://BUCKET`: The name of the bucket you want to empty.

#### Flags

- `-n, --name`: Optional. Specify the name of a specific object to remove.
- `-c, --cid`: Optional. Specify the CID (Content Identifier) of a specific object to remove.
- `-f, --force`: Optional. Adding this flag will prompt for confirmation before emptying the bucket.

#### Example

Empty a bucket:

```shell
gcscmd rm cs://my-bucket
```

Empty a bucket and confirm deletion for each object:

```shell
gcscmd rm cs://my-bucket --force
```

Empty a bucket and remove a specific object:

```shell
gcscmd rm cs://my-bucket --name=my-object
```



### Rename Object (rn)

The `rn` command allows you to rename an object. Use the following syntax:

```shell
gcscmd rn cs://<BUCKET> <--name=<name>> [--cid=<cid>] <--rename=<rename>> [--force]
```

- `cs://BUCKET`: The name of the bucket where the object is located.

#### Flags

- `-n, --name`: Specify the current name of the object.
- `-c, --cid`: Optional. Specify the CID (Content Identifier) of the object.
- `-r, --rename`: Specify the new name for the object.
- `-f, --force`: Optional. Adding this flag will prompt for confirmation if there are any filename conflicts during the renaming process.

#### Example

Rename an object:

```shell
gcscmd rn cs://my-bucket --name=my-object --rename=new-object
```

Rename an object using its CID:

```shell
gcscmd rn cs://my-bucket --cid=<object-cid> --rename=new-object
```

Rename an object and force overwrite if there are conflicts:

```shell
gcscmd rn cs://my-bucket --name=my-object --rename=new-object --force
```



### Version

The `version` command displays information about the current version of IPFS.

```shell
gcscmd version
```

#### Example

To show the IPFS version information, use the following command:

```shell
gcscmd version
```

This will display the version number information about IPFS.



## Configuration

The CLI tool supports custom configuration to modify its behavior. Configuration can be specified using a TOML (Tom's Obvious, Minimal Language) file.

### Configuration File

The configuration file is a TOML file that contains key-value pairs representing various configuration options. By default, the CLI tool looks for a configuration file named `config.toml` in the current working directory. You can specify a different file using the `--config` flag.

Here is an example of a configuration file (`config.toml`) with available options:

```
[cli]
ipfsGateway = "test-ipfs-gateway.netwarps.com/ipfs/"
useHttpsProtocol = true
bucketPrefix = 'cs://'
listOffset = 20
cleanTmpData = true
maxRetries = 3
retryDelay = 3

[sdk]
defaultRegion = 'hk-1'
timeZone = 'UTC +08:00'
chainStorageApiEndpoint = 'http://127.0.0.1:8821'
useHttpsProtocol = true
carFileWorkPath = './tmp/carfile'
carFileShardingThreshold = 46137344
chainStorageApiToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...'
httpRequestUserAgent = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36'
httpRequestOvertime = 30
carVersion = 1

[logger]
logPath = './logs'
mode = 'debug'
level = 'info'
isOutPutFile = true
maxAgeDay = 7
rotationTime = 1
useJson = true
loggerFile= 'chainstorage-cli-log'
```



### CLI Configuration

The CLI configuration allows you to specify various settings related to the tool's functionality.

#### ipfsGateway

- Description: The IPFS gateway URL to use for fetching IPFS content.
- Type: String
- Default Value: `test-ipfs-gateway.netwarps.com/ipfs/`

#### useHttpsProtocol

- Description: Indicates whether to use HTTPS protocol.
- Type: Boolean
- Default Value: `true`

#### bucketPrefix

- Description: The prefix to add to bucket names.
- Type: String
- Default Value: `cs://`

#### listOffset

- Description: The number of items to skip when listing objects.
- Type: Integer
- Default Value: `20`

#### cleanTmpData

- Description: Indicates whether to clean temporary data after each operation.
- Type: Boolean
- Default Value: `true`

#### maxRetries

- Description: The maximum number of retries for failed operations.
- Type: Integer
- Default Value: `3`

#### retryDelay

- Description: The delay (in seconds) between retry attempts.
- Type: Integer
- Default Value: `3`



### SDK Configuration

The SDK configuration allows you to specify various settings related to the SDK's functionality.

#### defaultRegion

- Description: The default region for SDK operations.
- Type: String
- Default Value: `hk-1`

#### timeZone

- Description: The time zone used for SDK operations.
- Type: String
- Default Value: `UTC +08:00`

#### chainStorageApiEndpoint

- Description: The API endpoint for the Chain Storage service.
- Type: String
- Default Value: `http://127.0.0.1:8821`

#### useHttpsProtocol

- Description: Indicates whether to use HTTPS protocol.
- Type: Boolean
- Default Value: `true`

#### carFileWorkPath

- Description: The working directory for CAR files.
- Type: String
- Default Value: `./tmp/carfile`

#### carFileShardingThreshold

- Description: The threshold (in bytes) for CAR file sharding.
- Type: Integer
- Default Value: `46137344`

#### chainStorageApiToken

- Description: The API token for accessing the Chain Storage service.
- Type: String
- Default Value: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`

#### httpRequestUserAgent

- Description: The user agent to use in HTTP requests.
- Type: String
- Default Value: `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)...`

#### httpRequestOvertime

- Description: The timeout (in seconds) for HTTP requests.
- Type: Integer
- Default Value: `30`

#### carVersion

- Description: The version of CAR files.
- Type: Integer
- Default Value: `1`



### Logger Configuration

The logger configuration allows you to specify various settings related to logging.

#### logPath

- Description: The path where log files will be stored.
- Type: String
- Default Value: `./logs`

#### mode

- Description: The mode of the logger.
- Type: String
- Default Value: `debug`

#### level

- Description: The logging level.
- Type: String
- Default Value: `info`

#### isOutPutFile

- Description: Indicates whether logging should be output to a file.
- Type: Boolean
- Default Value: `true`

#### maxAgeDay

- Description: The maximum age of log files in days before they are rotated.
- Type: Integer
- Default Value: `7`

#### rotationTime

- Description: The time interval (in days) at which log files should be rotated.
- Type: Integer
- Default Value: `1`

#### useJson

- Description: Indicates whether logging should be output in JSON format.
- Type: Boolean
- Default Value: `true`

#### loggerFile

- Description: The name of the log file.
- Type: String
- Default Value: `chainstorage-cli-log`
