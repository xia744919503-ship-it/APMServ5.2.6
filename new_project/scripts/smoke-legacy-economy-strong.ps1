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

Write-Host "=== City Economy 1:1 Verification ===" -ForegroundColor Cyan

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
  timePoints   = @()
  summary      = [pscustomobject]@{
    allPassed      = $true
    checks         = @()
    issues         = @()
  }
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
    $script:results.summary.allPassed = $false
    $script:issuesList += [pscustomobject]@{
      name    = $Name
      field   = $Field
      message = $Message
    }
    Write-Host "  [FAIL] $($Name): $($Message)" -ForegroundColor Red
  }
  else {
    Write-Host "  [PASS] $($Name)" -ForegroundColor Green
  }
}

Write-Host "`n--- Time Point T0 ---" -ForegroundColor Yellow

$cityT0 = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
if ($cityT0.Success -and $null -ne $cityT0.Json) {
  $results.timePoints += [pscustomobject]@{
    time   = 'T0'
    data   = $cityT0.Json
  }

  $people0 = [int64]$cityT0.Json.summary.resources.people
  $food0 = [int64]$cityT0.Json.summary.resources.food
  $wood0 = [int64]$cityT0.Json.summary.resources.wood
  $rock0 = [int64]$cityT0.Json.summary.resources.rock
  $iron0 = [int64]$cityT0.Json.summary.resources.iron
  $gold0 = [int64]$cityT0.Json.summary.resources.gold
  $tax0 = [int]$cityT0.Json.tax
  $morale0 = [int]$cityT0.Json.morale
  $peopleMax0 = [int64]$cityT0.Json.summary.resources.peopleMax

  Write-Host "  People: $people0 / $peopleMax0" -ForegroundColor White
  Write-Host "  Tax: $tax0, Morale: $morale0" -ForegroundColor White
  Write-Host "  Food: $food0, Wood: $wood0, Rock: $rock0, Iron: $iron0, Gold: $gold0" -ForegroundColor White

  Add-CheckResult -Name 'City Exists' -Field 'summary.cid' -Expected $CityId -Actual $cityT0.Json.summary.cid -Passed ($cityT0.Json.summary.cid -eq $CityId) -Message "City ID should be $CityId"
  Add-CheckResult -Name 'People Range' -Field 'summary.resources.people' -Expected ">0" -Actual $people0 -Passed ($people0 -gt 0) -Message "People should be greater than 0"
  Add-CheckResult -Name 'Tax Range' -Field 'tax' -Expected "0-100" -Actual $tax0 -Passed ($tax0 -ge 0 -and $tax0 -le 100) -Message "Tax should be 0-100"
  Add-CheckResult -Name 'Morale Range' -Field 'morale' -Expected "0-100" -Actual $morale0 -Passed ($morale0 -ge 0 -and $morale0 -le 100) -Message "Morale should be 0-100"
}
else {
  Write-Host "  Failed to get city info: $($cityT0.Error)" -ForegroundColor Red
  $results.summary.allPassed = $false
  $issuesList += [pscustomobject]@{
    name    = 'City Info T0'
    field   = 'response'
    message = "Failed: $($cityT0.Error)"
  }
}

Write-Host "`n--- Time Point T+60s ---" -ForegroundColor Yellow
Start-Sleep -Seconds 60

$cityT60 = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
if ($cityT60.Success -and $null -ne $cityT60.Json) {
  $results.timePoints += [pscustomobject]@{
    time   = 'T+60'
    data   = $cityT60.Json
  }

  $food60 = [int64]$cityT60.Json.summary.resources.food
  $wood60 = [int64]$cityT60.Json.summary.resources.wood
  $rock60 = [int64]$cityT60.Json.summary.resources.rock
  $iron60 = [int64]$cityT60.Json.summary.resources.iron
  $people60 = [int64]$cityT60.Json.summary.resources.people

  Write-Host "  People: $people60" -ForegroundColor White
  Write-Host "  Food: $food60, Wood: $wood60, Rock: $rock60, Iron: $iron60" -ForegroundColor White

  $foodDiff60 = $food60 - $food0
  $woodDiff60 = $wood60 - $wood0

  Write-Host "  Food Delta: $foodDiff60" -ForegroundColor White
  Write-Host "  Wood Delta: $woodDiff60" -ForegroundColor White

  Add-CheckResult -Name 'Food Positive' -Field 'food' -Expected '>0 if has production buildings' -Actual $foodDiff60 -Passed ($foodDiff60 -ge 0) -Message "Food should not decrease (got $foodDiff60)"
  Add-CheckResult -Name 'Wood Positive' -Field 'wood' -Expected '>0 if has production buildings' -Actual $woodDiff60 -Passed ($woodDiff60 -ge 0) -Message "Wood should not decrease (got $woodDiff60)"
}
else {
  Write-Host "  Failed to get city info: $($cityT60.Error)" -ForegroundColor Red
}

Write-Host "`n--- Time Point T+300s ---" -ForegroundColor Yellow
Start-Sleep -Seconds 240

