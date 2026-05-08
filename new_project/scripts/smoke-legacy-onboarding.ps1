param(
  [string]$BaseUrl = 'http://127.0.0.1:8080',
  [string]$Passport = 'test',
  [string]$Password = 'test',
  [string]$CommanderName = '',
  [string]$CityName = '',
  [int]$Province = 0,
  [string]$FlagChar = 'A',
  [int]$Sex = 0,
  [int]$Face = 0,
  [int]$QueueMaxPoll = 20,
  [int]$RoleCreateMaxRetry = 5
)

$ErrorActionPreference = 'Stop'

function New-RandomName {
  param([string]$Prefix = 'u')
  $suffix = Get-Random -Minimum 100000 -Maximum 999999
  return "$Prefix$suffix"
}

function Legacy-FailureCode {
  param([object]$Raw)
  if ($null -eq $Raw -or $Raw.Count -eq 0) {
    return ''
  }
  if ([int]$Raw[0] -ne 0) {
    return ''
  }
  if ($Raw.Count -ge 2) {
    return [string]$Raw[1]
  }
  return 'legacy_failed'
}

$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession

$loginPayload = @{
  version = 0
  loginType = 0
  passType = 'local'
  passport = $Passport
  password = $Password
  auth = ''
}

$login = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/legacy/login" -WebSession $session -ContentType 'application/json' -Body ($loginPayload | ConvertTo-Json)
$loginFailure = Legacy-FailureCode $login.raw
if ($loginFailure) {
  throw "legacy_login_failed:$loginFailure"
}

$uid = [int]$login.uid
$sid = [int64]$login.sid
$queued = [bool]$login.queued
$logged = [bool]$login.logged
$queuePollCount = 0

while ($queued -and -not $logged -and $queuePollCount -lt $QueueMaxPoll) {
  Start-Sleep -Seconds 2
  $queuePollCount += 1
  $queueResp = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/legacy/login/queue" -WebSession $session -ContentType 'application/json' -Body (@{
      uid = $uid
      sid = $sid
    } | ConvertTo-Json)

  $queueFailure = Legacy-FailureCode $queueResp.raw
  if ($queueFailure) {
    throw "legacy_queue_failed:$queueFailure"
  }

  $queued = [bool]$queueResp.queued
  $logged = [bool]$queueResp.logged
}

if (-not $logged) {
  throw 'legacy_login_not_completed'
}

$meBefore = Invoke-RestMethod -Method Get -Uri "$BaseUrl/api/auth/me" -WebSession $session
if ($null -eq $meBefore.user) {
  throw 'auth_me_empty_after_login'
}

$roleCreateAttempts = 0
$roleCreateLastFailure = ''
$roleCreateResult = $null
$createdRole = $false
$finalCommander = if ([string]::IsNullOrWhiteSpace($CommanderName)) { New-RandomName 'u' } else { $CommanderName.Trim() }
$finalCity = if ([string]::IsNullOrWhiteSpace($CityName)) { New-RandomName 'c' } else { $CityName.Trim() }

while ($meBefore.user.cityCount -le 0 -and $roleCreateAttempts -lt $RoleCreateMaxRetry) {
  $roleCreateAttempts += 1
  $rolePayload = @{
    uid = $uid
    sid = $sid
    userName = $finalCommander
    cityName = $finalCity
    province = $Province
    flagChar = $FlagChar
    sex = $Sex
    face = $Face
    code = ''
  }

  $roleCreateResult = Invoke-RestMethod -Method Post -Uri "$BaseUrl/api/legacy/role/create" -WebSession $session -ContentType 'application/json' -Body ($rolePayload | ConvertTo-Json)
  $roleFailure = Legacy-FailureCode $roleCreateResult.raw

  if (-not $roleFailure -and [int]$roleCreateResult.cid -gt 0) {
    $createdRole = $true
    break
  }

  $roleCreateLastFailure = $roleFailure
  if ($roleFailure -in @('invalid_char', 'name_illegal', 'used_city_holder_name')) {
    $finalCommander = New-RandomName 'u'
    $finalCity = New-RandomName 'c'
    continue
  }

  break
}

$meAfter = Invoke-RestMethod -Method Get -Uri "$BaseUrl/api/auth/me" -WebSession $session
$myCities = Invoke-RestMethod -Method Get -Uri "$BaseUrl/api/me/cities?limit=5" -WebSession $session

[pscustomobject]@{
  passport = $Passport
  uid = $uid
  sid = $sid
  queue_poll_count = $queuePollCount
  me_before_city_count = [int]$meBefore.user.cityCount
  role_create_attempts = $roleCreateAttempts
  role_create_created = $createdRole
  role_create_last_failure = $roleCreateLastFailure
  role_create_commander = $finalCommander
  role_create_city = $finalCity
  me_after_city_count = if ($null -ne $meAfter.user) { [int]$meAfter.user.cityCount } else { -1 }
  me_after_default_cid = if ($null -ne $meAfter.user) { [int]$meAfter.user.defaultCid } else { 0 }
  my_cities_count = if ($null -ne $myCities.items) { [int]$myCities.items.Count } else { 0 }
} | ConvertTo-Json -Depth 4
