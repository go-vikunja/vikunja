$ErrorActionPreference = "Stop"
$ROOT = "C:\Users\antho\Downloads\vikunja-task-duplicate"
$PATCH = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host " Phase 2: Cascade + Gantt Arrows + Tooltips" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path "$ROOT\frontend\src")) {
    Write-Host "[!] Vikunja source not found at $ROOT" -ForegroundColor Red; exit 1
}

Write-Host "[1/5] Backend: bi-directional chain relations..." -ForegroundColor Green
Copy-Item "$PATCH\task_from_chain.go" "$ROOT\pkg\models\task_from_chain.go" -Force
Write-Host "  ~ task_from_chain.go (inserts both precedes + follows)"

Write-Host ""
Write-Host "[2/5] Gantt: dependency arrows..." -ForegroundColor Green
Copy-Item "$PATCH\GanttDependencyArrows.vue" "$ROOT\frontend\src\components\gantt\GanttDependencyArrows.vue" -Force
Copy-Item "$PATCH\GanttChart.vue" "$ROOT\frontend\src\components\gantt\GanttChart.vue" -Force
Write-Host "  + GanttDependencyArrows.vue (SVG bezier arrows)"
Write-Host "  ~ GanttChart.vue (arrows overlay in rows area)"

Write-Host ""
Write-Host "[3/5] Gantt: bar tooltips..." -ForegroundColor Green
Copy-Item "$PATCH\GanttRowBars.vue" "$ROOT\frontend\src\components\gantt\GanttRowBars.vue" -Force
Write-Host "  ~ GanttRowBars.vue (SVG title tooltip on hover)"

Write-Host ""
Write-Host "[4/5] Date cascade logic..." -ForegroundColor Green
Copy-Item "$PATCH\useGanttTaskList.ts" "$ROOT\frontend\src\views\project\helpers\useGanttTaskList.ts" -Force
Write-Host "  ~ useGanttTaskList.ts (cascade prompt, recursive shift, TaskModel fix)"

Write-Host ""
Write-Host "[5/6] Chain editor + i18n..." -ForegroundColor Green
Copy-Item "$PATCH\ChainEditor.vue" "$ROOT\frontend\src\components\tasks\partials\ChainEditor.vue" -Force
Copy-Item "$PATCH\CreateFromChainModal.vue" "$ROOT\frontend\src\components\tasks\partials\CreateFromChainModal.vue" -Force
Copy-Item "$PATCH\en.json" "$ROOT\frontend\src\i18n\lang\en.json" -Force

Write-Host "[6/6] Bugfix: subproject filter..." -ForegroundColor Green
Copy-Item "$PATCH\SubprojectFilter.vue" "$ROOT\frontend\src\components\project\partials\SubprojectFilter.vue" -Force
Write-Host "  ~ ChainEditor.vue (cumulative Day X + timespan)"
Write-Host "  ~ CreateFromChainModal.vue (cumulative preview)"
Write-Host "  ~ en.json (Before/After labels)"

Write-Host ""
Write-Host "[BUILD]..." -ForegroundColor Green
Set-Location $ROOT
docker buildx build --tag vikunja-custom:latest --load .

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host " BUILD SUCCESSFUL!" -ForegroundColor Green
    Write-Host ""
    Write-Host " All features:" -ForegroundColor Cyan
    Write-Host "  Phase 1: Chain workflows, Templates tab, From Chain buttons" -ForegroundColor White
    Write-Host "  Phase 2: Date cascade, Gantt arrows, bar tooltips" -ForegroundColor White
    Write-Host "  Fix: Bi-directional relations (Before + After on tasks)" -ForegroundColor White
    Write-Host "  Fix: Cumulative offsets (days after previous step)" -ForegroundColor White
    Write-Host ""
    Write-Host "  IMPORTANT: Delete old chain tasks and recreate." -ForegroundColor Yellow
    Write-Host "  Old tasks have one-way relations only." -ForegroundColor Yellow
} else {
    Write-Host " BUILD FAILED" -ForegroundColor Red
}
