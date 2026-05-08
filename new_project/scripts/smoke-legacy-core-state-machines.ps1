param(
  [string]$BaseUrl = 'http://127.0.0.1:8080',
  [string]$TestPassport = 'test',
  [string]$TestPassword = 'test',
  [int]$ResearchCommanderUID = 687,
  [int]$ResearchCityCID = 204014,
  [int]$ResearchPosition = 153,
  [int]$DraftCommanderUID = 895,
  [int]$DraftCityCID = 5035,
  [int]$DraftPosition = 114
)

$ErrorActionPreference = 'Stop'

function Get-Json {
  param(
    [Microsoft.PowerShell.Commands.WebRequestSession]$Session,
    [string]$Url
  )

  Invoke-RestMethod -Method Get -WebSession $Session -Uri $Url
}

function Post-Json {
  param(
    [Microsoft.PowerShell.Commands.WebRequestSession]$Session,
    [string]$Url,
    [object]$Body
  )

  $payload = if ($null -eq $Body) { '{}' } else { $Body | ConvertTo-Json -Depth 8 }
  Invoke-RestMethod -Method Post -WebSession $Session -Uri $Url -ContentType 'application/json' -Body $payload
}

function Patch-Json {
  param(
    [Microsoft.PowerShell.Commands.WebRequestSession]$Session,
    [string]$Url,
    [object]$Body
  )

  Invoke-RestMethod -Method Patch -WebSession $Session -Uri $Url -ContentType 'application/json' -Body ($Body | ConvertTo-Json -Depth 8)
}

function New-Session {
  New-Object Microsoft.PowerShell.Commands.WebRequestSession
}

function Ensure-TestAccountCity {
  param(
    [Microsoft.PowerShell.Commands.WebRequestSession]$Session
  )

  $login = Post-Json -Session $Session -Url "$BaseUrl/api/auth/login" -Body @{
    passport = $TestPassport
    password = $TestPassword
  }
  if ($null -eq $login.user) {
    throw 'test_account_login_failed'
  }
  if ([int]$login.user.cityCount -gt 0) {
    return [int]$login.user.defaultCid
  }

  $legacyLogin = Post-Json -Session $Session -Url "$BaseUrl/api/legacy/login" -Body @{
    version   = 0
    loginType = 0
    passType  = 'local'
    passport  = $TestPassport
    password  = $TestPassword
    auth      = ''
  }
  if (-not $legacyLogin.logged) {
    throw 'legacy_login_for_test_account_failed'
  }

  $uid = [int]$legacyLogin.uid
  $sid = [int64]$legacyLogin.sid
  if ($uid -le 0 -or $sid -le 0) {
    throw 'legacy_login_missing_uid_or_sid'
  }

  $suffix = Get-Random -Minimum 100000 -Maximum 999999
  $roleCreate = Post-Json -Session $Session -Url "$BaseUrl/api/legacy/role/create" -Body @{
    uid      = $uid
    sid      = $sid
    userName = "u$suffix"
    cityName = "c$suffix"
    province = 0
    flagChar = 'A'
    sex      = 0
    face     = 0
    code     = ''
  }

  if ([int]$roleCreate.cid -le 0) {
    throw ('legacy_role_create_failed:' + (($roleCreate.raw | ConvertTo-Json -Compress)))
  }

  $me = Get-Json -Session $Session -Url "$BaseUrl/api/auth/me"
  if ($null -eq $me.user -or [int]$me.user.cityCount -le 0) {
    throw 'test_account_still_has_no_city_after_role_create'
  }
  return [int]$me.user.defaultCid
}

function Find-BuildableSlot {
  param(
    [Microsoft.PowerShell.Commands.WebRequestSession]$Session,
    [int]$CID
  )

  $slots = 100..155
  foreach ($position in $slots) {
    try {
      $options = Get-Json -Session $Session -Url "$BaseUrl/api/cities/$CID/buildings/options?position=$position"
      if ($options.slot.occupied) {
        continue
      }
      $candidate = $options.options | Where-Object { $_.canBuild -eq $true } | Select-Object -First 1
      if ($null -ne $candidate) {
        return [pscustomobject]@{
          position = [int]$position
          bid      = [int]$candidate.bid
          name     = [string]$candidate.name
        }
      }
    } catch {
    }
  }
  return $null
}

