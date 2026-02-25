# ============================================================
# patch-phase2.ps1 - Vikunja Custom Build: Full Patch + Build
# Run from: your vikunja source directory
# Usage: .\vikunja-phase2\patch-phase2.ps1              (build only)
#        .\vikunja-phase2\patch-phase2.ps1 -Deploy      (build + deploy)
# ============================================================

param(
    [switch]$Deploy
)

$ErrorActionPreference = "Stop"
$ROOT = $PSScriptRoot | Split-Path -Parent
$PATCH = Split-Path -Parent $MyInvocation.MyCommand.Path

$stepTotal = 20
if ($Deploy) { $stepTotal = 21 }
$step = 0

function Step($msg) {
    $script:step++
    Write-Host "[$step/$stepTotal] $msg" -ForegroundColor Green
}

Write-Host ""
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  Vikunja Custom Build - Phase 2 Full Patch" -ForegroundColor Cyan
$ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
Write-Host "  $ts" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan

# --- Preflight ---
if (-not (Test-Path "$ROOT\frontend\src")) {
    Write-Host "[!] Source not found at $ROOT" -ForegroundColor Red
    exit 1
}

# --- Ensure directories exist ---
$dirs = @(
    "$ROOT\pkg\models",
    "$ROOT\pkg\routes\api\v1",
    "$ROOT\pkg\migration",
    "$ROOT\pkg\routes",
    "$ROOT\frontend\src\components\gantt",
    "$ROOT\frontend\src\components\tasks\partials",
    "$ROOT\frontend\src\components\project\partials",
    "$ROOT\frontend\src\views\templates",
    "$ROOT\frontend\src\views\tasks",
    "$ROOT\frontend\src\views\labels",
    "$ROOT\frontend\src\views\teams",
    "$ROOT\frontend\src\views\project",
    "$ROOT\frontend\src\views\project\helpers",
    "$ROOT\frontend\src\services",
    "$ROOT\frontend\src\stores",
    "$ROOT\frontend\src\composables",
    "$ROOT\frontend\src\i18n\lang",
    "$ROOT\docs"
)
foreach ($d in $dirs) {
    New-Item -ItemType Directory -Path $d -Force | Out-Null
}

# ===========================
#  BACKEND - Go Models
# ===========================
Step "Backend: chain model + task creation"
Copy-Item "$PATCH\task_chain.go"      "$ROOT\pkg\models\task_chain.go" -Force
Copy-Item "$PATCH\task_from_chain.go" "$ROOT\pkg\models\task_from_chain.go" -Force

Step "Backend: auto-task model + creation logic"
Copy-Item "$PATCH\auto_task_template.go" "$ROOT\pkg\models\auto_task_template.go" -Force
Copy-Item "$PATCH\auto_task_create.go"   "$ROOT\pkg\models\auto_task_create.go" -Force

Step "Backend: core task model (done_by_id tracking)"
Copy-Item "$PATCH\tasks.go" "$ROOT\pkg\models\tasks.go" -Force

# ===========================
#  BACKEND - Handlers (echo v5)
# ===========================
Step "Backend: chain step attachment handler"
Copy-Item "$PATCH\chain_step_attachment.go" "$ROOT\pkg\routes\api\v1\chain_step_attachment.go" -Force

Step "Backend: auto-task trigger + check handler"
Copy-Item "$PATCH\auto_task_handler.go" "$ROOT\pkg\routes\api\v1\auto_task_handler.go" -Force

# ===========================
#  BACKEND - Migrations
# ===========================
Step "Migrations (5 files)"
Copy-Item "$PATCH\20260224050000.go" "$ROOT\pkg\migration\20260224050000.go" -Force
Copy-Item "$PATCH\20260224060000.go" "$ROOT\pkg\migration\20260224060000.go" -Force
Copy-Item "$PATCH\20260224070000.go" "$ROOT\pkg\migration\20260224070000.go" -Force
Copy-Item "$PATCH\20260224080000.go" "$ROOT\pkg\migration\20260224080000.go" -Force
Copy-Item "$PATCH\20260224090000.go" "$ROOT\pkg\migration\20260224090000.go" -Force

# ===========================
#  BACKEND - Routes
# ===========================
Step "Routes: all API endpoints"
Copy-Item "$PATCH\routes.go" "$ROOT\pkg\routes\routes.go" -Force

