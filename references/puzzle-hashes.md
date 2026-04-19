# Known puzzle mod hashes

Reference table of well-known puzzle template hashes. Use these with `coinset clvm curry` and `coinset clvm uncurry` to identify and compose puzzles. Full serialized mod bytes are available at `https://raw.githubusercontent.com/Chia-Network/chia_puzzles/main/src/programs.rs`.

## Core puzzles

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

## Commonly encountered puzzles

**Pay-to-singleton** (`P2_SINGLETON`)
`0x40f828d8dd55603f4ff9fbf6b73271e904e69406982f4fbefae2c8dcceaf9834`
Used as pool farming reward target.

**Notification** (`NOTIFICATION`)
`0xb8b9d8ffca6d5cba5422ead7f477ecfc8f6aaaa1c024b8c3aeb1956b24a0ab1e`

**Pool member inner puzzle** (`POOL_MEMBER_INNERPUZ`)
`0xa8490702e333ddd831a3ac9c22d0fa26d2bfeaf2d33608deb22f0e0123eb0494`

**Pool waiting room** (`POOL_WAITINGROOM_INNERPUZ`)
`0xa317541a765bf8375e1c6e7c13503d0d2cbf56cacad5182befe947e78e2c0307`

## Common TAIL programs

**Everything with signature** (`EVERYTHING_WITH_SIGNATURE`)
`0xf26fe751e5a9b13e87e4c19e4c383e713cec8c854cf8e591053e179e60e3b856`
Args: `PUBLIC_KEY` (48-byte atom). Most common TAIL for standard CAT issuance.

**Genesis by coin ID** (`GENESIS_BY_COIN_ID`)
`0x493afb89eed93ab86741b2aa61b8f5de495d33ff9b781dfc8919e602b2571571`
Args: `GENESIS_COIN_ID` (atom). Single-issuance tokens.
