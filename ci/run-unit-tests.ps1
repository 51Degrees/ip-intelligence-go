param (
    [Parameter(Mandatory)][string]$RepoName
)

$env:DATA_FILE = "$PWD/assets/51Degrees-EnterpriseIpiV41.ipi"
./go/run-unit-tests.ps1 @PSBoundParameters
