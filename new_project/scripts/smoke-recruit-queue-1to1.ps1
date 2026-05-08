param(
  [string]$BaseUrl = 'http://127.0.0.1:8080',
  [int]$UID = 895,
  [int]$CID = 5035,
  [int]$Position = 114
)

$ErrorActionPreference = 'Stop'
$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$null = New-Item -ItemType Directory -Force -Path 'artifacts' | Out-Null

$script:allPassed = $true
$script:checks = @()

function Add-Check {
  param(
    [string]$Name,
    [bool]$Passed,
    [object]$Detail = $null
  )
  if (-not $Passed) { $script:allPassed = $false }
  $script:checks += [pscustomobject]@{
    name = $Name
    passed = $Passed
    detail = $Detail
  }
}

function Invoke-Api {
  param(
    [string]$Method,
    [string]$Url,
    [object]$Body = $null,
    [Microsoft.PowerShell.Commands.WebRequestSession]$Session = $null
  )
  $args = @{
    Method = $Method
    Uri = $Url
    TimeoutSec = 30
  }
  if ($null -ne $Session) { $args.WebSession = $Session }
  if ($Method -ne 'GET') {
    $args.ContentType = 'application/json'
    $args.Body = if ($null -eq $Body) { '{}' } else { $Body | ConvertTo-Json -Depth 10 }
  }
  try {
    $resp = Invoke-WebRequest -UseBasicParsing @args
    $json = $null
    try { $json = $resp.Content | ConvertFrom-Json } catch { }
    return [pscustomobject]@{
      ok = $true
      status = [int]$resp.StatusCode
      json = $json
      error = $null
    }
  } catch {
    $status = 0
    $body = $null
    if ($_.Exception.Response) {
      $status = [int]$_.Exception.Response.StatusCode
      try {
        $sr = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $raw = $sr.ReadToEnd()
        $sr.Close()
        $body = if ([string]::IsNullOrWhiteSpace($raw)) { $null } else { $raw | ConvertFrom-Json }
      } catch { }
    }
    return [pscustomobject]@{
      ok = $false
      status = $status
      json = $body
      error = $_.Exception.Message
    }
  }
}

function Get-QueueItemById {
  param([object]$Snapshot, [int]$QueueID)
  if ($null -eq $Snapshot -or $null -eq $Snapshot.queue) { return $null }
  return ($Snapshot.queue | Where-Object { [int]$_.id -eq $QueueID } | Select-Object -First 1)
}

Write-Host "=== Recruit Queue 1:1 Smoke ===" -ForegroundColor Cyan
Write-Host "Target city: $CID position: $Position uid: $UID" -ForegroundColor DarkGray

$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$login = Invoke-Api -Method 'POST' -Url "$BaseUrl/api/auth/login" -Body @{ uid = $UID } -Session $session
Add-Check -Name 'login-by-uid' -Passed ($login.ok -and $login.status -eq 200 -and $null -ne $login.json.user) -Detail @{
  status = $login.status
  uid = if ($login.json.user) { [int]$login.json.user.uid } else { 0 }
}
if (-not $script:allPassed) { throw 'login failed' }

$before = Invoke-Api -Method 'GET' -Url "$BaseUrl/api/cities/$CID/barracks?position=$Position" -Session $session
Add-Check -Name 'barracks-snapshot-before' -Passed ($before.ok -and $before.status -eq 200) -Detail @{ status = $before.status }
if (-not $before.ok) { throw 'cannot load barracks snapshot' }

$draftOption = $null
foreach ($opt in $before.json.options) {
  if ($opt.canDraft -eq $true) {
    $draftOption = $opt
    break
  }
}
Add-Check -Name 'draft-option-available' -Passed ($null -ne $draftOption) -Detail @{
  optionCount = if ($before.json.options) { [int]$before.json.options.Count } else { 0 }
}
if ($null -eq $draftOption) { throw 'no draft option available' }

$resourceBefore = if ($before.json.resources) { $before.json.resources } else { $null }

