param(
  [switch]$Apply,
  [switch]$KeepUploads,
  [switch]$KeepDatabase,
  [switch]$KeepSessions
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$backendDir = Join-Path $repoRoot "backend"
$envPath = Join-Path $backendDir ".env"

function Read-DotEnvValue($path, $key, $fallback) {
  if (-not (Test-Path -LiteralPath $path)) {
    return $fallback
  }

  $line = Get-Content -LiteralPath $path | Where-Object { $_ -match "^\s*$([regex]::Escape($key))=" } | Select-Object -First 1
  if (-not $line) {
    return $fallback
  }

  return ($line -replace "^\s*$([regex]::Escape($key))=", "").Trim()
}

function Set-DotEnvValue($path, $key, $value) {
  $lines = @()
  if (Test-Path -LiteralPath $path) {
    $lines = @(Get-Content -LiteralPath $path)
  }

  $found = $false
  $next = foreach ($line in $lines) {
    if ($line -match "^\s*$([regex]::Escape($key))=") {
      $found = $true
      "$key=$value"
    } else {
      $line
    }
  }

  if (-not $found) {
    $next += "$key=$value"
  }

  Set-Content -LiteralPath $path -Value $next -Encoding UTF8
}

$dbPathRaw = Read-DotEnvValue $envPath "DB_PATH" "./videoplatform.db"
$uploadPathRaw = Read-DotEnvValue $envPath "UPLOAD_PATH" "./uploads"

$dbPath = if ([System.IO.Path]::IsPathRooted($dbPathRaw)) {
  $dbPathRaw
} else {
  Join-Path $backendDir $dbPathRaw
}

$uploadPath = if ([System.IO.Path]::IsPathRooted($uploadPathRaw)) {
  $uploadPathRaw
} else {
  Join-Path $backendDir $uploadPathRaw
}

Write-Host "Reset dev target:"
Write-Host "  Database : $dbPath"
Write-Host "  Uploads  : $uploadPath"
Write-Host "  Sessions : $(if ($KeepSessions) { 'keep JWT_SECRET' } else { 'rotate JWT_SECRET' })"

if (-not $Apply) {
  Write-Host ""
  Write-Host "Dry run only. Re-run with -Apply to execute."
  Write-Host "Examples:"
  Write-Host "  .\scripts\reset-dev.ps1 -Apply"
  Write-Host "  .\scripts\reset-dev.ps1 -Apply -KeepUploads"
  Write-Host "  .\scripts\reset-dev.ps1 -Apply -KeepDatabase"
  exit 0
}

if (-not $KeepDatabase -and (Test-Path -LiteralPath $dbPath)) {
  Remove-Item -LiteralPath $dbPath -Force
  Write-Host "Deleted database."
}

if (-not $KeepUploads) {
  New-Item -ItemType Directory -Force -Path $uploadPath | Out-Null
  Get-ChildItem -LiteralPath $uploadPath -Force | Remove-Item -Recurse -Force
  Write-Host "Cleared uploads."
}

if (-not $KeepSessions) {
  $secretBytes = New-Object byte[] 32
  [System.Security.Cryptography.RandomNumberGenerator]::Fill($secretBytes)
  $newSecret = [Convert]::ToBase64String($secretBytes)
  Set-DotEnvValue $envPath "JWT_SECRET" $newSecret
  Write-Host "Rotated JWT_SECRET. Existing admin cookies are now invalid."
}

Write-Host ""
Write-Host "Done. Restart the backend so the database can be recreated and the new JWT_SECRET is loaded."
