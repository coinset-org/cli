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

### Coinset API Host

All RPC endpoints (full node and Coinset query APIs) use the same API host configuration.

- Override with `--api`

## Available Commands

The following table shows all available CLI commands organized by functionality:

| Category | Command | Description |
|----------|---------|-------------|
| **Address Operations** | `address encode <puzzle_hash>` | Encode puzzle hash to address |
| | `address decode <address>` | Decode address to puzzle hash |
| **Coin Operations** | `coin_id <parent_coin_id> <puzzle_hash> <amount>` | Compute a coin ID from parent, puzzle hash and amount |
| | `get_coin_record_by_name <coin_name>` | Retrieve a coin record by its name |
| | `get_coin_records_by_hint <hint>` | Retrieve coin records by hint |
| | `get_coin_records_by_hints <hints>` | Retrieve coin records by multiple hints |
| | `get_coin_records_by_names <coin_names>` | Retrieve coin records by multiple names |
| | `get_coin_records_by_parent_ids <parent_ids>` | Retrieve coin records by parent IDs |
| | `get_coin_records_by_puzzle_hash <puzzle_hash_or_address>` | Retrieve coin records by puzzle hash or address |
| | `get_coin_records_by_puzzle_hashes <puzzle_hash_or_address> ...` | Retrieve coin records by multiple puzzle hashes or addressees |
| **Block Operations** | `get_block <header_hash>` | Retrieve a full block by header hash |
| | `get_blocks <start_height> <end_height>` | Retrieve multiple blocks in a height range |
| | `get_block_record <header_hash>` | Retrieve a block record by header hash |
| | `get_block_records <start_height> <end_height>` | Retrieve multiple block records in a height range |
| | `get_block_record_by_height <height>` | Retrieve a block record by height |
| | `get_block_spends <header_hash>` | Retrieve block spends by header hash |
| | `get_block_spends_with_conditions <header_hash>` | Retrieve block spends with conditions by header hash |
| | `get_unfinished_block_headers` | Retrieve unfinished block headers |
| **Blockchain State & Network** | `get_blockchain_state` | Retrieve the current blockchain state |
| | `get_network_info` | Retrieve information about the current network |
| | `get_network_space <older_block_header_hash> <newer_block_header_hash>` | Retrieve network space between two blocks |
| | `get_block_count_metrics` | Retrieve block count metrics |
| **Mempool Operations** | `get_all_mempool_items` | Retrieve all mempool items |
| | `get_all_mempool_tx_ids` | Retrieve all mempool transaction IDs |
| | `get_mempool_item_by_tx_id <tx_id>` | Retrieve a specific mempool item by transaction ID |
| | `get_mempool_items_by_coin_name <coin_name>` | Retrieve mempool items by coin name |
| **Transaction Operations (Full Node)** | `push_tx [spend_bundle_json]` | Push a spend bundle to the mempool |
| | `get_fee_estimate [cost_or_spend_bundle]` | Get fee estimate by cost, spend bundle, or spend type |
| | `get_fee_tiers_by_cost <cost>` | Get economy/standard/urgent fee tiers |
| | `get_additions_and_removals <header_hash>` | Get additions and removals for a block |
| | `get_aggsig_additional_data <spend_bundle>` | Get aggregate signature additional data |
| **Coinset Transactions** | `get_transaction <tx_id>` | Get a transaction by tx id |
| | `get_transactions_by_coin_name <coin_name>` | Get created/spent transaction references for a coin |
| | `get_raw_transaction_by_id <tx_id>` | Get raw mempool-style transaction details |
| | `get_block_transactions <height_or_header_hash>` | List transactions for a block |
| | `get_transactions_by_p2 <p2_or_address>` | List confirmed transactions by p2 |
| | `get_pending_transactions_by_p2 <p2_or_address>` | List pending transactions by p2 |
| | `get_transactions_by_puzzle_hash <puzzle_hash_or_address>` | Alias for transactions by p2 |
| | `get_pending_transactions_by_puzzle_hash <puzzle_hash_or_address>` | Alias for pending transactions by p2 |
| | `get_transactions_by_cat_asset_id <asset_id>` | List transactions by CAT asset id |
| | `get_transactions_by_nft_id <nft_id>` | List transactions by NFT id |
| **Coinset Offers** | `get_offer <offer_id>` | Get offer summary/state by offer id |
| | `get_offers_by_p2 <p2_or_address>` | List offers by maker p2 |
| | `get_offers_by_cat_asset_id <asset_id>` | List offers by CAT asset id |
| | `get_offers_by_nft_id <nft_id>` | List offers by NFT id |
| **Coinset Balances, Assets & NFT** | `get_xch_balance_by_p2 <p2_or_address>` | Get XCH balance by p2 |
| | `get_cat_balances_by_p2 <p2_or_address>` | List CAT balances by p2 |
| | `get_nft_balance_by_p2 <p2_or_address>` | Get NFT balance by p2 |
| | `get_clawback_coins_by_receiver <p2_or_address>` | List clawback coins by receiver p2 |
| | `get_cat_assets_by_p2 <p2_or_address>` | List CAT asset ids by p2 |
| | `get_cat_coins_by_asset_id <asset_id>` | List CAT coins by asset id |
| | `get_cat_coins_by_p2_and_asset_id <p2_or_address> <asset_id>` | List CAT coins by p2 and asset id |
| | `get_nft_ids_by_p2 <p2_or_address>` | List NFT ids by p2 |
| | `get_nft_coins_by_nft_id <nft_id>` | List NFT coins by NFT id |
| | `get_latest_nft_coin_by_nft_id <nft_id>` | Get latest NFT coin by NFT id |
| | `get_singleton_info <launcher_id>` | Get singleton info by launcher id |
| **Puzzle & Solution** | `get_puzzle_and_solution <coin_name>` | Retrieve puzzle and solution for a coin |
| | `get_puzzle_and_solution_with_conditions <coin_name>` | Retrieve puzzle and solution with conditions for a coin |
| **Memos & Metadata** | `get_memos_by_coin_name <coin_name>` | Retrieve memos associated with a coin |
| **Events & Real-time** | `events [type]` | Connect to WebSocket and display events (`peak`, `transaction`, `offer`, `balance`, `dashboard`) |
| **Utility** | `version` | Display the version number of coinset |

### Smoke Test Checklist

- `coinset get_coin_records_by_puzzle_hash <puzzle_hash_or_address> --limit 5 --order desc`
- `coinset get_additions_and_removals <height_or_header_hash> --limit 10 --filter additions`
- `coinset get_fee_estimate --spend-type send_xch_transaction --times 60,300 --spend-count 1`
- `coinset get_fee_tiers_by_cost 9401710`
- `coinset --api https://api.coinset.org get_transaction <tx_id>`
- `coinset get_offers_by_p2 <p2_or_address> --status open --limit 20`
- `coinset get_cat_balances_by_p2 <p2_or_address> --limit 20`
- `coinset events --events peak,transaction --tx-status confirmed`


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




