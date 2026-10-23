# Spec Delta

## ADDED Requirements

### Requirement: License files of the module

The site SHALL serve the `LICENSE` and `NOTICE` of the repository next to the WebAssembly module, as `wasm/LICENSE` and `wasm/NOTICE`, both in the deployed site and in the local preview. A push to `main` that changes them SHALL deploy the site again.

#### Scenario: License files on the site

- **WHEN** the site is assembled with the WASM archive of a release or with a local WebAssembly build
- **THEN** its `wasm` directory contains `LICENSE` and `NOTICE`, identical to those of the repository

#### Scenario: Change of the notices

- **WHEN** a commit that changes `NOTICE` is pushed to `main`
- **THEN** the site is deployed again, with that `NOTICE` in its `wasm` directory
