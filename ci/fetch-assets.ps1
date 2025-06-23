param (
    [Parameter(Mandatory)][string]$DeviceDetectionUrl
)
$ErrorActionPreference = "Stop"

$assets = New-Item -ItemType Directory -Path assets -Force
$dataFile = "$assets/51Degrees-EnterpriseIpiV41.ipi"

if (Test-Path $assets/$dataFile) {
    Write-Host "Data file exists, skipping download"
} else {
    Write-Host "Downloading data file ($dataFile.gz)..."
    ./steps/download-data-file.ps1 -FullFilePath "$dataFile.gz" -Url $DeviceDetectionUrl
    ./steps/gunzip-file.ps1 "$dataFile.gz"
}
