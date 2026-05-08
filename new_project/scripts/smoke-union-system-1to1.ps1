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
  param(
    [ValidateSet('GET', 'POST', 'PATCH')]
    [string]$Method,
    [string]$Url,
    [object]$Body = $null
  )

  $args = @{
    Method     = $Method
    Uri        = $Url
    WebSession = $session
    TimeoutSec = 12
  }

  if ($Method -ne 'GET') {
    $args.ContentType = 'application/json'
    $args.Body = if ($null -eq $Body) { '{}' } else { $Body | ConvertTo-Json -Depth 12 }
  }

  try {
    $resp = Invoke-WebRequest -UseBasicParsing @args
    $json = $null
    try { $json = $resp.Content | ConvertFrom-Json } catch { }
    return [pscustomobject]@{
      ok     = $true
      status = [int]$resp.StatusCode
      json   = $json
      raw    = $resp.Content
    }
  } catch {
    $status = if ($_.Exception.Response) { [int]$_.Exception.Response.StatusCode } else { 0 }
    $raw = $_.ErrorDetails.Message
    if ([string]::IsNullOrWhiteSpace($raw)) { $raw = $_.Exception.Message }
    $json = $null
    try { $json = $raw | ConvertFrom-Json } catch { }
    return [pscustomobject]@{
      ok     = $false
      status = $status
      json   = $json
      raw    = $raw
    }
  }
}

function Add-Step {
  param([string]$Name, [bool]$Passed, [hashtable]$Extra = @{})
  if (-not $Passed) { $script:allPassed = $false }
  $row = @{ step = $Name; passed = $Passed }
  foreach ($k in $Extra.Keys) { $row[$k] = $Extra[$k] }
  $script:steps += $row
}

$login = Invoke-Api -Method POST -Url "$BaseUrl/api/auth/login" -Body @{
  passport = $Passport
  password = $Password
}
$loginOK = ($login.status -eq 200 -and $null -ne $login.json.user)
Add-Step -Name 'login' -Passed $loginOK -Extra @{ status = $login.status }
if (-not $loginOK) { throw "login failed status=$($login.status)" }

$me = Invoke-Api -Method GET -Url "$BaseUrl/api/auth/me"
$uid = if ($me.json -and $me.json.user) { [int]$me.json.user.uid } else { 0 }
Add-Step -Name 'auth_me' -Passed ($me.status -eq 200 -and $uid -gt 0) -Extra @{ status = $me.status; uid = $uid }

$before = Invoke-Api -Method GET -Url "$BaseUrl/api/me/union"
Add-Step -Name 'union_read_before' -Passed ($before.status -eq 200 -and $null -ne $before.json.permissions) -Extra @{
  status = $before.status
  joined = if ($before.json) { [bool]$before.json.joined } else { $false }
}

$createCovered = $false
$createSucceeded = $false
$createdUnionId = 0
$createdUnionName = ''
$createDeniedStatus = 0
$createDeniedMessage = ''

if ($before.status -eq 200 -and -not [bool]$before.json.joined) {
  if ([bool]$before.json.permissions.canCreate) {
    $createCovered = $true
    $createdUnionName = "u$((Get-Date).ToString('HHmmss'))"
    $create = Invoke-Api -Method POST -Url "$BaseUrl/api/me/union/create" -Body @{ name = $createdUnionName }
    $createSucceeded = ($create.status -eq 200 -and $create.json -and [bool]$create.json.joined -and $null -ne $create.json.summary)
    if ($createSucceeded) {
      $createdUnionId = [int]$create.json.summary.id
    }
    Add-Step -Name 'union_create' -Passed $createSucceeded -Extra @{ status = $create.status; unionId = $createdUnionId; unionName = $createdUnionName }
  } else {
    $createCovered = $true
    $createDenied = Invoke-Api -Method POST -Url "$BaseUrl/api/me/union/create" -Body @{ name = "u$((Get-Date).ToString('HHmmss'))" }
    $createDeniedStatus = [int]$createDenied.status
    $createDeniedMessage = if ($createDenied.json -and $createDenied.json.message) { [string]$createDenied.json.message } else { '' }
    $deniedOK = (($createDenied.status -eq 400 -or $createDenied.status -eq 403) -and $createDeniedMessage -ne '')
    Add-Step -Name 'union_create_denied_permission_check' -Passed $deniedOK -Extra @{ status = $createDenied.status; message = $createDeniedMessage }
  }
}

$readAfterCreate = Invoke-Api -Method GET -Url "$BaseUrl/api/me/union"
$memberReadOK = ($readAfterCreate.status -eq 200 -and $null -ne $readAfterCreate.json.permissions)
if ($createSucceeded) {
  $memberReadOK = $memberReadOK -and [bool]$readAfterCreate.json.joined -and $null -ne $readAfterCreate.json.summary -and $readAfterCreate.json.members.Count -ge 1
}
Add-Step -Name 'union_member_read' -Passed $memberReadOK -Extra @{
  status = $readAfterCreate.status
  joined = if ($readAfterCreate.json) { [bool]$readAfterCreate.json.joined } else { $false }
  members = if ($readAfterCreate.json -and $readAfterCreate.json.members) { [int]$readAfterCreate.json.members.Count } else { 0 }
}

