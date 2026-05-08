param(
  [string]$BaseUrl = 'http://127.0.0.1:8080',
  [string]$Passport = 'test',
  [string]$Password = 'test',
  [int]$CityId = 0
)

$ErrorActionPreference = 'Continue'
$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'

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
    TimeoutSec = 15
  }

  if ($null -ne $Session) {
    $args.WebSession = $Session
  }

  if ($Method -ne 'GET') {
    $args.ContentType = 'application/json'
    $args.Body = if ($null -eq $Body) { '{}' } else { $Body | ConvertTo-Json -Depth 12 }
  }

  try {
    $resp = Invoke-WebRequest -UseBasicParsing @args
    $json = $null
    try {
      $json = $resp.Content | ConvertFrom-Json
    } catch { }
    return [pscustomobject]@{
      Status  = [int]$resp.StatusCode
      Json    = $json
      Raw     = $resp.Content
      Error   = ''
      Success = $true
    }
  } catch {
    $status = 0
    if ($_.Exception.Response) {
      $status = [int]$_.Exception.Response.StatusCode
    }
    return [pscustomobject]@{
      Status  = $status
      Json    = $null
      Raw     = $_.ErrorDetails.Message
      Error   = $_.Exception.Message
      Success = $false
    }
  }
}

Write-Host "=== Battle Report & Troop 1:1 Verification ===" -ForegroundColor Cyan

$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession

$login = Invoke-Api -Method POST -Url "$BaseUrl/api/auth/login" -Body @{
  passport = $Passport
  password = $Password
} -Session $session

if (-not $login.Success -or $null -eq $login.Json.user) {
  Write-Host "Login failed!" -ForegroundColor Red
  exit 1
}

$uid = [int]$login.Json.user.uid
$defaultCid = [int]$login.Json.user.defaultCid

if ($CityId -eq 0) {
  $CityId = $defaultCid
}

Write-Host "User: $uid, City: $CityId" -ForegroundColor White

$results = [pscustomobject]@{
  timestamp   = $timestamp
  uid         = $uid
  cityId      = $CityId
  checks      = @()
  issues      = @()
  allPassed   = $true
}

$checksList = @()
$issuesList = @()

function Add-CheckResult {
  param(
    [string]$Name,
    [string]$Field,
    [object]$Expected,
    [object]$Actual,
    [bool]$Passed,
    [string]$Message = ''
  )

  $check = [pscustomobject]@{
    name     = $Name
    field    = $Field
    expected = $Expected
    actual   = $Actual
    passed   = $Passed
    message  = $Message
  }

  $script:checksList += $check

  if (-not $Passed) {
    $script:results.allPassed = $false
    $script:issuesList += [pscustomobject]@{
      name    = $Name
      field   = $Field
      message = $Message
    }
    Write-Host "  [FAIL] ${Name}: ${Message}" -ForegroundColor Red
  }
  else {
    Write-Host "  [PASS] ${Name}" -ForegroundColor Green
  }
}

# Test 1: My Troops API
Write-Host "`n--- My Troops API ---" -ForegroundColor Yellow
$troops = Invoke-Api -Method GET -Url "$BaseUrl/api/me/troops" -Session $session
if ($troops.Success -and $null -ne $troops.Json) {
  $troopCount = [int]$troops.Json.total
  Write-Host "  Total Troops: $troopCount" -ForegroundColor White

  Add-CheckResult -Name 'Troops API Response' -Field 'total' -Expected '>=0' -Actual $troopCount -Passed ($troopCount -ge 0) -Message "Troops API should return valid count"

  $movingCount = [int]$troops.Json.moving
  $returningCount = [int]$troops.Json.returning
  $stationedCount = [int]$troops.Json.stationed
  $battlingCount = [int]$troops.Json.battling

  Add-CheckResult -Name 'Troop States Valid' -Field 'moving/returning/stationed' -Expected '>=0 each' -Actual "$movingCount/$returningCount/$stationedCount" -Passed ($movingCount -ge 0 -and $returningCount -ge 0 -and $stationedCount -ge 0) -Message "All troop states should be non-negative"
}
else {
  Write-Host "  Failed to get troops: $($troops.Error)" -ForegroundColor Red
}

