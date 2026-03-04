# coinset CLI

Use this tool to query, inspect, and investigate the Chia blockchain. `coinset` wraps the Chia Full Node RPC (hosted at api.coinset.org or any full node) and includes built-in spend inspection and local CLVM utilities.

## When to use this skill

- Query blockchain state, blocks, coins, mempool items
- Inspect spend bundles and coin spends (conditions, cost, puzzle recognition)
- Convert between addresses and puzzle hashes
- Compute coin IDs
- Decompile, compile, or run CLVM programs
- Monitor real-time blockchain events

## Command grammar

```
coinset <command> [args...] [flags...]
```

### Global flags

| Flag | Short | Description |
|------|-------|-------------|
| `--query <jq>` | `-q` | Apply a jq filter to output (default: `.`) |
| `--raw` | `-r` | Output raw JSON (no color/formatting) |
| `--describe` | `-d` | Add human-readable `_description` fields (XCH amounts, relative timestamps) |
| `--inspect` | | Replace output with interpreted spend analysis (only works when output contains spend data) |
| `--testnet` | `-t` | Use testnet11 (`https://testnet11.api.coinset.org`) |
| `--local` | `-l` | Use local full node (auto-config) |
| `--api <url>` | `-a` | Custom API endpoint (mutually exclusive with `--testnet`/`--mainnet`) |
| `--mainnet` | | Use mainnet (default) |

### Input normalization

- **Puzzle hash inputs** accept either `xch…`/`txch…` addresses or `0x…` hex. The CLI normalizes to `0x…` hex.
- **Block selectors** accept either a numeric height or a `0x…` header hash. Heights are resolved to header hashes automatically.
- **Hex values** accept with or without `0x` prefix.

## Commands by intent

### Find coins

```bash
coinset get_coin_record_by_name <0xCOIN_ID>
coinset get_coin_records_by_puzzle_hash <puzzle_hash_or_address> [--include-spent-coins] [--start-height N] [--end-height N]
coinset get_coin_records_by_puzzle_hashes <ph1> <ph2> ... [--include-spent-coins]
coinset get_coin_records_by_parent_ids <parent_id> ... [--include-spent-coins]
coinset get_coin_records_by_hint <hint> [--include-spent-coins]
coinset get_coin_records_by_hints <hint1> <hint2> ... [--include-spent-coins]
coinset get_coin_records_by_names <name1> <name2> ... [--include-spent-coins]
```

By default, coin record queries return only **unspent** coins. Pass `--include-spent-coins` / `-s` to include spent coins.

### Inspect spends

```bash
coinset get_puzzle_and_solution <0xCOIN_ID>
coinset get_puzzle_and_solution_with_conditions <0xCOIN_ID>
coinset get_block_spends <height_or_header_hash>
coinset get_block_spends_with_conditions <height_or_header_hash>
coinset get_mempool_item_by_tx_id <0xTX_ID>
```

All of these support `--inspect` to replace the raw JSON output with a structured spend analysis.

### Block and chain state

```bash
coinset get_blockchain_state
coinset get_network_info
coinset get_block_record_by_height <height>
coinset get_block_record <header_hash>
coinset get_block_records <start_height> <end_height>
coinset get_block <header_hash>
coinset get_blocks <start_height> <end_height>
coinset get_block_count_metrics
coinset get_network_space <newer_header_hash> <older_header_hash>
coinset get_additions_and_removals <header_hash>
```

### Mempool

```bash
coinset get_all_mempool_tx_ids
coinset get_all_mempool_items
coinset get_mempool_item_by_tx_id <0xTX_ID>
coinset get_mempool_items_by_coin_name <0xCOIN_NAME>
```

### Transaction submission

```bash
coinset push_tx '<spend_bundle_json>'
coinset push_tx ./spend_bundle.json
coinset push_tx -f ./spend_bundle.json
```

### Fee estimation

```bash
coinset get_fee_estimate <target_times> <cost>
```

### Utilities

```bash
coinset address encode <0xPUZZLE_HASH>
coinset address decode <xch_address>
coinset coin_id <0xPARENT_COIN_ID> <0xPUZZLE_HASH> <amount>
coinset get_memos_by_coin_name <0xCOIN_NAME>
coinset get_aggsig_additional_data <spend_bundle>
```

### CLVM tools

```bash
coinset clvm decompile <hex_bytes>
coinset clvm compile "<clvm_s_expression>"
coinset clvm run "<program>" "<env>" [--cost] [--max-cost N]
coinset clvm run --program "<program>" --env "<env>" [--cost] [--max-cost N]
coinset clvm tree_hash <program>
coinset clvm curry <mod> [arg1] [arg2] ... [--atom val] [--tree-hash val] [--program val]
coinset clvm uncurry <program>
```

