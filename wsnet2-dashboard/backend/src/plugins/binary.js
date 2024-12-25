import "./wasm_exec.js";
import fs from "node:fs";
import path from "node:path";

const dir = path.dirname(new URL(import.meta.url).pathname);
const go = new Go();

const wasm = await WebAssembly.instantiate(fs.readFileSync(path.resolve(dir, "binary.wasm")), go.importObject);
go.run(wasm.instance);
const binary = globalThis.binary;

export function UnmarshalRecursive(src) {
    const ret = binary.UnmarshalRecursive(src);
    if (ret.err != "") {
        return [null, ret.err]
    }
    return [JSON.parse(ret.val), null];
}
