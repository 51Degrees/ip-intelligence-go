param (
    [Parameter(Mandatory)][string]$IpIntelligenceUrl
)
$ErrorActionPreference = "Stop"

./steps/fetch-assets.ps1 -IpIntelligenceUrl:$IpIntelligenceUrl -Assets "51Degrees-EnterpriseIpiV41.ipi"
