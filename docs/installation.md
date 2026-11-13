# Installation

## Prebuilt Binaries (GitHub Releases)

Prebuilt static binaries for Linux, macOS, and Windows (`amd64` and `arm64`) are published on the [GitHub Releases](https://github.com/Carter907/qst/releases) page.

1. Download the archive for your platform (`qst_<version>_<os>_<arch>.tar.gz` or `.zip` on Windows) along with `checksums.txt`.
2. Verify the checksum:
   - **macOS / Linux**: `sha256sum -c --ignore-missing checksums.txt`
   - **Windows (PowerShell)**: `Get-FileHash .\qst_*.zip -Algorithm SHA256`
3. Extract the archive and move `qst` (or `qst.exe`) into a directory on your `PATH`.

## Building from Source

`qst` cli can run on all major operating systems as long as you have Go installed and can clone this repository.

**Install locally using `go build`**:

Clone the repository:
```sh
git clone https://github.com/Carter907/qst.git
cd qst
```

Build and install to your `GOPATH/bin`:

- **macOS / Linux**:
  ```sh
  go build -o $(go env GOPATH)/bin/qst .
  ```

- **Windows (PowerShell)**:
  ```powershell
  go build -o "$((go env GOPATH))\bin\qst.exe" .
  ```

- **Windows (Command Prompt)**:
  ```cmd
  go build -o "%USERPROFILE%\go\bin\qst.exe" .
  ```

*(Optional)* If you prefer to build the executable directly in the current directory:
- **macOS / Linux**: `go build -o qst .`
- **Windows**: `go build -o qst.exe .`

Make sure your Go bin directory (e.g., `$GOPATH/bin` or `%USERPROFILE%\go\bin`) is included in your system's `PATH`.