// SPDX-License-Identifier: MIT
// Provenance: original

// ES module loader of the PlantUML WebAssembly build, for browsers and
// Node.js. Usage:
//
//   import { load } from './plantuml.mjs';
//   const plantuml = await load();
//   const result = await plantuml.render('@startuml\nAlice -> Bob\n@enduml', { format: 'svg' });
//
// result has data (a Uint8Array or null), width, height, description,
// pages, diagrams and error (null, or { kind, message, line }).

let loaded;

export function load(wasmURL = new URL('./plantuml.wasm', import.meta.url)) {
  loaded ??= start(wasmURL);
  return loaded;
}

async function start(wasmURL) {
  if (!globalThis.Go) {
    await import('./wasm_exec.js');
  }
  const go = new globalThis.Go();
  let instance;
  if (wasmURL.protocol === 'file:' && typeof process !== 'undefined' && process.versions?.node) {
    const { readFile } = await import('node:fs/promises');
    ({ instance } = await WebAssembly.instantiate(await readFile(wasmURL), go.importObject));
  } else {
    ({ instance } = await WebAssembly.instantiateStreaming(fetch(wasmURL), go.importObject));
  }
  go.run(instance);
  const api = globalThis.plantuml;
  return { render: (source, options = {}) => api.render(source, options) };
}
