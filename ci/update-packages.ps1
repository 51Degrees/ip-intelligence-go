param (
    [Parameter(Mandatory)][string]$RepoName,
    [Parameter(Mandatory)][string]$OrgName,
    [bool]$DryRun = $false
)
$ErrorActionPreference = "Stop"
$PSNativeCommandUseErrorActionPreference = $true

Push-Location $RepoName
try {
    go get -u ./...
    go mod tidy

    Write-Output "Cloning latest ip-intelligence-cxx..."
    git clone --depth=1 --recurse-submodules --shallow-submodules "https://github.com/$OrgName/ip-intelligence-cxx.git"

    Write-Output "Generating amalgamation..."
    $src = "ip-intelligence-cxx/src"
    awk -f ci/amalgamate.awk $src/fiftyone.h $src/ip-graph-cxx/graph.h >ipi_interop/ip-intelligence-cxx.h
    awk -f ci/amalgamate.awk $src/common-cxx/*.c $src/ip-graph-cxx/*.c $src/*.c >ipi_interop/ip-intelligence-cxx.c
} finally {
    Write-Output "Cleaning up..."
    Remove-Item -Recurse -Force ip-intelligence-cxx
    Pop-Location
}