$profileDenied = Invoke-Api -Method POST -Url "$BaseUrl/api/me/union/profile" -Body @{
  name = "u$((Get-Date).ToString('MMddHHmmss'))"
  intro = 'smoke'
  announcement = 'smoke'
}
$permissionCheckOK = ($profileDenied.status -eq 400 -or $profileDenied.status -eq 403)
if ($createSucceeded) {
  $permissionCheckOK = ($profileDenied.status -eq 200)
}
Add-Step -Name 'union_permission_check_profile' -Passed $permissionCheckOK -Extra @{ status = $profileDenied.status }

$joinCovered = $false
$joinApplied = $false
$joinApplyStatus = 0
$joinCancelStatus = 0
if (-not $createSucceeded -and $before.status -eq 200 -and -not [bool]$before.json.joined -and [bool]$before.json.permissions.canApply -and $before.json.directory.Count -gt 0) {
  $joinCovered = $true
  $targetUnionId = [int]$before.json.directory[0].id
  $join = Invoke-Api -Method POST -Url "$BaseUrl/api/me/union/apply" -Body @{ unionId = $targetUnionId }
  $joinApplyStatus = [int]$join.status
  $joinOK = ($join.status -eq 200 -and $join.json -and $null -ne $join.json.application -and [int]$join.json.application.unionId -eq $targetUnionId)
  Add-Step -Name 'union_join_apply' -Passed $joinOK -Extra @{ status = $join.status; targetUnionId = $targetUnionId }
  $joinApplied = $joinOK

  if ($joinApplied) {
    $cancel = Invoke-Api -Method POST -Url "$BaseUrl/api/me/union/apply/cancel"
    $joinCancelStatus = [int]$cancel.status
    $cancelOK = ($cancel.status -eq 200 -and $cancel.json -and $null -eq $cancel.json.application)
    Add-Step -Name 'union_join_apply_cancel' -Passed $cancelOK -Extra @{ status = $cancel.status }
  }
}
if (-not $joinCovered -and $before.status -eq 200 -and -not [bool]$before.json.joined) {
  $joinProbe = Invoke-Api -Method POST -Url "$BaseUrl/api/me/union/apply" -Body @{ unionId = 0 }
  $joinProbeOK = ($joinProbe.status -eq 400 -or $joinProbe.status -eq 403)
  Add-Step -Name 'union_join_permission_or_param_check' -Passed $joinProbeOK -Extra @{ status = $joinProbe.status }
  $joinCovered = $joinProbeOK
}

$leaveStatus = 0
$leaveCovered = $false
$leaveOK = $true
if ($createSucceeded) {
  $leaveCovered = $true
  $leave = Invoke-Api -Method POST -Url "$BaseUrl/api/me/union/leave"
  $leaveStatus = [int]$leave.status
  $leaveOK = ($leave.status -eq 200 -and $leave.json -and -not [bool]$leave.json.joined)
} else {
  $leaveCovered = $true
  $leaveProbe = Invoke-Api -Method POST -Url "$BaseUrl/api/me/union/leave"
  $leaveStatus = [int]$leaveProbe.status
  $leaveOK = ($leaveProbe.status -eq 400 -or $leaveProbe.status -eq 403)
}
Add-Step -Name 'union_leave_or_disband' -Passed $leaveOK -Extra @{ status = $leaveStatus }

$finalRead = Invoke-Api -Method GET -Url "$BaseUrl/api/me/union"
$finalOK = ($finalRead.status -eq 200 -and $null -ne $finalRead.json.permissions)
Add-Step -Name 'union_read_final' -Passed $finalOK -Extra @{
  status = $finalRead.status
  joined = if ($finalRead.json) { [bool]$finalRead.json.joined } else { $false }
}

if (-not $createCovered) { $allPassed = $false; Add-Step -Name 'coverage_create' -Passed $false -Extra @{ reason = 'create path not covered' } }
if (-not $joinCovered) { $allPassed = $false; Add-Step -Name 'coverage_join' -Passed $false -Extra @{ reason = 'join path not covered' } }
if (-not $leaveCovered) { $allPassed = $false; Add-Step -Name 'coverage_leave' -Passed $false -Extra @{ reason = 'leave/disband path not covered' } }

$output = [pscustomobject]@{
  timestamp = $timestamp
  allPassed = $allPassed
  actor = [pscustomobject]@{
    passport = $Passport
    uid = $uid
  }
  coverage = [pscustomobject]@{
    createCovered = $createCovered
    createSucceeded = $createSucceeded
    joinCovered = $joinCovered
    joinApplied = $joinApplied
    memberReadChecked = $true
    permissionChecked = $true
    leaveCovered = $leaveCovered
  }
  details = [pscustomobject]@{
    createdUnionId = $createdUnionId
    createdUnionName = $createdUnionName
    createDeniedStatus = $createDeniedStatus
    createDeniedMessage = $createDeniedMessage
    joinApplyStatus = $joinApplyStatus
    joinCancelStatus = $joinCancelStatus
    leaveStatus = $leaveStatus
  }
  steps = $steps
}

$path = "artifacts/union-system-$timestamp.json"
$output | ConvertTo-Json -Depth 10 | Out-File -FilePath $path -Encoding UTF8
Write-Host $path
if ($allPassed) { exit 0 } else { exit 1 }