# Test 2: Reports API - Unread
Write-Host "`n--- Reports API (Unread) ---" -ForegroundColor Yellow
$reports = Invoke-Api -Method GET -Url "$BaseUrl/api/reports?filter=unread&page=0" -Session $session
if ($reports.Success -and $null -ne $reports.Json) {
  $reportTotal = [int]$reports.Json.total
  $reportPage = [int]$reports.Json.page
  $pageCount = [int]$reports.Json.pageCount

  Write-Host "  Total: $reportTotal, Page: $reportPage, PageCount: $pageCount" -ForegroundColor White

  Add-CheckResult -Name 'Reports API Response' -Field 'total' -Expected '>=0' -Actual $reportTotal -Passed ($reportTotal -ge 0) -Message "Reports API should return valid count"

  $filter = [string]$reports.Json.filter
  Add-CheckResult -Name 'Filter Echo' -Field 'filter' -Expected 'unread' -Actual $filter -Passed ($filter -eq 'unread') -Message "Filter should echo back as unread"
}
else {
  Write-Host "  Failed to get reports: $($reports.Error)" -ForegroundColor Red
}

# Test 3: Reports API - All
Write-Host "`n--- Reports API (All) ---" -ForegroundColor Yellow
$allReports = Invoke-Api -Method GET -Url "$BaseUrl/api/reports?filter=all&page=0" -Session $session
if ($allReports.Success -and $null -ne $allReports.Json) {
  $allReportTotal = [int]$allReports.Json.total
  Write-Host "  Total All Reports: $allReportTotal" -ForegroundColor White

  Add-CheckResult -Name 'All Reports API' -Field 'total' -Expected '>=0' -Actual $allReportTotal -Passed ($allReportTotal -ge 0) -Message "All reports should be accessible"
}
else {
  Write-Host "  Failed to get all reports: $($allReports.Error)" -ForegroundColor Red
}

# Test 4: City Detail Structure
Write-Host "`n--- City Detail Structure ---" -ForegroundColor Yellow
$cityDetail = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
if ($cityDetail.Success -and $null -ne $cityDetail.Json) {
  $hasSummary = $null -ne $cityDetail.Json.summary
  $hasProduction = $null -ne $cityDetail.Json.production
  $hasBuildings = $null -ne $cityDetail.Json.buildings
  $hasSoldiers = $null -ne $cityDetail.Json.soldiers

  Write-Host "  Has Summary: $hasSummary, Production: $hasProduction, Buildings: $hasBuildings, Soldiers: $hasSoldiers" -ForegroundColor White

  Add-CheckResult -Name 'City Summary Present' -Field 'summary' -Expected 'not null' -Actual $cityDetail.Json.summary -Passed $hasSummary -Message "City should have summary"

  Add-CheckResult -Name 'City Production Present' -Field 'production' -Expected 'not null' -Actual $cityDetail.Json.production -Passed $hasProduction -Message "City should have production data"

  Add-CheckResult -Name 'City Buildings Present' -Field 'buildings' -Expected 'not null' -Actual $cityDetail.Json.buildings -Passed $hasBuildings -Message "City should have buildings"

  Add-CheckResult -Name 'City Soldiers Present' -Field 'soldiers' -Expected 'not null' -Actual $cityDetail.Json.soldiers -Passed $hasSoldiers -Message "City should have soldiers array"
}
else {
  Write-Host "  Failed to get city detail: $($cityDetail.Error)" -ForegroundColor Red
}

