param(
  [string]$SpecDir = "proto/specs",
  [string]$OutDir  = "proto"
)

if (-not (Test-Path $SpecDir)) {
  Write-Error "Spec dir not found: $SpecDir"
  exit 1
}

Push-Location $SpecDir
protoc --proto_path=. --go_out=paths=source_relative:../ --go-grpc_out=paths=source_relative:../ *.proto
Pop-Location

Write-Host "Protos generated to $OutDir"
