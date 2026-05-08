param(
  [string]$BaseUrl = 'http://127.0.0.1:8080',
  [string]$Passport = 'test',
  [string]$Password = 'test'
)

$ErrorActionPreference = 'Stop'
$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$null = New-Item -ItemType Directory -Force -Path 'artifacts' | Out-Null

$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$script:allPassed = $true
$script:checks = @()

function Add-Check {
  param(
    [string]$Name,
    [bool]$Passed,
    [string]$Detail
  )
  $script:checks += [pscustomobject]@{
    name = $Name
    passed = $Passed
    detail = $Detail
  }
  if ($Passed) {
    Write-Host "[PASS] $Name - $Detail" -ForegroundColor Green
  } else {
    Write-Host "[FAIL] $Name - $Detail" -ForegroundColor Red
    $script:allPassed = $false
  }
}

function Invoke-Api {
  param(
    [ValidateSet('GET', 'POST', 'PATCH')]
    [string]$Method,
    [string]$Url,
    [object]$Body = $null
  )

  $args = @{
    Method = $Method
    Uri = $Url
    WebSession = $session
    TimeoutSec = 20
  }

  if ($Method -ne 'GET') {
    $args.ContentType = 'application/json'
    $args.Body = if ($null -eq $Body) { '{}' } else { $Body | ConvertTo-Json -Depth 12 }
  }

  try {
    $resp = Invoke-WebRequest -UseBasicParsing @args
    $json = $null
    try { $json = $resp.Content | ConvertFrom-Json } catch {}
    return [pscustomobject]@{
      status = [int]$resp.StatusCode
      json = $json
      raw = $resp.Content
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
    try { $json = $raw | ConvertFrom-Json } catch {}
    return [pscustomobject]@{
      status = $status
      json = $json
      raw = $raw
    }
  }
}

function Get-FirstNormalPricedItem {
  param([object]$ShopSnapshot)
  foreach ($group in $ShopSnapshot.groups) {
    foreach ($item in $group.items) {
      if (-not [bool]$item.battleShop -and [int64]$item.price -gt 0) {
        return $item
      }
    }
  }
  return $null
}

Write-Host "=== Shop/Payment 1:1 Smoke ===" -ForegroundColor Cyan

$login = Invoke-Api -Method POST -Url "$BaseUrl/api/auth/login" -Body @{
  passport = $Passport
  password = $Password
}
Add-Check -Name 'login' -Passed ($login.status -eq 200 -and $null -ne $login.json.user) -Detail "status=$($login.status)"
if ($login.status -ne 200 -or $null -eq $login.json.user) {
  throw "login failed"
}

$me = Invoke-Api -Method GET -Url "$BaseUrl/api/auth/me"
Add-Check -Name 'auth me' -Passed ($me.status -eq 200 -and $null -ne $me.json.user) -Detail "status=$($me.status)"
if ($me.status -ne 200 -or $null -eq $me.json.user) {
  throw "auth me failed"
}
$cityId = [int]$me.json.user.defaultCid

$shopBefore = Invoke-Api -Method GET -Url "$BaseUrl/api/me/shop"
$shopGroups = if ($null -ne $shopBefore.json -and $null -ne $shopBefore.json.groups) { [int]$shopBefore.json.groups.Count } else { 0 }
Add-Check -Name 'shop list' -Passed ($shopBefore.status -eq 200 -and $shopGroups -gt 0) -Detail "status=$($shopBefore.status),groups=$shopGroups"
if ($shopBefore.status -ne 200 -or $shopGroups -le 0) {
  throw "shop list failed"
}

$buyItem = Get-FirstNormalPricedItem -ShopSnapshot $shopBefore.json
Add-Check -Name 'buy candidate' -Passed ($null -ne $buyItem) -Detail ($(if ($null -ne $buyItem) { "itemId=$($buyItem.id),price=$($buyItem.price)" } else { "none" }))
if ($null -eq $buyItem) {
  throw "no buy candidate"
}

$yuanbaoBeforeBuy = [int64]$shopBefore.json.wallet.yuanbao
$buyOne = Invoke-Api -Method POST -Url "$BaseUrl/api/me/shop/buy" -Body @{
  itemId = [int]$buyItem.id
  count = 1
  payType = 0
  cityId = $cityId
}
Add-Check -Name 'shop buy status' -Passed ($buyOne.status -eq 200 -and $null -ne $buyOne.json.wallet) -Detail "status=$($buyOne.status)"
$yuanbaoAfterBuy = if ($null -ne $buyOne.json -and $null -ne $buyOne.json.wallet) { [int64]$buyOne.json.wallet.yuanbao } else { $yuanbaoBeforeBuy }
$expectedAfterBuy = $yuanbaoBeforeBuy - [int64]$buyItem.price
Add-Check -Name 'shop buy deduct' -Passed ($buyOne.status -eq 200 -and $yuanbaoAfterBuy -eq $expectedAfterBuy) -Detail "before=$yuanbaoBeforeBuy,after=$yuanbaoAfterBuy,price=$($buyItem.price)"

$countForInsufficient = [int]([Math]::Floor(($yuanbaoAfterBuy / [double][int64]$buyItem.price) + 1))
if ($countForInsufficient -lt 1) { $countForInsufficient = 1 }
$buyInsufficient = Invoke-Api -Method POST -Url "$BaseUrl/api/me/shop/buy" -Body @{
  itemId = [int]$buyItem.id
  count = $countForInsufficient
  payType = 0
  cityId = $cityId
}
$insufficientMsg = ''
if ($null -ne $buyInsufficient.json -and $null -ne $buyInsufficient.json.message) {
  $insufficientMsg = [string]$buyInsufficient.json.message
}
$insufficientByStatus = ($buyInsufficient.status -eq 400)
$insufficientByText = (-not [string]::IsNullOrWhiteSpace($insufficientMsg))
Add-Check -Name 'insufficient reject' -Passed ($insufficientByStatus -and $insufficientByText) -Detail "status=$($buyInsufficient.status),message=$insufficientMsg"

$chargeBefore = Invoke-Api -Method GET -Url "$BaseUrl/api/me/charge"
Add-Check -Name 'charge read before' -Passed ($chargeBefore.status -eq 200 -and $null -ne $chargeBefore.json.summary) -Detail "status=$($chargeBefore.status)"
if ($chargeBefore.status -ne 200 -or $null -eq $chargeBefore.json.summary) {
  throw "charge read before failed"
}

$exchangeCount = 1
$rate = [int64]$chargeBefore.json.summary.exchangeRate
$yuanbaoBeforeExchange = [int64]$chargeBefore.json.summary.yuanbao
$exchange = Invoke-Api -Method POST -Url "$BaseUrl/api/me/charge/exchange" -Body @{
  exchangeCount = $exchangeCount
}
Add-Check -Name 'charge exchange status' -Passed ($exchange.status -eq 200 -and $null -ne $exchange.json.summary) -Detail "status=$($exchange.status)"
$yuanbaoAfterExchange = if ($null -ne $exchange.json -and $null -ne $exchange.json.summary) { [int64]$exchange.json.summary.yuanbao } else { $yuanbaoBeforeExchange }
$expectedAfterExchange = $yuanbaoBeforeExchange + ([int64]$exchangeCount * $rate)
Add-Check -Name 'charge arrived' -Passed ($exchange.status -eq 200 -and $yuanbaoAfterExchange -eq $expectedAfterExchange) -Detail "before=$yuanbaoBeforeExchange,after=$yuanbaoAfterExchange,rate=$rate,count=$exchangeCount"

$result = [pscustomobject]@{
  timestamp = $timestamp
  allPassed = $script:allPassed
  baseUrl = $BaseUrl
  passport = $Passport
  uid = [int]$me.json.user.uid
  cityId = $cityId
  checks = $script:checks
  evidence = [pscustomobject]@{
    shop = [pscustomobject]@{
      groups = $shopGroups
      buyItemId = [int]$buyItem.id
      buyPrice = [int64]$buyItem.price
      yuanbaoBeforeBuy = $yuanbaoBeforeBuy
      yuanbaoAfterBuy = $yuanbaoAfterBuy
      insufficientStatus = [int]$buyInsufficient.status
      insufficientMessage = $insufficientMsg
      insufficientCountTried = $countForInsufficient
    }
    charge = [pscustomobject]@{
      exchangeCount = $exchangeCount
      exchangeRate = $rate
      yuanbaoBeforeExchange = $yuanbaoBeforeExchange
      yuanbaoAfterExchange = $yuanbaoAfterExchange
    }
  }
}

$outputFile = "artifacts/shop-payment-$timestamp.json"
$result | ConvertTo-Json -Depth 12 | Out-File -FilePath $outputFile -Encoding UTF8

Write-Host "artifact: $outputFile" -ForegroundColor Cyan
if ($script:allPassed) {
  exit 0
}
exit 1
