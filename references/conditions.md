# Conditions reference

When a puzzle executes, it returns a list of conditions. Each condition is a list starting with an opcode number.

## Coin creation

| Opcode | Name | Format | Description |
|---|---|---|---|
| 51 | `CREATE_COIN` | `(51 puzzle_hash amount (...memos)?)` | Create a new coin. Optional memos list; first 32-byte memo is the hint. Cost: 1,800,000. |
| 52 | `RESERVE_FEE` | `(52 amount)` | Assert minimum fee in this transaction. |

## Signatures

All AGG_SIG conditions cost 1,200,000 each.

| Opcode | Name | Description |
|---|---|---|
| 49 | `AGG_SIG_UNSAFE` | Verify signature on raw message. No domain separation -- can be replayed. |
| 50 | `AGG_SIG_ME` | Verify signature on message + coin_id + genesis_id. Recommended for most uses. |
| 43-48 | `AGG_SIG_PARENT` / `AGG_SIG_PUZZLE` / `AGG_SIG_AMOUNT` / `AGG_SIG_PUZZLE_AMOUNT` / `AGG_SIG_PARENT_AMOUNT` / `AGG_SIG_PARENT_PUZZLE` | CHIP-11 domain-separated signature variants binding to specific coin attributes. |

## Announcements (legacy)

| Opcode | Name | Description |
|---|---|---|
| 60 | `CREATE_COIN_ANNOUNCEMENT` | Create announcement bound to this coin's ID. |
| 61 | `ASSERT_COIN_ANNOUNCEMENT` | Assert `sha256(coin_id + message)` was announced. |
| 62 | `CREATE_PUZZLE_ANNOUNCEMENT` | Create announcement bound to this coin's puzzle hash. |
| 63 | `ASSERT_PUZZLE_ANNOUNCEMENT` | Assert `sha256(puzzle_hash + message)` was announced. |

## Messages (modern replacement for announcements)

| Opcode | Name | Description |
|---|---|---|
| 66 | `SEND_MESSAGE` | `(66 mode message ...)` -- Send message with sender/receiver commitment via mode bitmask. |
| 67 | `RECEIVE_MESSAGE` | `(67 mode message ...)` -- Assert receipt of matching message. |

The `mode` byte is 6 bits: 3 bits for sender commitment, 3 bits for receiver commitment. Each 3-bit group encodes which coin attributes (parent, puzzle, amount) to commit to. For example, `mode=0b111110` means sender commits to all three (coin ID) and receiver commits to parent+puzzle.

## Self-assertions

| Opcode | Name | Description |
|---|---|---|
| 70 | `ASSERT_MY_COIN_ID` | Assert this coin's ID matches. |
| 71 | `ASSERT_MY_PARENT_ID` | Assert this coin's parent ID matches. |
| 72 | `ASSERT_MY_PUZZLEHASH` | Assert this coin's puzzle hash matches. |
| 73 | `ASSERT_MY_AMOUNT` | Assert this coin's amount matches. |
| 74 | `ASSERT_MY_BIRTH_SECONDS` | Assert coin creation timestamp. |
| 75 | `ASSERT_MY_BIRTH_HEIGHT` | Assert coin creation height. |
| 76 | `ASSERT_EPHEMERAL` | Assert this coin was created in the current block. |

## Timelocks

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

## Concurrency

| Opcode | Name | Description |
|---|---|---|
| 64 | `ASSERT_CONCURRENT_SPEND` | Assert another specific coin is spent in same block. |
| 65 | `ASSERT_CONCURRENT_PUZZLE` | Assert a coin with specific puzzle hash is spent in same block. |
| 1 | `REMARK` | No-op; always valid. |
| 90 | `SOFTFORK` | Soft-fork guard for future conditions. Cost is specified in ten-thousands. |
