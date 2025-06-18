param (
    [Parameter(Mandatory)][string]$RepoName,
    [Parameter(Mandatory)][string]$OrgName,
    [Parameter(Mandatory)][string]$Name,
    [string]$Branch = "main"
)

$env:DATA_FILE = "$PWD/assets/51Degrees-EnterpriseIpiV41.ipi"
./go/run-performance-tests.ps1 @PSBoundParameters -ExamplesRepo ip-intelligence-go-examples -Example ./ipi_onpremise/performance/performance.go
