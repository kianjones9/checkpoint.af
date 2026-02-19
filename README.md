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
- `--only_on_diff` if --overwrite specified, skip overwriting if file is unchanged
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


## Usage examples

### Run server with access to GCS
```
docker run -p 8080:8080 \
    -v /tmp/sa-key.json:/gcloud/sa-key.json \
    -e GOOGLE_APPLICATION_CREDENTIALS=/gcloud/sa-key.json \
    -d kianjones9/checkpoint.af
```

### Run as cli tool
```
# from source
LETTA_API_KEY="sk-let-MTN***...***YQ==" go run cmd/cli/checkpoint.go --dest="gs://agentfiles/coding/"  --agent_id="agent-226dd8d4-09bf-4536-920e-aee9d91d14cb"

# from tagged release
LETTA_API_KEY="sk-let-MTN***...***YQ==" checkpoint --dest="gs://agentfiles/coding/"  --agent_id="agent-226dd8d4-09bf-4536-920e-aee9d91d14cb"
```

### Or even over local filesystem
```
LETTA_API_KEY="sk-let-MTN***...***YQ==" checkpoint --dest="file:///agentfiles/coding/"  --agent_id="agent-226dd8d4-09bf-4536-920e-aee9d91d14cb"
```