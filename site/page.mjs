// SPDX-License-Identifier: MIT
// Provenance: original

// DOM part of the playground of the landing page. It loads the module when
// the playground is on screen or gets the focus, renders the source 300 ms
// after the last edit, and shows the SVG in an <img> from a blob: URL, where
// the scripts and event attributes of an SVG do not run.
import { createPlayground, examples, download } from './playground.mjs';

const panel = document.querySelector('#playground .pg');
const source = document.getElementById('pg-source');
const select = document.getElementById('pg-example');
const status = document.getElementById('pg-status');
const image = document.getElementById('pg-image');
const save = document.getElementById('pg-download');

const playground = createPlayground(async () => {
  const { load } = await import('./wasm/plantuml.mjs');
  return load();
});

function setStatus(text, error = false) {
  status.textContent = text;
  status.classList.toggle('error', error);
}

// url is the blob: URL of the shown image.
let url;

function show(result) {
  if (!result) {
    // A newer render is running.
    return;
  }
  if (url) {
    URL.revokeObjectURL(url);
  }
  url = result.svg ? URL.createObjectURL(result.svg) : undefined;
  image.hidden = !url;
  if (url) {
    image.src = url;
    image.alt = result.error ? `Error image: ${result.error.text}` : 'The rendered diagram';
  }
  if (result.error) {
    setStatus(result.error.text, true);
  } else {
    setStatus(`Rendered in this page, ${result.width} x ${result.height} pixels.`);
  }
  const file = download(result);
  save.hidden = !file;
  if (file) {
    save.href = url;
    save.download = file.filename;
  }
}

async function render() {
  if (playground.status !== 'ready') {
    // While the module loads, the source is rendered once it is loaded.
    start();
    return;
  }
  try {
    show(await playground.render(source.value));
  } catch (err) {
    setStatus(`The diagram could not be rendered: ${err.message}`, true);
  }
}

function start() {
  if (playground.status !== 'idle') {
    return;
  }
  setStatus('Loading the WebAssembly module...');
  playground.activate().then(render, (err) => {
    setStatus(`The playground could not be loaded (${err.message}). Reload the page to try again.`, true);
  });
}

for (const [i, e] of examples.entries()) {
  select.add(new Option(e.name, String(i)));
}
source.value = examples[0].source;
panel.hidden = false;
document.getElementById('pg-start').hidden = true;

let timer;
source.addEventListener('input', () => {
  clearTimeout(timer);
  timer = setTimeout(render, 300);
});
select.addEventListener('change', () => {
  clearTimeout(timer);
  source.value = examples[Number(select.value)].source;
  render();
});
source.addEventListener('focus', start);
select.addEventListener('focus', start);

if ('IntersectionObserver' in window) {
  const observer = new IntersectionObserver((entries) => {
    if (entries.some((e) => e.isIntersecting)) {
      observer.disconnect();
      start();
    }
  });
  observer.observe(panel);
}