# ===========================
#  FRONTEND - Gantt
# ===========================
Step "Gantt: arrows + settings + tooltips + grid lines + header"
Copy-Item "$PATCH\GanttDependencyArrows.vue"  "$ROOT\frontend\src\components\gantt\GanttDependencyArrows.vue" -Force
Copy-Item "$PATCH\GanttArrowSettings.vue"     "$ROOT\frontend\src\components\gantt\GanttArrowSettings.vue" -Force
Copy-Item "$PATCH\GanttChart.vue"             "$ROOT\frontend\src\components\gantt\GanttChart.vue" -Force
Copy-Item "$PATCH\GanttRowBars.vue"           "$ROOT\frontend\src\components\gantt\GanttRowBars.vue" -Force
Copy-Item "$PATCH\GanttVerticalGridLines.vue" "$ROOT\frontend\src\components\gantt\GanttVerticalGridLines.vue" -Force
Copy-Item "$PATCH\GanttTimelineHeader.vue"    "$ROOT\frontend\src\components\gantt\GanttTimelineHeader.vue" -Force
Copy-Item "$PATCH\useGanttBar.ts"             "$ROOT\frontend\src\composables\useGanttBar.ts" -Force
Copy-Item "$PATCH\useGanttArrowConfig.ts"     "$ROOT\frontend\src\composables\useGanttArrowConfig.ts" -Force
Copy-Item "$PATCH\useGanttFilters.ts"         "$ROOT\frontend\src\views\project\helpers\useGanttFilters.ts" -Force
Copy-Item "$PATCH\ProjectGantt.vue"           "$ROOT\frontend\src\components\project\views\ProjectGantt.vue" -Force

# ===========================
#  FRONTEND - Chain System
# ===========================
Step "Chain editor + API + create-from-chain modal"
Copy-Item "$PATCH\taskChainApi.ts"          "$ROOT\frontend\src\services\taskChainApi.ts" -Force
Copy-Item "$PATCH\ChainEditor.vue"          "$ROOT\frontend\src\components\tasks\partials\ChainEditor.vue" -Force
Copy-Item "$PATCH\CreateFromChainModal.vue" "$ROOT\frontend\src\components\tasks\partials\CreateFromChainModal.vue" -Force

# ===========================
#  FRONTEND - Auto-Generated Tasks
# ===========================
Step "Auto-task API + editor (rich text, labels, log modal)"
Copy-Item "$PATCH\autoTaskApi.ts"     "$ROOT\frontend\src\services\autoTaskApi.ts" -Force
Copy-Item "$PATCH\AutoTaskEditor.vue" "$ROOT\frontend\src\components\tasks\partials\AutoTaskEditor.vue" -Force

# ===========================
#  FRONTEND - Stores, Composables, Helpers
# ===========================
Step "Stores + composables + helpers"
Copy-Item "$PATCH\tasks.ts"            "$ROOT\frontend\src\stores\tasks.ts" -Force
Copy-Item "$PATCH\useDragReorder.ts"   "$ROOT\frontend\src\composables\useDragReorder.ts" -Force
Copy-Item "$PATCH\useGanttTaskList.ts" "$ROOT\frontend\src\views\project\helpers\useGanttTaskList.ts" -Force

# ===========================
#  FRONTEND - Views
# ===========================
Step "Template manager page (3 tabs)"
Copy-Item "$PATCH\ListTemplates.vue" "$ROOT\frontend\src\views\templates\ListTemplates.vue" -Force

Step "Layout consistency: Labels, Teams, Projects"
Copy-Item "$PATCH\ListLabels.vue"   "$ROOT\frontend\src\views\labels\ListLabels.vue" -Force
Copy-Item "$PATCH\ListTeams.vue"    "$ROOT\frontend\src\views\teams\ListTeams.vue" -Force
Copy-Item "$PATCH\ListProjects.vue" "$ROOT\frontend\src\views\project\ListProjects.vue" -Force

Step "Upcoming page: filters + assigned-to-me"
Copy-Item "$PATCH\ShowTasks.vue" "$ROOT\frontend\src\views\tasks\ShowTasks.vue" -Force

Step "Home page: tasks first + auto-task check"
Copy-Item "$PATCH\Home.vue" "$ROOT\frontend\src\views\Home.vue" -Force

Step "Task row: title left, project right"
Copy-Item "$PATCH\SingleTaskInProject.vue" "$ROOT\frontend\src\components\tasks\partials\SingleTaskInProject.vue" -Force