All CLVM subcommands accept both hex bytes and s-expressions as input. The CLI auto-detects the format.

#### `curry` arg type handling

The `curry` command supports typed arguments for computing curried puzzle hashes:

- **Mod (first positional arg)**: if 32 bytes hex, treated as the mod's tree hash (hash-only mode). Otherwise treated as a full serialized program.
- **Positional curry args**: if 32 bytes hex, treated as an atom. Otherwise treated as a serialized program.
- **`--atom <val>`**: explicitly mark an arg as a raw atom value. Tree hash = `sha256(1, bytes)`.
- **`--tree-hash <val>`**: arg is already a tree hash. Used directly -- no hashing applied.
- **`--program <val>`**: explicitly mark an arg as a serialized CLVM program.

Flag args are appended after positional args in order: `--atom`, then `--tree-hash`, then `--program`.

**Output**: when all inputs have full bytes, `curried.bytes` and `curried.tree_hash` are both returned. When any input is a tree hash (mod is 32 bytes, or `--tree-hash` is used), only `curried.tree_hash` is returned (full bytes cannot be computed).

The key distinction: an atom arg like a TAIL hash is a 32-byte data value curried into the puzzle as a leaf node. A tree-hash arg like an inner puzzle hash represents a full subtree. Even if both are 32-byte hex, they contribute differently to the curried puzzle hash.

### Real-time events

```bash
coinset events              # all events
coinset events peak         # new block peaks only
coinset events transaction  # mempool transactions
coinset events offer        # offer activity
```

## `--inspect` output schema

When `--inspect` is used, the output follows `coinset.inspect.v1`:

```
{
  "schema_version": "coinset.inspect.v1",
  "tool": { "name": "coinset-inspect", "version": "..." },
  "input": { "kind": "mempool|block|coin", "notes": [...] },
  "result": {
    "status": "ok|failed",
    "error": null | { "kind": "...", "message": "..." },
    "summary": {
      "removals": [{ "coin_id", "parent_coin_id", "puzzle_hash", "amount" }],
      "additions": [{ "coin_id", "parent_coin_id", "puzzle_hash", "amount" }],
      "fee_mojos": <number>,
      "net_xch_delta_by_puzzle_hash": [{ "puzzle_hash", "delta_mojos" }]
    },
    "spends": [{
      "coin_spend": { "coin": {...}, "puzzle_reveal": "0x...", "solution": "0x..." },
      "evaluation": {
        "status": "ok|failed",
        "cost": <number>,
        "conditions": [{ "opcode": "CREATE_COIN", "args": [...] }, ...],
        "additions": [...],
        "failure": null | { "kind", "message" }
      },
      "clvm": {
        "puzzle_opd": "<disassembled puzzle>",
        "solution_opd": "<disassembled solution>",
        "tree_hash": "0x...",
        "uses_backrefs": true|false
      },
      "puzzle_recognition": {
        "recognized": true|false,
        "candidates": [{ "name", "confidence", "source_repo", "source_path", "source_ref" }],
        "wrappers": [{
          "name": "cat_layer|singleton_layer|...",
          "mod_hash": "0x...",
          "params": { ... },
          "inner_puzzle_tree_hash": "0x...",
          "parse_error": null | "..."
        }],
        "parsed_solution": { "layers": [...] }
      }
    }],
    "signatures": {
      "aggregated_signature": "0x...",
      "agg_sig_me": [{ "pubkey", "msg" }],
      "agg_sig_unsafe": [{ "pubkey", "msg" }]
    }
  }
}
```

### Supported `--inspect` input shapes

`--inspect` works when the RPC response contains one of:
- `mempool_item` (with nested `spend_bundle` or `spend_bundle_bytes`)
- `spend_bundle` / `spend_bundle_bytes`
- `coin_spends` / `block_spends` (array of coin spends)
- `coin_spend` or a single object with `puzzle_reveal` + `solution`
- `mempool_items` (map of tx_id -> mempool_item; each is inspected individually)

If the output shape doesn't match, `--inspect` returns an error. Fall back to the command without `--inspect`.

### Recognized puzzle layers

The inspector peels puzzle layers from outside in:

