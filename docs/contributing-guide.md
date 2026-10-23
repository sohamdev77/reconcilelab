# Development guide

See [CONTRIBUTING](../CONTRIBUTING.md). Unit tests use local fixtures; integration tests exercise the CLI in process. CI separately compiles platform binaries. Regenerate demo assets using `make demo`; output comes from the actual compiled executable. Run `make docs` to check local Markdown links and required assets.
