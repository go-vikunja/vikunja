$ErrorActionPreference = "Stop"
$ROOT = "C:\Users\antho\Downloads\vikunja-task-duplicate"
$PATCH = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "`n========== Phase 2: Templates + Chains + Auto-Tasks + UI ==========`n" -ForegroundColor Cyan

if (-not (Test-Path "$ROOT\frontend\src")) {
    Write-Host "[!] Source not found at $ROOT" -ForegroundColor Red; exit 1
}

Write-Host "[1/19] Backend: chain model + task creation..." -ForegroundColor Green
Copy-Item "$PATCH\task_chain.go"      "$ROOT\pkg\models\task_chain.go" -Force
Copy-Item "$PATCH\task_from_chain.go" "$ROOT\pkg\models\task_from_chain.go" -Force

Write-Host "[2/19] Backend: step attachment handler..." -ForegroundColor Green
New-Item -ItemType Directory -Path "$ROOT\pkg\routes\api\v1" -Force | Out-Null
Copy-Item "$PATCH\chain_step_attachment.go" "$ROOT\pkg\routes\api\v1\chain_step_attachment.go" -Force

Write-Host "[3/19] Backend: auto-task model + creation logic + handler..." -ForegroundColor Green
Copy-Item "$PATCH\auto_task_template.go" "$ROOT\pkg\models\auto_task_template.go" -Force
Copy-Item "$PATCH\auto_task_create.go"   "$ROOT\pkg\models\auto_task_create.go" -Force
Copy-Item "$PATCH\auto_task_handler.go"  "$ROOT\pkg\routes\api\v1\auto_task_handler.go" -Force

Write-Host "[4/19] Migrations..." -ForegroundColor Green
Copy-Item "$PATCH\20260224050000.go" "$ROOT\pkg\migration\20260224050000.go" -Force
Copy-Item "$PATCH\20260224060000.go" "$ROOT\pkg\migration\20260224060000.go" -Force
Copy-Item "$PATCH\20260224070000.go" "$ROOT\pkg\migration\20260224070000.go" -Force

Write-Host "[5/19] Routes: all endpoints..." -ForegroundColor Green
Copy-Item "$PATCH\routes.go" "$ROOT\pkg\routes\routes.go" -Force

Write-Host "[6/19] Gantt: dependency arrows + bar tooltips..." -ForegroundColor Green
Copy-Item "$PATCH\GanttDependencyArrows.vue" "$ROOT\frontend\src\components\gantt\GanttDependencyArrows.vue" -Force
Copy-Item "$PATCH\GanttChart.vue"            "$ROOT\frontend\src\components\gantt\GanttChart.vue" -Force
Copy-Item "$PATCH\GanttRowBars.vue"          "$ROOT\frontend\src\components\gantt\GanttRowBars.vue" -Force

Write-Host "[7/19] Chain editor + API + create-from-chain..." -ForegroundColor Green
Copy-Item "$PATCH\useGanttTaskList.ts"      "$ROOT\frontend\src\views\project\helpers\useGanttTaskList.ts" -Force
Copy-Item "$PATCH\tasks.ts"                 "$ROOT\frontend\src\stores\tasks.ts" -Force
Copy-Item "$PATCH\taskChainApi.ts"          "$ROOT\frontend\src\services\taskChainApi.ts" -Force
Copy-Item "$PATCH\ChainEditor.vue"          "$ROOT\frontend\src\components\tasks\partials\ChainEditor.vue" -Force
Copy-Item "$PATCH\CreateFromChainModal.vue" "$ROOT\frontend\src\components\tasks\partials\CreateFromChainModal.vue" -Force

Write-Host "[8/19] Auto-task API + editor..." -ForegroundColor Green
Copy-Item "$PATCH\autoTaskApi.ts"       "$ROOT\frontend\src\services\autoTaskApi.ts" -Force
Copy-Item "$PATCH\AutoTaskEditor.vue"   "$ROOT\frontend\src\components\tasks\partials\AutoTaskEditor.vue" -Force

Write-Host "[9/19] Drag-to-reorder composable..." -ForegroundColor Green
New-Item -ItemType Directory -Path "$ROOT\frontend\src\composables" -Force | Out-Null
Copy-Item "$PATCH\useDragReorder.ts" "$ROOT\frontend\src\composables\useDragReorder.ts" -Force

Write-Host "[10/19] Template manager page (3 tabs)..." -ForegroundColor Green
Copy-Item "$PATCH\ListTemplates.vue" "$ROOT\frontend\src\views\templates\ListTemplates.vue" -Force

Write-Host "[11/19] i18n + subproject filter fix..." -ForegroundColor Green
Copy-Item "$PATCH\en.json"              "$ROOT\frontend\src\i18n\lang\en.json" -Force
Copy-Item "$PATCH\SubprojectFilter.vue" "$ROOT\frontend\src\components\project\partials\SubprojectFilter.vue" -Force

Write-Host "[12/19] Layout consistency: Labels, Teams, Projects..." -ForegroundColor Green
Copy-Item "$PATCH\ListLabels.vue"   "$ROOT\frontend\src\views\labels\ListLabels.vue" -Force
Copy-Item "$PATCH\ListTeams.vue"    "$ROOT\frontend\src\views\teams\ListTeams.vue" -Force
Copy-Item "$PATCH\ListProjects.vue" "$ROOT\frontend\src\views\project\ListProjects.vue" -Force

Write-Host "[13/19] Upcoming page: filters + checkbox persistence..." -ForegroundColor Green
Copy-Item "$PATCH\ShowTasks.vue" "$ROOT\frontend\src\views\tasks\ShowTasks.vue" -Force

Write-Host "[14/19] Home page: tasks above last viewed + auto-check..." -ForegroundColor Green
Copy-Item "$PATCH\Home.vue" "$ROOT\frontend\src\views\Home.vue" -Force

Write-Host "[15/19] Task row: title left, project right..." -ForegroundColor Green
Copy-Item "$PATCH\SingleTaskInProject.vue" "$ROOT\frontend\src\components\tasks\partials\SingleTaskInProject.vue" -Force

Write-Host "[16/19] Building..." -ForegroundColor Green
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