| Layer name | What it means |
|---|---|
| `cat_layer` | CAT (Chia Asset Token). Params include `asset_id`. |
| `singleton_layer` | Singleton wrapper. Params include `launcher_id`. |
| `did_layer` | DID (Decentralized Identity). Params include `launcher_id`. |
| `nft_state_layer` | NFT state metadata layer. |
| `nft_ownership_layer` | NFT ownership/transfer layer. Params include `current_owner`. |
| `royalty_transfer_layer` | NFT royalty enforcement. Params include `royalty_basis_points`. |
| `augmented_condition_layer` | Prepends extra conditions to inner puzzle output. |
| `bulletin_layer` | On-chain bulletin board data layer. |
| `option_contract_layer` | Option contract wrapper. |
| `revocation_layer` | Clawback/revocation custody. Params include `hidden_puzzle_hash`. |
| `p2_singleton_layer` | Pay-to-singleton (pool farming reward target). |
| `p2_curried_layer` | Pay-to-curried-puzzle. |
| `p2_one_of_many_layer` | Merkle-tree multisig or multi-path spend. |
| `p2_delegated_conditions_layer` | Direct delegated conditions with public key. |
| `settlement_layer` | Offer settlement payments puzzle. |
| `stream_layer` | Payment streaming / vesting. |
| `standard_layer` | Standard transaction (`p2_delegated_puzzle_or_hidden_puzzle`). Params include `synthetic_key`. This is typically the innermost layer. |

Recognition is best-effort. If a `parse_error` appears on a wrapper, keep the successfully parsed layers but treat confidence as partial. If recognition returns `recognized: false`, rely on raw conditions and CLVM disassembly.

## Workflows

### Look up a coin and check if it was spent

```bash
coinset get_coin_record_by_name 0xCOIN_ID -q '.coin_record'
```

Check `spent` (boolean) and `spent_block_index` in the result.

### Get puzzle and solution for a spent coin

```bash
coinset get_puzzle_and_solution 0xCOIN_ID --inspect
```

### Inspect all spends in a block

```bash
coinset get_block_spends 5000000 --inspect
```

### Inspect a mempool transaction

```bash
coinset get_mempool_item_by_tx_id 0xTX_ID --inspect
```

### Trace coin lineage (parent -> children)

```bash
# Find children of a coin
coinset get_coin_records_by_parent_ids 0xPARENT_COIN_ID --include-spent-coins

# Find the parent's spend (to see what created this coin)
coinset get_puzzle_and_solution 0xPARENT_COIN_ID --inspect
```

### Decompile a puzzle reveal

```bash
coinset clvm decompile 0xff02ffff01ff...
```

### Run a CLVM program and see cost

```bash
coinset clvm run "(+ (q . 2) (q . 3))" "()" --cost
```

### Compute a puzzle hash

```bash
coinset clvm tree_hash ff02ffff01ff02ffff03...
```

`tree_hash` computes the sha256 tree hash of a CLVM program. This is the puzzle hash used to identify coins on-chain. Accepts hex or s-expression input.

### Curry and uncurry puzzles

```bash
# Curry arguments into a puzzle template (full bytes)
coinset clvm curry <mod_hex_or_sexpr> <arg1> <arg2> ...

# Recover the mod and curried arguments from a full puzzle
coinset clvm uncurry <curried_program>
```

`uncurry` output includes `kind` ("atom" or "tree") and `value` (raw atom bytes without CLVM serialization prefix) for each arg. For atom args, `value` is directly matchable against known hashes and IDs without needing to strip CLVM serialization prefixes:

```bash
# 1. Get the puzzle reveal from a spend
coinset get_puzzle_and_solution <0xCOIN_ID> -q '.coin_solution.puzzle_reveal' --raw

# 2. Uncurry it to see the base mod and curried args
coinset clvm uncurry <puzzle_reveal_hex>

# 3. Match mod_.tree_hash against known puzzle hashes (see "Known puzzle mod hashes" section)
```

### Compute a CAT puzzle hash from known hashes

The most common use case for `curry` with `--tree-hash` is computing CAT puzzle hashes. When sending someone a CAT, you only need their address (inner puzzle hash) -- the full inner puzzle is never needed to compute the puzzle hash.

```bash
# Compute CAT puzzle hash for a specific asset + owner
# MOD_HASH and TAIL_HASH are atoms (32 bytes → auto-detected)
# Inner puzzle hash needs --tree-hash since it's a tree hash, not an atom
coinset clvm curry 0x37bef360ee858133b69d595a906dc45d01af50379dad515eb9518abb7c1d2a7a \
  0x37bef360ee858133b69d595a906dc45d01af50379dad515eb9518abb7c1d2a7a \
  0x<TAIL_HASH> \
  --tree-hash 0x<INNER_PUZZLE_HASH> \
  -q '.curried.tree_hash'
```

### Compute a CAT settlement puzzle hash

Find all offer coins for a specific CAT asset:

```bash
# Settlement payments has no curried args, so its mod hash IS its tree hash
coinset clvm curry 0x37bef360ee858133b69d595a906dc45d01af50379dad515eb9518abb7c1d2a7a \
  0x37bef360ee858133b69d595a906dc45d01af50379dad515eb9518abb7c1d2a7a \
  0x<TAIL_HASH> \
  --tree-hash 0xcfbfdeed5c4ca2de3d0bf520b9cb4bb7743a359bd2e6a188d19ce7dffc21d3e7 \
  -q '.curried.tree_hash'

# Then search for offer coins
coinset get_coin_records_by_puzzle_hash <result>
```

