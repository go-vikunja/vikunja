$ErrorActionPreference = "Stop"
$ROOT = "C:\Users\antho\Downloads\vikunja-task-duplicate"
$PATCH = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "`n========== Phase 2c: Templates + Chains + Cascade + Arrows + Descriptions ==========`n" -ForegroundColor Cyan

if (-not (Test-Path "$ROOT\frontend\src")) {
    Write-Host "[!] Source not found at $ROOT" -ForegroundColor Red; exit 1
}

Write-Host "[1/9] Backend: chain model + bidirectional relations + cumulative offsets..." -ForegroundColor Green
Copy-Item "$PATCH\task_chain.go"      "$ROOT\pkg\models\task_chain.go" -Force
Copy-Item "$PATCH\task_from_chain.go" "$ROOT\pkg\models\task_from_chain.go" -Force

Write-Host "[2/9] Backend: step attachment handler (new file)..." -ForegroundColor Green
New-Item -ItemType Directory -Path "$ROOT\pkg\routes\api\v1" -Force | Out-Null
Copy-Item "$PATCH\chain_step_attachment.go" "$ROOT\pkg\routes\api\v1\chain_step_attachment.go" -Force

Write-Host "[3/9] Migration: step attachments table..." -ForegroundColor Green
Copy-Item "$PATCH\20260224050000.go" "$ROOT\pkg\migration\20260224050000.go" -Force

Write-Host "[4/9] Routes: attachment endpoints..." -ForegroundColor Green
Copy-Item "$PATCH\routes.go" "$ROOT\pkg\routes\routes.go" -Force

Write-Host "[5/9] Gantt: dependency arrows + bar tooltips..." -ForegroundColor Green
Copy-Item "$PATCH\GanttDependencyArrows.vue" "$ROOT\frontend\src\components\gantt\GanttDependencyArrows.vue" -Force
Copy-Item "$PATCH\GanttChart.vue"            "$ROOT\frontend\src\components\gantt\GanttChart.vue" -Force
Copy-Item "$PATCH\GanttRowBars.vue"          "$ROOT\frontend\src\components\gantt\GanttRowBars.vue" -Force

Write-Host "[6/10] Cascade + chain editor + attachments + create-from-chain..." -ForegroundColor Green
Copy-Item "$PATCH\useGanttTaskList.ts"      "$ROOT\frontend\src\views\project\helpers\useGanttTaskList.ts" -Force
Copy-Item "$PATCH\tasks.ts"                 "$ROOT\frontend\src\stores\tasks.ts" -Force
Copy-Item "$PATCH\taskChainApi.ts"          "$ROOT\frontend\src\services\taskChainApi.ts" -Force
Copy-Item "$PATCH\ChainEditor.vue"          "$ROOT\frontend\src\components\tasks\partials\ChainEditor.vue" -Force
Copy-Item "$PATCH\CreateFromChainModal.vue" "$ROOT\frontend\src\components\tasks\partials\CreateFromChainModal.vue" -Force

Write-Host "[7/10] Drag-to-reorder composable..." -ForegroundColor Green
New-Item -ItemType Directory -Path "$ROOT\frontend\src\composables" -Force | Out-Null
Copy-Item "$PATCH\useDragReorder.ts" "$ROOT\frontend\src\composables\useDragReorder.ts" -Force

Write-Host "[8/10] Template manager: tabs + NEW TEMPLATE button..." -ForegroundColor Green
Copy-Item "$PATCH\ListTemplates.vue" "$ROOT\frontend\src\views\templates\ListTemplates.vue" -Force

Write-Host "[9/13] i18n + subproject filter fix..." -ForegroundColor Green
Copy-Item "$PATCH\en.json"              "$ROOT\frontend\src\i18n\lang\en.json" -Force
Copy-Item "$PATCH\SubprojectFilter.vue" "$ROOT\frontend\src\components\project\partials\SubprojectFilter.vue" -Force

Write-Host "[10/13] Layout consistency: Labels, Teams, Projects..." -ForegroundColor Green
Copy-Item "$PATCH\ListLabels.vue"   "$ROOT\frontend\src\views\labels\ListLabels.vue" -Force
Copy-Item "$PATCH\ListTeams.vue"    "$ROOT\frontend\src\views\teams\ListTeams.vue" -Force
Copy-Item "$PATCH\ListProjects.vue" "$ROOT\frontend\src\views\project\ListProjects.vue" -Force

Write-Host "[11/13] Fix: Upcoming page checkbox persistence..." -ForegroundColor Green
Copy-Item "$PATCH\ShowTasks.vue" "$ROOT\frontend\src\views\tasks\ShowTasks.vue" -Force

Write-Host "[12/13] Building..." -ForegroundColor Green
Set-Location $ROOT
docker buildx build --tag vikunja-custom:latest --load .

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n BUILD OK" -ForegroundColor Green
    Write-Host "`n--- Deploy Steps (manual) ---" -ForegroundColor Yellow
    Write-Host "1. docker save vikunja-custom:latest -o vikunja-custom.tar"
    Write-Host "2. scp vikunja-custom.tar root@<SERVER_IP>:/tmp/"
    Write-Host "3. ssh root@<SERVER_IP>"
    Write-Host "4. docker load -i /tmp/vikunja-custom.tar"
    Write-Host "5. cd /path/to/vikunja-compose && docker compose down && docker compose up -d"
    Write-Host "6. rm /tmp/vikunja-custom.tar"
} else {
    Write-Host "`n BUILD FAILED" -ForegroundColor Red
}
