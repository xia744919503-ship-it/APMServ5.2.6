param(
  [string]$BaseUrl = 'http://127.0.0.1:8080',
  [string]$Passport = 'test',
  [string]$Password = 'test'
)

$ErrorActionPreference = 'Stop'
$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession

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
    TimeoutSec = 10
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
    } catch {
    }
    return [pscustomobject]@{
      Status = [int]$resp.StatusCode
      Json   = $json
      Raw    = $resp.Content
      Error  = ''
    }
  } catch {
    $status = 0
    if ($_.Exception.Response) {
      $status = [int]$_.Exception.Response.StatusCode
    }
    $raw = $_.ErrorDetails.Message
    if ([string]::IsNullOrWhiteSpace($raw)) {
      $raw = $_.Exception.Message
    }
    $json = $null
    try {
      $json = $raw | ConvertFrom-Json
    } catch {
    }
    return [pscustomobject]@{
      Status = $status
      Json   = $json
      Raw    = $raw
      Error  = $_.Exception.Message
    }
  }
}

function Assert-True {
  param(
    [bool]$Condition,
    [string]$Code
  )

  if (-not $Condition) {
    throw $Code
  }
}

$login = Invoke-Api -Method POST -Url "$BaseUrl/api/auth/login" -Body @{
  passport = $Passport
  password = $Password
}
Assert-True ($login.Status -eq 200) 'login_status_not_200'
Assert-True ($null -ne $login.Json.user) 'login_user_missing'

$me = Invoke-Api -Method GET -Url "$BaseUrl/api/auth/me"
Assert-True ($me.Status -eq 200) 'auth_me_status_not_200'
Assert-True ($null -ne $me.Json.user) 'auth_me_user_missing'

$tasks = Invoke-Api -Method GET -Url "$BaseUrl/api/me/tasks"
Assert-True ($tasks.Status -eq 200) 'tasks_status_not_200'
$firstTask = $null
if ($tasks.Json.categories.Count -gt 0 -and $tasks.Json.categories[0].groups.Count -gt 0 -and $tasks.Json.categories[0].groups[0].tasks.Count -gt 0) {
  $firstTask = $tasks.Json.categories[0].groups[0].tasks[0]
}

$taskClaim = $null
if ($null -ne $firstTask) {
  $taskClaim = Invoke-Api -Method POST -Url "$BaseUrl/api/me/tasks/claim" -Body @{
    taskId = [int]$firstTask.id
  }
  $claimOk = ($taskClaim.Status -eq 200 -and $null -ne $taskClaim.Json.summary) -or (($taskClaim.Status -eq 400 -or $taskClaim.Status -eq 403) -and $null -ne $taskClaim.Json.message)
  Assert-True $claimOk 'task_claim_semantic_mismatch'
}

$union = Invoke-Api -Method GET -Url "$BaseUrl/api/me/union"
Assert-True ($union.Status -eq 200) 'union_status_not_200'
Assert-True ($null -ne $union.Json.permissions) 'union_permissions_missing'

$unionCreateDenied = $null
if (-not [bool]$union.Json.permissions.canCreate) {
  $unionCreateDenied = Invoke-Api -Method POST -Url "$BaseUrl/api/me/union/create" -Body @{
    name = 'smokeu'
  }
  Assert-True (($unionCreateDenied.Status -eq 400 -or $unionCreateDenied.Status -eq 403) -and $null -ne $unionCreateDenied.Json.message) 'union_create_denied_semantic_mismatch'
}

$shop = Invoke-Api -Method GET -Url "$BaseUrl/api/me/shop"
Assert-True ($shop.Status -eq 200) 'shop_status_not_200'
Assert-True ($null -ne $shop.Json.wallet) 'shop_wallet_missing'

$shopBuyInvalid = Invoke-Api -Method POST -Url "$BaseUrl/api/me/shop/buy" -Body @{
  itemId  = 9
  count   = 0
  payType = 0
  cityId  = [int]$me.Json.user.defaultCid
}
Assert-True ($shopBuyInvalid.Status -eq 400 -and $null -ne $shopBuyInvalid.Json.message) 'shop_buy_invalid_semantic_mismatch'

$charge = Invoke-Api -Method GET -Url "$BaseUrl/api/me/charge"
Assert-True ($charge.Status -eq 200) 'charge_status_not_200'
Assert-True ($null -ne $charge.Json.summary) 'charge_summary_missing'

