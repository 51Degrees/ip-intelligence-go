param (
    [Parameter(Mandatory)][string]$RepoName
)
$ErrorActionPreference = "Stop"

# TODO: remove
Write-Warning "No tests yet, not fetching assets"
exit 0

$assets = New-Item -ItemType Directory -Path assets -Force
$assetsDestination = "$RepoName"
$files = "51Degrees-LiteV4.1.hash", "20000 Evidence Records.yml"

foreach ($file in $files) {
    if (!(Test-Path $assets/$file)) {
        Write-Host "Downloading $file"
        Invoke-WebRequest -Uri "https://github.com/51Degrees/device-detection-data/raw/main/$file" -OutFile $assets/$file
    } else {
        Write-Host "'$file' exists, skipping download"
    }

    New-Item -ItemType SymbolicLink -Force -Target "$assets/$file" -Path "$assetsDestination/$file"
}

