$ErrorActionPreference = "Stop"
$ROOT = "C:\Users\antho\Downloads\vikunja-task-duplicate"
$PATCH = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host " Phase 1: Task Chain Workflows" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path "$ROOT\frontend\src")) {
    Write-Host "[!] Vikunja source not found at $ROOT" -ForegroundColor Red; exit 1
}

Write-Host "[1/5] Backend models..." -ForegroundColor Green
Copy-Item "$PATCH\task_chain.go" "$ROOT\pkg\models\task_chain.go" -Force
Copy-Item "$PATCH\task_from_chain.go" "$ROOT\pkg\models\task_from_chain.go" -Force
Write-Host "  + task_chain.go (chain + step models, CRUD)"
Write-Host "  + task_from_chain.go (task generation, cumulative offsets, bi-directional relations)"

Write-Host ""
Write-Host "[2/5] Migration + routes..." -ForegroundColor Green
Copy-Item "$PATCH\20260224040000.go" "$ROOT\pkg\migration\20260224040000.go" -Force
Copy-Item "$PATCH\routes.go" "$ROOT\pkg\routes\routes.go" -Force
Write-Host "  + 20260224040000.go (task_chains + task_chain_steps tables)"
Write-Host "  ~ routes.go (chain API endpoints)"

Write-Host ""
Write-Host "[3/5] Frontend API + components..." -ForegroundColor Green
Copy-Item "$PATCH\taskChainApi.ts" "$ROOT\frontend\src\services\taskChainApi.ts" -Force
Copy-Item "$PATCH\ChainEditor.vue" "$ROOT\frontend\src\components\tasks\partials\ChainEditor.vue" -Force
Copy-Item "$PATCH\CreateFromChainModal.vue" "$ROOT\frontend\src\components\tasks\partials\CreateFromChainModal.vue" -Force
Write-Host "  + taskChainApi.ts"
Write-Host "  + ChainEditor.vue (step editor, Day X indicators, timespan)"
Write-Host "  + CreateFromChainModal.vue (chain picker, anchor date, preview)"

Write-Host ""
Write-Host "[4/5] View integration..." -ForegroundColor Green
Copy-Item "$PATCH\ListTemplates.vue" "$ROOT\frontend\src\views\templates\ListTemplates.vue" -Force
Copy-Item "$PATCH\ProjectList.vue" "$ROOT\frontend\src\components\project\views\ProjectList.vue" -Force
Copy-Item "$PATCH\ProjectTable.vue" "$ROOT\frontend\src\components\project\views\ProjectTable.vue" -Force
Write-Host "  ~ ListTemplates.vue (Templates | Chains tabs)"
Write-Host "  ~ ProjectList.vue (From Chain button)"
Write-Host "  ~ ProjectTable.vue (From Chain button)"

Write-Host ""
Write-Host "[5/5] i18n..." -ForegroundColor Green
Copy-Item "$PATCH\en.json" "$ROOT\frontend\src\i18n\lang\en.json" -Force
Write-Host "  ~ en.json (30 chain keys, Before/After relation labels)"

Write-Host ""
Write-Host "Phase 1 complete. Run Phase 2 patch next, then build." -ForegroundColor Yellow
