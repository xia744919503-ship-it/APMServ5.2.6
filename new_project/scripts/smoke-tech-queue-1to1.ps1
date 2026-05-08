param(
  [string]$BaseUrl = 'http://127.0.0.1:8080',
  [string]$Passport = '',
  [string]$Password = '',
  [int]$UID = 687,
  [int]$CityId = 0,
  [int]$CollegePosition = 0
)

$ErrorActionPreference = 'Stop'
$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$null = New-Item -ItemType Directory -Force -Path 'artifacts' | Out-Null
$script:allPassed = $true

function Write-Check {
  param([bool]$Passed, [string]$Message)
  if ($Passed) {
    Write-Host "  [PASS] $Message" -ForegroundColor Green
  } else {
    Write-Host "  [FAIL] $Message" -ForegroundColor Red
    $script:allPassed = $false
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
    Method     = $Method
    Uri        = $Url
    TimeoutSec = 30
  }
  if ($null -ne $Session) { $args.WebSession = $Session }
  if ($Method -ne 'GET') {
    $args.ContentType = 'application/json'
    $args.Body = if ($null -eq $Body) { '{}' } else { $Body | ConvertTo-Json -Depth 12 }
  }
  try {
    $resp = Invoke-WebRequest -UseBasicParsing @args
    $json = $null
    try { $json = $resp.Content | ConvertFrom-Json } catch { }
    return [pscustomobject]@{ Status = [int]$resp.StatusCode; Json = $json; Success = $true; Raw = $resp.Content }
  } catch {
    $status = if ($_.Exception.Response) { [int]$_.Exception.Response.StatusCode } else { 0 }
    return [pscustomobject]@{ Status = $status; Json = $null; Success = $false; Error = $_.Exception.Message }
  }
}

function Find-ResearchOptionByTid {
  param($Options, [int]$Tid)
  foreach ($item in $Options) {
    if ([int]$item.tid -eq $Tid) { return $item }
  }
  return $null
}

function Find-ShortestResearchCandidate {
  param(
    [Microsoft.PowerShell.Commands.WebRequestSession]$Session,
    [int]$PreferredCityId = 0,
    [int]$PreferredPosition = 0
  )
  $cityIds = @()
  if ($PreferredCityId -gt 0) {
    $cityIds += $PreferredCityId
  } else {
    $myCities = Invoke-Api -Method GET -Url "$BaseUrl/api/me/cities?limit=80" -Session $Session
    if (-not $myCities.Success) { throw "list_my_cities_failed: $($myCities.Error)" }
    foreach ($c in $myCities.Json.items) { $cityIds += [int]$c.cid }
  }

  $best = $null
  foreach ($cid in $cityIds) {
    $city = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$cid" -Session $Session
    if (-not $city.Success) { continue }

    $position = $PreferredPosition
    if ($position -le 0) {
      foreach ($b in $city.Json.buildings) {
        if ([int]$b.bid -eq 7) {
          $position = [int]$b.position
          break
        }
      }
    }
    if ($position -le 0) { continue }

    $research = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$cid/research?position=$position" -Session $Session
    if (-not $research.Success) { continue }

    foreach ($opt in $research.Json.options) {
      if ($opt.canUpgrade -ne $true -or [int]$opt.state -ne 0) { continue }
      $item = [ordered]@{
        cid = $cid
        position = $position
        tid = [int]$opt.tid
        duration = [int64]$opt.upgradeDuration
        option = $opt
      }
      if ($null -eq $best -or [int64]$item.duration -lt [int64]$best.duration) {
        $best = $item
      }
    }
  }
  return $best
}

Write-Host "=== Tech Queue 1:1 Smoke ===" -ForegroundColor Cyan
$results = [ordered]@{
  timestamp = $timestamp
  baseUrl = $BaseUrl
  checks = @()
  allPassed = $true
}

$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession

# 1) login + city baseline
$loginBody = @{}
if ($UID -gt 0) {
  $loginBody.uid = $UID
} else {
  $loginBody.passport = $Passport
  $loginBody.password = $Password
}
$login = Invoke-Api -Method POST -Url "$BaseUrl/api/auth/login" -Body $loginBody -Session $session
if (-not $login.Success -or $null -eq $login.Json.user) {
  throw "login_failed: $($login.Error)"
}
$uid = [int]$login.Json.user.uid
$selected = Find-ShortestResearchCandidate -Session $session -PreferredCityId $CityId -PreferredPosition $CollegePosition
if ($null -eq $selected) { throw "no_research_candidate_can_upgrade" }
$CityId = [int]$selected.cid
$CollegePosition = [int]$selected.position

$cityBefore = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
if (-not $cityBefore.Success) { throw "city_before_failed: $($cityBefore.Error)" }
$snapshotBefore = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId/research?position=$CollegePosition" -Session $session
if (-not $snapshotBefore.Success) { throw "research_snapshot_before_failed: $($snapshotBefore.Error)" }

