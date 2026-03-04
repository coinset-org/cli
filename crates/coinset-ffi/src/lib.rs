use std::ffi::{CStr, CString, c_char};

use coinset_clvm_core::{
    CurryArg, compile_to_hex, curry_typed, decompile_from_hex, run_program_text, tree_hash_input,
    uncurry,
};
use serde::Deserialize;
use coinset_inspect_core::{ExplainLevel, inspect_json_string};
use serde_json::json;

#[derive(Deserialize)]
struct CurryPayload {
    mod_is_hash: bool,
    args: Vec<CurryArg>,
}

const FLAG_PRETTY: u32 = 1 << 0;
const FLAG_CONDITIONS_ONLY: u32 = 1 << 1;
const FLAG_INCLUDE_COST: u32 = 1 << 2;

#[unsafe(no_mangle)]
pub extern "C" fn coinset_inspect(input_ptr: *const u8, input_len: usize, flags: u32) -> *mut c_char {
    let pretty = (flags & FLAG_PRETTY) != 0;
    let explain_level = if (flags & FLAG_CONDITIONS_ONLY) != 0 {
        ExplainLevel::Conditions
    } else {
        ExplainLevel::Deep
    };
    match with_utf8_input(input_ptr, input_len, |s| inspect_json_string(s, pretty, explain_level)) {
        Ok(s) => CString::new(s).unwrap_or_else(|_| render_error("invalid_output", "output contained NUL")),
        Err(e) => render_error("inspect_failed", &format!("{e:#}")),
    }
    .into_raw()
}

#[unsafe(no_mangle)]
pub extern "C" fn coinset_clvm_decompile(input_ptr: *const u8, input_len: usize, flags: u32) -> *mut c_char {
    let pretty = (flags & FLAG_PRETTY) != 0;
    match with_utf8_input(input_ptr, input_len, |s| {
        let out = decompile_from_hex(s)?;
        if pretty {
            Ok(serde_json::to_string_pretty(&out)?)
        } else {
            Ok(serde_json::to_string(&out)?)
        }
    }) {
        Ok(s) => CString::new(s).unwrap_or_else(|_| render_error("invalid_output", "output contained NUL")),
        Err(e) => render_error("clvm_decompile_failed", &format!("{e:#}")),
    }
    .into_raw()
}

#[unsafe(no_mangle)]
pub extern "C" fn coinset_clvm_compile(input_ptr: *const u8, input_len: usize, flags: u32) -> *mut c_char {
    let pretty = (flags & FLAG_PRETTY) != 0;
    match with_utf8_input(input_ptr, input_len, |s| {
        let out = compile_to_hex(s)?;
        if pretty {
            Ok(serde_json::to_string_pretty(&out)?)
        } else {
            Ok(serde_json::to_string(&out)?)
        }
    }) {
        Ok(s) => CString::new(s).unwrap_or_else(|_| render_error("invalid_output", "output contained NUL")),
        Err(e) => render_error("clvm_compile_failed", &format!("{e:#}")),
    }
    .into_raw()
}

#[unsafe(no_mangle)]
pub extern "C" fn coinset_clvm_run(
    program_ptr: *const u8,
    program_len: usize,
    env_ptr: *const u8,
    env_len: usize,
    max_cost: u64,
    flags: u32,
) -> *mut c_char {
    let pretty = (flags & FLAG_PRETTY) != 0;
    let include_cost = (flags & FLAG_INCLUDE_COST) != 0;

    let program = match unsafe_bytes_to_str(program_ptr, program_len) {
        Ok(s) => s,
        Err(e) => return render_error("invalid_utf8", &format!("{e:#}")).into_raw(),
    };
    let env = match unsafe_bytes_to_str(env_ptr, env_len) {
        Ok(s) => s,
        Err(e) => return render_error("invalid_utf8", &format!("{e:#}")).into_raw(),
    };

    let max_cost = if max_cost == 0 { None } else { Some(max_cost) };
    match run_program_text(program, env, max_cost, include_cost) {
        Ok(out) => {
            let s = if pretty {
                serde_json::to_string_pretty(&out)
            } else {
                serde_json::to_string(&out)
            };
            match s {
                Ok(s) => CString::new(s)
                    .unwrap_or_else(|_| render_error("invalid_output", "output contained NUL")),
                Err(e) => render_error("invalid_output", &format!("{e:#}")),
            }
        }
        Err(e) => render_error("clvm_run_failed", &format!("{e:#}")),
    }
    .into_raw()
}

