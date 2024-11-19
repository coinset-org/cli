# Coinset CLI

This command-line interface gives you quick access to the Chia blockchain information. The CLI can be used with the Full Node RPC provided by coinset.org, or any full node.

![Demo GIF](./docs/demo.gif)


## Installation

Prebuilt binaries can be downloaded from the [Releases](https://github.com/coinset-org/cli/releases) page, or installed with *brew* or *go*.

### Using Homebrew

```bash
brew install coinset-org/cli/coinset
```

### Go Install

```bash
go install github.com/coinset-org/cli/cmd/coinset@latest
```

### Build from Source
See the detailed instructions for your operating system:
* [MacOS](docs/install_from_source.md#macos)
* [Linux](docs/install_from_source.md#linux)
* [Windows](docs/install_from_source.md#windows)

## Usage
Once installed you can access the Full Node RPC hosted at coinset.org. The first argument is the RPC name and output is automatically pretty printed. For example

```bash
$ coinset get_network_info
```
```json
{
  "network_name": "mainnet",
  "network_prefix": "xch",
  "success": true
}	
```

Required RPC parameters can be passed in as arguments and optional parameters can be set with flags. For example:

```bash
$ coinset get_coin_records_by_parent_ids 0xa908ee64a5821b7bda5d798c053a79c8b3d7c608bb7735f4cefc7833ead4f6cd --include-spent-coins
```
```json
{
  "coin_records": [
    {
      "coin": {
        "amount": 300000,
        "parent_coin_info": "0xa908ee64a5821b7bda5d798c053a79c8b3d7c608bb7735f4cefc7833ead4f6cd",
        "puzzle_hash": "0x5e2bb312cff523f00b286865fedf78209755109c627022d68ccc891ede1d5da9"
      },
      "coinbase": false,
      "confirmed_block_index": 4907529,
      "spent": true,
      "spent_block_index": 4907534,
      "timestamp": 1707250407
    },
    {
      "coin": {
        "amount": 10000,
        "parent_coin_info": "0xa908ee64a5821b7bda5d798c053a79c8b3d7c608bb7735f4cefc7833ead4f6cd",
        "puzzle_hash": "0x352928569cc16f72c212d95912adca3e634a3de136b85ed396a76b19e684e2f6"
      },
      "coinbase": false,
      "confirmed_block_index": 4907529,
      "spent": false,
      "spent_block_index": 0,
      "timestamp": 1707250407
    }
  ],
  "success": true
}
```

### Coin IDs

Using the `coin_id` command you can encode coin IDs. For example:

```bash
coinset coin_id 0xeca65946d1b80b527bcab5e94673f30bb3fd8a9466b31379fa5fa1f49c492031 0x66e55285340258cb79e6eda4d16f230bec2df7a2d7b40b8c6268247be9e659cb 2000000007
```
```bash
0xce2a2dd052bdbcf7fffc309e2a5e1f8589513335f85232a6c8cfb4a7d49ee32b
```

### Address Encoding

Using the `address` command you can encode and decode adresses. For example:

```bash
coinset address encode 0xbf3d35bba83d984be6cc4db0d6c84922e275a39ca4f8e1dd3cddfe2fa5eb2e2f
```
```bash
xch1hu7ntwag8kvyhekvfkcddjzfyt38tguu5nuwrhfumhlzlf0t9chs6cj5k8
```

```bash
coinset address decode xch1hu7ntwag8kvyhekvfkcddjzfyt38tguu5nuwrhfumhlzlf0t9chs6cj5k8
```
```bash
0xbf3d35bba83d984be6cc4db0d6c84922e275a39ca4f8e1dd3cddfe2fa5eb2e2f
```

### JQ Filtering

Using the `-q` option you can pass in a jq filter to be used on the output. For example:

```bash
coinset get_blockchain_state -q .blockchain_state.peak.height
```
```json
4911276
```

### Autocomplete

Autocompletions are provided and can be generated with `coinset completion <shell>`. Completions are automatically installed when using *brew* but brew completions for your shell need to be enabled. [Instructions can be found here](https://docs.brew.sh/Shell-Completion).



Manual installation instructions can be found with the help flag. For example:

```bash
coinset completion zsh --help
```
```
Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(coinset completion zsh); compdef _coinset coinset

To load completions for every new session, execute once:

#### Linux:

	coinset completion zsh > "${fpath[1]}/_coinset"

#### macOS:

	coinset completion zsh > $(brew --prefix)/share/zsh/site-functions/_coinset

You will need to start a new shell for this setup to take effect.
```