### Lineage proofs in CAT solutions

When investigating CAT spends, the solution includes a **lineage proof** to verify the parent was a legitimate CAT of the same type. The proof contains `(parent_parent_coin_id, parent_inner_puzzle_hash, parent_amount)`. This allows tracing CAT coin lineage backwards through the chain.

### Filter output with jq

```bash
# Get current peak height
coinset get_blockchain_state -q .blockchain_state.peak.height

# Get just the fee from a mempool item
coinset get_mempool_item_by_tx_id 0xTX_ID -q .mempool_item.fee

# Count coins for an address
coinset get_coin_records_by_puzzle_hash xch1... -q '.coin_records | length'
```

### Human-readable output

```bash
coinset get_coin_record_by_name 0xCOIN_ID --describe
```

Adds `amount_description` (e.g., `"0.000300000000 XCH"`), `timestamp_description` (e.g., `"2 hours ago, 2024-02-06 15:33:27"`), and `confirmed_block_index_description` fields.

### Direct HTTP fallback

If you need to call the Full Node RPC directly:

```bash
# Mainnet
curl -s https://api.coinset.org/get_blockchain_state -d '{}'

# Testnet11
curl -s https://testnet11.api.coinset.org/get_blockchain_state -d '{}'
```

## Chia blockchain concepts

### The coin model

Chia uses a **coin set** model (analogous to Bitcoin's UTXO model). There are no accounts or balances at the protocol level -- only coins.

A **coin** is defined by three values:
- `parent_coin_info` -- the coin ID of the coin that created this one
- `puzzle_hash` -- the hash of the CLVM program (puzzle) that locks this coin
- `amount` -- value in mojos (1 XCH = 1,000,000,000,000 mojos)

The **coin ID** (also called coin name) is `sha256(parent_coin_info + puzzle_hash + amount)`. It uniquely identifies a coin.

A **coin record** is the blockchain's view of a coin: the coin itself plus metadata (confirmed block index, timestamp, spent status, spent block index).

### Spending coins

To spend a coin, you must provide:
- **puzzle reveal** -- the full CLVM program whose hash matches the coin's `puzzle_hash`
- **solution** -- CLVM data passed as input to the puzzle

The puzzle is executed with the solution as its environment. If execution succeeds, it returns a list of **conditions** that dictate what happens.

A **coin spend** is the tuple `(coin, puzzle_reveal, solution)`.

A **spend bundle** is a group of coin spends plus an `aggregated_signature` (BLS12-381). All coin spends in a bundle are atomic -- they either all succeed or all fail.

### Additions, removals, and fees

- **Removals** are the coins being spent (inputs)
- **Additions** are the coins being created via `CREATE_COIN` conditions (outputs)
- **Fee** = sum of removal amounts - sum of addition amounts

### Addresses and puzzle hashes

A Chia **address** (e.g., `xch1...`) is the bech32m encoding of a puzzle hash. `xch` prefix is mainnet, `txch` is testnet. They are interchangeable representations of the same 32-byte value.

### Hints and memos

When a `CREATE_COIN` condition includes an optional third argument (a list), the first element is the **hint** if it's exactly 32 bytes. Hints typically contain the inner puzzle hash (p2) of the created coin, allowing wallets and indexers to find coins that belong to them even when the outer puzzle hash differs (as with CATs, NFTs, etc.).

`get_coin_records_by_hint` uses these hints to find coins. `get_memos_by_coin_name` retrieves the full memo list.

### Mempool vs confirmed

**Mempool items** are unconfirmed transactions waiting to be included in a block. They can be evicted, replaced, or reorganized away. **Block spends** are confirmed and considered final (modulo chain reorgs, which are extremely rare on Chia).

Prefer confirmed block data for definitive analysis. Use mempool data for investigating pending transactions.

## CLVM essentials

### Data model

CLVM (Chialisp Virtual Machine) operates on a binary tree of **atoms** and **cons pairs**:
- An **atom** is a sequence of bytes (numbers, hashes, public keys, or empty)
- A **cons pair** is `(first . rest)` -- two children
- A **list** is a chain of cons pairs ending in nil: `(a . (b . (c . nil)))`
- **Nil** is the empty atom, represented as `()` in text or `0x80` in serialized form

### Programs and evaluation

A CLVM program is itself a tree. When evaluated:
- If it's an **atom**, it's a path lookup into the environment (1 = whole env, 2 = first, 3 = rest, 5 = first of rest, etc.)
- If it's a **cons pair**, the left element is the operator and the right elements are operands
- The **quote operator** `q` returns its argument as a literal value without evaluating it
- The **apply operator** `a` runs a program with a given environment

### Serialization

Puzzle reveals and solutions are hex-encoded serialized CLVM:
- `0xFF` prefix = cons pair (followed by first, then rest)
- `0x80` = nil
- `0x00`-`0x7F` = single-byte atom
- Larger atoms use length-prefixed encoding

Use `coinset clvm decompile <hex>` to convert to readable s-expression form. Use `coinset clvm tree_hash <program>` to compute the sha256 tree hash of a serialized program -- this is how puzzle hashes and coin IDs are derived.

### Currying

Currying pre-bakes parameters into a puzzle template before deployment. In Chialisp source, `SCREAMING_SNAKE_CASE` parameters are conventionally meant to be curried in. For example, the standard transaction curries in a `SYNTHETIC_PUBLIC_KEY`. The CAT puzzle curries in `MOD_HASH`, `TAIL_PROGRAM_HASH`, and `INNER_PUZZLE`.

The curried puzzle is a new program with a different tree hash than the uncurried template. Crucially, the curried puzzle hash is deterministic from the mod hash and the tree hashes of the curried arguments -- you don't need the full serialized programs to compute it. This is how wallets derive puzzle hashes for coins they create: they know the mod hash (a constant) and the tree hashes of the args (derived from the recipient's address, asset ID, etc.).