# 0) health
$health = Get-Json -Session (New-Session) -Url "$BaseUrl/api/health"

# 1) test account city baseline + tax/production/building state machine
$testSession = New-Session
$testCID = Ensure-TestAccountCity -Session $testSession
$testBefore = Get-Json -Session $testSession -Url "$BaseUrl/api/cities/$testCID"

$taxRoundtrip = Patch-Json -Session $testSession -Url "$BaseUrl/api/cities/$testCID/tax" -Body @{
  tax = [int]$testBefore.tax
}
$productionRoundtrip = Patch-Json -Session $testSession -Url "$BaseUrl/api/cities/$testCID/production" -Body @{
  foodRate = [int]$testBefore.production.settings.foodRate
  woodRate = [int]$testBefore.production.settings.woodRate
  rockRate = [int]$testBefore.production.settings.rockRate
  ironRate = [int]$testBefore.production.settings.ironRate
}

$buildSlot = Find-BuildableSlot -Session $testSession -CID $testCID
if ($null -eq $buildSlot) {
  throw 'no_buildable_slot_found_for_test_city'
}

$createResp = Post-Json -Session $testSession -Url "$BaseUrl/api/cities/$testCID/buildings/create" -Body @{
  position = [int]$buildSlot.position
  bid      = [int]$buildSlot.bid
}
$createdBuilding = $createResp.buildings | Where-Object { $_.position -eq [int]$buildSlot.position } | Select-Object -First 1
if ($null -eq $createdBuilding -or [int]$createdBuilding.state -ne 1 -or [int]$createdBuilding.level -ne 0) {
  throw 'building_create_state_machine_mismatch'
}

$cancelCreateResp = Post-Json -Session $testSession -Url "$BaseUrl/api/cities/$testCID/buildings/cancel" -Body @{
  position = [int]$buildSlot.position
}
$createdAfterCancel = $cancelCreateResp.buildings | Where-Object { $_.position -eq [int]$buildSlot.position } | Select-Object -First 1
if ($null -ne $createdAfterCancel) {
  throw 'building_create_cancel_should_remove_level0_building'
}

$govUpgradeResp = Post-Json -Session $testSession -Url "$BaseUrl/api/cities/$testCID/buildings/upgrade" -Body @{
  position = 120
}
$govAfterUpgrade = $govUpgradeResp.buildings | Where-Object { $_.position -eq 120 } | Select-Object -First 1
if ($null -eq $govAfterUpgrade -or [int]$govAfterUpgrade.state -ne 1) {
  throw 'government_upgrade_did_not_enter_busy_state'
}

$govCancelResp = Post-Json -Session $testSession -Url "$BaseUrl/api/cities/$testCID/buildings/cancel" -Body @{
  position = 120
}
$govAfterCancel = $govCancelResp.buildings | Where-Object { $_.position -eq 120 } | Select-Object -First 1
if ($null -eq $govAfterCancel -or [int]$govAfterCancel.state -ne 0 -or [int]$govAfterCancel.level -ne [int]$govAfterUpgrade.level) {
  throw 'government_upgrade_cancel_state_machine_mismatch'
}

# 2) research state machine
$researchSession = New-Session
Post-Json -Session $researchSession -Url "$BaseUrl/api/auth/login" -Body @{ uid = $ResearchCommanderUID } | Out-Null
$researchBefore = Get-Json -Session $researchSession -Url "$BaseUrl/api/cities/$ResearchCityCID/research?position=$ResearchPosition"
$researchCandidate = $researchBefore.options | Where-Object { $_.canUpgrade -eq $true } | Select-Object -First 1
if ($null -eq $researchCandidate) {
  throw 'no_research_candidate_can_upgrade'
}

