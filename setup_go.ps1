$ErrorActionPreference = 'Stop'

Write-Host "Downloading Go 1.26.3..."
$url = "https://dl.google.com/go/go1.26.3.windows-amd64.zip"
$output = "go.zip"

# Download using WebClient for speed and reliability
$clnt = New-Object System.Net.WebClient
$clnt.DownloadFile($url, $output)

Write-Host "Extracting Go using tar..."
tar -xf $output

Write-Host "Cleaning up zip file..."
Remove-Item $output

Write-Host "Go compiler installed successfully!"
& .\go\bin\go.exe version