$targetTid = [int]$selected.tid
$targetDuration = [int64]$selected.duration
$targetCost = [ordered]@{
  wood = [int64]$selected.option.woodNeed
  rock = [int64]$selected.option.rockNeed
  iron = [int64]$selected.option.ironNeed
  food = [int64]$selected.option.foodNeed
  gold = [int64]$selected.option.goldNeed
}
Write-Host "Target tid=$targetTid duration=$targetDuration s" -ForegroundColor White

$baselineOption = Find-ResearchOptionByTid -Options $snapshotBefore.Json.options -Tid $targetTid
$baselineLevel = [int64]$baselineOption.level

# 2) precondition: invalid tid should be rejected
$invalidStart = Invoke-Api -Method POST -Url "$BaseUrl/api/cities/$CityId/research/start" -Body @{
  position = $CollegePosition
  tid = -1
} -Session $session
$invalidRejected = ($invalidStart.Status -eq 400)
Write-Check $invalidRejected "Precondition invalid tid rejected (400)"
$results.checks += [ordered]@{ name = "precondition_invalid_tid"; passed = $invalidRejected; status = $invalidStart.Status }

# 3) start research -> queue created + resources deducted
$start = Invoke-Api -Method POST -Url "$BaseUrl/api/cities/$CityId/research/start" -Body @{
  position = $CollegePosition
  tid = $targetTid
} -Session $session
if (-not $start.Success -or $start.Status -ne 200) {
  throw "research_start_failed: status=$($start.Status) err=$($start.Error)"
}
$startedOption = Find-ResearchOptionByTid -Options $start.Json.options -Tid $targetTid
$queueCreated = ([int]$start.Json.activeTid -eq $targetTid) -and ($null -ne $startedOption) -and ([int]$startedOption.state -eq 1)
Write-Check $queueCreated "Queue created (activeTid/state)"
$results.checks += [ordered]@{ name = "queue_created"; passed = $queueCreated; activeTid = [int]$start.Json.activeTid; optionState = if ($null -eq $startedOption) { -1 } else { [int]$startedOption.state } }

$cityAfterStart = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
if (-not $cityAfterStart.Success) { throw "city_after_start_failed: $($cityAfterStart.Error)" }
$deductWood = ([int64]$cityAfterStart.Json.summary.resources.wood -eq ([int64]$cityBefore.Json.summary.resources.wood - [int64]$targetCost.wood))
$deductRock = ([int64]$cityAfterStart.Json.summary.resources.rock -eq ([int64]$cityBefore.Json.summary.resources.rock - [int64]$targetCost.rock))
$deductIron = ([int64]$cityAfterStart.Json.summary.resources.iron -eq ([int64]$cityBefore.Json.summary.resources.iron - [int64]$targetCost.iron))
$deductFood = ([int64]$cityAfterStart.Json.summary.resources.food -eq ([int64]$cityBefore.Json.summary.resources.food - [int64]$targetCost.food))
$deductGold = ([int64]$cityAfterStart.Json.summary.resources.gold -eq ([int64]$cityBefore.Json.summary.resources.gold - [int64]$targetCost.gold))
$deductAll = $deductWood -and $deductRock -and $deductIron -and $deductFood -and $deductGold
Write-Check $deductAll "Resources deducted exactly once on start"
$results.checks += [ordered]@{
  name = "resource_deduction"
  passed = $deductAll
  before = $cityBefore.Json.summary.resources
  afterStart = $cityAfterStart.Json.summary.resources
  expectedCost = $targetCost
}

# 4) precondition: busy queue blocks 2nd start
$secondCandidate = $null
foreach ($opt in $start.Json.options) {
  if ([int]$opt.tid -ne $targetTid) {
    $secondCandidate = $opt
    break
  }
}
if ($null -eq $secondCandidate) { $secondCandidate = $startedOption }
$busyStart = Invoke-Api -Method POST -Url "$BaseUrl/api/cities/$CityId/research/start" -Body @{
  position = $CollegePosition
  tid = [int]$secondCandidate.tid
} -Session $session
$busyRejected = ($busyStart.Status -eq 400)
Write-Check $busyRejected "Precondition busy queue rejects second start"
$results.checks += [ordered]@{ name = "precondition_busy_queue"; passed = $busyRejected; status = $busyStart.Status }