$researchStart = Post-Json -Session $researchSession -Url "$BaseUrl/api/cities/$ResearchCityCID/research/start" -Body @{
  position = $ResearchPosition
  tid      = [int]$researchCandidate.tid
}
$researchDuring = $researchStart.options | Where-Object { $_.tid -eq [int]$researchCandidate.tid } | Select-Object -First 1
if ([int]$researchStart.activeTid -ne [int]$researchCandidate.tid -or [int]$researchDuring.state -ne 1) {
  throw 'research_start_state_machine_mismatch'
}

$researchCancel = Post-Json -Session $researchSession -Url "$BaseUrl/api/cities/$ResearchCityCID/research/cancel" -Body @{
  position = $ResearchPosition
  tid      = [int]$researchCandidate.tid
}
$researchAfterCancel = $researchCancel.options | Where-Object { $_.tid -eq [int]$researchCandidate.tid } | Select-Object -First 1
if ([int]$researchCancel.activeTid -ne 0 -or [int]$researchAfterCancel.state -ne 0) {
  throw 'research_cancel_state_machine_mismatch'
}

# 3) draft state machine
$draftSession = New-Session
Post-Json -Session $draftSession -Url "$BaseUrl/api/auth/login" -Body @{ uid = $DraftCommanderUID } | Out-Null
$draftBefore = Get-Json -Session $draftSession -Url "$BaseUrl/api/cities/$DraftCityCID/barracks?position=$DraftPosition"
$draftCandidate = $draftBefore.options | Where-Object { $_.canDraft -eq $true } | Select-Object -First 1
if ($null -eq $draftCandidate) {
  throw 'no_draft_candidate_can_draft'
}

$draftStart = Post-Json -Session $draftSession -Url "$BaseUrl/api/cities/$DraftCityCID/barracks/draft/start" -Body @{
  position = $DraftPosition
  sid      = [int]$draftCandidate.sid
  count    = 1
}
$draftQueueRow = $draftStart.queue | Sort-Object -Property id -Descending | Select-Object -First 1
if ($null -eq $draftQueueRow -or [int]$draftStart.queueCount -le [int]$draftBefore.queueCount) {
  throw 'draft_start_state_machine_mismatch'
}

$draftCancel = Post-Json -Session $draftSession -Url "$BaseUrl/api/cities/$DraftCityCID/barracks/draft/cancel" -Body @{
  position = $DraftPosition
  queueId  = [int]$draftQueueRow.id
}
$draftQueueRemaining = $draftCancel.queue | Where-Object { $_.id -eq [int]$draftQueueRow.id } | Select-Object -First 1
if ($null -ne $draftQueueRemaining -or [int]$draftCancel.queueCount -ge [int]$draftStart.queueCount) {
  throw 'draft_cancel_state_machine_mismatch'
}

# 4) troop dispatch/callback state machine
$dispatchCities = (Get-Json -Session $draftSession -Url "$BaseUrl/api/me/cities?limit=40").items
if ($null -eq $dispatchCities -or $dispatchCities.Count -lt 2) {
  throw 'not_enough_cities_for_troop_dispatch_smoke'
}

$dispatchFromCID = 0
$dispatchToCID = 0
$dispatchSoldierSID = 0
$dispatchSoldierCount = 0

foreach ($city in $dispatchCities) {
  $candidateCID = [int]$city.cid
  $candidateDetail = Get-Json -Session $draftSession -Url "$BaseUrl/api/cities/$candidateCID"
  $candidateSoldier = $candidateDetail.soldiers | Where-Object { [int64]$_.count -gt 0 } | Select-Object -First 1
  if ($null -eq $candidateSoldier) {
    continue
  }
  $target = $dispatchCities | Where-Object { [int]$_.cid -ne $candidateCID } | Select-Object -First 1
  if ($null -eq $target) {
    continue
  }

  $dispatchFromCID = $candidateCID
  $dispatchToCID = [int]$target.cid
  $dispatchSoldierSID = [int]$candidateSoldier.sid
  $dispatchSoldierCount = [int64]$candidateSoldier.count
  break
}

