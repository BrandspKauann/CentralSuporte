[CmdletBinding()]
param(
    [ValidateSet('GET', 'POST', 'PUT', 'PATCH')][string]$Method = 'GET',
    [Parameter(Mandatory)][string]$Path,
    [string]$BodyFile
)
$ErrorActionPreference = 'Stop'
$baseUrl = 'https://hirayama-suporte.vercel.app/api/v2/'
if ($Path -notmatch '^[a-zA-Z0-9]' -or $Path -match '[:\\#\s]' -or $Path.Contains('..')) {
    throw 'Informe somente um caminho relativo da API.'
}
$secretFile = Join-Path $env:LOCALAPPDATA 'HirayamaSuporte\api-token.dpapi'
if (!(Test-Path -LiteralPath $secretFile)) {
    throw 'Conecte primeiro com tools/hirayama/Connect-Hirayama.ps1. Nenhuma tarefa foi gravada.'
}
$secureToken = Get-Content -LiteralPath $secretFile -Raw | ConvertTo-SecureString
$plainToken = [System.Net.NetworkCredential]::new('', $secureToken).Password
try {
    $request = @{
        Uri = $baseUrl + $Path
        Method = $Method
        Headers = @{ Authorization = 'Bearer ' + $plainToken }
        MaximumRedirection = 0
        ErrorAction = 'Stop'
    }
    if ($BodyFile) {
        $body = Get-Content -LiteralPath $BodyFile -Raw -Encoding UTF8
        $null = $body | ConvertFrom-Json
        $request.Body = [System.Text.Encoding]::UTF8.GetBytes($body)
        $request.ContentType = 'application/json; charset=utf-8'
    }
    Invoke-RestMethod @request
} catch {
    $status = $_.Exception.Response.StatusCode
    throw "Falha na API Hirayama (HTTP $([int]$status)). A gravação precisa ser conferida antes de repetir."
} finally {
    $plainToken = $null
    $request = $null
    $secureToken.Dispose()
}