# ===========================
#  FRONTEND - i18n + Misc
# ===========================
Step "i18n + subproject filter + version display"
Copy-Item "$PATCH\en.json"              "$ROOT\frontend\src\i18n\lang\en.json" -Force
Copy-Item "$PATCH\SubprojectFilter.vue" "$ROOT\frontend\src\components\project\partials\SubprojectFilter.vue" -Force
Copy-Item "$PATCH\PoweredByLink.vue"    "$ROOT\frontend\src\components\home\PoweredByLink.vue" -Force

# ===========================
#  DOCUMENTATION
# ===========================
Step "Documentation: changelog, architecture, manifest"
Copy-Item "$PATCH\CHANGELOG.md"       "$ROOT\CHANGELOG.md" -Force
Copy-Item "$PATCH\AUTO_TASKS.md"      "$ROOT\docs\AUTO_TASKS.md" -Force
Copy-Item "$PATCH\PATCH_MANIFEST.md"  "$ROOT\docs\PATCH_MANIFEST.md" -Force

# ===========================
#  PATCH SUMMARY
# ===========================
Write-Host ""
Write-Host "--- Patch Summary ---" -ForegroundColor Yellow
Write-Host "  Backend Go models    : 5 files (incl. core tasks.go)" -ForegroundColor Gray
Write-Host "  Backend handlers     : 2 files (echo v5)" -ForegroundColor Gray
Write-Host "  Migrations           : 5 files" -ForegroundColor Gray
Write-Host "  Routes               : 1 file" -ForegroundColor Gray
Write-Host "  Gantt components     : 7 files (arrows, settings, chart, bars, grid, header, config)" -ForegroundColor Gray
Write-Host "  Chain components     : 3 files" -ForegroundColor Gray
Write-Host "  Auto-task components : 2 files" -ForegroundColor Gray
Write-Host "  Stores/composables   : 4 files" -ForegroundColor Gray
Write-Host "  View pages           : 7 files" -ForegroundColor Gray
Write-Host "  i18n + misc          : 3 files (en.json, SubprojectFilter, PoweredByLink)" -ForegroundColor Gray
Write-Host "  Documentation        : 3 files" -ForegroundColor Gray
Write-Host "  --------------------------------" -ForegroundColor DarkGray
Write-Host "  TOTAL                : 45 files" -ForegroundColor White

# ===========================
#  BUILD
# ===========================
Step "Docker build"
Write-Host ""
Set-Location $ROOT
$gitHash = (git rev-parse --short HEAD 2>$null)
if (-not $gitHash) { $gitHash = "unknown" }
$buildTag = "custom-$gitHash"
Write-Host "  Build version: $buildTag" -ForegroundColor Cyan
$buildStart = Get-Date
docker buildx build --build-arg RELEASE_VERSION=$buildTag --tag vikunja-custom:latest --load .
$buildEnd = Get-Date
$buildSec = [math]::Round(($buildEnd - $buildStart).TotalSeconds)

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "  BUILD FAILED  ($buildSec sec)" -ForegroundColor Red
    Write-Host "  Fix errors above and re-run." -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "  BUILD OK  ($buildSec sec)" -ForegroundColor Green

