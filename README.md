# Readme

## Introduction
checkpoint.af is a convenient CLI Utilty and light-weight API server for saving, versioning, and migrating snapshots of a Letta agent's state as an [agentfile](https://github.com/letta-ai/agent-file).

## Functionality

### Save
Save a snapshot of the agent as an agentfile.

Supports the following parameters:
- `--base_url` URL of Letta server running API (default: api.letta.com)
- `--api_key` LETTA_API_KEY to authenticate to Letta server
- `--agent_id` pass agent id to export
- `--overwrite` for in-place overwriting of existing agentfile
- `--dest` pass directory or blob storage prefix to which the export should be saved

Supports the following destinations:
- In-memory
- File
- S3
- GCS
- Azure Blob
- Any blob storage compatible with the above interfaces

[Full list here](https://gocloud.dev/howto/blob/#services)

### rollback (unimplemented)
Return the agent to a previously snapshotted state (optionally in place)

### migrate (unimplemented)
Migrate a "rolled-back" agent to a different snapshotted state in it's future