#[unsafe(no_mangle)]
pub extern "C" fn coinset_clvm_tree_hash(input_ptr: *const u8, input_len: usize, flags: u32) -> *mut c_char {
    let pretty = (flags & FLAG_PRETTY) != 0;
    match with_utf8_input(input_ptr, input_len, |s| {
        let out = tree_hash_input(s)?;
        if pretty {
            Ok(serde_json::to_string_pretty(&out)?)
        } else {
            Ok(serde_json::to_string(&out)?)
        }
    }) {
        Ok(s) => CString::new(s).unwrap_or_else(|_| render_error("invalid_output", "output contained NUL")),
        Err(e) => render_error("clvm_tree_hash_failed", &format!("{e:#}")),
    }
    .into_raw()
}

#[unsafe(no_mangle)]
pub extern "C" fn coinset_clvm_uncurry(input_ptr: *const u8, input_len: usize, flags: u32) -> *mut c_char {
    let pretty = (flags & FLAG_PRETTY) != 0;
    match with_utf8_input(input_ptr, input_len, |s| {
        let out = uncurry(s)?;
        if pretty {
            Ok(serde_json::to_string_pretty(&out)?)
        } else {
            Ok(serde_json::to_string(&out)?)
        }
    }) {
        Ok(s) => CString::new(s).unwrap_or_else(|_| render_error("invalid_output", "output contained NUL")),
        Err(e) => render_error("clvm_uncurry_failed", &format!("{e:#}")),
    }
    .into_raw()
}

#[unsafe(no_mangle)]
pub extern "C" fn coinset_clvm_curry(
    mod_ptr: *const u8,
    mod_len: usize,
    payload_ptr: *const u8,
    payload_len: usize,
    flags: u32,
) -> *mut c_char {
    let pretty = (flags & FLAG_PRETTY) != 0;

    let mod_str = match unsafe_bytes_to_str(mod_ptr, mod_len) {
        Ok(s) => s,
        Err(e) => return render_error("invalid_utf8", &format!("{e:#}")).into_raw(),
    };
    let payload_json = match unsafe_bytes_to_str(payload_ptr, payload_len) {
        Ok(s) => s,
        Err(e) => return render_error("invalid_utf8", &format!("{e:#}")).into_raw(),
    };

    let payload: CurryPayload = match serde_json::from_str(payload_json) {
        Ok(v) => v,
        Err(e) => return render_error("invalid_args", &format!("{e}")).into_raw(),
    };

    match curry_typed(mod_str, payload.mod_is_hash, &payload.args) {
        Ok(out) => {
            let s = if pretty {
                serde_json::to_string_pretty(&out)
            } else {
                serde_json::to_string(&out)
            };
            match s {
                Ok(s) => CString::new(s).unwrap_or_else(|_| render_error("invalid_output", "output contained NUL")),
                Err(e) => render_error("invalid_output", &format!("{e:#}")),
            }
        }
        Err(e) => render_error("clvm_curry_failed", &format!("{e:#}")),
    }
    .into_raw()
}

#[unsafe(no_mangle)]
pub extern "C" fn coinset_free(s: *mut c_char) {
    if s.is_null() {
        return;
    }
    unsafe {
        drop(CString::from_raw(s));
    }
}

#[unsafe(no_mangle)]
pub extern "C" fn coinset_version() -> *const c_char {
    static VERSION: &CStr = unsafe {
        CStr::from_bytes_with_nul_unchecked(concat!(env!("CARGO_PKG_VERSION"), "\0").as_bytes())
    };
    VERSION.as_ptr()
}

fn with_utf8_input<F>(input_ptr: *const u8, input_len: usize, f: F) -> anyhow::Result<String>
where
    F: FnOnce(&str) -> anyhow::Result<String>,
{
    let s = unsafe_bytes_to_str(input_ptr, input_len)?;
    f(s)
}

fn unsafe_bytes_to_str<'a>(ptr: *const u8, len: usize) -> anyhow::Result<&'a str> {
    if ptr.is_null() && len != 0 {
        anyhow::bail!("invalid_argument: input_ptr was null");
    }
    let bytes = unsafe { std::slice::from_raw_parts(ptr, len) };
    Ok(std::str::from_utf8(bytes)?)
}

