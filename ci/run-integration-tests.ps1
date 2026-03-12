param (
    [Parameter(Mandatory)][string]$IpIntelligence,
    [Parameter(Mandatory)][string]$IpIntelligenceUrl
)
$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $true

$results = New-Item -ItemType directory -Force -Path "$PSScriptRoot/../test-results/integration"

$env:IPI_KEY = $IpIntelligence
$env:IPI_DATA_FILE_URL = $IpIntelligenceUrl
$env:DATA_FILE = "$PWD/assets/51Degrees-EnterpriseIpiV41.ipi"
$env:EVIDENCE_YAML = "$PWD/assets/ip-intelligence-evidence.yml"

Push-Location "$PSScriptRoot/.."
try {
    go test -v ./examples 2>&1 | go-junit-report -set-exit-code -iocopy -out "$results/results.xml"
} finally {
    Pop-Location
}