Use `coinset clvm curry <mod> [args...]` to curry arguments into a puzzle template, and `coinset clvm uncurry <program>` to reverse the process -- recovering the original mod and the curried arguments. Uncurrying a `puzzle_reveal` is one of the most useful investigation techniques: it separates the reusable puzzle template from the instance-specific parameters (public keys, asset IDs, launcher IDs, etc.).

When currying, each arg is either an **atom** (a data value like a hash or key) or a **tree** (a full CLVM program like an inner puzzle). Atoms contribute `sha256(1, atom_bytes)` to the curried tree hash, while trees contribute their recursive tree hash directly. The `--atom`, `--tree-hash`, and `--program` flags on `curry` control this distinction.

### Cost

Every CLVM operation has a cost. Key numbers:
- **Max cost per block**: 11,000,000,000 (11 billion)
- **Per-byte cost**: 12,000 per byte of serialized puzzle + solution
- **AGG_SIG_\* condition**: 1,200,000 each
- **CREATE_COIN condition**: 1,800,000 each
- **Fee priority**: mojos per cost. When the mempool is full, ~5 mojos/cost is needed to displace existing transactions.

`coinset clvm run ... --cost` and `--inspect` both report execution cost.

### Common operators

| Operator | Description |
|---|---|
| `q` (quote) | Return value literally |
| `a` (apply) | Run program with environment |
| `i` (if) | Conditional branch |
| `c` (cons) | Create pair |
| `f` (first) | First element of pair |
| `r` (rest) | Rest element of pair |
| `+`, `-`, `*`, `/` | Arithmetic |
| `=` | Equality test |
| `sha256` | SHA-256 hash |
| `concat` | Concatenate bytes |
| `substr` | Slice bytes |
| `logand`, `logior`, `logxor`, `lognot` | Bitwise operations |
| `pubkey_for_exp` | Secret key to G1 public key |
| `g1_add`, `g1_multiply` | BLS G1 point operations |
| `coinid` | Compute coin ID with validation (CHIP-11) |
| `softfork` | Future-proof operator for soft-forked features |
| `x` (exit) | Terminate program / raise error |

## Conditions reference

When a puzzle executes, it returns a list of conditions. Each condition is a list starting with an opcode number.

### Coin creation

| Opcode | Name | Format | Description |
|---|---|---|---|
| 51 | `CREATE_COIN` | `(51 puzzle_hash amount (...memos)?)` | Create a new coin. Optional memos list; first 32-byte memo is the hint. Cost: 1,800,000. |
| 52 | `RESERVE_FEE` | `(52 amount)` | Assert minimum fee in this transaction. |

### Signatures

All AGG_SIG conditions cost 1,200,000 each.

| Opcode | Name | Description |
|---|---|---|
| 49 | `AGG_SIG_UNSAFE` | Verify signature on raw message. No domain separation -- can be replayed. |
| 50 | `AGG_SIG_ME` | Verify signature on message + coin_id + genesis_id. Recommended for most uses. |
| 43-48 | `AGG_SIG_PARENT` / `AGG_SIG_PUZZLE` / `AGG_SIG_AMOUNT` / `AGG_SIG_PUZZLE_AMOUNT` / `AGG_SIG_PARENT_AMOUNT` / `AGG_SIG_PARENT_PUZZLE` | CHIP-11 domain-separated signature variants binding to specific coin attributes. |

