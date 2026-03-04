use anyhow::{Context, Result, bail};
use chialisp::classic::clvm::OPERATORS_LATEST_VERSION;
use chialisp::classic::clvm_tools::binutils::{assemble, disassemble};
use clvm_traits::{FromClvm, ToClvm, Raw};
use clvm_utils::{CurriedProgram, TreeHash, curry_tree_hash, tree_hash_atom};
use clvm_utils::tree_hash;
use clvmr::allocator::{Allocator, NodePtr, SExp};
use clvmr::chia_dialect::{ChiaDialect, MEMPOOL_MODE};
use clvmr::cost::Cost;
use clvmr::run_program::run_program;
use clvmr::serde::{node_from_bytes_backrefs, node_to_bytes};
use serde::{Deserialize, Serialize};

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

#[derive(Debug, Clone, Serialize)]
pub struct TreeHashOutput {
    pub tree_hash: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct CurryOutput {
    pub curried: CurriedInfo,
    pub mod_: ModInfo,
    pub args: Vec<ArgInfo>,
}

#[derive(Debug, Clone, Serialize)]
pub struct UncurryOutput {
    pub curried: bool,
    pub mod_: Option<ModBytes>,
    pub args: Option<Vec<UncurriedArg>>,
}

#[derive(Debug, Clone, Serialize)]
pub struct CurriedInfo {
    #[serde(skip_serializing_if = "Option::is_none")]
    pub bytes: Option<String>,
    pub tree_hash: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct ModInfo {
    pub tree_hash: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct ArgInfo {
    pub tree_hash: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct ModBytes {
    pub bytes: String,
    pub tree_hash: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct UncurriedArg {
    pub bytes: String,
    pub tree_hash: String,
    pub kind: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub value: Option<String>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct CurryArg {
    pub kind: CurryArgKind,
    pub value: String,
}

#[derive(Debug, Clone, Deserialize, PartialEq)]
#[serde(rename_all = "snake_case")]
pub enum CurryArgKind {
    Atom,
    TreeHash,
    Program,
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

pub fn tree_hash_input(input: &str) -> Result<TreeHashOutput> {
    let mut allocator = Allocator::new();
    let node = parse_program_input(&mut allocator, input)?;
    let h = tree_hash(&allocator, node);
    Ok(TreeHashOutput {
        tree_hash: encode_hex_prefixed(h.as_ref()),
    })
}

pub fn curry_typed(mod_input: &str, mod_is_hash: bool, args: &[CurryArg]) -> Result<CurryOutput> {
    let hash_mode = mod_is_hash || args.iter().any(|a| a.kind == CurryArgKind::TreeHash);

    if hash_mode {
        return curry_hash_mode(mod_input, mod_is_hash, args);
    }

    let mut allocator = Allocator::new();
    let mod_node = parse_program_input(&mut allocator, mod_input)?;

    let mut arg_nodes = Vec::<NodePtr>::with_capacity(args.len());
    for a in args {
        let node = match a.kind {
            CurryArgKind::Atom => {
                let bytes = decode_hex_prefixed(&a.value)?;
                allocator.new_atom(&bytes)?
            }
            CurryArgKind::Program => parse_program_input(&mut allocator, &a.value)?,
            CurryArgKind::TreeHash => unreachable!(),
        };
        arg_nodes.push(node);
    }

    let curried_args = build_curried_args(&mut allocator, &arg_nodes)?;
    let curried = CurriedProgram {
        program: Raw(mod_node),
        args: Raw(curried_args),
    }
    .to_clvm(&mut allocator)?;

    let bytes = node_to_bytes(&allocator, curried)?;
    let curried_tree = tree_hash(&allocator, curried);

    Ok(CurryOutput {
        curried: CurriedInfo {
            bytes: Some(encode_hex_prefixed(&bytes)),
            tree_hash: encode_hex_prefixed(curried_tree.as_ref()),
        },
        mod_: ModInfo {
            tree_hash: encode_hex_prefixed(tree_hash(&allocator, mod_node).as_ref()),
        },
        args: arg_nodes
            .iter()
            .map(|n| ArgInfo {
                tree_hash: encode_hex_prefixed(tree_hash(&allocator, *n).as_ref()),
            })
            .collect(),
    })
}

fn curry_hash_mode(mod_input: &str, mod_is_hash: bool, args: &[CurryArg]) -> Result<CurryOutput> {
    let mod_hash = if mod_is_hash {
        let bytes = decode_hex_prefixed(mod_input)?;
        if bytes.len() != 32 {
            bail!("mod hash must be exactly 32 bytes");
        }
        TreeHash::from(<[u8; 32]>::try_from(bytes.as_slice()).unwrap())
    } else {
        let mut allocator = Allocator::new();
        let mod_node = parse_program_input(&mut allocator, mod_input)?;
        tree_hash(&allocator, mod_node)
    };

    let mut arg_hashes = Vec::<TreeHash>::with_capacity(args.len());
    let mut allocator = Allocator::new();
    for a in args {
        let h = match a.kind {
            CurryArgKind::Atom => {
                let bytes = decode_hex_prefixed(&a.value)?;
                tree_hash_atom(&bytes)
            }
            CurryArgKind::TreeHash => {
                let bytes = decode_hex_prefixed(&a.value)?;
                if bytes.len() != 32 {
                    bail!("tree hash arg must be exactly 32 bytes");
                }
                TreeHash::from(<[u8; 32]>::try_from(bytes.as_slice()).unwrap())
            }
            CurryArgKind::Program => {
                let node = parse_program_input(&mut allocator, &a.value)?;
                tree_hash(&allocator, node)
            }
        };
        arg_hashes.push(h);
    }

    let curried_hash = curry_tree_hash(mod_hash, &arg_hashes);

    Ok(CurryOutput {
        curried: CurriedInfo {
            bytes: None,
            tree_hash: encode_hex_prefixed(curried_hash.as_ref()),
        },
        mod_: ModInfo {
            tree_hash: encode_hex_prefixed(mod_hash.as_ref()),
        },
        args: arg_hashes
            .iter()
            .map(|h| ArgInfo {
                tree_hash: encode_hex_prefixed(h.as_ref()),
            })
            .collect(),
    })
}

pub fn uncurry(program: &str) -> Result<UncurryOutput> {
    let mut allocator = Allocator::new();
    let node = parse_program_input(&mut allocator, program)?;

    let parsed = CurriedProgram::<Raw<NodePtr>, Raw<NodePtr>>::from_clvm(&allocator, node);
    let Ok(curried) = parsed else {
        return Ok(UncurryOutput {
            curried: false,
            mod_: None,
            args: None,
        });
    };

    let mod_node = curried.program.0;
    let args_node = curried.args.0;
    let args_list = parse_curried_args(&allocator, args_node)?;

    let mod_bytes = node_to_bytes(&allocator, mod_node)?;
    let mod_tree = tree_hash(&allocator, mod_node);

    let mut args_out = Vec::with_capacity(args_list.len());
    for a in args_list {
        let b = node_to_bytes(&allocator, a)?;
        let h = tree_hash(&allocator, a);
        let (kind, value) = match allocator.sexp(a) {
            SExp::Atom => {
                let raw = allocator.atom(a);
                ("atom".to_string(), Some(encode_hex_prefixed(raw.as_ref())))
            }
            SExp::Pair(..) => ("tree".to_string(), None),
        };
        args_out.push(UncurriedArg {
            bytes: encode_hex_prefixed(&b),
            tree_hash: encode_hex_prefixed(h.as_ref()),
            kind,
            value,
        });
    }

    Ok(UncurryOutput {
        curried: true,
        mod_: Some(ModBytes {
            bytes: encode_hex_prefixed(&mod_bytes),
            tree_hash: encode_hex_prefixed(mod_tree.as_ref()),
        }),
        args: Some(args_out),
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

fn build_curried_args(allocator: &mut Allocator, args: &[NodePtr]) -> Result<NodePtr> {
    let mut rest = allocator.new_small_number(1)?;
    let nil = allocator.nil();
    for arg in args.iter().rev() {
        let q = allocator.new_small_number(1)?;
        let quote = allocator.new_pair(q, *arg)?;
        let tail = allocator.new_pair(rest, nil)?;
        let inner = allocator.new_pair(quote, tail)?;
        let op = allocator.new_small_number(4)?;
        rest = allocator.new_pair(op, inner)?;
    }
    Ok(rest)
}

fn parse_curried_args(allocator: &Allocator, mut node: NodePtr) -> Result<Vec<NodePtr>> {
    let mut out = Vec::<NodePtr>::new();
    loop {
        match allocator.sexp(node) {
            clvmr::allocator::SExp::Atom => {
                // termination is atom 1
                if allocator.atom(node).as_ref() == [1_u8] {
                    break;
                }
                break;
            }
            clvmr::allocator::SExp::Pair(op, rest) => {
                // op must be atom 4
                if !matches!(allocator.sexp(op), clvmr::allocator::SExp::Atom)
                    || allocator.atom(op).as_ref() != [4_u8]
                {
                    break;
                }
                let clvmr::allocator::SExp::Pair(quoted_arg, rest_pair) = allocator.sexp(rest) else {
                    break;
                };
                // quoted_arg is (1 . arg)
                let clvmr::allocator::SExp::Pair(q, arg) = allocator.sexp(quoted_arg) else {
                    break;
                };
                if !matches!(allocator.sexp(q), clvmr::allocator::SExp::Atom) || allocator.atom(q).as_ref() != [1_u8] {
                    break;
                }
                out.push(arg);

                let clvmr::allocator::SExp::Pair(next, nil) = allocator.sexp(rest_pair) else {
                    break;
                };
                if !matches!(allocator.sexp(nil), clvmr::allocator::SExp::Atom) || !allocator.atom(nil).as_ref().is_empty() {
                    break;
                }
                node = next;
            }
        }
    }
    Ok(out)
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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_curry_hash_mode_cat() {
        let result = curry_typed(
            "37bef360ee858133b69d595a906dc45d01af50379dad515eb9518abb7c1d2a7a",
            true,
            &[
                CurryArg { kind: CurryArgKind::Atom, value: "37bef360ee858133b69d595a906dc45d01af50379dad515eb9518abb7c1d2a7a".to_string() },
                CurryArg { kind: CurryArgKind::Atom, value: "e1c77e2f2bc1758cfd4f3f875adc9b96db322a558072a5b0b8e697d0931bd7fc".to_string() },
                CurryArg { kind: CurryArgKind::Program, value: "ff02ffff01ff02ffff01ff02ffff03ff0bffff01ff02ffff03ffff09ff05ffff1dff0bffff1effff0bff0bffff02ff06ffff04ff02ffff04ff17ff8080808080808080ffff01ff02ff17ff2f80ffff01ff088080ff0180ffff01ff04ffff04ff04ffff04ff05ffff04ffff02ff06ffff04ff02ffff04ff17ff80808080ff80808080ffff02ff17ff2f808080ff0180ffff04ffff01ff32ff02ffff03ffff07ff0580ffff01ff0bffff0102ffff02ff06ffff04ff02ffff04ff09ff80808080ffff02ff06ffff04ff02ffff04ff0dff8080808080ffff01ff0bffff0101ff058080ff0180ff018080ffff04ffff01b0b4301d5383702bb6824712a96425d343aaa4448e288de4d2b243425cdce9fd5fd19e0d4ddaacac858216644c05bac4d0ff018080".to_string() },
            ],
        );
        let out = result.expect("curry_typed should succeed");
        println!("curried tree_hash: {}", out.curried.tree_hash);
    }
}

fn looks_like_hex(input: &str) -> bool {
    let raw = input
        .strip_prefix("0x")
        .or_else(|| input.strip_prefix("0X"))
        .unwrap_or(input);
    !raw.is_empty() && raw.len() % 2 == 0 && raw.bytes().all(|b| b.is_ascii_hexdigit())
}
