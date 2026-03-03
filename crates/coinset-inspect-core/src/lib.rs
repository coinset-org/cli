use std::collections::BTreeMap;

use anyhow::{Context, Result, bail};
use chia_bls::PublicKey;
use chia_consensus::allocator::make_allocator;
use chia_consensus::consensus_constants::TEST_CONSTANTS;
use chia_consensus::owned_conditions::{OwnedSpendBundleConditions, OwnedSpendConditions};
use chia_consensus::spendbundle_conditions::get_conditions_from_spendbundle;
use chia_protocol::{Bytes, Coin, CoinSpend, SpendBundle};
use chia_traits::Streamable;
use chialisp::classic::clvm::OPERATORS_LATEST_VERSION;
use chialisp::classic::clvm_tools::binutils::disassemble;
use clvm_utils::tree_hash_from_bytes;
use clvmr::allocator::Allocator as ClvmAllocator;
use clvmr::serde::node_from_bytes_backrefs;
use clvmr::LIMIT_HEAP;
use serde::Serialize;
use serde_json::{Map, Value, json};

const DEFAULT_MAX_COST: u64 = 11_000_000_000;
const DEFAULT_PREV_TX_HEIGHT: u32 = 10_000_000;

#[derive(Debug, Clone, Copy, Eq, PartialEq)]
pub enum ExplainLevel {
    Conditions,
    Deep,
}