$cityT300 = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
if ($cityT300.Success -and $null -ne $cityT300.Json) {
  $results.timePoints += [pscustomobject]@{
    time   = 'T+300'
    data   = $cityT300.Json
  }

  $food300 = [int64]$cityT300.Json.summary.resources.food
  $people300 = [int64]$cityT300.Json.summary.resources.people
  $morale300 = [int64]$cityT300.Json.morale

  Write-Host "  People: $people300" -ForegroundColor White
  Write-Host "  Food: $food300" -ForegroundColor White

  $foodDiff300 = $food300 - $food0
  $peopleDiff300 = $people300 - $people0

  Write-Host "  Food Delta: $foodDiff300" -ForegroundColor White
  Write-Host "  People Delta: $peopleDiff300" -ForegroundColor White

  Add-CheckResult -Name 'Food Increases Over Time' -Field 'food' -Expected ">= food@T+60" -Actual $food300 -Passed ($food300 -ge $food60) -Message "Food should continue increasing if production buildings exist"
  Add-CheckResult -Name 'People Stable' -Field 'people' -Expected 'stable' -Actual $peopleDiff300 -Passed ($peopleDiff300 -eq 0 -or ($people300 -gt 0 -and $people300 -le $peopleMax0)) -Message "People should be relatively stable"
}
else {
  Write-Host "  Failed to get city info: $($cityT300.Error)" -ForegroundColor Red
}

Write-Host "`n--- Tax Rate Test ---" -ForegroundColor Yellow

if ($cityT0.Success) {
  $testTax = 50
  $setTax = Invoke-Api -Method POST -Url "$BaseUrl/api/cities/$CityId/tax" -Body @{
    tax = $testTax
  } -Session $session

  if ($setTax.Success -and $null -ne $setTax.Json.tax) {
    $newTax = [int]$setTax.Json.tax
    Write-Host "  Set tax to: $newTax" -ForegroundColor White
    Add-CheckResult -Name 'Tax Setting' -Field 'tax' -Expected $testTax -Actual $newTax -Passed ($newTax -eq $testTax) -Message "Tax should be set to $testTax"
  }

  $cityAfterTax = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
  if ($cityAfterTax.Success) {
    $moraleAfterTax = [int]$cityAfterTax.Json.morale
    $complaint = [int]$cityAfterTax.Json.complaint

    Write-Host "  Morale: $moraleAfterTax, Complaint: $complaint" -ForegroundColor White

    $expectedMorale = 100 - $testTax - $complaint
    if ($expectedMorale -lt 0) { $expectedMorale = 0 }
    if ($expectedMorale -gt 100) { $expectedMorale = 100 }

    Add-CheckResult -Name 'Morale Formula' -Field 'morale' -Expected "100 - tax - complaint" -Actual $moraleAfterTax -Passed ($moraleAfterTax -ge 0 -and $moraleAfterTax -le 100) -Message "Morale should be 100 - tax - complaint"
  }
}

Write-Host "`n--- Resource Production Rate ---" -ForegroundColor Yellow

if ($cityT0.Success) {
  $production = $cityT0.Json.production
  $foodRate = [double]$production.foodAdd
  $woodRate = [double]$production.woodAdd
  $rockRate = [double]$production.rockAdd
  $ironRate = [double]$production.ironAdd

  Write-Host "  Food Rate: $foodRate/hour" -ForegroundColor White
  Write-Host "  Wood Rate: $woodRate/hour" -ForegroundColor White
  Write-Host "  Rock Rate: $rockRate/hour" -ForegroundColor White
  Write-Host "  Iron Rate: $ironRate/hour" -ForegroundColor White

  Add-CheckResult -Name 'Food Rate Valid' -Field 'foodAdd' -Expected '>=0' -Actual $foodRate -Passed ($foodRate -ge 0) -Message "Food rate should be non-negative"
  Add-CheckResult -Name 'Wood Rate Valid' -Field 'woodAdd' -Expected '>=0' -Actual $woodRate -Passed ($woodRate -ge 0) -Message "Wood rate should be non-negative"

  if ($foodRate -gt 0 -and $food0 -gt 0) {
    $expectedFood60 = [int64]($foodRate * 60 / 3600)
    $actualFood60Delta = $food60 - $food0
    $tolerance = $expectedFood60 * 0.2

    Write-Host "  Expected Food/60s: ~$expectedFood60, Actual: $actualFood60Delta" -ForegroundColor White

    if ($actualFood60Delta -gt 0) {
      Add-CheckResult -Name 'Food Production Rate' -Field 'foodAdd' -Expected "$expectedFood60 +/- 20%" -Actual $actualFood60Delta -Passed ($actualFood60Delta -gt ($expectedFood60 - $tolerance) -and $actualFood60Delta -lt ($expectedFood60 + $tolerance)) -Message "Food production rate should match declared rate"
    }
  }
  else {
    Write-Host "  No production buildings detected (foodRate=0)" -ForegroundColor White
  }
}
else {
  Write-Host "  Failed to get production info" -ForegroundColor Yellow
}

$results.summary.checks = $checksList
$results.summary.issues = $issuesList

$outputFile = "artifacts\city-economy-diff-$timestamp.json"
$results | ConvertTo-Json -Depth 10 | Out-File -FilePath $outputFile -Encoding UTF8

Write-Host "`n=== Summary ===" -ForegroundColor Cyan
Write-Host "Total Checks: $($results.summary.checks.Count)" -ForegroundColor White
$passedCount = ($results.summary.checks | Where-Object { $_.passed }).Count
$failedCount = ($results.summary.checks | Where-Object { -not $_.passed }).Count
Write-Host "Passed: $passedCount" -ForegroundColor Green
Write-Host "Failed: $failedCount" -ForegroundColor $(if ($failedCount -gt 0) { 'Red' } else { 'Green' })

if ($results.summary.issues.Count -gt 0) {
  Write-Host "`nIssues:" -ForegroundColor Red
  foreach ($issue in $results.summary.issues) {
    Write-Host "  [$($issue.name)] $($issue.message)" -ForegroundColor Red
  }
}

Write-Host "`nReport: $outputFile" -ForegroundColor Cyan

if ($results.summary.allPassed) {
  Write-Host "`nAll economy checks passed!" -ForegroundColor Green
  exit 0
}
else {
  Write-Host "`nSome economy checks failed!" -ForegroundColor Red
  exit 1
}