# 5) settle path: short duration -> wait complete; long duration -> cancel path
if ([int64]$targetDuration -le 120) {
  $waitSec = [Math]::Max(2, [int]$targetDuration + 2)
  Write-Host "Waiting $waitSec s for settlement..." -ForegroundColor Gray
  Start-Sleep -Seconds $waitSec

  $snapshotAfter = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId/research?position=$CollegePosition" -Session $session
  if (-not $snapshotAfter.Success) { throw "research_snapshot_after_failed: $($snapshotAfter.Error)" }
  $afterOption = Find-ResearchOptionByTid -Options $snapshotAfter.Json.options -Tid $targetTid
  $settled = ([int]$snapshotAfter.Json.activeTid -eq 0) -and ($null -ne $afterOption) -and ([int]$afterOption.state -eq 0) -and ([int64]$afterOption.level -ge ($baselineLevel + 1))
  Write-Check $settled "Completion settlement applied (idle + level advanced)"
  $results.checks += [ordered]@{
    name = "completion_settlement"
    passed = $settled
    baselineLevel = $baselineLevel
    afterLevel = if ($null -eq $afterOption) { -1 } else { [int64]$afterOption.level }
    activeTidAfter = [int]$snapshotAfter.Json.activeTid
  }

  $cityAfterSettle = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
  if (-not $cityAfterSettle.Success) { throw "city_after_settle_failed: $($cityAfterSettle.Error)" }
  $noExtraDeduct = (
    [int64]$cityAfterSettle.Json.summary.resources.wood -eq [int64]$cityAfterStart.Json.summary.resources.wood -and
    [int64]$cityAfterSettle.Json.summary.resources.rock -eq [int64]$cityAfterStart.Json.summary.resources.rock -and
    [int64]$cityAfterSettle.Json.summary.resources.iron -eq [int64]$cityAfterStart.Json.summary.resources.iron -and
    [int64]$cityAfterSettle.Json.summary.resources.food -eq [int64]$cityAfterStart.Json.summary.resources.food -and
    [int64]$cityAfterSettle.Json.summary.resources.gold -eq [int64]$cityAfterStart.Json.summary.resources.gold
  )
  Write-Check $noExtraDeduct "Settlement does not re-deduct resources"
  $results.checks += [ordered]@{ name = "settlement_resource_stability"; passed = $noExtraDeduct }
} else {
  Write-Host "Duration=$targetDuration s is long; using cancel-closure path." -ForegroundColor Yellow
  $cancel = Invoke-Api -Method POST -Url "$BaseUrl/api/cities/$CityId/research/cancel" -Body @{
    position = $CollegePosition
    tid = $targetTid
  } -Session $session
  $cancelOk = ($cancel.Status -eq 200)
  Write-Check $cancelOk "Cancel queued research returns 200"

  $snapshotAfterCancel = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId/research?position=$CollegePosition" -Session $session
  if (-not $snapshotAfterCancel.Success) { throw "research_snapshot_after_cancel_failed: $($snapshotAfterCancel.Error)" }
  $afterCancelOption = Find-ResearchOptionByTid -Options $snapshotAfterCancel.Json.options -Tid $targetTid
  $cancelSettled = ([int]$snapshotAfterCancel.Json.activeTid -eq 0) -and ($null -ne $afterCancelOption) -and ([int]$afterCancelOption.state -eq 0)
  Write-Check $cancelSettled "Cancel settlement applied (idle + state reset)"
  $results.checks += [ordered]@{
    name = "completion_settlement"
    passed = ($cancelOk -and $cancelSettled)
    baselineLevel = $baselineLevel
    afterLevel = if ($null -eq $afterCancelOption) { -1 } else { [int64]$afterCancelOption.level }
    activeTidAfter = [int]$snapshotAfterCancel.Json.activeTid
    mode = "cancel"
  }

  $cityAfterCancel = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
  if (-not $cityAfterCancel.Success) { throw "city_after_cancel_failed: $($cityAfterCancel.Error)" }
  # Legacy semantics differ by branch: some servers keep deduction after cancel, some refund.
  # We only require no *additional* deduction after cancel settlement.
  $resourceStable = (
    [int64]$cityAfterCancel.Json.summary.resources.wood -ge [int64]$cityAfterStart.Json.summary.resources.wood -and
    [int64]$cityAfterCancel.Json.summary.resources.rock -ge [int64]$cityAfterStart.Json.summary.resources.rock -and
    [int64]$cityAfterCancel.Json.summary.resources.iron -ge [int64]$cityAfterStart.Json.summary.resources.iron -and
    [int64]$cityAfterCancel.Json.summary.resources.food -ge [int64]$cityAfterStart.Json.summary.resources.food -and
    [int64]$cityAfterCancel.Json.summary.resources.gold -ge [int64]$cityAfterStart.Json.summary.resources.gold
  )
  Write-Check $resourceStable "Cancel settlement keeps or refunds resources (no extra deduction)"
  $results.checks += [ordered]@{ name = "settlement_resource_stability"; passed = $resourceStable; mode = "cancel" }
}

$results.allPassed = $script:allPassed
$output = "artifacts/tech-queue-$timestamp.json"
$results | ConvertTo-Json -Depth 12 | Out-File -FilePath $output -Encoding utf8

Write-Host ""
Write-Host "Artifact: $output" -ForegroundColor Cyan
if ($script:allPassed) { exit 0 } else { exit 1 }
