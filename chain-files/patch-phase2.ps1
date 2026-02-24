$ErrorActionPreference = "Stop"
$ROOT = "C:\Users\antho\Downloads\vikunja-task-duplicate"
$PATCH = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host " Vikunja Phase 2 - Cascade + Gantt Arrows" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path "$ROOT\frontend\src")) {
    Write-Host "[!] Vikunja source not found at $ROOT" -ForegroundColor Red; exit 1
}

Write-Host "[1/4] Backend: cumulative offset fix..." -ForegroundColor Green
Copy-Item "$PATCH\task_from_chain.go" "$ROOT\pkg\models\task_from_chain.go" -Force
Write-Host "  ~ task_from_chain.go (offsets now cumulative)"

Write-Host ""
Write-Host "[2/4] Gantt: dependency arrows + tooltips..." -ForegroundColor Green
Copy-Item "$PATCH\GanttDependencyArrows.vue" "$ROOT\frontend\src\components\gantt\GanttDependencyArrows.vue" -Force
Write-Host "  + GanttDependencyArrows.vue (new)"
Copy-Item "$PATCH\GanttChart.vue" "$ROOT\frontend\src\components\gantt\GanttChart.vue" -Force
Write-Host "  ~ GanttChart.vue (arrows overlay integrated)"
Copy-Item "$PATCH\GanttRowBars.vue" "$ROOT\frontend\src\components\gantt\GanttRowBars.vue" -Force
Write-Host "  ~ GanttRowBars.vue (hover tooltip on bars)"

Write-Host ""
Write-Host "[3/4] Date cascade logic..." -ForegroundColor Green
Copy-Item "$PATCH\useGanttTaskList.ts" "$ROOT\frontend\src\views\project\helpers\useGanttTaskList.ts" -Force
Write-Host "  ~ useGanttTaskList.ts (cascade prompt + recursive shift)"

Write-Host ""
Write-Host "[4/4] Chain editor + i18n fixes..." -ForegroundColor Green
Copy-Item "$PATCH\ChainEditor.vue" "$ROOT\frontend\src\components\tasks\partials\ChainEditor.vue" -Force
Write-Host "  ~ ChainEditor.vue (cumulative day indicators + timespan)"
Copy-Item "$PATCH\CreateFromChainModal.vue" "$ROOT\frontend\src\components\tasks\partials\CreateFromChainModal.vue" -Force
Write-Host "  ~ CreateFromChainModal.vue (cumulative preview)"
Copy-Item "$PATCH\en.json" "$ROOT\frontend\src\i18n\lang\en.json" -Force
Write-Host "  ~ en.json"

Write-Host ""
Write-Host "[BUILD]..." -ForegroundColor Green
Set-Location $ROOT
docker buildx build --tag vikunja-custom:latest --load .

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host " BUILD SUCCESSFUL!" -ForegroundColor Green
    Write-Host ""
    Write-Host " Phase 2 features:" -ForegroundColor Cyan
    Write-Host "  - Date cascade: move a chain task, confirm to shift downstream" -ForegroundColor White
    Write-Host "  - Gantt dependency arrows: dashed lines between chain tasks" -ForegroundColor White
    Write-Host "  - Cumulative offsets: 'days after prev' not 'days from anchor'" -ForegroundColor White
    Write-Host "  - Chain editor: live Day X indicators + total timespan" -ForegroundColor White
} else {
    Write-Host " BUILD FAILED" -ForegroundColor Red
}
