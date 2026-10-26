# Install the incident-patched compiler without changing the Windows container OS.
$ErrorActionPreference = "Stop"
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
$GoVersion = "1.26.10"
$InstallRoot = Join-Path $env:TEMP "go-$GoVersion-patched"
$GoRoot = Join-Path $InstallRoot "go"
$GoExe = Join-Path $GoRoot "bin\go.exe"
if (!(Test-Path $GoExe)) {
    New-Item -ItemType Directory -Force -Path $InstallRoot | Out-Null
    $Archive = Join-Path $InstallRoot "go.windows-amd64.zip"
    Invoke-WebRequest -UseBasicParsing -Uri "https://go.dev/dl/go1.26.10.windows-amd64.zip" -OutFile $Archive
    Expand-Archive -Path $Archive -DestinationPath $InstallRoot -Force
    Remove-Item $Archive
}
$env:GOROOT = $GoRoot
$env:PATH = "$GoRoot\bin;$env:PATH"
$env:GOTOOLCHAIN = "local"
& $GoExe version
if ($LASTEXITCODE -ne 0) { throw "Patched Go SDK could not start" }
