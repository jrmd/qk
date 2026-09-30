# QK Command Runner

QK finds full-stack projects below the current directory and runs commands in
parallel. A target is a directory containing both `package.json` and
`composer.json`.

## Usage

```sh
qk ls                         # preview discovered targets
qk install                    # install JavaScript and Composer dependencies
qk build                      # run build:prod
qk cmd <command> [args...]    # run any executable in every target
qk npm <args...>
qk yarn <args...>
qk composer <args...>
qk watch                      # run dev, watch:dev, or start when available
qk dev                        # install, build, then watch
```

All commands accept these global flags:

```text
-d, --depth int         maximum directory depth to search (default 3, -1 unlimited)
-x, --exclude-current   exclude the current directory when it is a project
-j, --joined            show interleaved command output (default true)
```

QK skips `.git`, `.idea`, `node_modules`, and `vendor` while discovering
projects.

## Configuration

Create `~/.qk.json` to change display defaults:

```json
{
  "ShowTimer": true,
  "ShowScripts": true,
  "ShowStdout": false
}
```