# ===========================
#  DEPLOY
# ===========================
if (-not $Deploy) {
    Write-Host ""
    Write-Host "--- Deploy skipped (use -Deploy to push to server) ---" -ForegroundColor Yellow
    Write-Host "  Tip: .\vikunja-phase2\patch-phase2.ps1 -Deploy" -ForegroundColor Gray
} else {
    Step "Export + upload + restart"

    $configFile = "$ROOT\deploy-config.json"
    $cfg = $null

    # Load saved config if it exists
    if (Test-Path $configFile) {
        try {
            $cfg = Get-Content $configFile -Raw | ConvertFrom-Json
            Write-Host "  Loaded saved deploy config from deploy-config.json" -ForegroundColor DarkGray
        } catch {
            $cfg = $null
        }
    }

    # Prompt for details (with defaults from saved config or sensible fallbacks)
    $defaultHost   = if ($cfg.host)       { $cfg.host }       else { "" }
    $defaultUser   = if ($cfg.user)       { $cfg.user }       else { "root" }
    $defaultPort   = if ($cfg.port)       { $cfg.port }       else { "22" }
    $defaultRemote = if ($cfg.remotePath) { $cfg.remotePath } else { "/tmp" }
    $defaultKey    = if ($cfg.keyPath)    { $cfg.keyPath }    else { "$env:USERPROFILE\.ssh\id_ed25519" }

    Write-Host ""
    Write-Host "--- Deploy Configuration ---" -ForegroundColor Yellow

    $sshHost = Read-Host "  Server hostname or IP [$defaultHost]"
    if ([string]::IsNullOrWhiteSpace($sshHost)) { $sshHost = $defaultHost }
    if ([string]::IsNullOrWhiteSpace($sshHost)) {
        Write-Host "  [!] Server hostname is required" -ForegroundColor Red
        exit 1
    }

    $sshUser = Read-Host "  SSH username [$defaultUser]"
    if ([string]::IsNullOrWhiteSpace($sshUser)) { $sshUser = $defaultUser }

    $sshPort = Read-Host "  SSH port [$defaultPort]"
    if ([string]::IsNullOrWhiteSpace($sshPort)) { $sshPort = $defaultPort }

    $sshKey = Read-Host "  SSH key path [$defaultKey]"
    if ([string]::IsNullOrWhiteSpace($sshKey)) { $sshKey = $defaultKey }

    if (-not (Test-Path $sshKey)) {
        Write-Host "  [!] SSH key not found at $sshKey" -ForegroundColor Red
        Write-Host "  Generate one with: ssh-keygen -t ed25519" -ForegroundColor Yellow
        Write-Host "  Then copy it with: ssh-copy-id -p $sshPort ${sshUser}@${sshHost}" -ForegroundColor Yellow
        exit 1
    }

    $remoteFolder = Read-Host "  Remote folder to SCP tar into [$defaultRemote]"
    if ([string]::IsNullOrWhiteSpace($remoteFolder)) { $remoteFolder = $defaultRemote }

    # Ask to save for next time
    $saveConfig = Read-Host "  Save for next time? [Y/n]"
    if ($saveConfig -ne "n" -and $saveConfig -ne "N") {
        @{
            host       = $sshHost
            user       = $sshUser
            port       = $sshPort
            keyPath    = $sshKey
            remotePath = $remoteFolder
        } | ConvertTo-Json | Set-Content $configFile -Encoding UTF8
        Write-Host "  Saved to deploy-config.json" -ForegroundColor DarkGray
    }

    $server     = "${sshUser}@${sshHost}"
    $tarFile    = "$ROOT\vikunja-custom.tar"
    $remoteTar  = "${remoteFolder}/vikunja-custom.tar"

    Write-Host ""
    Write-Host "  Saving image..." -ForegroundColor Gray
    docker save vikunja-custom:latest -o $tarFile

    if (-not (Test-Path $tarFile)) {
        Write-Host "  [!] docker save failed" -ForegroundColor Red
        exit 1
    }

    $sizeMB = [math]::Round((Get-Item $tarFile).Length / 1MB, 1)
    Write-Host "  Image: $sizeMB MB" -ForegroundColor Gray

    Write-Host "  Uploading to ${server}:${remoteTar} ..." -ForegroundColor Gray
    scp -P $sshPort -i $sshKey -o StrictHostKeyChecking=no -o PasswordAuthentication=no $tarFile "${server}:${remoteTar}"

    if ($LASTEXITCODE -ne 0) {
        Write-Host ""
        Write-Host "  [!] SCP upload failed - SSH key auth may not be set up" -ForegroundColor Red
        Write-Host "  Copy your key to the server with:" -ForegroundColor Yellow
        Write-Host "    ssh-copy-id -i $sshKey -p $sshPort ${server}" -ForegroundColor Cyan
        Remove-Item $tarFile -Force -ErrorAction SilentlyContinue
        exit 1
    }

    Write-Host "  Loading image on server..." -ForegroundColor Gray
    ssh -t -p $sshPort -i $sshKey -o StrictHostKeyChecking=no -o PasswordAuthentication=no $server "docker load -i $remoteTar; rm $remoteTar"

    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "  DEPLOYED - image loaded on server" -ForegroundColor Green
        Write-Host "  Restart the stack in Portainer to apply." -ForegroundColor Yellow
    } else {
        Write-Host ""
        Write-Host "  [!] Remote load failed - check server" -ForegroundColor Red
        Write-Host "  To eliminate the sudo password prompt, run on the server:" -ForegroundColor Yellow
        Write-Host "    sudo usermod -aG docker $sshUser" -ForegroundColor Cyan
        Write-Host "  Then log out and back in. After that, change the load command to:" -ForegroundColor Yellow
        Write-Host "    docker load (no sudo needed)" -ForegroundColor Cyan
    }

    Remove-Item $tarFile -Force -ErrorAction SilentlyContinue
}

Write-Host ""
Write-Host "==========================================================" -ForegroundColor Cyan
$ts2 = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
Write-Host "  Finished: $ts2" -ForegroundColor Cyan
Write-Host "==========================================================" -ForegroundColor Cyan