if ($dispatchFromCID -le 0 -or $dispatchToCID -le 0 -or $dispatchSoldierSID -le 0) {
  throw 'no_troop_dispatch_sample_found'
}

$dispatchResp = Post-Json -Session $draftSession -Url "$BaseUrl/api/cities/$dispatchFromCID/troops/dispatch" -Body @{
  targetCid    = $dispatchToCID
  soldierSid   = $dispatchSoldierSID
  soldierCount = 1
  task         = 1
}
$dispatchTroop = $dispatchResp.items | Sort-Object -Property id -Descending | Select-Object -First 1
if ($null -eq $dispatchTroop -or [int]$dispatchTroop.state -ne 0) {
  throw 'troop_dispatch_state_machine_mismatch'
}

$callbackResp = Post-Json -Session $draftSession -Url "$BaseUrl/api/troops/$([int]$dispatchTroop.id)/callback" -Body @{}
$callbackTroop = $callbackResp.items | Where-Object { [int]$_.id -eq [int]$dispatchTroop.id } | Select-Object -First 1
if ($null -ne $callbackTroop -and [int]$callbackTroop.state -ne 1) {
  throw 'troop_callback_state_machine_mismatch'
}

[pscustomobject]@{
  ok = $true
  health_connected = [bool]$health.database.connected
  building = [pscustomobject]@{
    cid = [int]$testCID
    slot_position = [int]$buildSlot.position
    slot_bid = [int]$buildSlot.bid
    slot_name = [string]$buildSlot.name
    create_state = [int]$createdBuilding.state
    create_level = [int]$createdBuilding.level
    exists_after_cancel = ($null -ne $createdAfterCancel)
    gov_state_after_upgrade = [int]$govAfterUpgrade.state
    gov_level_after_upgrade = [int]$govAfterUpgrade.level
    gov_state_after_cancel = [int]$govAfterCancel.state
    gov_level_after_cancel = [int]$govAfterCancel.level
  }
  economy = [pscustomobject]@{
    tax = [int]$taxRoundtrip.tax
    foodRate = [int]$productionRoundtrip.production.settings.foodRate
    woodRate = [int]$productionRoundtrip.production.settings.woodRate
    rockRate = [int]$productionRoundtrip.production.settings.rockRate
    ironRate = [int]$productionRoundtrip.production.settings.ironRate
  }
  research = [pscustomobject]@{
    uid = [int]$ResearchCommanderUID
    cid = [int]$ResearchCityCID
    position = [int]$ResearchPosition
    tid = [int]$researchCandidate.tid
    active_after_start = [int]$researchStart.activeTid
    active_after_cancel = [int]$researchCancel.activeTid
    state_after_start = [int]$researchDuring.state
    state_after_cancel = [int]$researchAfterCancel.state
  }
  draft = [pscustomobject]@{
    uid = [int]$DraftCommanderUID
    cid = [int]$DraftCityCID
    position = [int]$DraftPosition
    sid = [int]$draftCandidate.sid
    queue_before = [int]$draftBefore.queueCount
    queue_after_start = [int]$draftStart.queueCount
    queue_after_cancel = [int]$draftCancel.queueCount
    queue_id = [int]$draftQueueRow.id
  }
  troop = [pscustomobject]@{
    uid = [int]$DraftCommanderUID
    from_cid = [int]$dispatchFromCID
    to_cid = [int]$dispatchToCID
    sid = [int]$dispatchSoldierSID
    soldier_count_in_city = [int64]$dispatchSoldierCount
    troop_id = [int]$dispatchTroop.id
    state_after_dispatch = [int]$dispatchTroop.state
    state_after_callback = if ($null -eq $callbackTroop) { -1 } else { [int]$callbackTroop.state }
    callback_total = [int]$callbackResp.total
    callback_returning = [int]$callbackResp.returning
  }
} | ConvertTo-Json -Depth 8
