use std::ffi::{CStr, CString, c_char};

use coinset_clvm_core::{compile_to_hex, decompile_from_hex, run_program_text};
use coinset_inspect_core::{ExplainLevel, inspect_json_string};
use serde_json::json;

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
