param (
    [Parameter(Mandatory)][string]$IpIntelligence
)

$ErrorActionPreference = 'Stop'

$env:IPI_KEY = $IpIntelligence
$env:DATA_FILE = "$PWD/assets/51Degrees-EnterpriseIpiV41.ipi"
$env:EVIDENCE_YAML = "$PWD/assets/ip-intelligence-evidence.yml"

$examples = Get-ChildItem -Directory -Depth 1 -Exclude 'common', 'performance' $PSScriptRoot/../examples

Push-Location "$PSScriptRoot/.."
try {
    $failed = foreach ($example in $examples) {
        Write-Host "Running example $($example.Name)..."
        go run $example || $example.Name
    }
} finally {
    Pop-Location
}

if ($failed) {
    Write-Host "Failed: $failed"
    exit 1
}
