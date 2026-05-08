param(
  [string]$OldBaseUrl = 'http://127.0.0.1:8088',
  [string]$NewBaseUrl = 'http://127.0.0.1:8080',
  [string]$Passport = 'test',
  [string]$Password = 'test',
  [string]$OutputPath = 'artifacts'
)

$ErrorActionPreference = 'Continue'
$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'

$null = New-Item -ItemType Directory -Force -Path $OutputPath

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

function Compare-Json {
  param(
    [object]$Old,
    [object]$New,
    [string]$Path = ''
  )

  $diffs = @()

  if ($null -eq $Old -and $null -eq $New) {
    return $diffs
  }

  if ($null -eq $Old -or $null -eq $New) {
    $diffs += [pscustomobject]@{
      Path     = $Path
      Type     = 'null_mismatch'
      OldValue = $Old
      NewValue = $New
      Severity = 'warning'
    }
    return $diffs
  }

  if ($Old.GetType().Name -ne $New.GetType().Name) {
    $diffs += [pscustomobject]@{
      Path     = $Path
      Type     = 'type_mismatch'
      OldValue = $Old.GetType().Name
      NewValue = $New.GetType().Name
      Severity = 'warning'
    }
    return $diffs
  }

  if ($Old -is [System.Collections.IDictionary] -or $Old -is [hashtable]) {
    $allKeys = @($Old.Keys) + @($New.Keys) | Sort-Object -Unique
    foreach ($key in $allKeys) {
      $oldVal = $Old[$key]
      $newVal = $New[$key]
      $diffs += Compare-Json -Old $oldVal -New $newVal -Path "$Path/$key"
    }
  }
  elseif ($Old -is [System.Collections.IEnumerable] -and $Old -isnot [string]) {
    $oldArr = @($Old)
    $newArr = @($New)
    $maxLen = [Math]::Max($oldArr.Count, $newArr.Count)
    for ($i = 0; $i -lt $maxLen; $i++) {
      $oldVal = if ($i -lt $oldArr.Count) { $oldArr[$i] } else { $null }
      $newVal = if ($i -lt $newArr.Count) { $newArr[$i] } else { $null }
      $diffs += Compare-Json -Old $oldVal -New $newVal -Path "$Path[$i]"
    }
  }
  else {
    if ($Old -ne $New) {
      $severity = if ($Path -match 'id|Id|ID') { 'low' } else { 'medium' }
      $diffs += [pscustomobject]@{
        Path     = $Path
        Type     = 'value_mismatch'
        OldValue = $Old
        NewValue = $New
        Severity = $severity
      }
    }
  }

  return $diffs
}

$results = [pscustomobject]@{
  timestamp         = $timestamp
  oldServer         = $OldBaseUrl
  newServer         = $NewBaseUrl
  passport          = $Passport
  startTime         = (Get-Date).ToString('o')
  endTime           = ''
  oldServerReachable = $false
  loginSuccess      = $false
  uid               = 0
  defaultCid        = 0
  apiResults        = @()
  summary           = [pscustomobject]@{
    totalTests      = 0
    passed          = 0
    failed          = 0
    mismatches      = 0
    errors          = 0
    criticalIssues  = @()
  }
}

Write-Host "=== Login Test ===" -ForegroundColor Cyan

$oldSession = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$newSession = New-Object Microsoft.PowerShell.Commands.WebRequestSession