impl Default for ExplainLevel {
    fn default() -> Self {
        Self::Deep
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct InspectionOutput {
    pub schema_version: String,
    pub tool: ToolInfo,
    pub input: InputInfo,
    pub result: ResultInfo,
}

#[derive(Debug, Clone, Serialize)]
pub struct ToolInfo {
    pub name: String,
    pub version: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct InputInfo {
    pub kind: String,
    pub notes: Vec<String>,
}

#[derive(Debug, Clone, Serialize)]
pub struct ResultInfo {
    pub status: String,
    pub error: Option<ErrorInfo>,
    pub summary: Summary,
    pub spends: Vec<SpendAnalysis>,
    pub signatures: SignatureSummary,
}

#[derive(Debug, Clone, Serialize)]
pub struct ErrorInfo {
    pub kind: String,
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Default)]
pub struct Summary {
    pub removals: Vec<CoinRef>,
    pub additions: Vec<CoinRef>,
    pub fee_mojos: u64,
    pub net_xch_delta_by_puzzle_hash: Vec<NetDelta>,
}

#[derive(Debug, Clone, Serialize)]
pub struct NetDelta {
    pub puzzle_hash: String,
    pub delta_mojos: i128,
}

#[derive(Debug, Clone, Serialize)]
pub struct SpendAnalysis {
    pub coin_spend: CoinSpendView,
    pub evaluation: EvaluationInfo,
    pub clvm: Option<ClvmInfo>,
}

#[derive(Debug, Clone, Serialize)]
pub struct CoinSpendView {
    pub coin: CoinRef,
    pub puzzle_reveal: String,
    pub solution: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct CoinRef {
    pub coin_id: String,
    pub parent_coin_id: String,
    pub puzzle_hash: String,
    pub amount: u64,
}

#[derive(Debug, Clone, Serialize)]
pub struct EvaluationInfo {
    pub status: String,
    pub cost: u64,
    pub conditions: Vec<ConditionInfo>,
    pub additions: Vec<CoinRef>,
    pub failure: Option<FailureInfo>,
}

#[derive(Debug, Clone, Serialize)]
pub struct ConditionInfo {
    pub opcode: String,
    pub args: Vec<Value>,
}

#[derive(Debug, Clone, Serialize)]
pub struct FailureInfo {
    pub kind: String,
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Default)]
pub struct SignatureSummary {
    pub aggregated_signature: String,
    pub agg_sig_me: Vec<AggSigInfo>,
    pub agg_sig_unsafe: Vec<AggSigInfo>,
}

#[derive(Debug, Clone, Serialize)]
pub struct AggSigInfo {
    pub pubkey: String,
    pub msg: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct ClvmInfo {
    pub puzzle_opd: String,
    pub solution_opd: String,
    pub tree_hash: String,
    pub uses_backrefs: bool,
}

#[derive(Debug, Clone)]
enum InputKind {
    Mempool,
    Block,
    Coin,
}

impl InputKind {
    fn as_str(&self) -> &'static str {
        match self {
            Self::Mempool => "mempool",
            Self::Block => "block",
            Self::Coin => "coin",
        }
    }
}

pub fn inspect_json(input_json: &str, explain_level: ExplainLevel) -> Result<InspectionOutput> {
    let (kind, bundle, notes) = load_input(input_json)?;
    inspect_bundle(kind, bundle, notes, explain_level)
}

pub fn inspect_json_string(input_json: &str, pretty: bool, explain_level: ExplainLevel) -> Result<String> {
    let output = inspect_json(input_json, explain_level)?;
    if pretty {
        Ok(serde_json::to_string_pretty(&output)?)
    } else {
        Ok(serde_json::to_string(&output)?)
    }
}

fn inspect_bundle(
    kind: InputKind,
    spend_bundle: SpendBundle,
    notes: Vec<String>,
    explain_level: ExplainLevel,
) -> Result<InspectionOutput> {
    let mut allocator = make_allocator(LIMIT_HEAP);
    let eval = get_conditions_from_spendbundle(
        &mut allocator,
        &spend_bundle,
        DEFAULT_MAX_COST,
        DEFAULT_PREV_TX_HEIGHT,
        &TEST_CONSTANTS,
    );

    match eval {
        Ok(conditions) => {
            let owned = OwnedSpendBundleConditions::from(&allocator, conditions);
            Ok(build_success_output(kind, notes, spend_bundle, owned, explain_level))
        }
        Err(err) => Ok(build_error_output(
            kind,
            notes,
            spend_bundle,
            &format!("{err:?}"),
            explain_level,
        )),
    }
}

fn build_success_output(
    kind: InputKind,
    notes: Vec<String>,
    spend_bundle: SpendBundle,
    owned: OwnedSpendBundleConditions,
    explain_level: ExplainLevel,
) -> InspectionOutput {
    let mut spends = Vec::<SpendAnalysis>::new();
    let mut removals = Vec::<CoinRef>::new();
    let mut additions = Vec::<CoinRef>::new();
    let mut agg_sig_me = Vec::<AggSigInfo>::new();
    let mut agg_sig_unsafe = Vec::<AggSigInfo>::new();

    let spend_count = spend_bundle.coin_spends.len().min(owned.spends.len());
    for idx in 0..spend_count {
        let spend = &spend_bundle.coin_spends[idx];
        let conds = &owned.spends[idx];
        let spend_analysis = analyze_single_spend(spend, conds, explain_level, &mut agg_sig_me);
        removals.push(coin_ref_from_coin(&spend.coin));
        additions.extend(spend_analysis.evaluation.additions.iter().cloned());
        spends.push(spend_analysis);
    }

    for (pk, msg) in &owned.agg_sig_unsafe {
        agg_sig_unsafe.push(AggSigInfo {
            pubkey: encode_hex_prefixed(&pk.to_bytes()),
            msg: encode_hex_prefixed(msg.as_ref()),
        });
    }

    removals.sort_by(|a, b| a.coin_id.cmp(&b.coin_id));
    additions.sort_by(|a, b| a.coin_id.cmp(&b.coin_id));
    agg_sig_me.sort_by(|a, b| a.pubkey.cmp(&b.pubkey).then(a.msg.cmp(&b.msg)));
    agg_sig_unsafe.sort_by(|a, b| a.pubkey.cmp(&b.pubkey).then(a.msg.cmp(&b.msg)));

    let fee_mojos = owned
        .removal_amount
        .saturating_sub(owned.addition_amount)
        .try_into()
        .unwrap_or(u64::MAX);
    let net_xch_delta_by_puzzle_hash = compute_net_delta(&removals, &additions);

    InspectionOutput {
        schema_version: "coinset.inspect.v1".to_string(),
        tool: ToolInfo {
            name: "coinset-inspect".to_string(),
            version: env!("CARGO_PKG_VERSION").to_string(),
        },
        input: InputInfo {
            kind: kind.as_str().to_string(),
            notes,
        },
        result: ResultInfo {
            status: "ok".to_string(),
            error: None,
            summary: Summary {
                removals,
                additions,
                fee_mojos,
                net_xch_delta_by_puzzle_hash,
            },
            spends,
            signatures: SignatureSummary {
                aggregated_signature: encode_hex_prefixed(
                    &spend_bundle.aggregated_signature.to_bytes(),
                ),
                agg_sig_me,
                agg_sig_unsafe,
            },
        },
    }
}

fn build_error_output(
    kind: InputKind,
    notes: Vec<String>,
    spend_bundle: SpendBundle,
    message: &str,
    explain_level: ExplainLevel,
) -> InspectionOutput {
    let mut spends = Vec::new();
    let mut removals = Vec::new();
    for spend in &spend_bundle.coin_spends {
        removals.push(coin_ref_from_coin(&spend.coin));
        let clvm = if explain_level == ExplainLevel::Deep {
            Some(analyze_clvm(spend))
        } else {
            None
        };
        spends.push(SpendAnalysis {
            coin_spend: CoinSpendView {
                coin: coin_ref_from_coin(&spend.coin),
                puzzle_reveal: encode_hex_prefixed(spend.puzzle_reveal.as_ref()),
                solution: encode_hex_prefixed(spend.solution.as_ref()),
            },
            evaluation: EvaluationInfo {
                status: "failed".to_string(),
                cost: 0,
                conditions: Vec::new(),
                additions: Vec::new(),
                failure: Some(FailureInfo {
                    kind: "validation_error".to_string(),
                    message: message.to_string(),
                }),
            },
            clvm,
        });
    }

    InspectionOutput {
        schema_version: "coinset.inspect.v1".to_string(),
        tool: ToolInfo {
            name: "coinset-inspect".to_string(),
            version: env!("CARGO_PKG_VERSION").to_string(),
        },
        input: InputInfo {
            kind: kind.as_str().to_string(),
            notes,
        },
        result: ResultInfo {
            status: "failed".to_string(),
            error: Some(ErrorInfo {
                kind: "validation_error".to_string(),
                message: message.to_string(),
            }),
            summary: Summary {
                removals,
                additions: Vec::new(),
                fee_mojos: 0,
                net_xch_delta_by_puzzle_hash: Vec::new(),
            },
            spends,
            signatures: SignatureSummary {
                aggregated_signature: encode_hex_prefixed(
                    &spend_bundle.aggregated_signature.to_bytes(),
                ),
                agg_sig_me: Vec::new(),
                agg_sig_unsafe: Vec::new(),
            },
        },
    }
}

fn analyze_single_spend(
    spend: &CoinSpend,
    conds: &OwnedSpendConditions,
    explain_level: ExplainLevel,
    agg_sig_me_out: &mut Vec<AggSigInfo>,
) -> SpendAnalysis {
    let mut conditions = Vec::<ConditionInfo>::new();
    let mut additions = Vec::<CoinRef>::new();

    add_signature_conditions(
        &conds.agg_sig_me,
        "AGG_SIG_ME",
        &mut conditions,
        agg_sig_me_out,
    );
    add_signature_conditions(
        &conds.agg_sig_parent,
        "AGG_SIG_PARENT",
        &mut conditions,
        &mut Vec::new(),
    );
    add_signature_conditions(
        &conds.agg_sig_puzzle,
        "AGG_SIG_PUZZLE",
        &mut conditions,
        &mut Vec::new(),
    );
    add_signature_conditions(
        &conds.agg_sig_amount,
        "AGG_SIG_AMOUNT",
        &mut conditions,
        &mut Vec::new(),
    );
    add_signature_conditions(
        &conds.agg_sig_puzzle_amount,
        "AGG_SIG_PUZZLE_AMOUNT",
        &mut conditions,
        &mut Vec::new(),
    );
    add_signature_conditions(
        &conds.agg_sig_parent_amount,
        "AGG_SIG_PARENT_AMOUNT",
        &mut conditions,
        &mut Vec::new(),
    );
    add_signature_conditions(
        &conds.agg_sig_parent_puzzle,
        "AGG_SIG_PARENT_PUZZLE",
        &mut conditions,
        &mut Vec::new(),
    );

    let mut create_coin = conds.create_coin.clone();
    create_coin.sort_by(|a, b| {
        a.0.as_ref()
            .cmp(b.0.as_ref())
            .then(a.1.cmp(&b.1))
            .then(a.2.as_ref().map(Bytes::as_ref).cmp(&b.2.as_ref().map(Bytes::as_ref)))
    });
    for (puzzle_hash, amount, hint) in create_coin {
        let new_coin = Coin::new(conds.coin_id, puzzle_hash, amount);
        let coin_ref = coin_ref_from_coin(&new_coin);
        let mut args = vec![json!(encode_hex_prefixed(puzzle_hash.as_ref())), json!(amount)];
        if let Some(ref hint) = hint {
            args.push(json!([encode_hex_prefixed(hint.as_ref())]));
        }
        conditions.push(ConditionInfo {
            opcode: "CREATE_COIN".to_string(),
            args,
        });
        additions.push(coin_ref);
    }

    add_optional_assertion(
        "ASSERT_HEIGHT_RELATIVE",
        conds.height_relative.map(u64::from),
        &mut conditions,
    );
    add_optional_assertion(
        "ASSERT_SECONDS_RELATIVE",
        conds.seconds_relative,
        &mut conditions,
    );
    add_optional_assertion(
        "ASSERT_BEFORE_HEIGHT_RELATIVE",
        conds.before_height_relative.map(u64::from),
        &mut conditions,
    );
    add_optional_assertion(
        "ASSERT_BEFORE_SECONDS_RELATIVE",
        conds.before_seconds_relative,
        &mut conditions,
    );
    add_optional_assertion(
        "ASSERT_MY_BIRTH_HEIGHT",
        conds.birth_height.map(u64::from),
        &mut conditions,
    );
    add_optional_assertion(
        "ASSERT_MY_BIRTH_SECONDS",
        conds.birth_seconds,
        &mut conditions,
    );

    SpendAnalysis {
        coin_spend: CoinSpendView {
            coin: coin_ref_from_coin(&spend.coin),
            puzzle_reveal: encode_hex_prefixed(spend.puzzle_reveal.as_ref()),
            solution: encode_hex_prefixed(spend.solution.as_ref()),
        },
        evaluation: EvaluationInfo {
            status: "ok".to_string(),
            cost: conds.execution_cost + conds.condition_cost,
            conditions,
            additions,
            failure: None,
        },
        clvm: if explain_level == ExplainLevel::Deep {
            Some(analyze_clvm(spend))
        } else {
            None
        },
    }
}

fn analyze_clvm(spend: &CoinSpend) -> ClvmInfo {
    let mut allocator = ClvmAllocator::new();
    let (puzzle_opd, uses_backrefs_p) = disasm_bytes(&mut allocator, spend.puzzle_reveal.as_ref());
    let (solution_opd, uses_backrefs_s) = disasm_bytes(&mut allocator, spend.solution.as_ref());
    let uses_backrefs = uses_backrefs_p || uses_backrefs_s;
    let tree_hash = tree_hash_from_bytes(spend.puzzle_reveal.as_ref())
        .map(|h| encode_hex_prefixed(h.as_ref()))
        .unwrap_or_else(|_| encode_hex_prefixed(spend.coin.puzzle_hash.as_ref()));
    ClvmInfo {
        puzzle_opd,
        solution_opd,
        tree_hash,
        uses_backrefs,
    }
}

fn disasm_bytes(allocator: &mut ClvmAllocator, bytes: &[u8]) -> (String, bool) {
    let uses_backrefs = bytes.contains(&0xfe);
    match node_from_bytes_backrefs(allocator, bytes) {
        Ok(node) => (
            disassemble(allocator, node, Some(OPERATORS_LATEST_VERSION)),
            uses_backrefs,
        ),
        Err(err) => (format!("<failed to disassemble: {err}>"), uses_backrefs),
    }
}

fn add_signature_conditions(
    pairs: &[(PublicKey, Bytes)],
    opcode: &str,
    conditions: &mut Vec<ConditionInfo>,
    agg_sig_out: &mut Vec<AggSigInfo>,
) {
    for (pk, msg) in pairs {
        let pk_hex = encode_hex_prefixed(&pk.to_bytes());
        let msg_hex = encode_hex_prefixed(msg.as_ref());
        conditions.push(ConditionInfo {
            opcode: opcode.to_string(),
            args: vec![json!(pk_hex.clone()), json!(msg_hex.clone())],
        });
        agg_sig_out.push(AggSigInfo {
            pubkey: pk_hex,
            msg: msg_hex,
        });
    }
}

fn add_optional_assertion(opcode: &str, value: Option<u64>, conditions: &mut Vec<ConditionInfo>) {
    if let Some(v) = value {
        conditions.push(ConditionInfo {
            opcode: opcode.to_string(),
            args: vec![json!(v)],
        });
    }
}

fn compute_net_delta(removals: &[CoinRef], additions: &[CoinRef]) -> Vec<NetDelta> {
    let mut map = BTreeMap::<String, i128>::new();
    for coin in removals {
        *map.entry(coin.puzzle_hash.clone()).or_insert(0) -= i128::from(coin.amount);
    }
    for coin in additions {
        *map.entry(coin.puzzle_hash.clone()).or_insert(0) += i128::from(coin.amount);
    }
    map.into_iter()
        .map(|(puzzle_hash, delta_mojos)| NetDelta {
            puzzle_hash,
            delta_mojos,
        })
        .collect()
}

fn coin_ref_from_coin(coin: &Coin) -> CoinRef {
    CoinRef {
        coin_id: encode_hex_prefixed(coin.coin_id().as_ref()),
        parent_coin_id: encode_hex_prefixed(coin.parent_coin_info.as_ref()),
        puzzle_hash: encode_hex_prefixed(coin.puzzle_hash.as_ref()),
        amount: coin.amount,
    }
}

fn encode_hex_prefixed(bytes: &[u8]) -> String {
    format!("0x{}", hex::encode(bytes))
}

fn decode_hex_prefixed(s: &str) -> Result<Vec<u8>> {
    let raw = s
        .strip_prefix("0x")
        .or_else(|| s.strip_prefix("0X"))
        .unwrap_or(s)
        .trim();
    if raw.is_empty() {
        return Ok(Vec::new());
    }
    if raw.len() % 2 != 0 {
        bail!("hex input must have even length");
    }
    Ok(hex::decode(raw)?)
}

fn normalize_hex_no_prefix(s: &str) -> Result<String> {
    let bytes = decode_hex_prefixed(s)?;
    Ok(hex::encode(bytes))
}

fn load_input(input_json: &str) -> Result<(InputKind, SpendBundle, Vec<String>)> {
    let value: Value = serde_json::from_str(input_json)?;
    let mut notes = Vec::<String>::new();

    // Mempool-ish wrappers
    let scope = if let Some(wrapper) = value.get("mempool_item") {
        notes.push("input contained mempool_item wrapper; using nested payload".to_string());
        wrapper
    } else {
        &value
    };

    // Spend bundle shapes
    if let Some(sb) = scope.get("spend_bundle") {
        let bundle = parse_spend_bundle_object(sb)?;
        return Ok((InputKind::Mempool, bundle, notes));
    }
    if let Some(sb_bytes) = scope.get("spend_bundle_bytes").and_then(Value::as_str) {
        notes.push("input contained spend_bundle_bytes; decoded with streamable parser".to_string());
        let bundle = parse_spend_bundle_bytes(sb_bytes)?;
        return Ok((InputKind::Mempool, bundle, notes));
    }
    if scope.get("coin_spends").is_some() && scope.is_object() {
        // Ambiguous shape: could be a full SpendBundle or a block-style coin spend list.
        // Treat it as a SpendBundle only if it includes an aggregate signature.
        if scope.get("aggregated_signature").is_some() {
            let bundle = parse_spend_bundle_object(scope)?;
            return Ok((InputKind::Mempool, bundle, notes));
        }
    }

    // Block-ish shapes
    if let Some(items) = value.get("block_spends") {
        let spends = parse_coin_spend_list(items)?;
        notes.push("block input normalized to SpendBundle with default aggregate signature".to_string());
        return Ok((InputKind::Block, SpendBundle::new(spends, Default::default()), notes));
    }
    if let Some(items) = value.get("coin_spends") {
        // Some endpoints return block objects with coin_spends.
        let spends = parse_coin_spend_list(items)?;
        notes.push("block input normalized to SpendBundle with default aggregate signature".to_string());
        return Ok((InputKind::Block, SpendBundle::new(spends, Default::default()), notes));
    }
    if value.is_array() {
        let spends = parse_coin_spend_list(&value)?;
        notes.push("array input normalized to SpendBundle with default aggregate signature".to_string());
        return Ok((InputKind::Block, SpendBundle::new(spends, Default::default()), notes));
    }

    // Single coin spend shapes
    if value.get("coin_spend").is_some() || value.get("puzzle_reveal").is_some() {
        let spend_value = value.get("coin_spend").unwrap_or(&value);
        let spend = parse_coin_spend(spend_value)?;
        notes.push("coin input normalized to single-spend SpendBundle".to_string());
        return Ok((InputKind::Coin, SpendBundle::new(vec![spend], Default::default()), notes));
    }

    bail!("unsupported input shape (no spend bundle/spends found)")
}

fn parse_spend_bundle_object(value: &Value) -> Result<SpendBundle> {
    let normalized = normalize_spend_bundle_value(value)?;
    serde_json::from_value(normalized).context("failed to parse spend bundle JSON")
}

fn parse_spend_bundle_bytes(hex_value: &str) -> Result<SpendBundle> {
    let bytes = decode_hex_prefixed(hex_value)?;
    SpendBundle::from_bytes(&bytes).context("failed to parse spend bundle bytes")
}

fn parse_coin_spend_list(value: &Value) -> Result<Vec<CoinSpend>> {
    let arr = value.as_array().context("coin spend list must be an array")?;
    let mut ret = Vec::with_capacity(arr.len());
    for item in arr {
        ret.push(parse_coin_spend(item)?);
    }
    Ok(ret)
}

fn parse_coin_spend(value: &Value) -> Result<CoinSpend> {
    let normalized = normalize_coin_spend_value(value)?;
    serde_json::from_value(normalized).context("failed to parse coin spend")
}

fn normalize_spend_bundle_value(value: &Value) -> Result<Value> {
    let mut obj = value
        .as_object()
        .context("spend bundle JSON must be an object")?
        .clone();
    let coin_spends = obj
        .get("coin_spends")
        .cloned()
        .context("spend bundle is missing coin_spends")?;
    let normalized_spends = normalize_coin_spend_array(&coin_spends)?;
    obj.insert("coin_spends".to_string(), normalized_spends);
    if let Some(sig) = obj.get("aggregated_signature").and_then(Value::as_str) {
        // Some JSON uses 0x prefix; chia-protocol serde expects hex without 0x.
        obj.insert(
            "aggregated_signature".to_string(),
            Value::String(normalize_hex_no_prefix(sig)?),
        );
    }
    Ok(Value::Object(obj))
}

fn normalize_coin_spend_array(value: &Value) -> Result<Value> {
    let arr = value
        .as_array()
        .context("coin_spends must be an array of coin spend objects")?;
    let mut out = Vec::with_capacity(arr.len());
    for item in arr {
        out.push(normalize_coin_spend_value(item)?);
    }
    Ok(Value::Array(out))
}

fn normalize_coin_spend_value(value: &Value) -> Result<Value> {
    let obj = value
        .as_object()
        .context("coin spend entry must be an object")?;
    let mut out = Map::new();
    for (k, v) in obj {
        if k == "puzzle_reveal" || k == "solution" {
            let s = v
                .as_str()
                .with_context(|| format!("{k} must be a hex string"))?;
            out.insert(k.clone(), json!(normalize_hex_no_prefix(s)?));
        } else {
            out.insert(k.clone(), v.clone());
        }
    }
    Ok(Value::Object(out))
}

#[cfg(test)]
mod tests {
    use super::*;
    use chia_protocol::{Coin, Program};

    fn sample_spend_bundle() -> SpendBundle {
        let parent = [0x11_u8; 32];
        let puzzle = Program::from(vec![0x01_u8]);
        let puzzle_hash = tree_hash_from_bytes(puzzle.as_ref()).expect("tree hash");
        let coin = Coin::new(parent.into(), puzzle_hash.into(), 1);
        let solution = Program::from(vec![0x80]);
        let spend = CoinSpend::new(coin, puzzle, solution);
        SpendBundle::new(vec![spend], Default::default())
    }

    #[test]
    fn mempool_wrapper_spend_bundle_parses() {
        let bundle = sample_spend_bundle();
        let blob = json!({ "spend_bundle": bundle });
        let (kind, parsed, _notes) =
            load_input(&serde_json::to_string(&blob).expect("json")).expect("parse");
        assert_eq!(kind.as_str(), "mempool");
        assert_eq!(parsed.coin_spends.len(), 1);
    }

    #[test]
    fn block_spends_object_parses() {
        let bundle = sample_spend_bundle();
        let blob = json!({ "coin_spends": bundle.coin_spends });
        let (kind, parsed, _notes) =
            load_input(&serde_json::to_string(&blob).expect("json")).expect("parse");
        assert_eq!(kind.as_str(), "block");
        assert_eq!(parsed.coin_spends.len(), 1);
    }
}
