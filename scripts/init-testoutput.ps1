# 初始化测试输出目录依赖脚本
# 为 testoutput 下的 logger 和 metrics 目录初始化 go.mod 并下载依赖

$ErrorActionPreference = "Stop"

# 获取脚本所在目录
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectDir = Split-Path -Parent $scriptDir

# 定义测试输出目录
$testDirs = @(
    Join-Path $projectDir "testoutput/logger"
    Join-Path $projectDir "testoutput/metrics"
)

foreach ($dir in $testDirs) {
    if (-not (Test-Path $dir)) {
        Write-Host "Creating directory: $dir"
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
    }

    $goModFile = Join-Path $dir "go.mod"
    
    # 获取目录名作为模块名
    $dirName = Split-Path -Leaf $dir
    $moduleName = "testoutput/$dirName"
    
    # 创建 go.mod 文件内容（无 BOM UTF-8）
    $goModContent = @"
module $moduleName

go 1.25

require github.com/zeromicro/go-zero v1.8.2
"@
    
    Write-Host "Setting up $dirName..."
    [System.IO.File]::WriteAllText($goModFile, $goModContent, [System.Text.UTF8Encoding]::new($false))
    
    # 进入目录并下载依赖
    Push-Location $dir
    try {
        Write-Host "  Downloading dependencies for $dirName..."
        $output = go mod tidy 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  OK Dependencies ready for $dirName" -ForegroundColor Green
        } else {
            Write-Warning "  Failed to tidy modules for $dirName, but go.mod is created"
            Write-Warning "  Error: $output"
        }
    }
    finally {
        Pop-Location
    }
}

Write-Host ""
Write-Host "All test output directories are ready!" -ForegroundColor Green