$start = Invoke-Api -Method 'POST' -Url "$BaseUrl/api/cities/$CID/barracks/draft/start" -Body @{
  position = $Position
  sid = [int]$draftOption.sid
  count = 1
} -Session $session
Add-Check -Name 'draft-start-200' -Passed ($start.ok -and $start.status -eq 200) -Detail @{ status = $start.status; sid = [int]$draftOption.sid }
if (-not $start.ok) { throw 'draft start failed' }

$queueNew = $start.json.queue | Sort-Object -Property id -Descending | Select-Object -First 1
Add-Check -Name 'queue-created' -Passed ($null -ne $queueNew -and [int]$start.json.queueCount -gt [int]$before.json.queueCount) -Detail @{
  queueBefore = [int]$before.json.queueCount
  queueAfter = [int]$start.json.queueCount
  queueId = if ($queueNew) { [int]$queueNew.id } else { 0 }
}
if ($null -eq $queueNew) { throw 'queue row missing after start' }

$resourceAfterStart = if ($start.json.resources) { $start.json.resources } else { $null }
$deducted = $true
if ($null -ne $resourceBefore -and $null -ne $resourceAfterStart) {
  $goldBefore = [int64]$resourceBefore.gold
  $goldAfter = [int64]$resourceAfterStart.gold
  $foodBefore = [int64]$resourceBefore.food
  $foodAfter = [int64]$resourceAfterStart.food
  $deducted = ($goldAfter -le $goldBefore -and $foodAfter -le $foodBefore)
}
Add-Check -Name 'resource-deducted' -Passed $deducted -Detail @{
  before = $resourceBefore
  after = $resourceAfterStart
}

Start-Sleep -Seconds 4
$afterSettle = Invoke-Api -Method 'GET' -Url "$BaseUrl/api/cities/$CID/barracks?position=$Position" -Session $session
Add-Check -Name 'barracks-snapshot-after-wait' -Passed ($afterSettle.ok -and $afterSettle.status -eq 200) -Detail @{ status = $afterSettle.status }
$completedQueueMissing = $null -eq (Get-QueueItemById -Snapshot $afterSettle.json -QueueID ([int]$queueNew.id))
Add-Check -Name 'queue-completed' -Passed $completedQueueMissing -Detail @{
  queueId = [int]$queueNew.id
  queueCountAfterWait = [int]$afterSettle.json.queueCount
}

$invalidStart = Invoke-Api -Method 'POST' -Url "$BaseUrl/api/cities/$CID/barracks/draft/start" -Body @{
  position = $Position
  sid = 0
  count = 0
} -Session $session
Add-Check -Name 'invalid-start-rejected' -Passed ($invalidStart.status -eq 400) -Detail @{
  status = $invalidStart.status
  error = if ($invalidStart.json) { $invalidStart.json.error } else { $invalidStart.error }
}

$invalidCancel = Invoke-Api -Method 'POST' -Url "$BaseUrl/api/cities/$CID/barracks/draft/cancel" -Body @{
  position = $Position
  queueId = 0
} -Session $session
Add-Check -Name 'invalid-cancel-rejected' -Passed ($invalidCancel.status -eq 400) -Detail @{
  status = $invalidCancel.status
  error = if ($invalidCancel.json) { $invalidCancel.json.error } else { $invalidCancel.error }
}

$artifact = [pscustomobject]@{
  timestamp = $timestamp
  baseUrl = $BaseUrl
  uid = $UID
  cid = $CID
  position = $Position
  sid = [int]$draftOption.sid
  queueId = [int]$queueNew.id
  allPassed = $script:allPassed
  checks = $script:checks
}
$artifactPath = "artifacts/recruit-queue-$timestamp.json"
$artifact | ConvertTo-Json -Depth 12 | Out-File -FilePath $artifactPath -Encoding UTF8

Write-Host "Artifact: $artifactPath" -ForegroundColor Gray
if (-not $script:allPassed) {
  Write-Host 'Recruit queue smoke failed.' -ForegroundColor Red
  exit 1
}
Write-Host 'Recruit queue smoke passed.' -ForegroundColor Green