### Announcements (legacy)

| Opcode | Name | Description |
|---|---|---|
| 60 | `CREATE_COIN_ANNOUNCEMENT` | Create announcement bound to this coin's ID. |
| 61 | `ASSERT_COIN_ANNOUNCEMENT` | Assert `sha256(coin_id + message)` was announced. |
| 62 | `CREATE_PUZZLE_ANNOUNCEMENT` | Create announcement bound to this coin's puzzle hash. |
| 63 | `ASSERT_PUZZLE_ANNOUNCEMENT` | Assert `sha256(puzzle_hash + message)` was announced. |

### Messages (modern replacement for announcements)

| Opcode | Name | Description |
|---|---|---|
| 66 | `SEND_MESSAGE` | `(66 mode message ...)` -- Send message with sender/receiver commitment via mode bitmask. |
| 67 | `RECEIVE_MESSAGE` | `(67 mode message ...)` -- Assert receipt of matching message. |

The `mode` byte is 6 bits: 3 bits for sender commitment, 3 bits for receiver commitment. Each 3-bit group encodes which coin attributes (parent, puzzle, amount) to commit to. For example, `mode=0b111110` means sender commits to all three (coin ID) and receiver commits to parent+puzzle.

### Self-assertions

| Opcode | Name | Description |
|---|---|---|
| 70 | `ASSERT_MY_COIN_ID` | Assert this coin's ID matches. |
| 71 | `ASSERT_MY_PARENT_ID` | Assert this coin's parent ID matches. |
| 72 | `ASSERT_MY_PUZZLEHASH` | Assert this coin's puzzle hash matches. |
| 73 | `ASSERT_MY_AMOUNT` | Assert this coin's amount matches. |
| 74 | `ASSERT_MY_BIRTH_SECONDS` | Assert coin creation timestamp. |
| 75 | `ASSERT_MY_BIRTH_HEIGHT` | Assert coin creation height. |
| 76 | `ASSERT_EPHEMERAL` | Assert this coin was created in the current block. |

### Timelocks

| Opcode | Name | Description |
|---|---|---|
| 80 | `ASSERT_SECONDS_RELATIVE` | Min seconds since coin creation. |
| 81 | `ASSERT_SECONDS_ABSOLUTE` | Min absolute timestamp. |
| 82 | `ASSERT_HEIGHT_RELATIVE` | Min blocks since coin creation. |
| 83 | `ASSERT_HEIGHT_ABSOLUTE` | Min absolute block height. |
| 84 | `ASSERT_BEFORE_SECONDS_RELATIVE` | Max seconds since coin creation. |
| 85 | `ASSERT_BEFORE_SECONDS_ABSOLUTE` | Max absolute timestamp. |
| 86 | `ASSERT_BEFORE_HEIGHT_RELATIVE` | Max blocks since coin creation. |
| 87 | `ASSERT_BEFORE_HEIGHT_ABSOLUTE` | Max absolute block height. |

### Concurrency

| Opcode | Name | Description |
|---|---|---|
| 64 | `ASSERT_CONCURRENT_SPEND` | Assert another specific coin is spent in same block. |
| 65 | `ASSERT_CONCURRENT_PUZZLE` | Assert a coin with specific puzzle hash is spent in same block. |

### Other

| Opcode | Name | Description |
|---|---|---|
| 1 | `REMARK` | No-op; always valid. |
| 90 | `SOFTFORK` | Soft-fork guard for future conditions. Cost is specified in ten-thousands. |

## Puzzle composition

Chia puzzles compose through **outer/inner puzzle wrapping**. An outer puzzle enforces protocol rules and delegates ownership decisions to an inner puzzle. The inner puzzle is typically the **standard transaction** at the bottom of the stack.

When the inner puzzle emits `CREATE_COIN` conditions, the outer puzzle intercepts them and wraps the created coin's puzzle hash with itself, ensuring output coins maintain the same wrapper structure.

### Standard transaction

The innermost puzzle for most coins. Curries in a `SYNTHETIC_PUBLIC_KEY`. Spending requires either a delegated puzzle signed by the synthetic key, or revealing a hidden puzzle. This is what `xch…` addresses resolve to.

### CAT (Chia Asset Token)

Fungible token wrapper. Curries in `MOD_HASH`, `TAIL_PROGRAM_HASH`, and `INNER_PUZZLE`. Enforces supply conservation via a ring of coin announcements across all CAT spends in a bundle. The **TAIL** (Token and Asset Issuance Limitations) program controls minting and melting rules. The `asset_id` is the TAIL program hash.

Amounts: 1 CAT = 1,000 mojos of the underlying coin.

### Singleton

