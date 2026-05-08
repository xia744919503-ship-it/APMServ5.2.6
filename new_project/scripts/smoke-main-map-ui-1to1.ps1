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

$me = Invoke-Api -Method GET -Url "$BaseUrl/api/auth/me"
$cid = if ($me.json -and $me.json.user) { [int]$me.json.user.defaultCid } else { 0 }
Add-Step -Name 'auth_me' -Passed ($me.status -eq 200 -and $cid -gt 0) -Extra @{ status = $me.status; defaultCid = $cid }

$overview = Invoke-Api -Method GET -Url "$BaseUrl/api/dashboard/overview"
$overviewOk = ($overview.status -eq 200 -and $overview.json -and $overview.json.counts)
Add-Step -Name 'dashboard_overview' -Passed $overviewOk -Extra @{ status = $overview.status }

$city = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$cid"
$cityOk = ($city.status -eq 200 -and $city.json -and $city.json.summary -and $city.json.production -and $city.json.buildings)
Add-Step -Name 'city_detail' -Passed $cityOk -Extra @{ status = $city.status; cityId = $cid }

$map = Invoke-Api -Method GET -Url "$BaseUrl/api/world/map?radius=6"
$mapOk = ($map.status -eq 200 -and $map.json -and $null -ne $map.json.focusX -and $null -ne $map.json.focusY -and $null -ne $map.json.tiles)
Add-Step -Name 'world_map' -Passed $mapOk -Extra @{ status = $map.status; tileCount = if ($map.json -and $map.json.tiles) { [int]$map.json.tiles.Count } else { -1 } }

$cityList = Invoke-Api -Method GET -Url "$BaseUrl/api/me/cities?limit=40"
$cityListOk = ($cityList.status -eq 200 -and $cityList.json -and $null -ne $cityList.json.items)
Add-Step -Name 'city_list' -Passed $cityListOk -Extra @{ status = $cityList.status; count = if ($cityList.json -and $cityList.json.items) { [int]$cityList.json.items.Count } else { -1 } }

$reports = Invoke-Api -Method GET -Url "$BaseUrl/api/reports?filter=all&page=0"
$reportsOk = ($reports.status -eq 200 -and $reports.json -and $null -ne $reports.json.items)
Add-Step -Name 'reports_list' -Passed $reportsOk -Extra @{ status = $reports.status; count = if ($reports.json -and $reports.json.items) { [int]$reports.json.items.Count } else { -1 } }

$mail = Invoke-Api -Method GET -Url "$BaseUrl/api/mail?folder=inbox&page=0"
$mailOk = ($mail.status -eq 200 -and $mail.json -and $null -ne $mail.json.items)
Add-Step -Name 'mail_inbox' -Passed $mailOk -Extra @{ status = $mail.status; count = if ($mail.json -and $mail.json.items) { [int]$mail.json.items.Count } else { -1 } }

$output = [pscustomobject]@{
  timestamp = $timestamp
  allPassed = $allPassed
  steps = $steps
}
$path = "artifacts/main-map-ui-$timestamp.json"
$output | ConvertTo-Json -Depth 8 | Out-File -FilePath $path -Encoding UTF8
Write-Host $path
if ($allPassed) { exit 0 } else { exit 1 }