fn render_error(kind: &str, message: &str) -> CString {
    let obj = json!({
        "schema_version": "coinset.ffi.v1",
        "tool": {
            "name": "coinset-ffi",
            "version": env!("CARGO_PKG_VERSION"),
        },
        "result": {
            "status": "failed",
            "error": { "kind": kind, "message": message },
        }
    });
    let s = serde_json::to_string(&obj).unwrap_or_else(|_| "{\"status\":\"failed\"}".to_string());
    CString::new(s).unwrap_or_else(|_| CString::new("{\"status\":\"failed\"}").expect("cstr"))
}

#[cfg(test)]
mod tests {
    use super::*;
    use chia_protocol::{Coin, CoinSpend, Program, SpendBundle};
    use clvm_utils::tree_hash_from_bytes;
    use serde_json::json;

    #[test]
    fn ffi_inspect_includes_puzzle_recognition_key() {
        let parent = [0x11_u8; 32];
        let puzzle = Program::from(vec![0x01_u8]);
        let puzzle_hash = tree_hash_from_bytes(puzzle.as_ref()).expect("tree hash");
        let coin = Coin::new(parent.into(), puzzle_hash.into(), 1);
        let solution = Program::from(vec![0x80]);
        let spend = CoinSpend::new(coin, puzzle, solution);
        let bundle = SpendBundle::new(vec![spend], Default::default());

        let input = json!({ "spend_bundle": bundle });
        let input_str = serde_json::to_string(&input).expect("json");
        let ptr = input_str.as_ptr();
        let len = input_str.len();

        let out_ptr = coinset_inspect(ptr, len, 0);
        assert!(!out_ptr.is_null());
        let out = unsafe { CStr::from_ptr(out_ptr) }.to_string_lossy().to_string();
        coinset_free(out_ptr);

        let v: serde_json::Value = serde_json::from_str(&out).expect("json parse");
        let spend0 = &v["result"]["spends"][0];
        assert!(spend0.get("puzzle_recognition").is_some(), "missing puzzle_recognition key");
    }

    #[test]
    fn ffi_curry_hash_mode() {
        let mod_input = "37bef360ee858133b69d595a906dc45d01af50379dad515eb9518abb7c1d2a7a";
        let payload = r#"{"mod_is_hash":true,"args":[{"kind":"atom","value":"37bef360ee858133b69d595a906dc45d01af50379dad515eb9518abb7c1d2a7a"},{"kind":"atom","value":"e1c77e2f2bc1758cfd4f3f875adc9b96db322a558072a5b0b8e697d0931bd7fc"},{"kind":"program","value":"ff02ffff01ff02ffff01ff02ffff03ff0bffff01ff02ffff03ffff09ff05ffff1dff0bffff1effff0bff0bffff02ff06ffff04ff02ffff04ff17ff8080808080808080ffff01ff02ff17ff2f80ffff01ff088080ff0180ffff01ff04ffff04ff04ffff04ff05ffff04ffff02ff06ffff04ff02ffff04ff17ff80808080ff80808080ffff02ff17ff2f808080ff0180ffff04ffff01ff32ff02ffff03ffff07ff0580ffff01ff0bffff0102ffff02ff06ffff04ff02ffff04ff09ff80808080ffff02ff06ffff04ff02ffff04ff0dff8080808080ffff01ff0bffff0101ff058080ff0180ff018080ffff04ffff01b0b4301d5383702bb6824712a96425d343aaa4448e288de4d2b243425cdce9fd5fd19e0d4ddaacac858216644c05bac4d0ff018080"}]}"#;

        let mod_ptr = mod_input.as_ptr();
        let mod_len = mod_input.len();
        let payload_ptr = payload.as_ptr();
        let payload_len = payload.len();

        let out_ptr = coinset_clvm_curry(mod_ptr, mod_len, payload_ptr, payload_len, 0);
        assert!(!out_ptr.is_null());
        let out = unsafe { CStr::from_ptr(out_ptr) }.to_string_lossy().to_string();
        coinset_free(out_ptr);

        let v: serde_json::Value = serde_json::from_str(&out).expect("json parse");
        assert_eq!(
            v["curried"]["tree_hash"].as_str().unwrap(),
            "0x46a759d7e88c9f5fd43e4e09956656ea4325c947e37b8730bbc23c3be9406985"
        );
    }
}