Guarantees a unique, traceable on-chain identity via a `launcher_id`. Always produces exactly one odd-valued output coin (the continuation), or an output of -113 to self-destruct. Used as the foundation for NFTs, DIDs, plot NFTs, and vaults.

### NFT

An NFT is a singleton containing multiple nested layers:
- `singleton_layer` (uniqueness)
- `nft_state_layer` (metadata: URIs, hash, updater puzzle)
- `nft_ownership_layer` (current owner DID, transfer program)
- `royalty_transfer_layer` (royalty enforcement on trades)
- Inner puzzle (typically standard transaction)

### DID (Decentralized Identity)

A singleton-based identity. The `did_layer` curries in a `launcher_id`, optional recovery list, and metadata. Used as an ownership identity for NFTs and other assets.

### Offers

Offers enable peer-to-peer, trustless, atomic asset exchange on Chia. A **maker** creates a partial spend bundle describing what they're offering and what they want in return. A **taker** completes the spend bundle with their side and pushes it to the network. Both sides execute in the same block or not at all -- there is no counterparty risk.

**Settlement payments puzzle**: Offers work through the `settlement_payments` puzzle (recognized by `--inspect` as `settlement_layer`). Its solution is a list of `notarized_payments`, each structured as `(nonce . ((puzzle_hash amount ...memos) ...))`. For each entry, the puzzle creates a `CREATE_PUZZLE_ANNOUNCEMENT` of the tree hash of the notarized payment, and a `CREATE_COIN` for each payment within the entry. The maker's coins assert these puzzle announcements, binding the two sides together -- if either side is altered, the assertions fail and the entire bundle is invalid.

The settlement puzzle is versatile: it can be used as an inner puzzle inside a CAT or NFT, or standalone for XCH, enabling cross-asset trades.

**Nonce**: The nonce is the tree hash of a sorted list of the coin IDs being offered. This binds the settlement to specific coins and prevents a maker from creating two offers that could both be fulfilled with a single payment.

**Offer file**: An offer file is a bech32-encoded incomplete spend bundle containing only the maker's side. It can be shared freely -- anyone who sees it can only accept or ignore it. No private keys are exposed. Any alteration to the file invalidates the offer.

**Offer lifecycle**:
- **PENDING_ACCEPT** -- maker created the offer; not yet taken
- **PENDING_CONFIRM** -- taker completed the bundle and pushed to mempool
- **CONFIRMED** -- included in a block; trade complete
- **PENDING_CANCEL** -- maker is spending the offered coins to invalidate
- **CANCELLED** -- maker's coins were spent, invalidating the offer
- **FAILED** -- taker's acceptance failed (maker cancelled or another taker was first)

**Cancellation**: An offer is cancelled by spending the coins it references, which invalidates the partial spend bundle. If the offer was never published, simply deleting the file is sufficient.

**Aggregation**: Multiple offers can be aggregated into a single settlement by an Automated Market Maker (AMM). Each asset type gets one `settlement_payments` ephemeral coin, and multiple offers' notarized payments are combined. This is relevant when investigating complex multi-party swaps on-chain.

## Known puzzle mod hashes

Reference table of well-known puzzle template hashes. Use these with `coinset clvm curry` and `coinset clvm uncurry` to identify and compose puzzles. Full serialized mod bytes are available at `https://raw.githubusercontent.com/Chia-Network/chia_puzzles/main/src/programs.rs`.

### Core puzzles

**Standard transaction** (`P2_DELEGATED_PUZZLE_OR_HIDDEN_PUZZLE`)
`0xe9aaa49f45bad5c889b86ee3341550c155cfdd10c3a6757de618d20612fffd52`
Args: `SYNTHETIC_PUBLIC_KEY` (48-byte atom)

**CAT v2** (`CAT_PUZZLE`)
`0x37bef360ee858133b69d595a906dc45d01af50379dad515eb9518abb7c1d2a7a`
Args: `MOD_HASH` (atom — the CAT mod hash itself), `TAIL_PROGRAM_HASH` (atom — the asset ID), `INNER_PUZZLE` (tree)

**Singleton v1.1** (`SINGLETON_TOP_LAYER_V1_1`)
`0x7faa3253bfddd1e0decb0906b2dc6247bbc4cf608f58345d173adb63e8b47c9f`
Args: `SINGLETON_STRUCT` (tree — `(MOD_HASH . (LAUNCHER_ID . LAUNCHER_PUZZLE_HASH))`), `INNER_PUZZLE` (tree)

**Singleton launcher** (`SINGLETON_LAUNCHER`)
`0xeff07522495060c066f66f32acc2a77e3a3e737aca8baea4d1a64ea4cdc13da9`
No curried args. The launcher coin ID becomes the `launcher_id`.

