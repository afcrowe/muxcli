# muxcli

Small CLI for basic Mux Video operations: create/upload, delete, get asset details, create/list static renditions, and request temporary master download links.

This repository contains a single Go binary `muxcli` that talks to the Mux Video REST API using your Token ID and Secret.

## Build

From the project root:

```bash
go build -o muxcli
```

This produces a `muxcli` executable in the current directory.

## Required flags

- `--key-id` : your Mux Token ID
- `--secret-key` : your Mux Token Secret

Currently the CLI expects these values as flags on each command (no environment variable support yet).

## Commands & examples

Create (create a direct upload, upload a local file, then wait for the asset to be ready):

```bash
./muxcli create --key-id <TOKEN_ID> --secret-key <SECRET> --input-file /path/to/video.mp4
```

Delete a single asset by id:

```bash
./muxcli delete --key-id <TOKEN_ID> --secret-key <SECRET> --asset-id <ASSET_ID>
```

Get asset details (prints pretty JSON):

```bash
./muxcli get --key-id <TOKEN_ID> --secret-key <SECRET> --asset-id <ASSET_ID>
```

Create a static rendition (resolution defaults to `highest`):

```bash
./muxcli create-rendition --key-id <TOKEN_ID> --secret-key <SECRET> --asset-id <ASSET_ID> --resolution highest
```

List static renditions for an asset:

```bash
./muxcli list-renditions --key-id <TOKEN_ID> --secret-key <SECRET> --asset-id <ASSET_ID>
```

Get temporary master MP4 URL (one-shot):

```bash
./muxcli get-master --key-id <TOKEN_ID> --secret-key <SECRET> --asset-id <ASSET_ID>
```

Get temporary master MP4 URL but poll until it appears (poll interval 5s, default timeout 300s):

```bash
# poll until master.url is present (default timeout 300s)
./muxcli get-master --key-id <TOKEN_ID> --secret-key <SECRET> --asset-id <ASSET_ID> --wait

# poll with a 10 minute timeout
./muxcli get-master --key-id <TOKEN_ID> --secret-key <SECRET> --asset-id <ASSET_ID> --wait --timeout 600
```

Notes
- The CLI uses basic polling for a few operations (create/upload -> waits for asset to be created, and `get-master --wait` polls for master URL). You can adjust the timeout with the `--timeout` flag for `--wait`.
- There's a helper `DeleteAssetsFromFile` in the codebase (in `commands.go`) for batch deletion, but a `--batch-file` flag is not yet wired into the CLI.

Contributing / next steps
- Add `--batch-file` to `delete` to delete multiple asset IDs from a file.
- Add environment variable support for MUX credentials (convenience) and optional config file.
- Add unit and integration tests (mocking Mux API responses).
