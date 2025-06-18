param (
    [Parameter(Mandatory)][string]$RepoName,
    [Parameter(Mandatory)][string]$OrgName,
    [Parameter(Mandatory)][string]$Name,
    [string]$Branch = "main"
)

./go/run-performance-tests.ps1 @PSBoundParameters -ExamplesRepo ip-intelligence-go-examples -Example ./ipi_onpremise/performance/performance.go
