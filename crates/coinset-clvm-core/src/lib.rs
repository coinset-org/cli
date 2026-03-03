use anyhow::{Context, Result, bail};
use chialisp::classic::clvm::OPERATORS_LATEST_VERSION;
use chialisp::classic::clvm_tools::binutils::{assemble, disassemble};
use clvmr::allocator::Allocator;
use clvmr::chia_dialect::{ChiaDialect, MEMPOOL_MODE};
use clvmr::cost::Cost;
use clvmr::run_program::run_program;
use clvmr::serde::{node_from_bytes_backrefs, node_to_bytes};
use serde::Serialize;

#[derive(Debug, Clone, Serialize)]
pub struct CompileOutput {
    pub bytes: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct DecompileOutput {
    pub clvm: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct RunOutput {
    pub result: String,
    pub cost: Option<u64>,
}

pub fn compile_to_hex(program: &str) -> Result<CompileOutput> {
    let mut allocator = Allocator::new();
    let node = assemble(&mut allocator, program)
        .map_err(|e| anyhow::anyhow!("failed to assemble CLVM: {e}"))?;
    let bytes = node_to_bytes(&allocator, node)?;
    Ok(CompileOutput {
        bytes: encode_hex_prefixed(&bytes),
    })
}

pub fn decompile_from_hex(hex_bytes: &str) -> Result<DecompileOutput> {
    let bytes = decode_hex_prefixed(hex_bytes)?;
    let mut allocator = Allocator::new();
    let node = node_from_bytes_backrefs(&mut allocator, &bytes)?;
    Ok(DecompileOutput {
        clvm: disassemble(&allocator, node, Some(OPERATORS_LATEST_VERSION)),
    })
}

pub fn run_program_text(
    program: &str,
    env: &str,
    max_cost: Option<u64>,
    include_cost: bool,
) -> Result<RunOutput> {
    let mut allocator = Allocator::new();
    let program_node = parse_program_input(&mut allocator, program)?;
    let env_node = parse_program_input(&mut allocator, env)?;

    let dialect = ChiaDialect::new(MEMPOOL_MODE);
    let cost_limit: Cost = match max_cost {
        Some(v) => v as Cost,
        None => 0,
    };

    let reduction = run_program(&mut allocator, &dialect, program_node, env_node, cost_limit)
        .map_err(|e| anyhow::anyhow!("CLVM runtime error: {e:?}"))?;
    let cost = reduction.0 as u64;
    let result_node = reduction.1;

    Ok(RunOutput {
        result: disassemble(&allocator, result_node, Some(OPERATORS_LATEST_VERSION)),
        cost: if include_cost { Some(cost) } else { None },
    })
}

fn parse_program_input(allocator: &mut Allocator, input: &str) -> Result<clvmr::allocator::NodePtr> {
    if looks_like_hex(input) {
        let bytes = decode_hex_prefixed(input)?;
        let node = node_from_bytes_backrefs(allocator, &bytes)?;
        return Ok(node);
    }
    let node = assemble(allocator, input).map_err(|e| anyhow::anyhow!("failed to assemble CLVM: {e}"))?;
    Ok(node)
}

fn decode_hex_prefixed(input: &str) -> Result<Vec<u8>> {
    let raw = input
        .strip_prefix("0x")
        .or_else(|| input.strip_prefix("0X"))
        .unwrap_or(input)
        .trim();
    if raw.is_empty() {
        return Ok(Vec::new());
    }
    if raw.len() % 2 != 0 {
        bail!("hex input must have even length");
    }
    Ok(hex::decode(raw).with_context(|| "failed to decode hex input")?)
}

fn encode_hex_prefixed(bytes: &[u8]) -> String {
    format!("0x{}", hex::encode(bytes))
}

fn looks_like_hex(input: &str) -> bool {
    let raw = input
        .strip_prefix("0x")
        .or_else(|| input.strip_prefix("0X"))
        .unwrap_or(input);
    !raw.is_empty() && raw.len() % 2 == 0 && raw.bytes().all(|b| b.is_ascii_hexdigit())
}
