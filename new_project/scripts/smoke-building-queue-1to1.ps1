param(
  [string]$BaseUrl = 'http://127.0.0.1:8080',
  [string]$Passport = 'test',
  [string]$Password = 'test',
  [int]$CityId = 0
)

$ErrorActionPreference = 'Continue'
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
    return [pscustomobject]@{ Status = [int]$resp.StatusCode; Json = $json; Success = $true }
  } catch {
    $status = if ($_.Exception.Response) { [int]$_.Exception.Response.StatusCode } else { 0 }
    return [pscustomobject]@{ Status = $status; Json = $null; Success = $false; Error = $_.Exception.Message }
  }
}

Write-Host "=== Building Queue 1:1 Test ===" -ForegroundColor Cyan
Write-Host "Testing: building options -> create -> cancel -> upgrade -> complete" -ForegroundColor White

$results = @{ timestamp = $timestamp; steps = @(); checks = @(); allPassed = $true }
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

# Step 2: Get initial building state
Write-Host "`n--- Step 2: Initial Building State ---" -ForegroundColor Yellow
$cityBefore = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
Write-Check $cityBefore.Success "City API accessible (status=$($cityBefore.Status))"

# Find an unlocked empty position with buildable options for testing
$occupiedPositions = @{}
if ($cityBefore.Success -and $cityBefore.Json.buildings) {
  foreach ($b in $cityBefore.Json.buildings) {
    $occupiedPositions[[int]$b.position] = $true
  }
}

$testPosition = 0
$buildOption = $null
$options = $null
for ($pos = 1; $pos -le 200; $pos++) {
  if ($occupiedPositions.ContainsKey($pos)) { continue }
  $opts = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId/buildings/options?position=$pos" -Session $session
  if ($opts.Success -and $opts.Json.slot -and $null -ne $opts.Json.options) {
    $slot = $opts.Json.slot
    $optList = @($opts.Json.options)
    if ($slot.unlocked -eq $true -and $slot.occupied -eq $false -and $optList.Count -gt 0) {
      foreach ($opt in $optList) {
        if ($opt.canBuild -eq $true) {
          $buildOption = $opt
          break
        }
      }
      if ($null -ne $buildOption) {
        $testPosition = $pos
        $options = $opts
        break
      }
    }
  }
}

if ($testPosition -eq 0) {
  for ($pos = 1; $pos -le 200; $pos++) {
    if ($occupiedPositions.ContainsKey($pos)) { continue }
    $opts = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId/buildings/options?position=$pos" -Session $session
    if ($opts.Success -and $opts.Json.slot) {
      $slot = $opts.Json.slot
      if ($slot.unlocked -eq $true -and $slot.occupied -eq $false) {
        $testPosition = $pos
        $options = $opts
        break
      }
    }
  }
}
Write-Host "  Test position: $testPosition" -ForegroundColor White
$results.steps += @{ step = "find_slot"; passed = ($testPosition -gt 0); position = $testPosition }

# Step 3: Building Options API
Write-Host "`n--- Step 3: Building Options ---" -ForegroundColor Yellow
if ($null -eq $options) {
  $options = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId/buildings/options?position=$testPosition" -Session $session
}
Write-Check $options.Success "Building options accessible (status=$($options.Status))"
Write-Check ($options.Json.slot -ne $null) "Slot info returned"
Write-Check ($null -ne $options.Json.options) "Options array returned"
$optionsCount = if ($null -eq $options.Json.options) { 0 } else { @($options.Json.options).Count }
Write-Check ($optionsCount -gt 0) "Options have items (count=$optionsCount)"

# Find a canBuild option
if ($null -eq $buildOption -and $options.Json.options) {
  foreach ($opt in $options.Json.options) {
    if ($opt.canBuild -eq $true) {
      $buildOption = $opt
      break
    }
  }
}
$results.steps += @{ step = "options"; passed = ($buildOption -ne $null); optionCount = $optionsCount }

# Step 4: Create Building
Write-Host "`n--- Step 4: Create Building ---" -ForegroundColor Yellow
if ($null -eq $buildOption) {
  Write-Host "  No buildable option found, using bid=1 (鍐滅敯) at unlocked slot" -ForegroundColor Yellow
  $buildOption = @{ bid = 1; name = "鍐滅敯" }
}

