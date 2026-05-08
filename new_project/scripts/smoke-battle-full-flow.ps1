param(
  [string]$BaseUrl = 'http://127.0.0.1:8080',
  [string]$Passport = 'test',
  [string]$Password = 'test',
  [int]$CityId = 0
)

$ErrorActionPreference = 'Continue'
$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'

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
    return [pscustomobject]@{ Status = [int]$resp.StatusCode; Json = $json; Success = $true }
  } catch {
    $status = if ($_.Exception.Response) { [int]$_.Exception.Response.StatusCode } else { 0 }
    return [pscustomobject]@{ Status = $status; Json = $null; Success = $false; Error = $_.Exception.Message }
  }
}

Write-Host "=== Battle Flow Full Test ===" -ForegroundColor Cyan
Write-Host "Testing: dispatch -> callback -> report read" -ForegroundColor White

$results = @{ timestamp = $timestamp; steps = @(); checks = @() }
$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession

# Step 1: Login
Write-Host "`n--- Step 1: Login ---" -ForegroundColor Yellow
$login = Invoke-Api -Method POST -Url "$BaseUrl/api/auth/login" -Body @{ passport = $Passport; password = $Password } -Session $session

if (-not $login.Success -or $null -eq $login.Json.user) {
  Write-Host "Login failed!" -ForegroundColor Red
  exit 1
}

$uid = [int]$login.Json.user.uid
$defaultCid = [int]$login.Json.user.defaultCid
if ($CityId -eq 0) { $CityId = $defaultCid }

Write-Host "User: $uid, City: $CityId" -ForegroundColor White
$results.steps += @{ step = "login"; passed = $true }

# Setup: Ensure test city has scouts (sid=3) and valid target
Write-Host "`n--- Setup: Ensuring scouts and valid target ---" -ForegroundColor Yellow

# Ensure scout soldiers (sid=3) exist
$scoutCountRaw = mysql -u root -proot bloodwar -sN -e "SELECT COALESCE(count, 0) FROM sys_city_soldier WHERE cid=$CityId AND sid=3;" 2>$null
$scoutCount = [int]($scoutCountRaw -split "`n" | Select-Object -First 1)
if ($scoutCount -lt 50) {
  mysql -u root -proot bloodwar -e "INSERT INTO sys_city_soldier (cid, sid, count) VALUES ($CityId, 3, 50) ON DUPLICATE KEY UPDATE count=50;" 2>$null
  Write-Host "  Added 50 scouts (sid=3)" -ForegroundColor Gray
} else {
  Write-Host "  City has $scoutCount scouts" -ForegroundColor Gray
}

# Ensure ground building (bid=8) exists
$groundCheck = mysql -u root -proot bloodwar -sN -e "SELECT level FROM sys_building WHERE cid=$CityId AND bid=8 LIMIT 1;" 2>$null
if ([string]::IsNullOrWhiteSpace($groundCheck)) {
  mysql -u root -proot bloodwar -e "INSERT INTO sys_building (cid, bid, level, position, state) VALUES ($CityId, 8, 1, 14, 0) ON DUPLICATE KEY UPDATE bid=bid;" 2>$null
  Write-Host "  Added ground building (bid=8)" -ForegroundColor Gray
}

# Find valid target (non-owned, non-self)
$targetCid = 0
$otherCityRaw = mysql -u root -proot bloodwar -sN -e "SELECT cid FROM sys_city WHERE cid <> $CityId AND uid <> $uid LIMIT 1;" 2>$null
$otherCities = @($otherCityRaw -split "`n" | Where-Object { $_.Trim() -ne '' })
if ($otherCities.Count -gt 0) {
  $targetCid = [int]$otherCities[0]
}
# If no other city, try world field
if ($targetCid -eq 0) {
  $fieldRaw = mysql -u root -proot bloodwar -sN -e "SELECT wid, type FROM mem_world WHERE type > 0 LIMIT 1;" 2>$null
  $fieldLines = @($fieldRaw -split "`n" | Where-Object { $_.Trim() -ne '' })
  if ($fieldLines.Count -gt 0) {
    $parts = $fieldLines[0] -split '\s+'
    if ($parts.Count -ge 2) {
      $wid = [int]$parts[0]
      $y = [Math]::Floor($wid / 1000)
      $x = $wid % 1000
      $targetCid = $y * 1000 + $x
    }
  }
}
if ($targetCid -eq 0) {
  Write-Host "  ERROR: No valid target found!" -ForegroundColor Red
  exit 1
}
Write-Host "  Target CID: $targetCid" -ForegroundColor Green