$chargeExchangeInvalid = Invoke-Api -Method POST -Url "$BaseUrl/api/me/charge/exchange" -Body @{
  exchangeCount = 0
}
Assert-True ($chargeExchangeInvalid.Status -eq 400 -and $null -ne $chargeExchangeInvalid.Json.message) 'charge_exchange_invalid_semantic_mismatch'

$mailSystem = Invoke-Api -Method GET -Url "$BaseUrl/api/mail?folder=system&page=0"
$mailInbox = Invoke-Api -Method GET -Url "$BaseUrl/api/mail?folder=inbox&page=0"
$mailOutbox = Invoke-Api -Method GET -Url "$BaseUrl/api/mail?folder=outbox&page=0"
Assert-True ($mailSystem.Status -eq 200) 'mail_system_status_not_200'
Assert-True ($mailInbox.Status -eq 200) 'mail_inbox_status_not_200'
Assert-True ($mailOutbox.Status -eq 200) 'mail_outbox_status_not_200'

$reportsUnread = Invoke-Api -Method GET -Url "$BaseUrl/api/reports?filter=unread&page=0"
$reportsType0 = Invoke-Api -Method GET -Url "$BaseUrl/api/reports?filter=type0&page=0"
Assert-True ($reportsUnread.Status -eq 200) 'reports_unread_status_not_200'
Assert-True ($reportsType0.Status -eq 200) 'reports_type0_status_not_200'

$rankingKinds = @(
  'user',
  'union',
  'hero_level',
  'hero_affairs',
  'hero_bravery',
  'hero_wisdom',
  'city_people',
  'city_type',
  'jungong',
  'juanxian',
  'qinwang',
  'gongpin',
  'jungong_union',
  'juanxian_union',
  'qinwang_union',
  'gongpin_union',
  'military',
  'military_attack',
  'military_defence',
  'rich',
  'rich_day',
  'rich_month',
  'battle_total',
  'battle_week',
  'battle_day'
)

$rankingResults = @()
foreach ($kind in $rankingKinds) {
  $rankResp = Invoke-Api -Method GET -Url "$BaseUrl/api/rankings?kind=$kind&page=0"
  Assert-True ($rankResp.Status -eq 200) "ranking_status_not_200:$kind"
  Assert-True ($rankResp.Json.kind -eq $kind) "ranking_kind_mismatch:$kind"
  $rankingResults += [pscustomobject]@{
    kind    = $kind
    total   = [int]$rankResp.Json.total
    columns = [int]$rankResp.Json.columns.Count
  }
}

[pscustomobject]@{
  ok          = $true
  passport    = $Passport
  uid         = [int]$me.Json.user.uid
  default_cid = [int]$me.Json.user.defaultCid
  tasks       = [pscustomobject]@{
    total         = [int]$tasks.Json.summary.taskCount
    first_task_id = if ($null -ne $firstTask) { [int]$firstTask.id } else { 0 }
    claim_status  = if ($null -ne $taskClaim) { [int]$taskClaim.Status } else { 0 }
    claim_message = if ($null -ne $taskClaim -and $null -ne $taskClaim.Json.message) { [string]$taskClaim.Json.message } else { '' }
  }
  union       = [pscustomobject]@{
    canCreate             = [bool]$union.Json.permissions.canCreate
    canApply              = [bool]$union.Json.permissions.canApply
    denied_create_status  = if ($null -ne $unionCreateDenied) { [int]$unionCreateDenied.Status } else { 0 }
    denied_create_message = if ($null -ne $unionCreateDenied -and $null -ne $unionCreateDenied.Json.message) { [string]$unionCreateDenied.Json.message } else { '' }
  }
  shop        = [pscustomobject]@{
    yuanbao             = [int64]$shop.Json.wallet.yuanbao
    groups              = [int]$shop.Json.groups.Count
    buy_invalid_status  = [int]$shopBuyInvalid.Status
    buy_invalid_message = [string]$shopBuyInvalid.Json.message
  }
  charge      = [pscustomobject]@{
    yuanbao                  = [int64]$charge.Json.summary.yuanbao
    exchange_invalid_status  = [int]$chargeExchangeInvalid.Status
    exchange_invalid_message = [string]$chargeExchangeInvalid.Json.message
  }
  mail        = [pscustomobject]@{
    system_total = [int]$mailSystem.Json.total
    inbox_total  = [int]$mailInbox.Json.total
    outbox_total = [int]$mailOutbox.Json.total
  }
  reports     = [pscustomobject]@{
    unread_total = [int]$reportsUnread.Json.total
    type0_total  = [int]$reportsType0.Json.total
  }
  rankings    = $rankingResults
} | ConvertTo-Json -Depth 8
