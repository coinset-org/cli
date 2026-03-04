use chialisp::classic::clvm::OPERATORS_LATEST_VERSION;
use chialisp::classic::clvm_tools::binutils::disassemble;
use chia_puzzle_types::did::DidSolution;
use chia_sdk_driver::{
    AugmentedConditionLayer, BulletinLayer, CatLayer, DidLayer, Layer, NftOwnershipLayer,
    NftStateLayer, OptionContractLayer, P2CurriedLayer, P2DelegatedConditionsLayer,
    P2OneOfManyLayer, P2SingletonLayer, Puzzle as DriverPuzzle, RevocationLayer,
    RoyaltyTransferLayer, SettlementLayer, SingletonLayer, StandardLayer, StreamLayer,
};
use clvm_utils::tree_hash;
use clvmr::allocator::NodePtr;
use clvmr::serde::node_from_bytes_backrefs;
use clvmr::Allocator;
use serde::Serialize;
use serde_json::{Value, json};

use crate::encode_hex_prefixed;

const SOURCE_REPO: &str = "xch-dev/chia-wallet-sdk";
const SOURCE_REF: &str = "0.33.0";
const MAX_LAYER_DEPTH: usize = 32;

#[derive(Debug, Clone, Serialize)]
pub struct PuzzleRecognition {
    pub recognized: bool,
    pub candidates: Vec<PuzzleCandidate>,
    pub wrappers: Vec<WrapperInfo>,
    pub parsed_solution: Option<Value>,
}

#[derive(Debug, Clone, Serialize)]
pub struct WrapperInfo {
    pub name: String,
    pub source_repo: String,
    pub source_ref: String,
    pub source_path: Option<String>,
    pub mod_hash: String,
    pub curried_args_tree_hash: Option<String>,
    pub inner_puzzle_tree_hash: Option<String>,
    pub params: Value,
    pub parse_error: Option<String>,
}

#[derive(Debug, Clone, Serialize)]
pub struct PuzzleCandidate {
    pub name: String,
    pub confidence: f64,
    pub source_repo: Option<String>,
    pub source_path: Option<String>,
    pub source_ref: Option<String>,
}

#[derive(Debug, Clone)]
struct LayerMatch {
    name: &'static str,
    source_path: &'static str,
    params: Value,
    next_puzzle: Option<DriverPuzzle>,
    next_solution: Option<NodePtr>,
    solution: Value,
    parse_error: Option<String>,
}

pub fn recognize_puzzle_and_solution(puzzle_reveal_bytes: &[u8], solution_bytes: &[u8]) -> PuzzleRecognition {
    let mut allocator = Allocator::new();

    let puzzle_ptr = match node_from_bytes_backrefs(&mut allocator, puzzle_reveal_bytes) {
        Ok(ptr) => ptr,
        Err(err) => {
            return PuzzleRecognition {
                recognized: false,
                candidates: Vec::new(),
                wrappers: Vec::new(),
                parsed_solution: Some(json!({
                    "layers": [],
                    "decode_error": format!("failed to decode puzzle_reveal bytes: {err}"),
                })),
            };
        }
    };

    let solution_ptr = node_from_bytes_backrefs(&mut allocator, solution_bytes).ok();
    let solution_decode_error = if solution_ptr.is_none() {
        Some("failed to decode solution bytes".to_string())
    } else {
        None
    };

    let mut current_puzzle = DriverPuzzle::parse(&allocator, puzzle_ptr);
    let mut current_solution = solution_ptr;
    let mut wrappers = Vec::<WrapperInfo>::new();
    let mut candidates = Vec::<PuzzleCandidate>::new();
    let mut solution_layers = Vec::<Value>::new();

    for _ in 0..MAX_LAYER_DEPTH {
        let matches = collect_matches(&allocator, current_puzzle, current_solution);
        if matches.is_empty() {
            break;
        }

        if matches.len() > 1 {
            for matched in &matches {
                candidates.push(candidate_from_match(matched, 0.5));
            }
            solution_layers.push(json!({
                "status": "ambiguous",
                "options": matches.iter().map(|m| m.name).collect::<Vec<_>>(),
            }));
            break;
        }

        let matched = matches[0].clone();
        let current_hash = current_puzzle.curried_puzzle_hash();

        wrappers.push(WrapperInfo {
            name: matched.name.to_string(),
            source_repo: SOURCE_REPO.to_string(),
            source_ref: SOURCE_REF.to_string(),
            source_path: Some(matched.source_path.to_string()),
            mod_hash: encode_tree_hash(current_puzzle.mod_hash().as_ref()),
            curried_args_tree_hash: current_puzzle
                .as_curried()
                .map(|curried| encode_tree_hash(tree_hash(&allocator, curried.args).as_ref())),
            inner_puzzle_tree_hash: matched
                .next_puzzle
                .map(|p| encode_tree_hash(p.curried_puzzle_hash().as_ref())),
            params: matched.params.clone(),
            parse_error: matched.parse_error.clone(),
        });

        candidates.push(candidate_from_match(
            &matched,
            if matched.parse_error.is_some() { 0.8 } else { 1.0 },
        ));

        solution_layers.push(json!({
            "layer": matched.name,
            "source_path": matched.source_path,
            "result": matched.solution,
        }));

        current_solution = matched.next_solution;

        let Some(next_puzzle) = matched.next_puzzle else {
            break;
        };

        if next_puzzle.curried_puzzle_hash() == current_hash {
            solution_layers.push(json!({
                "status": "stopped",
                "reason": "next puzzle hash equals current hash",
            }));
            break;
        }

        current_puzzle = next_puzzle;
    }

    let parsed_solution = if solution_layers.is_empty() && solution_decode_error.is_none() {
        None
    } else {
        Some(json!({
            "layers": solution_layers,
            "remaining_solution": current_solution.map(|ptr| node_summary(&allocator, ptr)),
            "decode_error": solution_decode_error,
        }))
    };

    PuzzleRecognition {
        recognized: !wrappers.is_empty(),
        candidates,
        wrappers,
        parsed_solution,
    }
}

