# svync
Svync is a tool designed to synchronize structural variant calls from different callers. It uses YAML configs to define how to handle the standardization. 

## Usage
```bash
svync --config <config.yaml> --input <input.vcf>
```

### Arguments
#### Required
| Argument | Description |
| --- | --- |
| `--config`/`-c` | Path to the YAML config file |
| `--input`/`-i` | Path to the input VCF file |

#### Optional
| Argument | Description | Default |
| --- | --- | --- |
| `--output`/`-o` | Path to the output VCF file | `stdout` |
| `--nodate`/`--nd` | Do not add the date to the output VCF file | `false` |
| `--mute-warnings`/`--mw` | Do not output warnings | `false` |

## Configuration
The configuration file is the core of the standardization in Svync. More information can be found in the [configuration documentation](docs/configuration.md).


## Installation
### Mamba/Conda
This is the preferred way of installing Svync.

```bash
mamba install -c bioconda svync
```

or with conda:
  
```bash 
conda install -c bioconda svync
```

### Precompiled binaries
Precompiled binaries are available for Linux and macOS on the [releases page](https://github.com/nvnieuwk/svync/releases).


### Installation from source
Make sure you have go installed on your machine (or [install](https://go.dev/doc/install) it if you don't currently have it)

Then run these commands to install svync:

```bash
go get .
go build .
sudo mv svync /usr/local/bin/
```

Next run this command to check if it was correctly installed:

```bash
svync --help
```

