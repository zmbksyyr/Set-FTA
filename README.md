# Set-FTA

Set-FTA is a generic Windows command-line tool for assigning file extensions to
portable applications. The Go executable contains no application-specific
paths or extension lists; callers provide one or more repeatable `--app`
arguments.

```text
SetFTA.exe --app "ProgId|Display name|C:\Path\App.exe|.txt,.log" [--app ...]
```

Each application argument contains four pipe-separated fields:

1. A safe, unique ProgId, such as `PortableAssoc.Editor`.
2. A display name.
3. The absolute path of the executable.
4. A comma-separated extension list.

Run the caller with administrator rights. By default, SetFTA validates all
inputs, previews every change, waits for Enter, applies and verifies each
association, then waits for Enter before closing. Use `--yes --no-pause` for
non-interactive automation.

`examples/Set-My-Associations.bat` is an example preset with reference paths.
Its CONFIG section can be replaced without rebuilding Go. Release archives keep
this directory layout so the example locates `SetFTA.exe` in its parent folder.

## Build

```powershell
go test ./...
go build -trimpath -ldflags="-s -w" -o SetFTA.exe ./cmd/setfta
```

GitHub Actions tests each change and publishes a timestamped Windows AMD64 ZIP
to GitHub Releases. The workflow can also be started manually.

## Implementation note

Windows does not expose a supported API for silently changing an existing
user's protected `UserChoice` default. This project implements the undocumented
UserChoice hash and invokes the Microsoft-supplied `System32\regini.exe` tool.
Windows updates may change that behavior. Review and test the tool before using
it on managed systems.