# Check if old PHP server is reachable
$oldHealthCheck = Invoke-Api -Method GET -Url "$OldBaseUrl/" -Session $oldSession
if (-not $oldHealthCheck.Success) {
  Write-Host "WARNING: Old PHP server ($OldBaseUrl) is unreachable!" -ForegroundColor Red
  Write-Host "  Cannot perform legacy comparison without old server." -ForegroundColor Yellow
  Write-Host "  To fix: Start old PHP server on port 8088, then re-run this script." -ForegroundColor Yellow

  # Generate report indicating old server unreachable
  $results.endTime = (Get-Date).ToString('o')
  $results.oldServerReachable = $false
  $outputFile = Join-Path $OutputPath "legacy-rule-diff-$timestamp.json"
  $results | ConvertTo-Json -Depth 10 | Out-File -FilePath $outputFile -Encoding UTF8

  Write-Host "`n=== Summary ===" -ForegroundColor Cyan
  Write-Host "Old PHP server: UNREACHABLE" -ForegroundColor Red
  Write-Host "Cannot compare without old server." -ForegroundColor Yellow
  Write-Host "`nReport: $outputFile" -ForegroundColor Cyan

  exit 1
}
$results.oldServerReachable = $true

$oldLogin = Invoke-Api -Method POST -Url "$OldBaseUrl/api/legacy/login" -Body @{ username = $Passport; password = $Password } -Session $oldSession
$newLogin = Invoke-Api -Method POST -Url "$NewBaseUrl/api/auth/login" -Body @{ passport = $Passport; password = $Password } -Session $newSession

Write-Host "Old login: $($oldLogin.Status)" -ForegroundColor $(if ($oldLogin.Success) { 'Green' } else { 'Red' })
Write-Host "New login: $($newLogin.Status)" -ForegroundColor $(if ($newLogin.Success) { 'Green' } else { 'Red' })

if ($newLogin.Success -and $null -ne $newLogin.Json.user) {
  $results.loginSuccess = $true
  $results.uid = [int]$newLogin.Json.user.uid
  $results.defaultCid = [int]$newLogin.Json.user.defaultCid
}

Write-Host "`n=== API Comparison Test ===" -ForegroundColor Cyan

$apis = @(
  @{ api = '/me/tasks'; name = 'Tasks'; urlOld = "$OldBaseUrl/api/me/tasks"; urlNew = "$NewBaseUrl/api/me/tasks" },
  @{ api = '/me/union'; name = 'Union'; urlOld = "$OldBaseUrl/api/me/union"; urlNew = "$NewBaseUrl/api/me/union" },
  @{ api = '/me/shop'; name = 'Shop'; urlOld = "$OldBaseUrl/api/me/shop"; urlNew = "$NewBaseUrl/api/me/shop" },
  @{ api = '/me/charge'; name = 'Charge'; urlOld = "$OldBaseUrl/api/me/charge"; urlNew = "$NewBaseUrl/api/me/charge" },
  @{ api = '/mail?folder=inbox&page=0'; name = 'Mail Inbox'; urlOld = "$OldBaseUrl/api/mail?folder=inbox&page=0"; urlNew = "$NewBaseUrl/api/mail?folder=inbox&page=0" },
  @{ api = '/reports?filter=unread&page=0'; name = 'Reports'; urlOld = "$OldBaseUrl/api/reports?filter=unread&page=0"; urlNew = "$NewBaseUrl/api/reports?filter=unread&page=0" }
)

$kinds = @('user', 'union', 'jungong', 'military', 'rich')
foreach ($kind in $kinds) {
  $apis += @{
    api = "/rankings?kind=$kind"
    name = "Rank-$kind"
    urlOld = "$OldBaseUrl/api/rankings?kind=$kind&page=0"
    urlNew = "$NewBaseUrl/api/rankings?kind=$kind&page=0"
  }
}

foreach ($item in $apis) {
  $oldResp = Invoke-Api -Method GET -Url $item.urlOld -Session $oldSession
  $newResp = Invoke-Api -Method GET -Url $item.urlNew -Session $newSession

  $apiResult = [pscustomobject]@{
    api        = $item.api
    name       = $item.name
    oldStatus  = $oldResp.Status
    newStatus  = $newResp.Status
    match      = $false
    diffs      = @()
    error      = ''
  }

  if ($oldResp.Success -and $newResp.Success) {
    $apiResult.match = ($oldResp.Status -eq $newResp.Status)
    if ($oldResp.Json -and $newResp.Json) {
      $apiResult.diffs = @(Compare-Json -Old $oldResp.Json -New $newResp.Json -Path $item.api)
    }
  }
  else {
    $apiResult.error = "Old: $($oldResp.Error), New: $($newResp.Error)"
  }

  $results.apiResults += $apiResult
  $results.summary.totalTests++

  Write-Host "[$($item.name)] Old=$($oldResp.Status) New=$($newResp.Status) Match=$($apiResult.match)" -ForegroundColor $(if ($apiResult.match) { 'Green' } else { 'Yellow' })
}

