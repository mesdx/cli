# MesDX Memory

MesDX “memory” is a directory of Markdown files (committed to your repo) that the CLI/MCP server indexes for fast search.

## Valid memories (what gets indexed)

- **Project-scoped memories** (`mesdx.scope: project`) are always indexed.
- **File-scoped memories** (`mesdx.scope: file`) are indexed **only if** `mesdx.file` points to an existing repo-relative file.
  - If the referenced file does not exist, the memory is **excluded from the search index** (and will be reconciled as `fileStatus: deleted`).

## Search chunking behavior

When indexing a memory Markdown file, MesDX splits it into chunks using **top-level Markdown headers**:

- A new chunk starts at each line beginning with **`# `** (single `#`).
- Each chunk is indexed as its own search document for better relevance.

This means `mesdx.memorySearch` results are effectively “best matching chunk per memory file”.

