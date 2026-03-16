param (
    [string]$DeviceDetection,
    [string]$DeviceDetectionUrl,
    [string]$IpIntelligenceUrl
)
$ErrorActionPreference = "Stop"

./steps/fetch-assets.ps1 -DeviceDetection:$DeviceDetection -DeviceDetectionUrl:$DeviceDetectionUrl -IpIntelligenceUrl:$IpIntelligenceUrl `
    -Assets "51Degrees-EnterpriseIpiV41.ipi", "ip-intelligence-evidence.yml", "TAC-HashV41.hash"
