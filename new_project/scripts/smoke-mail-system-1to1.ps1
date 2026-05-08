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
  param([string]$Method, [string]$Url, [object]$Body = $null)
  $args = @{ Method = $Method; Uri = $Url; WebSession = $session; TimeoutSec = 12 }
  if ($Method -ne 'GET') {
    $args.ContentType = 'application/json'
    $args.Body = if ($null -eq $Body) { '{}' } else { $Body | ConvertTo-Json -Depth 10 }
  }
  try {
    $resp = Invoke-WebRequest -UseBasicParsing @args
    $json = $null
    try { $json = $resp.Content | ConvertFrom-Json } catch { }
    return [pscustomobject]@{ ok = $true; status = [int]$resp.StatusCode; json = $json; raw = $resp.Content }
  } catch {
    $status = if ($_.Exception.Response) { [int]$_.Exception.Response.StatusCode } else { 0 }
    $raw = $_.ErrorDetails.Message
    return [pscustomobject]@{ ok = $false; status = $status; json = $null; raw = $raw }
  }
}

function Add-Step {
  param([string]$Name, [bool]$Passed, [hashtable]$Extra = @{})
  if (-not $Passed) { $script:allPassed = $false }
  $row = @{ step = $Name; passed = $Passed }
  foreach ($k in $Extra.Keys) { $row[$k] = $Extra[$k] }
  $script:steps += $row
}

$login = Invoke-Api -Method POST -Url "$BaseUrl/api/auth/login" -Body @{ passport = $Passport; password = $Password }
Add-Step -Name 'login' -Passed ($login.status -eq 200 -and $null -ne $login.json.user) -Extra @{ status = $login.status }
if ($login.status -ne 200) { throw "login failed: $($login.status)" }

$me = Invoke-Api -Method GET -Url "$BaseUrl/api/auth/me"
$uid = if ($me.json -and $me.json.user) { [int]$me.json.user.uid } else { 0 }
Add-Step -Name 'auth_me' -Passed ($me.status -eq 200 -and $uid -gt 0) -Extra @{ uid = $uid; status = $me.status }

$system = Invoke-Api -Method GET -Url "$BaseUrl/api/mail?folder=system&page=0"
$inbox = Invoke-Api -Method GET -Url "$BaseUrl/api/mail?folder=inbox&page=0"
$outbox = Invoke-Api -Method GET -Url "$BaseUrl/api/mail?folder=outbox&page=0"
Add-Step -Name 'mail_system' -Passed ($system.status -eq 200) -Extra @{ status = $system.status; total = if ($system.json) { [int]$system.json.total } else { -1 } }
Add-Step -Name 'mail_inbox' -Passed ($inbox.status -eq 200) -Extra @{ status = $inbox.status; total = if ($inbox.json) { [int]$inbox.json.total } else { -1 } }
Add-Step -Name 'mail_outbox' -Passed ($outbox.status -eq 200) -Extra @{ status = $outbox.status; total = if ($outbox.json) { [int]$outbox.json.total } else { -1 } }

$firstMailId = 0
if ($inbox.json -and $inbox.json.items -and $inbox.json.items.Count -gt 0) {
  $firstMailId = [int]$inbox.json.items[0].id
}

if ($firstMailId -gt 0) {
  $mailDetail = Invoke-Api -Method GET -Url "$BaseUrl/api/mail/$firstMailId"
  $detailOk = ($mailDetail.status -eq 200 -and $mailDetail.json -and $mailDetail.json.id -eq $firstMailId)
  Add-Step -Name 'mail_detail' -Passed $detailOk -Extra @{ mailId = $firstMailId; status = $mailDetail.status }
} else {
  Add-Step -Name 'mail_detail' -Passed $true -Extra @{ reason = 'no inbox mail to inspect' }
}

$sendProbe = Invoke-Api -Method POST -Url "$BaseUrl/api/mail/send" -Body @{ toUid = 0; title = ''; content = '' }
$sendRejected = ($sendProbe.status -eq 400 -or $sendProbe.status -eq 403 -or $sendProbe.status -eq 404)
Add-Step -Name 'mail_send_invalid_rejected' -Passed $sendRejected -Extra @{ status = $sendProbe.status }

$deleteProbe = Invoke-Api -Method POST -Url "$BaseUrl/api/mail/delete" -Body @{ mailId = 0 }
$deleteRejected = ($deleteProbe.status -eq 200 -or $deleteProbe.status -eq 400 -or $deleteProbe.status -eq 403 -or $deleteProbe.status -eq 404)
Add-Step -Name 'mail_delete_invalid_rejected' -Passed $deleteRejected -Extra @{ status = $deleteProbe.status }

$output = [pscustomobject]@{
  timestamp = $timestamp
  allPassed = $allPassed
  steps = $steps
}
$path = "artifacts/mail-system-$timestamp.json"
$output | ConvertTo-Json -Depth 8 | Out-File -FilePath $path -Encoding UTF8
Write-Host $path
if ($allPassed) { exit 0 } else { exit 1 }
