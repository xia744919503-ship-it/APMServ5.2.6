param(
  [string]$BaseUrl = 'http://127.0.0.1:8080',
  [int]$CommanderLimit = 5
)

$ErrorActionPreference = 'Stop'

$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession

function Get-Json {
  param(
    [string]$Url
  )

  Invoke-RestMethod -WebSession $session $Url
}

function Post-Json {
  param(
    [string]$Url,
    [object]$Body
  )

  $payload = if ($null -eq $Body) { '{}' } else { $Body | ConvertTo-Json }
  Invoke-RestMethod -Method Post -WebSession $session -ContentType 'application/json' -Body $payload $Url
}

function Patch-Json {
  param(
    [string]$Url,
    [object]$Body
  )

  Invoke-RestMethod -Method Patch -WebSession $session -ContentType 'application/json' -Body ($Body | ConvertTo-Json) $Url
}

$commanders = Get-Json "$BaseUrl/api/auth/commanders?limit=$CommanderLimit"
if (-not $commanders.items -or $commanders.items.Count -eq 0) {
  throw 'No commander options returned from /api/auth/commanders.'
}

$commander = $commanders.items[0]
$login = Post-Json "$BaseUrl/api/auth/login" @{ uid = $commander.uid }
$me = Get-Json "$BaseUrl/api/auth/me"
$myCities = Get-Json "$BaseUrl/api/me/cities?limit=5"

if (-not $myCities.items -or $myCities.items.Count -eq 0) {
  throw 'No owned cities returned from /api/me/cities after login.'
}

$city = $myCities.items[0]
$detail = Get-Json "$BaseUrl/api/cities/$($city.cid)"

$taxPatch = Patch-Json "$BaseUrl/api/cities/$($city.cid)/tax" @{
  tax = $detail.tax
}

$productionPatch = Patch-Json "$BaseUrl/api/cities/$($city.cid)/production" @{
  foodRate = $detail.production.settings.foodRate
  woodRate = $detail.production.settings.woodRate
  rockRate = $detail.production.settings.rockRate
  ironRate = $detail.production.settings.ironRate
}

$outboxBefore = Get-Json "$BaseUrl/api/mail?folder=outbox&page=0"
$inboxBefore = Get-Json "$BaseUrl/api/mail?folder=inbox&page=0"
$illegalMailStatus = 0
$illegalMailMessage = ''
$illegalMailPayload = $null
try {
  Post-Json "$BaseUrl/api/mail/send" @{
    toName = $me.user.name
    title = '__smoke_illegal_mail__'
    content = '91wanvip.com'
  } | Out-Null
  $illegalMailStatus = 200
} catch {
  if ($_.Exception.Response) {
    $illegalMailStatus = [int]$_.Exception.Response.StatusCode
  }
  $illegalMailMessage = $_.ErrorDetails.Message
  if (-not [string]::IsNullOrWhiteSpace($illegalMailMessage)) {
    try {
      $illegalMailPayload = $illegalMailMessage | ConvertFrom-Json
    } catch {
    }
  }
}
$outboxAfterIllegal = Get-Json "$BaseUrl/api/mail?folder=outbox&page=0"
$inboxAfterIllegal = Get-Json "$BaseUrl/api/mail?folder=inbox&page=0"
if ($illegalMailStatus -ne 400) {
  throw "Expected illegal mail send to fail with 400, got $illegalMailStatus."
}
$illegalMailText = if ($null -ne $illegalMailPayload -and $illegalMailPayload.message) {
  [string]$illegalMailPayload.message
} else {
  $illegalMailMessage
}
if ([string]::IsNullOrWhiteSpace($illegalMailText)) {
  throw 'Expected illegal mail send to return an error message.'
}
if ($outboxAfterIllegal.total -ne $outboxBefore.total) {
  throw 'Illegal mail send changed outbox count.'
}
if ($inboxAfterIllegal.total -ne $inboxBefore.total) {
  throw 'Illegal mail send changed inbox count.'
}
$sentMail = Post-Json "$BaseUrl/api/mail/send" @{
  toName = $me.user.name
  title = '__smoke_mail__'
  content = 'smoke-mail-content'
}
$outboxAfterSend = Get-Json "$BaseUrl/api/mail?folder=outbox&page=0"
$inboxAfterSend = Get-Json "$BaseUrl/api/mail?folder=inbox&page=0"
$mailDeleteOutbox = Post-Json "$BaseUrl/api/mail/delete" @{
  folder = 'outbox'
  ids = @($sentMail.summary.id)
  page = 0
}
$selfInboxCopy = $inboxAfterSend.items | Where-Object { $_.id -eq $sentMail.summary.id } | Select-Object -First 1
$mailDeleteInbox = $null
if ($null -ne $selfInboxCopy) {
  $mailDeleteInbox = Post-Json "$BaseUrl/api/mail/delete" @{
    folder = 'inbox'
    ids = @($selfInboxCopy.id)
    page = 0
  }
}
$outboxAfterDelete = Get-Json "$BaseUrl/api/mail?folder=outbox&page=0"
$inboxAfterDelete = Get-Json "$BaseUrl/api/mail?folder=inbox&page=0"

Post-Json "$BaseUrl/api/auth/logout" $null | Out-Null
$afterLogout = Get-Json "$BaseUrl/api/auth/me"

$myCitiesAfterLogoutStatus = try {
  Invoke-WebRequest -UseBasicParsing -WebSession $session "$BaseUrl/api/me/cities?limit=1" | Out-Null
  200
} catch {
  [int]$_.Exception.Response.StatusCode
}

$root = Invoke-WebRequest -UseBasicParsing "$BaseUrl/"

[pscustomobject]@{
  commander_uid = $commander.uid
  commander_name = $commander.name
  session_uid = $me.user.uid
  tested_city_cid = $city.cid
  tested_city_name = $city.name
  tax_roundtrip = $taxPatch.tax
  production_roundtrip = [pscustomobject]@{
    foodRate = $productionPatch.production.settings.foodRate
    woodRate = $productionPatch.production.settings.woodRate
    rockRate = $productionPatch.production.settings.rockRate
    ironRate = $productionPatch.production.settings.ironRate
  }
  mail_send_id = $sentMail.summary.id
  mail_send_title = $sentMail.summary.title
  mail_send_to = $sentMail.summary.toName
  mail_send_has_html = -not [string]::IsNullOrWhiteSpace($sentMail.htmlDocument)
  illegal_mail_status = $illegalMailStatus
  illegal_mail_message = $illegalMailText
  outbox_total_before = $outboxBefore.total
  outbox_total_after_illegal = $outboxAfterIllegal.total
  outbox_total_after_send = $outboxAfterSend.total
  outbox_total_after_delete = $outboxAfterDelete.total
  inbox_total_before = $inboxBefore.total
  inbox_total_after_illegal = $inboxAfterIllegal.total
  inbox_total_after_send = $inboxAfterSend.total
  inbox_total_after_delete = $inboxAfterDelete.total
  self_mail_deleted_from_inbox = $null -ne $mailDeleteInbox
  me_after_logout_is_null = $null -eq $afterLogout.user
  my_cities_status_after_logout = $myCitiesAfterLogoutStatus
  root_status = [int]$root.StatusCode
  root_title = if ($root.Content -match '<title>(.*?)</title>') { $matches[1] } else { '' }
  root_references_built_assets = $root.Content -match '/assets/'
  login_user_name = $login.user.name
} | ConvertTo-Json -Depth 6