# Refresh city data
Start-Sleep -Milliseconds 300
$cityCheck = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session

# Get scout count
$scoutsAvailable = 0
if ($cityCheck.Json.soldiers) {
  foreach ($s in $cityCheck.Json.soldiers) {
    if ([int]$s.sid -eq 3) {
      $scoutsAvailable = [int]$s.count
      Write-Host "  Scouts: $scoutsAvailable" -ForegroundColor White
      break
    }
  }
}

# Step 3: Initial troops status
Write-Host "`n--- Step 3: Initial Troops Status ---" -ForegroundColor Yellow
$troopsBefore = Invoke-Api -Method GET -Url "$BaseUrl/api/me/troops" -Session $session
if ($troopsBefore.Success) {
  Write-Host "  Moving: $($troopsBefore.Json.moving), Stationed: $($troopsBefore.Json.stationed), Battling: $($troopsBefore.Json.battling)" -ForegroundColor White
  $results.steps += @{ step = "troops_before"; passed = $true }
} else {
  Write-Check $false "Troops API accessible"
  $results.steps += @{ step = "troops_before"; passed = $false }
}

# Step 4: Scout Dispatch (task=2, sid=3 ONLY)
Write-Host "`n--- Step 4: Scout Dispatch (task=2, sid=3) ---" -ForegroundColor Yellow
$dispatchCount = [Math]::Min(10, $scoutsAvailable)
$dispatch = Invoke-Api -Method POST -Url "$BaseUrl/api/cities/$CityId/troops/dispatch" -Body @{
  targetCid    = $targetCid
  soldierSid   = 3  # MUST be scout for task=2
  soldierCount = $dispatchCount
  task         = 2  # MUST be scout task
} -Session $session

Write-Host "  Dispatch Status: $($dispatch.Status)" -ForegroundColor White
$troopId = 0
if ($dispatch.Json) {
  # Check items array first (dispatch returns TroopPage with items[])
  if ($dispatch.Json.items -and $dispatch.Json.items.Count -gt 0) {
    if ($dispatch.Json.items[0].id) { $troopId = [int]$dispatch.Json.items[0].id }
  }
  # Fallback: direct fields
  if ($troopId -eq 0 -and $dispatch.Json.troopId) { $troopId = [int]$dispatch.Json.troopId }
  if ($troopId -eq 0 -and $dispatch.Json.id) { $troopId = [int]$dispatch.Json.id }
  Write-Host "  Response: $($dispatch.Json | ConvertTo-Json -Depth 3 -Compress)" -ForegroundColor White
}
if ($dispatch.Json.message) {
  Write-Host "  Error: $($dispatch.Json.message)" -ForegroundColor Yellow
}

# 200 = success, 400 = validation error (means endpoint exists)
$dispatchOk = ($dispatch.Status -eq 200)
Write-Check $dispatchOk "Dispatch returns 200 (status=$($dispatch.Status))"
Write-Check ($troopId -gt 0) "Dispatch returns valid troopId: $troopId"

$results.steps += @{ step = "dispatch"; passed = $dispatchOk; troopId = $troopId; status = $dispatch.Status }

# Step 5: Wait for settle then verify DB record
Write-Host "`n--- Step 5: Verify Troop in DB ---" -ForegroundColor Yellow
Write-Host "  Waiting 6 seconds for troop path..." -ForegroundColor Gray
Start-Sleep -Seconds 6

