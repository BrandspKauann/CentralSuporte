[CmdletBinding()]
param()
$ErrorActionPreference = 'Stop'
if ($env:OS -ne 'Windows_NT') { throw 'Este armazenamento usa a proteção da conta Windows.' }
$baseUrl = 'https://hirayama-suporte.vercel.app/api/v2/'
$secretDirectory = Join-Path $env:LOCALAPPDATA 'HirayamaSuporte'
$secretFile = Join-Path $secretDirectory 'api-token.dpapi'
$secureToken = Read-Host 'Cole seu token de API Hirayama (entrada oculta)' -AsSecureString
$plainToken = [System.Net.NetworkCredential]::new('', $secureToken).Password
try {
    # Validate before saving. Never print a token or an authentication response.
    $user = Invoke-RestMethod -Uri ($baseUrl + 'user') -Headers @{ Authorization = 'Bearer ' + $plainToken } -MaximumRedirection 0
    if (!$user.id -or !$user.username) { throw 'A API não confirmou a conta autenticada.' }
    $null = New-Item -ItemType Directory -Path $secretDirectory -Force
    $secureToken | ConvertFrom-SecureString | Set-Content -LiteralPath $secretFile -Encoding UTF8
    Write-Output "API conectada à conta $($user.username), ID $($user.id). Token protegido nesta conta Windows."
} catch {
    throw 'Não consegui validar e salvar a conexão. Confira o token e a permissão de leitura do usuário.'
} finally {
    $plainToken = $null
    $secureToken.Dispose()
}
