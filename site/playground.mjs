// SPDX-License-Identifier: MIT
// Provenance: original

// Playground logic of the landing page, without the DOM, so that Node.js
// can test it: loading the module once, rendering sources as SVG, the
// examples, and the download of a result. page.mjs connects it to the page.

// examples are the diagrams the visitor can pick. They are all types that
// the WebAssembly build renders without Graphviz.
export const examples = [
  {
    name: 'Sequence diagram',
    source: `@startuml
Alice -> Bob: Authentication Request
Bob --> Alice: Authentication Response

Alice -> Bob: Another authentication Request
Alice <-- Bob: Another authentication Response
@enduml
`,
  },
  {
    name: 'Sequence diagram with groups and notes',
    source: `@startuml
actor Customer
participant Shop
database Stock

Customer -> Shop : order(item)
activate Shop
Shop -> Stock : reserve(item)
alt item in stock
  Stock --> Shop : reserved
  Shop --> Customer : confirmation
else out of stock
  Stock --> Shop : none left
  Shop --> Customer : apology
end
deactivate Shop
note right of Customer : groups, notes\\nand activations
@enduml
`,
  },
  {
    name: 'Activity diagram',
    source: `@startuml
start
:Read the diagram source;
if (Valid syntax?) then (yes)
  :Lay out the diagram;
  :Write the SVG;
else (no)
  :Draw the error image;
endif
stop
@enduml
`,
  },
  {
    name: 'Activity diagram with swimlanes and a loop',
    source: `@startuml
|Writer|
start
repeat
  :Edit the source;
  |gopuml|
  :Render the diagram;
  |Writer|
repeat while (Looks right?) is (no) not (yes)
:Commit the diagram;
stop
@enduml
`,
  },
];

// describe returns the text of a render error: its message, with its line
// when the message does not name it.
export function describe(error) {
  if (error.line && !error.message.includes(`line ${error.line}`)) {
    return `Line ${error.line}: ${error.message}`;
  }
  return error.message;
}

// download returns the file of a successful result, or null for an error.
export function download(result) {
  if (!result || result.error || !result.svg) {
    return null;
  }
  return { filename: 'diagram.svg', blob: result.svg };
}

// createPlayground returns a playground whose module comes from load, a
// function that resolves to an object with render(source, options), as
// load() of plantuml.mjs does. Nothing is loaded before activate or render.
export function createPlayground(load) {
  let module;
  let status = 'idle';
  let latest = 0;
  return {
    // status is idle, loading, ready or failed.
    get status() {
      return status;
    },

    // activate loads the module once. It resolves when it is loaded and
    // rejects when it cannot be.
    activate() {
      if (!module) {
        status = 'loading';
        module = Promise.resolve()
          .then(load)
          .then(
            (m) => {
              status = 'ready';
              return m;
            },
            (err) => {
              status = 'failed';
              throw err;
            },
          );
      }
      return module;
    },

    // render renders source as SVG. It resolves to null when a newer render
    // started in the meantime, else to { svg, error, width, height }: svg is
    // a Blob of the SVG (the error image for an error), error is null or
    // { kind, message, line, text }.
    async render(source) {
      const n = ++latest;
      const plantuml = await this.activate();
      const r = await plantuml.render(source, { format: 'svg' });
      if (n !== latest) {
        return null;
      }
      return {
        svg: r.data ? new Blob([r.data], { type: 'image/svg+xml' }) : null,
        error: r.error ? { ...r.error, text: describe(r.error) } : null,
        width: r.width,
        height: r.height,
      };
    },
  };
}