**Settlement payments** (`SETTLEMENT_PAYMENT`)
`0xcfbfdeed5c4ca2de3d0bf520b9cb4bb7743a359bd2e6a188d19ce7dffc21d3e7`
No curried args. Used as inner puzzle in offers.

**DID inner puzzle** (`DID_INNERPUZ`)
`0x33143d2bef64f14036742673afd158126b94284b4530a28c354fac202b0c910e`
Args: `INNER_PUZZLE` (tree), `RECOVERY_DID_LIST_HASH` (atom), `NUM_VERIFICATIONS_REQUIRED` (atom), `SINGLETON_STRUCT` (tree), `METADATA` (tree)

**NFT state layer** (`NFT_STATE_LAYER`)
`0xa04d9f57764f54a43e4030befb4d80026e870519aaa66334aef8304f5d0393c2`
Args: `MOD_HASH` (atom), `METADATA` (tree), `METADATA_UPDATER_PUZZLE_HASH` (atom), `INNER_PUZZLE` (tree)

**NFT ownership layer** (`NFT_OWNERSHIP_LAYER`)
`0xc5abea79afaa001b5427dfa0c8cf42ca6f38f5841b78f9b3c252733eb2de2726`
Args: `MOD_HASH` (atom), `CURRENT_OWNER` (atom), `TRANSFER_PROGRAM` (tree), `INNER_PUZZLE` (tree)

**Royalty transfer program** (`NFT_OWNERSHIP_TRANSFER_PROGRAM_ONE_WAY_CLAIM_WITH_ROYALTIES`)
`0x025dee0fb1e9fa110302a7e9bfb6e381ca09618e2778b0184fa5c6b275cfce1f`
Args: `SINGLETON_STRUCT` (tree), `ROYALTY_ADDRESS` (atom), `TRADE_PRICE_PERCENTAGE` (atom)

**NFT metadata updater** (`NFT_METADATA_UPDATER_DEFAULT`)
`0xfe8a4b4e27a2e29a4d3fc7ce9d527adbcaccbab6ada3903ccf3ba9a769d2d78b`
No curried args.

### Commonly encountered puzzles

**Pay-to-singleton** (`P2_SINGLETON`)
`0x40f828d8dd55603f4ff9fbf6b73271e904e69406982f4fbefae2c8dcceaf9834`
Used as pool farming reward target.

**Notification** (`NOTIFICATION`)
`0xb8b9d8ffca6d5cba5422ead7f477ecfc8f6aaaa1c024b8c3aeb1956b24a0ab1e`

**Pool member inner puzzle** (`POOL_MEMBER_INNERPUZ`)
`0xa8490702e333ddd831a3ac9c22d0fa26d2bfeaf2d33608deb22f0e0123eb0494`

**Pool waiting room** (`POOL_WAITINGROOM_INNERPUZ`)
`0xa317541a765bf8375e1c6e7c13503d0d2cbf56cacad5182befe947e78e2c0307`

### Common TAIL programs

**Everything with signature** (`EVERYTHING_WITH_SIGNATURE`)
`0xf26fe751e5a9b13e87e4c19e4c383e713cec8c854cf8e591053e179e60e3b856`
Args: `PUBLIC_KEY` (48-byte atom). Most common TAIL for standard CAT issuance.

**Genesis by coin ID** (`GENESIS_BY_COIN_ID`)
`0x493afb89eed93ab86741b2aa61b8f5de495d33ff9b781dfc8919e602b2571571`
Args: `GENESIS_COIN_ID` (atom). Single-issuance tokens.

## Agent operating rules

1. **Always capture identifiers**: coin ID, puzzle hash, height, header hash, tx ID. These are needed for follow-up queries.
2. **Check `success` field**: all Full Node RPC responses include `"success": true|false`. Always verify.
3. **Use `--inspect` first** for spend analysis. Drop to raw CLVM disassembly only when `--inspect` doesn't cover what you need.
4. **If `--inspect` errors with "not supported for this endpoint/output shape"**: re-run the command without `--inspect`, or try a different endpoint that returns spend bundle data.
5. **Use `--include-spent-coins`** on coin record queries when tracing history. Without it, you only see unspent coins.
6. **Prefer confirmed data**: block spends are final; mempool items can disappear.
7. **Explain from conditions outward**: start with `evaluation.conditions` and `evaluation.cost`, then `puzzle_recognition.wrappers`, then raw CLVM if needed.
8. **If recognition is ambiguous** (multiple candidates or `parse_error`): report the ambiguity, don't make definitive wrapper claims. Fall back to conditions and additions/removals.
9. **Don't infer transaction intent from incomplete data**: always ground explanations in the actual conditions, additions, removals, and fee.
10. **Use `-q` for targeted extraction** instead of parsing large JSON responses manually.