Write-Host "  Creating: bid=$($buildOption.bid) ($($buildOption.name)) at position $testPosition" -ForegroundColor White
$create = Invoke-Api -Method POST -Url "$BaseUrl/api/cities/$CityId/buildings/create" -Body @{
  position = $testPosition
  bid = $buildOption.bid
} -Session $session

Write-Host "  Create Status: $($create.Status)" -ForegroundColor White
if ($create.Json.message) {
  Write-Host "  Response: $($create.Json.message)" -ForegroundColor Yellow
}
if ($create.Json.ok -eq $false) {
  Write-Host "  Error: $($create.Json.message)" -ForegroundColor Red
}

$createOk = ($create.Status -eq 200)
Write-Check $createOk "Create returns 200 (status=$($create.Status))"

# Step 4b: Verify building in city response immediately after create
$cityAfterCreate = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
$foundAfterCreate = $false
if ($cityAfterCreate.Json.buildings) {
  foreach ($b in $cityAfterCreate.Json.buildings) {
    if ($b.position -eq $testPosition) {
      $foundAfterCreate = $true
      Write-Host "  Building confirmed in city: bid=$($b.bid), state=$($b.state)" -ForegroundColor Green
      break
    }
  }
}
Write-Check $foundAfterCreate "Building appears in city after create"
$results.steps += @{ step = "create"; passed = $createOk; bid = $buildOption.bid; position = $testPosition; confirmed = $foundAfterCreate }

# Step 5: Cancel Building
Write-Host "`n--- Step 5: Cancel Building ---" -ForegroundColor Yellow
$cancel = Invoke-Api -Method POST -Url "$BaseUrl/api/cities/$CityId/buildings/cancel" -Body @{
  position = $testPosition
} -Session $session

Write-Host "  Cancel Status: $($cancel.Status)" -ForegroundColor White
$cancelMsg = ""
if ($cancel.Json.message) { $cancelMsg = $cancel.Json.message }
if ($cancel.Json.error) { $cancelMsg = $cancel.Json.error }
Write-Host "  Response: $cancelMsg" -ForegroundColor Gray
$cancelOk = ($cancel.Status -eq 200)
Write-Check $cancelOk "Cancel returns 200 (status=$($cancel.Status))"
$results.steps += @{ step = "cancel"; passed = $cancelOk }

# Step 6: Re-create for completion test (skip cancel step this time)
Write-Host "`n--- Step 6: Re-create for Completion Test ---" -ForegroundColor Yellow
$create2 = Invoke-Api -Method POST -Url "$BaseUrl/api/cities/$CityId/buildings/create" -Body @{
  position = $testPosition
  bid = $buildOption.bid
} -Session $session

Write-Host "  Create Status: $($create2.Status)" -ForegroundColor White
$create2Ok = ($create2.Status -eq 200)
Write-Check $create2Ok "Re-create returns 200"
$results.steps += @{ step = "recreate"; passed = $create2Ok }

# Step 7: Wait for completion
Write-Host "`n--- Step 7: Wait for Completion ---" -ForegroundColor Yellow
$buildingCompleted = $false
$completedBuilding = $null
$cityAfter = $null
$maxPoll = 12
for ($i = 1; $i -le $maxPoll; $i++) {
  Start-Sleep -Seconds 5
  $cityAfter = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
  if (-not $cityAfter.Success) { continue }
  if ($cityAfter.Json.buildings) {
    foreach ($b in $cityAfter.Json.buildings) {
      if ($b.position -eq $testPosition) {
        $completedBuilding = $b
        if ($b.state -eq 0) {
          $buildingCompleted = $true
          Write-Host "  Building completed at poll#${i}: bid=$($b.bid), level=$($b.level), state=$($b.state)" -ForegroundColor Green
        } else {
          Write-Host "  Poll#$i still busy: state=$($b.state)" -ForegroundColor Gray
        }
        break
      }
    }
  }
  if ($buildingCompleted) { break }
}
Write-Check ($null -ne $cityAfter -and $cityAfter.Success) "City refresh accessible"
Write-Check $buildingCompleted "Building completed (state=0)"
$results.steps += @{ step = "complete"; passed = $buildingCompleted }

