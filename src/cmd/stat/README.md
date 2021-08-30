# Comparing reports between versions

A tool for comparing reports between versions. At the output, it creates a markdown file with the results.

Run:

```shell
cd src/cmd/stat
make compare -i PROJECT_DIR=/Users/petrmakhnev/psalm/ NOVERIFY_ARGS='--index-only-files="./stubs" ./src'
```

The current master is automatically cloned into the folder, compiles and runs the analysis of the project on it If the master is already cloned, it will be updated.