if ($troopId -gt 0) {
  $troopCheck = mysql -u root -proot bloodwar -sN -e "SELECT id, startcid, targetcid, task, state FROM sys_troops WHERE id=$troopId;" 2>$null
  if (-not [string]::IsNullOrWhiteSpace($troopCheck)) {
    Write-Host "  DB Record: $troopCheck" -ForegroundColor Gray
    $troopOk = $troopCheck -match "^\s*\d+\s+$CityId\s+"
    Write-Check $troopOk "Troop startcid=$CityId matches source"
  }
  $results.steps += @{ step = "verify_db"; passed = ($troopCheck -ne "") }
}

# Step 6: Trigger settleDueTroops and check reports
Write-Host "`n--- Step 6: Settle & Reports ---" -ForegroundColor Yellow
$trigger = Invoke-Api -Method GET -Url "$BaseUrl/api/me/troops" -Session $session
Write-Host "  Triggered settleDueTroops" -ForegroundColor Gray

$reports = Invoke-Api -Method GET -Url "$BaseUrl/api/reports?filter=unread&page=0" -Session $session
Write-Check $reports.Success "Reports API accessible (status=$($reports.Status))"

$hasScoutReport = $false
if ($reports.Success -and $reports.Json.items) {
  foreach ($r in $reports.Json.items) {
    # Scout reports: type=0, origin from this city
    if ($r.type -eq 0 -and $r.originCid -eq $CityId) {
      $hasScoutReport = $true
      $headline = $r.headline
      if ([string]::IsNullOrEmpty($headline)) { $headline = "report id=$($r.id)" }
      Write-Host "  Found scout report: $headline" -ForegroundColor Green
      break
    }
  }
}
Write-Check $hasScoutReport "Scout report exists"
$results.steps += @{ step = "reports"; passed = $reports.Success; scout_report = $hasScoutReport }

# Step 7: Verify soldier count decreased
Write-Host "`n--- Step 7: Verify Soldier Count ---" -ForegroundColor Yellow
$cityAfter = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
$scoutsAfter = 0
if ($cityAfter.Json.soldiers) {
  foreach ($s in $cityAfter.Json.soldiers) {
    if ([int]$s.sid -eq 3) {
      $scoutsAfter = [int]$s.count
      break
    }
  }
}
Write-Host "  Scouts before: $scoutsAvailable, after: $scoutsAfter" -ForegroundColor White
$decreased = $scoutsAfter -lt $scoutsAvailable
Write-Check $decreased "Scout count decreased ($scoutsAfter < $scoutsAvailable)"
$results.steps += @{ step = "soldier_change"; passed = $decreased }
if ($reports.Success -and $null -ne $reports.Json.items) {
  $reportCount = [int]$reports.Json.total
  Write-Host "  Total reports: $reportCount" -ForegroundColor White

  $requiredFields = @('id', 'type', 'read', 'title', 'createdAt')
  $allFieldsPresent = $true
  if ($reports.Json.items.Count -gt 0) {
    foreach ($field in $requiredFields) {
      $hasField = $null -ne $reports.Json.items[0].$field
      Write-Check $hasField "Report has '$field' field"
      if (-not $hasField) { $allFieldsPresent = $false }
    }
  }
  $results.steps += @{ step = "report_structure"; passed = $allFieldsPresent; count = $reportCount }
} else {
  Write-Check $false "Report structure verified (requires reports in test data)"
  $results.steps += @{ step = "report_structure"; passed = $false; reason = "no reports" }
}

# Output
$outputFile = "artifacts\battle-flow-test-$timestamp.json"
$results | ConvertTo-Json -Depth 10 | Out-File -FilePath $outputFile -Encoding UTF8

Write-Host "`n=== Summary ===" -ForegroundColor Cyan
$passedSteps = ($results.steps | Where-Object { $_.passed }).Count
$totalSteps = $results.steps.Count
Write-Host "Steps: $passedSteps/$totalSteps passed" -ForegroundColor White

if ($script:allPassed) {
  Write-Host "`nAll checks passed!" -ForegroundColor Green
  exit 0
} else {
  Write-Host "`nSome checks failed!" -ForegroundColor Red
  exit 1
}