fn collect_matches(
    allocator: &Allocator,
    puzzle: DriverPuzzle,
    solution: Option<NodePtr>,
) -> Vec<LayerMatch> {
    let mut matches = Vec::new();

    if let Some(matched) = try_cat_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_singleton_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_did_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_nft_state_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_nft_ownership_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_royalty_transfer_layer(allocator, puzzle) {
        matches.push(matched);
    }
    if let Some(matched) = try_augmented_condition_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_bulletin_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_option_contract_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_revocation_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_p2_singleton_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_p2_curried_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_p2_one_of_many_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_p2_delegated_conditions_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_settlement_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_stream_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }
    if let Some(matched) = try_standard_layer(allocator, puzzle, solution) {
        matches.push(matched);
    }

    matches
}

fn try_cat_layer(allocator: &Allocator, puzzle: DriverPuzzle, solution: Option<NodePtr>) -> Option<LayerMatch> {
    let layer = CatLayer::<DriverPuzzle>::parse_puzzle(allocator, puzzle).ok()??;
    let mut parse_error = None;
    let (next_solution, solution_json) = match solution {
        Some(ptr) => match CatLayer::<DriverPuzzle>::parse_solution(allocator, ptr) {
            Ok(parsed) => (
                Some(parsed.inner_puzzle_solution),
                json!({
                    "status": "ok",
                    "inner_solution_tree_hash": node_tree_hash_hex(allocator, parsed.inner_puzzle_solution),
                }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse CAT solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "cat_layer",
        source_path: "crates/chia-sdk-driver/src/layers/cat_layer.rs",
        params: json!({ "asset_id": encode_hex_prefixed(layer.asset_id.as_ref()) }),
        next_puzzle: Some(layer.inner_puzzle),
        next_solution,
        solution: solution_json,
        parse_error,
    })
}

fn try_singleton_layer(allocator: &Allocator, puzzle: DriverPuzzle, solution: Option<NodePtr>) -> Option<LayerMatch> {
    let layer = SingletonLayer::<DriverPuzzle>::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (next_solution, solution_json) = match solution {
        Some(ptr) => match SingletonLayer::<DriverPuzzle>::parse_solution(allocator, ptr) {
            Ok(parsed) => (
                Some(parsed.inner_solution),
                json!({ "status": "ok", "inner_solution_tree_hash": node_tree_hash_hex(allocator, parsed.inner_solution) }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse singleton solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "singleton_layer",
        source_path: "crates/chia-sdk-driver/src/layers/singleton_layer.rs",
        params: json!({ "launcher_id": encode_hex_prefixed(layer.launcher_id.as_ref()) }),
        next_puzzle: Some(layer.inner_puzzle),
        next_solution,
        solution: solution_json,
        parse_error,
    })
}

fn try_did_layer(allocator: &Allocator, puzzle: DriverPuzzle, solution: Option<NodePtr>) -> Option<LayerMatch> {
    let layer = DidLayer::<NodePtr, DriverPuzzle>::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (next_solution, solution_json) = match solution {
        Some(ptr) => match DidLayer::<NodePtr, DriverPuzzle>::parse_solution(allocator, ptr) {
            Ok(DidSolution::Spend(inner_solution)) => (
                Some(inner_solution),
                json!({ "status": "ok", "kind": "spend", "inner_solution_tree_hash": node_tree_hash_hex(allocator, inner_solution) }),
            ),
            Ok(DidSolution::Recover(recovery)) => (
                None,
                json!({ "status": "ok", "kind": "recover", "parsed_debug": format!("{recovery:?}") }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse DID solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "did_layer",
        source_path: "crates/chia-sdk-driver/src/layers/did_layer.rs",
        params: json!({
            "launcher_id": encode_hex_prefixed(layer.launcher_id.as_ref()),
            "recovery_list_hash": layer.recovery_list_hash.map(|h| encode_hex_prefixed(h.as_ref())),
            "num_verifications_required": layer.num_verifications_required,
            "metadata_tree_hash": node_tree_hash_hex(allocator, layer.metadata),
        }),
        next_puzzle: Some(layer.inner_puzzle),
        next_solution,
        solution: solution_json,
        parse_error,
    })
}

fn try_nft_state_layer(allocator: &Allocator, puzzle: DriverPuzzle, solution: Option<NodePtr>) -> Option<LayerMatch> {
    let layer = NftStateLayer::<NodePtr, DriverPuzzle>::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (next_solution, solution_json) = match solution {
        Some(ptr) => match NftStateLayer::<NodePtr, DriverPuzzle>::parse_solution(allocator, ptr) {
            Ok(parsed) => (
                Some(parsed.inner_solution),
                json!({ "status": "ok", "inner_solution_tree_hash": node_tree_hash_hex(allocator, parsed.inner_solution) }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse NFT state solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "nft_state_layer",
        source_path: "crates/chia-sdk-driver/src/layers/nft_state_layer.rs",
        params: json!({
            "metadata_updater_puzzle_hash": encode_hex_prefixed(layer.metadata_updater_puzzle_hash.as_ref()),
            "metadata_tree_hash": node_tree_hash_hex(allocator, layer.metadata),
        }),
        next_puzzle: Some(layer.inner_puzzle),
        next_solution,
        solution: solution_json,
        parse_error,
    })
}

fn try_nft_ownership_layer(
    allocator: &Allocator,
    puzzle: DriverPuzzle,
    solution: Option<NodePtr>,
) -> Option<LayerMatch> {
    let layer = NftOwnershipLayer::<DriverPuzzle, DriverPuzzle>::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (next_solution, solution_json) = match solution {
        Some(ptr) => match NftOwnershipLayer::<DriverPuzzle, DriverPuzzle>::parse_solution(allocator, ptr) {
            Ok(parsed) => (
                Some(parsed.inner_solution),
                json!({ "status": "ok", "inner_solution_tree_hash": node_tree_hash_hex(allocator, parsed.inner_solution) }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse NFT ownership solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "nft_ownership_layer",
        source_path: "crates/chia-sdk-driver/src/layers/nft_ownership_layer.rs",
        params: json!({
            "current_owner": layer.current_owner.map(|owner| encode_hex_prefixed(owner.as_ref())),
            "transfer_layer_tree_hash": encode_tree_hash(layer.transfer_layer.curried_puzzle_hash().as_ref()),
        }),
        next_puzzle: Some(layer.inner_puzzle),
        next_solution,
        solution: solution_json,
        parse_error,
    })
}

fn try_royalty_transfer_layer(allocator: &Allocator, puzzle: DriverPuzzle) -> Option<LayerMatch> {
    let layer = RoyaltyTransferLayer::parse_puzzle(allocator, puzzle).ok().flatten()?;
    Some(LayerMatch {
        name: "royalty_transfer_layer",
        source_path: "crates/chia-sdk-driver/src/layers/royalty_transfer_layer.rs",
        params: json!({
            "launcher_id": encode_hex_prefixed(layer.launcher_id.as_ref()),
            "royalty_puzzle_hash": encode_hex_prefixed(layer.royalty_puzzle_hash.as_ref()),
            "royalty_basis_points": layer.royalty_basis_points,
        }),
        next_puzzle: None,
        next_solution: None,
        solution: json!({ "status": "unsupported" }),
        parse_error: None,
    })
}

fn try_augmented_condition_layer(
    allocator: &Allocator,
    puzzle: DriverPuzzle,
    solution: Option<NodePtr>,
) -> Option<LayerMatch> {
    let layer = AugmentedConditionLayer::<NodePtr, DriverPuzzle>::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (next_solution, solution_json) = match solution {
        Some(ptr) => match AugmentedConditionLayer::<NodePtr, DriverPuzzle>::parse_solution(allocator, ptr) {
            Ok(parsed) => (
                Some(parsed.inner_solution),
                json!({ "status": "ok", "inner_solution_tree_hash": node_tree_hash_hex(allocator, parsed.inner_solution) }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse augmented condition solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "augmented_condition_layer",
        source_path: "crates/chia-sdk-driver/src/layers/augmented_condition_layer.rs",
        params: json!({}),
        next_puzzle: Some(layer.inner_puzzle),
        next_solution,
        solution: solution_json,
        parse_error,
    })
}

fn try_bulletin_layer(allocator: &Allocator, puzzle: DriverPuzzle, solution: Option<NodePtr>) -> Option<LayerMatch> {
    let layer = BulletinLayer::<DriverPuzzle>::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (next_solution, solution_json) = match solution {
        Some(ptr) => match BulletinLayer::<DriverPuzzle>::parse_solution(allocator, ptr) {
            Ok(inner_solution) => (
                Some(inner_solution),
                json!({ "status": "ok", "inner_solution_tree_hash": node_tree_hash_hex(allocator, inner_solution) }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse bulletin solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "bulletin_layer",
        source_path: "crates/chia-sdk-driver/src/layers/bulletin_layer.rs",
        params: json!({}),
        next_puzzle: Some(layer.inner_puzzle),
        next_solution,
        solution: solution_json,
        parse_error,
    })
}

fn try_option_contract_layer(
    allocator: &Allocator,
    puzzle: DriverPuzzle,
    solution: Option<NodePtr>,
) -> Option<LayerMatch> {
    let layer = OptionContractLayer::<DriverPuzzle>::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (next_solution, solution_json) = match solution {
        Some(ptr) => match OptionContractLayer::<DriverPuzzle>::parse_solution(allocator, ptr) {
            Ok(parsed) => (
                Some(parsed.inner_solution),
                json!({ "status": "ok", "inner_solution_tree_hash": node_tree_hash_hex(allocator, parsed.inner_solution) }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse option contract solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "option_contract_layer",
        source_path: "crates/chia-sdk-driver/src/layers/option_contract_layer.rs",
        params: json!({
            "underlying_coin_id": encode_hex_prefixed(layer.underlying_coin_id.as_ref()),
            "underlying_delegated_puzzle_hash": encode_hex_prefixed(layer.underlying_delegated_puzzle_hash.as_ref()),
        }),
        next_puzzle: Some(layer.inner_puzzle),
        next_solution,
        solution: solution_json,
        parse_error,
    })
}

fn try_revocation_layer(allocator: &Allocator, puzzle: DriverPuzzle, solution: Option<NodePtr>) -> Option<LayerMatch> {
    let layer = RevocationLayer::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (next_solution, solution_json) = match solution {
        Some(ptr) => match RevocationLayer::parse_solution(allocator, ptr) {
            Ok(parsed) => (
                None,
                json!({
                    "status": "ok",
                    "hidden": parsed.hidden,
                    "puzzle_tree_hash": node_tree_hash_hex(allocator, parsed.puzzle),
                    "solution_tree_hash": node_tree_hash_hex(allocator, parsed.solution),
                }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse revocation solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "revocation_layer",
        source_path: "crates/chia-sdk-driver/src/layers/revocation_layer.rs",
        params: json!({
            "hidden_puzzle_hash": encode_hex_prefixed(layer.hidden_puzzle_hash.as_ref()),
            "inner_puzzle_hash": encode_hex_prefixed(layer.inner_puzzle_hash.as_ref()),
        }),
        next_puzzle: None,
        next_solution,
        solution: solution_json,
        parse_error,
    })
}

fn try_p2_singleton_layer(
    allocator: &Allocator,
    puzzle: DriverPuzzle,
    solution: Option<NodePtr>,
) -> Option<LayerMatch> {
    let layer = P2SingletonLayer::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (_next_solution, solution_json): (Option<NodePtr>, Value) = match solution {
        Some(ptr) => match P2SingletonLayer::parse_solution(allocator, ptr) {
            Ok(parsed) => (
                None,
                json!({
                    "status": "ok",
                    "singleton_inner_puzzle_hash": encode_hex_prefixed(parsed.singleton_inner_puzzle_hash.as_ref()),
                    "my_id": encode_hex_prefixed(parsed.my_id.as_ref()),
                }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse p2_singleton solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "p2_singleton_layer",
        source_path: "crates/chia-sdk-driver/src/layers/p2_singleton_layer.rs",
        params: json!({ "launcher_id": encode_hex_prefixed(layer.launcher_id.as_ref()) }),
        next_puzzle: None,
        next_solution: None,
        solution: solution_json,
        parse_error,
    })
}

fn try_p2_curried_layer(
    allocator: &Allocator,
    puzzle: DriverPuzzle,
    solution: Option<NodePtr>,
) -> Option<LayerMatch> {
    let layer = P2CurriedLayer::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (_next_solution, solution_json): (Option<NodePtr>, Value) = match solution {
        Some(ptr) => match P2CurriedLayer::parse_solution(allocator, ptr) {
            Ok(parsed) => (
                None,
                json!({
                    "status": "ok",
                    "puzzle_tree_hash": node_tree_hash_hex(allocator, parsed.puzzle),
                    "solution_tree_hash": node_tree_hash_hex(allocator, parsed.solution),
                }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse p2_curried solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "p2_curried_layer",
        source_path: "crates/chia-sdk-driver/src/layers/p2_curried_layer.rs",
        params: json!({ "puzzle_hash": encode_hex_prefixed(layer.puzzle_hash.as_ref()) }),
        next_puzzle: None,
        next_solution: None,
        solution: solution_json,
        parse_error,
    })
}

fn try_p2_one_of_many_layer(
    allocator: &Allocator,
    puzzle: DriverPuzzle,
    solution: Option<NodePtr>,
) -> Option<LayerMatch> {
    let layer = P2OneOfManyLayer::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (_next_solution, solution_json): (Option<NodePtr>, Value) = match solution {
        Some(ptr) => match P2OneOfManyLayer::parse_solution(allocator, ptr) {
            Ok(parsed) => (
                None,
                json!({
                    "status": "ok",
                    "puzzle_tree_hash": node_tree_hash_hex(allocator, parsed.puzzle),
                    "solution_tree_hash": node_tree_hash_hex(allocator, parsed.solution),
                }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse p2_one_of_many solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "p2_one_of_many_layer",
        source_path: "crates/chia-sdk-driver/src/layers/p2_one_of_many_layer.rs",
        params: json!({ "merkle_root": encode_hex_prefixed(layer.merkle_root.as_ref()) }),
        next_puzzle: None,
        next_solution: None,
        solution: solution_json,
        parse_error,
    })
}

fn try_p2_delegated_conditions_layer(
    allocator: &Allocator,
    puzzle: DriverPuzzle,
    solution: Option<NodePtr>,
) -> Option<LayerMatch> {
    let layer = P2DelegatedConditionsLayer::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (_next_solution, solution_json): (Option<NodePtr>, Value) = match solution {
        Some(ptr) => match P2DelegatedConditionsLayer::parse_solution(allocator, ptr) {
            Ok(parsed) => (None, json!({ "status": "ok", "conditions_len": parsed.conditions.len() })),
            Err(err) => {
                parse_error = Some(format!("failed to parse p2_delegated_conditions solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "p2_delegated_conditions_layer",
        source_path: "crates/chia-sdk-driver/src/layers/p2_delegated_conditions_layer.rs",
        params: json!({ "public_key": encode_hex_prefixed(&layer.public_key.to_bytes()) }),
        next_puzzle: None,
        next_solution: None,
        solution: solution_json,
        parse_error,
    })
}

fn try_settlement_layer(
    allocator: &Allocator,
    puzzle: DriverPuzzle,
    solution: Option<NodePtr>,
) -> Option<LayerMatch> {
    let _layer = SettlementLayer::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (_next_solution, solution_json): (Option<NodePtr>, Value) = match solution {
        Some(ptr) => match SettlementLayer::parse_solution(allocator, ptr) {
            Ok(parsed) => (None, json!({ "status": "ok", "payments_len": parsed.notarized_payments.len() })),
            Err(err) => {
                parse_error = Some(format!("failed to parse settlement solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "settlement_layer",
        source_path: "crates/chia-sdk-driver/src/layers/settlement_layer.rs",
        params: json!({}),
        next_puzzle: None,
        next_solution: None,
        solution: solution_json,
        parse_error,
    })
}

fn try_stream_layer(allocator: &Allocator, puzzle: DriverPuzzle, solution: Option<NodePtr>) -> Option<LayerMatch> {
    let layer = StreamLayer::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (_next_solution, solution_json): (Option<NodePtr>, Value) = match solution {
        Some(ptr) => match StreamLayer::parse_solution(allocator, ptr) {
            Ok(parsed) => (
                None,
                json!({
                    "status": "ok",
                    "my_amount": parsed.my_amount,
                    "payment_time": parsed.payment_time,
                    "to_pay": parsed.to_pay,
                    "clawback": parsed.clawback,
                }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse stream solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "stream_layer",
        source_path: "crates/chia-sdk-driver/src/layers/streaming_layer.rs",
        params: json!({
            "recipient": encode_hex_prefixed(layer.recipient.as_ref()),
            "clawback_ph": layer.clawback_ph.map(|value| encode_hex_prefixed(value.as_ref())),
            "end_time": layer.end_time,
            "last_payment_time": layer.last_payment_time,
        }),
        next_puzzle: None,
        next_solution: None,
        solution: solution_json,
        parse_error,
    })
}

fn try_standard_layer(allocator: &Allocator, puzzle: DriverPuzzle, solution: Option<NodePtr>) -> Option<LayerMatch> {
    let layer = StandardLayer::parse_puzzle(allocator, puzzle).ok().flatten()?;
    let mut parse_error = None;
    let (_next_solution, solution_json): (Option<NodePtr>, Value) = match solution {
        Some(ptr) => match StandardLayer::parse_solution(allocator, ptr) {
            Ok(parsed) => (
                None,
                json!({
                    "status": "ok",
                    "has_original_public_key": parsed.original_public_key.is_some(),
                    "delegated_puzzle_tree_hash": node_tree_hash_hex(allocator, parsed.delegated_puzzle),
                    "delegated_solution_tree_hash": node_tree_hash_hex(allocator, parsed.solution),
                }),
            ),
            Err(err) => {
                parse_error = Some(format!("failed to parse standard solution: {err}"));
                (None, json!({ "status": "error", "message": err.to_string() }))
            }
        },
        None => (None, json!({ "status": "missing_solution" })),
    };

    Some(LayerMatch {
        name: "standard_layer",
        source_path: "crates/chia-sdk-driver/src/layers/standard_layer.rs",
        params: json!({ "synthetic_key": encode_hex_prefixed(&layer.synthetic_key.to_bytes()) }),
        next_puzzle: None,
        next_solution: None,
        solution: solution_json,
        parse_error,
    })
}

fn candidate_from_match(matched: &LayerMatch, confidence: f64) -> PuzzleCandidate {
    PuzzleCandidate {
        name: matched.name.to_string(),
        confidence,
        source_repo: Some(SOURCE_REPO.to_string()),
        source_path: Some(matched.source_path.to_string()),
        source_ref: Some(SOURCE_REF.to_string()),
    }
}

fn node_summary(allocator: &Allocator, ptr: NodePtr) -> Value {
    json!({
        "tree_hash": node_tree_hash_hex(allocator, ptr),
        "disasm": disassemble(allocator, ptr, Some(OPERATORS_LATEST_VERSION)),
    })
}

fn node_tree_hash_hex(allocator: &Allocator, ptr: NodePtr) -> String {
    encode_tree_hash(tree_hash(allocator, ptr).as_ref())
}

fn encode_tree_hash(bytes: &[u8]) -> String {
    encode_hex_prefixed(bytes)
}

