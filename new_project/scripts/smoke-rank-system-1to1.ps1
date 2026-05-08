param(
  [string]$BaseUrl = 'http://127.0.0.1:8080',
  [string]$Passport = 'test',
  [string]$Password = 'test'
)

$ErrorActionPreference = 'Stop'
$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$null = New-Item -ItemType Directory -Force -Path 'artifacts' | Out-Null
$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$allPassed = $true
$steps = @()

$kinds = @(
  'user','union','hero_level','hero_affairs','hero_bravery','hero_wisdom',
  'city_people','city_type','jungong','juanxian','qinwang','gongpin',
  'jungong_union','juanxian_union','qinwang_union','gongpin_union',
  'military','military_attack','military_defence','rich','rich_day','rich_month',
  'battle_total','battle_week','battle_day'
)

function Invoke-Api {
  param([string]$Method, [string]$Url, [object]$Body = $null)
  $args = @{ Method = $Method; Uri = $Url; WebSession = $session; TimeoutSec = 12 }
  if ($Method -ne 'GET') {
    $args.ContentType = 'application/json'
    $args.Body = if ($null -eq $Body) { '{}' } else { $Body | ConvertTo-Json -Depth 10 }
  }
  try {
    $resp = Invoke-WebRequest -UseBasicParsing @args
    $json = $null
    try { $json = $resp.Content | ConvertFrom-Json } catch { }
    return [pscustomobject]@{ ok = $true; status = [int]$resp.StatusCode; json = $json; raw = $resp.Content }
  } catch {
    $status = if ($_.Exception.Response) { [int]$_.Exception.Response.StatusCode } else { 0 }
    return [pscustomobject]@{ ok = $false; status = $status; json = $null; raw = $_.ErrorDetails.Message }
  }
}

function Add-Step {
  param([string]$Name, [bool]$Passed, [hashtable]$Extra = @{})
  if (-not $Passed) { $script:allPassed = $false }
  $row = @{ step = $Name; passed = $Passed }
  foreach ($k in $Extra.Keys) { $row[$k] = $Extra[$k] }
  $script:steps += $row
}

$login = Invoke-Api -Method POST -Url "$BaseUrl/api/auth/login" -Body @{ passport = $Passport; password = $Password }
Add-Step -Name 'login' -Passed ($login.status -eq 200 -and $null -ne $login.json.user) -Extra @{ status = $login.status }
if ($login.status -ne 200) { throw "login failed: $($login.status)" }

foreach ($kind in $kinds) {
  $resp = Invoke-Api -Method GET -Url "$BaseUrl/api/rankings?kind=$kind&page=0"
  $ok = ($resp.status -eq 200 -and $resp.json -and $resp.json.kind -eq $kind -and $null -ne $resp.json.columns)
  Add-Step -Name "ranking_$kind" -Passed $ok -Extra @{
    status = $resp.status
    kind = if ($resp.json) { [string]$resp.json.kind } else { '' }
    total = if ($resp.json) { [int]$resp.json.total } else { -1 }
    columns = if ($resp.json -and $resp.json.columns) { [int]$resp.json.columns.Count } else { -1 }
  }
}

$invalid = Invoke-Api -Method GET -Url "$BaseUrl/api/rankings?kind=user&page=-1"
$invalidSemantics = ($invalid.status -eq 200 -or $invalid.status -eq 400)
Add-Step -Name 'ranking_invalid_page_semantic' -Passed $invalidSemantics -Extra @{ status = $invalid.status }

$output = [pscustomobject]@{
  timestamp = $timestamp
  allPassed = $allPassed
  steps = $steps
}
$path = "artifacts/rank-system-$timestamp.json"
$output | ConvertTo-Json -Depth 8 | Out-File -FilePath $path -Encoding UTF8
Write-Host $path
if ($allPassed) { exit 0 } else { exit 1 }
