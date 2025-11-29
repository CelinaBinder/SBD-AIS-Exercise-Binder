Write-Host "Generating Go protobuf files..."

Get-ChildItem *.proto | ForEach-Object {
    Write-Host ("Processing: " + $_.Name)
    protoc --go_out=. --go-grpc_out=. $_.Name
}

Write-Host "Done generating protobuf files."