if ($results.defaultCid -gt 0) {
  $oldCity = Invoke-Api -Method GET -Url "$OldBaseUrl/api/cities/$($results.defaultCid)/info" -Session $oldSession
  $newCity = Invoke-Api -Method GET -Url "$NewBaseUrl/api/cities/$($results.defaultCid)/info" -Session $newSession

  $cityResult = [pscustomobject]@{
    api        = "/cities/$($results.defaultCid)/info"
    name       = 'City Info'
    oldStatus  = $oldCity.Status
    newStatus  = $newCity.Status
    match      = $false
    diffs      = @()
    error      = ''
  }

  if ($oldCity.Success -and $newCity.Success) {
    $cityResult.match = ($oldCity.Status -eq $newCity.Status)
    if ($oldCity.Json -and $newCity.Json) {
      $cityResult.diffs = @(Compare-Json -Old $oldCity.Json -New $newCity.Json -Path 'city')
    }
  }
  else {
    $cityResult.error = "Old: $($oldCity.Error), New: $($newCity.Error)"
  }

  $results.apiResults += $cityResult
  $results.summary.totalTests++

  Write-Host "[City] Old=$($oldCity.Status) New=$($newCity.Status) Match=$($cityResult.match)" -ForegroundColor $(if ($cityResult.match) { 'Green' } else { 'Yellow' })
}

$results.endTime = (Get-Date).ToString('o')

foreach ($api in $results.apiResults) {
  if ($api.match) {
    $results.summary.passed++
  }
  else {
    $results.summary.failed++
    $results.summary.mismatches += $api.diffs.Count

    if ($api.diffs.Count -gt 0) {
      foreach ($diff in $api.diffs) {
        if ($diff.Severity -eq 'medium' -or $diff.Severity -eq 'high') {
          $results.summary.criticalIssues += [pscustomobject]@{
            api   = $api.api
            path  = $diff.Path
            type  = $diff.Type
            value = "$($diff.OldValue) -> $($diff.NewValue)"
          }
        }
      }
    }
  }

  if ($api.error) {
    $results.summary.errors++
  }
}

$outputFile = Join-Path $OutputPath "legacy-rule-diff-$timestamp.json"
$results | ConvertTo-Json -Depth 10 | Out-File -FilePath $outputFile -Encoding UTF8

Write-Host "`n=== Summary ===" -ForegroundColor Cyan
Write-Host "Total: $($results.summary.totalTests)" -ForegroundColor White
Write-Host "Passed: $($results.summary.passed)" -ForegroundColor Green
Write-Host "Failed: $($results.summary.failed)" -ForegroundColor $(if ($results.summary.failed -gt 0) { 'Red' } else { 'Green' })
Write-Host "Mismatches: $($results.summary.mismatches)" -ForegroundColor Yellow
Write-Host "Errors: $($results.summary.errors)" -ForegroundColor $(if ($results.summary.errors -gt 0) { 'Red' } else { 'Green' })

if ($results.summary.criticalIssues.Count -gt 0) {
  Write-Host "`nCritical Issues:" -ForegroundColor Red
  foreach ($issue in $results.summary.criticalIssues) {
    Write-Host "  [$($issue.api)] $($issue.path): $($issue.value)" -ForegroundColor Red
  }
}

Write-Host "`nReport: $outputFile" -ForegroundColor Cyan

if ($results.summary.failed -gt 0) {
  exit 1
}
exit 0