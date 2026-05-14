# 版本检查脚本
# 检查 git tag 的最新版本与 version.json 中的版本是否一致

$ErrorActionPreference = "Stop"

# 获取脚本所在目录
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectDir = Split-Path -Parent $scriptDir

# 读取 version.json
$versionFile = Join-Path $projectDir "version.json"
if (-not (Test-Path $versionFile)) {
    Write-Error "version.json not found at $versionFile"
    exit 1
}

$versionContent = Get-Content $versionFile -Raw | ConvertFrom-Json
$jsonVersion = $versionContent.version
$suffix = $versionContent.suffix

# 构建完整版本号
if ($suffix -and $suffix -ne "") {
    $fullVersion = "$jsonVersion-$suffix"
}
else {
    $fullVersion = $jsonVersion
}

Write-Host "version.json version: $fullVersion"

# 获取最新的 git tag
try {
    $gitTag = git -C $projectDir describe --tags --abbrev=0 2>$null
}
catch {
    $gitTag = $null
}

if (-not $gitTag) {
    Write-Error "No git tag found"
    exit 1
}

Write-Host "git tag: $gitTag"

# 移除 git tag 的 v 前缀（如果有）
$normalizedGitTag = $gitTag -replace "^v", ""

# 比较版本
if ($normalizedGitTag -eq $jsonVersion) {
    Write-Host "Version check passed: git tag ($gitTag) matches version.json ($jsonVersion)" -ForegroundColor Green
    exit 0
}
else {
    Write-Error "Version mismatch! Git tag: $gitTag (normalized: $normalizedGitTag), version.json: $jsonVersion"
    exit 1
}
