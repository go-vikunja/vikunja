$ErrorActionPreference = "Stop"
$ROOT = "C:\Users\antho\Downloads\vikunja-task-duplicate"
$PATCH = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host " Vikunja Sub-project Roll-up - Patch" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path "$ROOT\frontend\src")) {
    Write-Host "[!] Vikunja source not found at $ROOT" -ForegroundColor Red; exit 1
}

Write-Host "[1/6] Patching backend..." -ForegroundColor Green
Copy-Item "$PATCH\project.go" "$ROOT\pkg\models\project.go" -Force
Write-Host "  ~ pkg/models/project.go (added GetAllChildProjectIDs)"
Copy-Item "$PATCH\task_collection.go" "$ROOT\pkg\models\task_collection.go" -Force
Write-Host "  ~ pkg/models/task_collection.go (include_subprojects param)"

Write-Host ""
Write-Host "[2/6] Patching frontend composables..." -ForegroundColor Green
Copy-Item "$PATCH\useTaskList.ts" "$ROOT\frontend\src\composables\useTaskList.ts" -Force
Write-Host "  ~ composables/useTaskList.ts"
Copy-Item "$PATCH\useSubprojectColors.ts" "$ROOT\frontend\src\composables\useSubprojectColors.ts" -Force
Write-Host "  + composables/useSubprojectColors.ts"
Copy-Item "$PATCH\useGanttTaskList.ts" "$ROOT\frontend\src\views\project\helpers\useGanttTaskList.ts" -Force
Write-Host "  ~ helpers/useGanttTaskList.ts"
Copy-Item "$PATCH\useGanttFilters.ts" "$ROOT\frontend\src\views\project\helpers\useGanttFilters.ts" -Force
Write-Host "  ~ helpers/useGanttFilters.ts"
Copy-Item "$PATCH\taskCollection.ts" "$ROOT\frontend\src\services\taskCollection.ts" -Force
Write-Host "  ~ services/taskCollection.ts"

Write-Host ""
Write-Host "[3/6] Adding/updating components..." -ForegroundColor Green
Copy-Item "$PATCH\SubprojectFilter.vue" "$ROOT\frontend\src\components\project\partials\SubprojectFilter.vue" -Force
Write-Host "  + components/project/partials/SubprojectFilter.vue"

Write-Host ""
Write-Host "[4/6] Patching Gantt..." -ForegroundColor Green
Copy-Item "$PATCH\GanttChart.vue" "$ROOT\frontend\src\components\gantt\GanttChart.vue" -Force
Write-Host "  ~ components/gantt/GanttChart.vue (subproject colors)"
Copy-Item "$PATCH\ProjectGantt.vue" "$ROOT\frontend\src\components\project\views\ProjectGantt.vue" -Force
Write-Host "  ~ ProjectGantt.vue"

Write-Host ""
Write-Host "[5/6] Patching List/Table views + i18n..." -ForegroundColor Green
Copy-Item "$PATCH\ProjectList.vue" "$ROOT\frontend\src\components\project\views\ProjectList.vue" -Force
Write-Host "  ~ ProjectList.vue"
Copy-Item "$PATCH\ProjectTable.vue" "$ROOT\frontend\src\components\project\views\ProjectTable.vue" -Force
Write-Host "  ~ ProjectTable.vue (project column default ON)"
Copy-Item "$PATCH\ProjectCard.vue" "$ROOT\frontend\src\components\project\partials\ProjectCard.vue" -Force
Write-Host "  ~ ProjectCard.vue (parent project name on cards)"
Copy-Item "$PATCH\en.json" "$ROOT\frontend\src\i18n\lang\en.json" -Force
Write-Host "  ~ en.json"

Write-Host ""
Write-Host "[6/6] Building Docker image..." -ForegroundColor Green
Set-Location $ROOT
docker buildx build --tag vikunja-custom:latest --load .

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host " BUILD SUCCESSFUL!" -ForegroundColor Green
    Write-Host "  docker save vikunja-custom:latest -o vikunja-custom.tar" -ForegroundColor Yellow
    Write-Host "  scp vikunja-custom.tar superuser@mail:/tmp/" -ForegroundColor Yellow
} else {
    Write-Host " BUILD FAILED" -ForegroundColor Red
}