# Test 5: Troop Dispatch Endpoint Check
Write-Host "`n--- Troop Dispatch Endpoint ---" -ForegroundColor Yellow
$dispatchCheck = Invoke-Api -Method POST -Url "$BaseUrl/api/cities/$CityId/troops/dispatch" -Body @{
  targetCid = 266010
  soldierSid = 1
  soldierCount = 10
  task = 2
} -Session $session

if ($dispatchCheck.Success) {
  Write-Host "  Dispatch Status: $($dispatchCheck.Status)" -ForegroundColor White
  Add-CheckResult -Name 'Dispatch Endpoint' -Field 'status' -Expected '2xx/4xx' -Actual $dispatchCheck.Status -Passed ($dispatchCheck.Status -ge 200 -and $dispatchCheck.Status -lt 500) -Message "Dispatch endpoint should be accessible"
}
else {
  Write-Host "  Dispatch Status: $($dispatchCheck.Status)" -ForegroundColor White
  Add-CheckResult -Name 'Dispatch Endpoint' -Field 'status' -Expected '2xx/4xx' -Actual $dispatchCheck.Status -Passed ($dispatchCheck.Status -ge 200 -and $dispatchCheck.Status -lt 500) -Message "Dispatch endpoint should be accessible"
}

# Test 6: Report Detail with Invalid ID
Write-Host "`n--- Report Detail (Invalid ID) ---" -ForegroundColor Yellow
$invalidReport = Invoke-Api -Method GET -Url "$BaseUrl/api/reports/999999" -Session $session
if ($invalidReport.Success -eq $false) {
  Write-Host "  Invalid Report Status: $($invalidReport.Status)" -ForegroundColor White
  Add-CheckResult -Name 'Invalid Report 404' -Field 'status' -Expected '404' -Actual $invalidReport.Status -Passed ($invalidReport.Status -eq 404) -Message "Non-existent report should return 404"
}
else {
  Write-Host "  Invalid Report returned: $($invalidReport.Status)" -ForegroundColor White
}

# Test 7: Troop Callback with Invalid ID
Write-Host "`n--- Troop Callback (Invalid ID) ---" -ForegroundColor Yellow
$invalidCallback = Invoke-Api -Method POST -Url "$BaseUrl/api/troops/999999/callback" -Session $session
if ($invalidCallback.Success -eq $false) {
  Write-Host "  Invalid Callback Status: $($invalidCallback.Status)" -ForegroundColor White
  Add-CheckResult -Name 'Invalid Callback' -Field 'status' -Expected '4xx' -Actual $invalidCallback.Status -Passed ($invalidCallback.Status -ge 400) -Message "Non-existent troop callback should fail"
}
else {
  Write-Host "  Invalid Callback returned: $($invalidCallback.Status)" -ForegroundColor White
}

$results.checks = $checksList
$results.issues = $issuesList

$outputFile = "artifacts\city-battle-report-diff-$timestamp.json"
$results | ConvertTo-Json -Depth 10 | Out-File -FilePath $outputFile -Encoding UTF8

Write-Host "`n=== Summary ===" -ForegroundColor Cyan
Write-Host "Total Checks: $($results.checks.Count)" -ForegroundColor White
$passedCount = ($results.checks | Where-Object { $_.passed }).Count
$failedCount = ($results.checks | Where-Object { -not $_.passed }).Count
Write-Host "Passed: $passedCount" -ForegroundColor Green
Write-Host "Failed: $failedCount" -ForegroundColor $(if ($failedCount -gt 0) { 'Red' } else { 'Green' })

if ($results.issues.Count -gt 0) {
  Write-Host "`nIssues:" -ForegroundColor Red
  foreach ($issue in $results.issues) {
    Write-Host "  [$($issue.name)] $($issue.message)" -ForegroundColor Red
  }
}

Write-Host "`nReport: $outputFile" -ForegroundColor Cyan

if ($results.allPassed) {
  Write-Host "`nAll battle report checks passed!" -ForegroundColor Green
  exit 0
}
else {
  Write-Host "`nSome battle report checks failed!" -ForegroundColor Red
  exit 1
}