# Step 8: Upgrade Building
Write-Host "`n--- Step 8: Upgrade Building ---" -ForegroundColor Yellow
if (-not $buildingCompleted) {
  Write-Host "  Building not completed in time; recreating completion baseline before upgrade..." -ForegroundColor Yellow
  $recreateBaseline = Invoke-Api -Method POST -Url "$BaseUrl/api/cities/$CityId/buildings/create" -Body @{
    position = $testPosition
    bid = $buildOption.bid
  } -Session $session
  if ($recreateBaseline.Status -eq 200) {
    for ($i = 1; $i -le $maxPoll; $i++) {
      Start-Sleep -Seconds 5
      $cityTmp = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session
      if ($cityTmp.Success -and $cityTmp.Json.buildings) {
        foreach ($b in $cityTmp.Json.buildings) {
          if ($b.position -eq $testPosition -and $b.state -eq 0) {
            $buildingCompleted = $true
            break
          }
        }
      }
      if ($buildingCompleted) { break }
    }
  }
}
$upgrade = Invoke-Api -Method POST -Url "$BaseUrl/api/cities/$CityId/buildings/upgrade" -Body @{
  position = $testPosition
} -Session $session

Write-Host "  Upgrade Status: $($upgrade.Status)" -ForegroundColor White
if ($upgrade.Json.message) {
  Write-Host "  Response: $($upgrade.Json.message)" -ForegroundColor Gray
}

$upgradeOk = ($upgrade.Status -eq 200)
Write-Check $upgradeOk "Upgrade returns 200 (status=$($upgrade.Status))"

# Check upgrade started (state should be 1=upgrading)
$upgradeStarted = $false
if ($upgrade.Json.detail) {
  foreach ($b in $upgrade.Json.detail.buildings) {
    if ($b.position -eq $testPosition -and $b.state -eq 1) {
      $upgradeStarted = $true
      Write-Host "  Upgrade started: state=$($b.state)" -ForegroundColor Green
      break
    }
  }
}
if (-not $upgradeStarted) {
  Write-Host "  Upgrade response: $($upgrade.Json | ConvertTo-Json -Compress)" -ForegroundColor Gray
}
$results.steps += @{ step = "upgrade"; passed = $upgradeOk }

# Step 9: Cancel Upgrade
Write-Host "`n--- Step 9: Cancel Upgrade ---" -ForegroundColor Yellow
$cancel2 = Invoke-Api -Method POST -Url "$BaseUrl/api/cities/$CityId/buildings/cancel" -Body @{
  position = $testPosition
} -Session $session

Write-Host "  Cancel Status: $($cancel2.Status)" -ForegroundColor White
$cancel2Ok = ($cancel2.Status -eq 200)
Write-Check $cancel2Ok "Cancel upgrade returns 200"
$results.steps += @{ step = "cancel_upgrade"; passed = $cancel2Ok }

# Step 10: Final state check
Write-Host "`n--- Step 10: Final State Check ---" -ForegroundColor Yellow
$cityFinal = Invoke-Api -Method GET -Url "$BaseUrl/api/cities/$CityId" -Session $session

$finalBuilding = $null
if ($cityFinal.Json.buildings) {
  foreach ($b in $cityFinal.Json.buildings) {
    if ($b.position -eq $testPosition) {
      $finalBuilding = $b
      break
    }
  }
}

$fieldChecks = @('bid', 'name', 'level', 'position', 'state', 'stateStartTime', 'stateEndTime')
$allFieldsPresent = $true
if ($finalBuilding) {
  Write-Host "  Final building: bid=$($finalBuilding.bid), level=$($finalBuilding.level), state=$($finalBuilding.state)" -ForegroundColor White
  foreach ($field in $fieldChecks) {
    $hasField = $null -ne $finalBuilding.$field
    Write-Check $hasField "Building has '$field' field"
    if (-not $hasField) { $allFieldsPresent = $false }
  }
} else {
  Write-Host "  Building not found at position $testPosition" -ForegroundColor Yellow
  Write-Check $false "Building exists at position $testPosition"
  $allFieldsPresent = $false
}
$results.steps += @{ step = "final_state"; passed = $allFieldsPresent }

# Output
$results.allPassed = $script:allPassed
$outputFile = "artifacts\building-queue-$timestamp.json"
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

