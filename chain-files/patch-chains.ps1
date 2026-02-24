$ErrorActionPreference = "Stop"
$ROOT = "C:\Users\antho\Downloads\vikunja-task-duplicate"
$PATCH = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host " Vikunja Task Chains - Full Patch" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path "$ROOT\frontend\src")) {
    Write-Host "[!] Vikunja source not found at $ROOT" -ForegroundColor Red; exit 1
}

Write-Host "[1/5] Backend models..." -ForegroundColor Green
Copy-Item "$PATCH\task_chain.go" "$ROOT\pkg\models\task_chain.go" -Force
Write-Host "  + pkg/models/task_chain.go"
Copy-Item "$PATCH\task_from_chain.go" "$ROOT\pkg\models\task_from_chain.go" -Force
Write-Host "  + pkg/models/task_from_chain.go"

Write-Host ""
Write-Host "[2/5] Backend migration + routes..." -ForegroundColor Green
Copy-Item "$PATCH\20260224040000.go" "$ROOT\pkg\migration\20260224040000.go" -Force
Write-Host "  + pkg/migration/20260224040000.go"
Copy-Item "$PATCH\routes.go" "$ROOT\pkg\routes\routes.go" -Force
Write-Host "  ~ pkg/routes/routes.go (chain routes added)"

Write-Host ""
Write-Host "[3/5] Frontend API + components..." -ForegroundColor Green
Copy-Item "$PATCH\taskChainApi.ts" "$ROOT\frontend\src\services\taskChainApi.ts" -Force
Write-Host "  + services/taskChainApi.ts"
Copy-Item "$PATCH\ChainEditor.vue" "$ROOT\frontend\src\components\tasks\partials\ChainEditor.vue" -Force
Write-Host "  + components/tasks/partials/ChainEditor.vue"
Copy-Item "$PATCH\CreateFromChainModal.vue" "$ROOT\frontend\src\components\tasks\partials\CreateFromChainModal.vue" -Force
Write-Host "  + components/tasks/partials/CreateFromChainModal.vue"

Write-Host ""
Write-Host "[4/5] Frontend views + templates page..." -ForegroundColor Green
Copy-Item "$PATCH\ListTemplates.vue" "$ROOT\frontend\src\views\templates\ListTemplates.vue" -Force
Write-Host "  ~ views/templates/ListTemplates.vue (Chains tab added)"
Copy-Item "$PATCH\ProjectList.vue" "$ROOT\frontend\src\components\project\views\ProjectList.vue" -Force
Write-Host "  ~ ProjectList.vue (From Chain button)"
Copy-Item "$PATCH\ProjectTable.vue" "$ROOT\frontend\src\components\project\views\ProjectTable.vue" -Force
Write-Host "  ~ ProjectTable.vue (From Chain button)"

Write-Host ""
Write-Host "[5/5] i18n..." -ForegroundColor Green
Copy-Item "$PATCH\en.json" "$ROOT\frontend\src\i18n\lang\en.json" -Force
Write-Host "  ~ en.json (29 chain keys added)"

Write-Host ""
Write-Host "[BUILD] Building Docker image..." -ForegroundColor Green
Set-Location $ROOT
docker buildx build --tag vikunja-custom:latest --load .

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host " BUILD SUCCESSFUL!" -ForegroundColor Green
    Write-Host ""
    Write-Host "  docker save vikunja-custom:latest -o vikunja-custom.tar" -ForegroundColor Yellow
    Write-Host "  scp vikunja-custom.tar superuser@192.168.2.102:/tmp/" -ForegroundColor Yellow
    Write-Host ""
    Write-Host " New features:" -ForegroundColor Cyan
    Write-Host "  - Templates page: Chains tab for creating chain workflows" -ForegroundColor White
    Write-Host "  - Chain editor: Define steps with title + offset + duration" -ForegroundColor White
    Write-Host "  - 'From Chain' button in List/Table views" -ForegroundColor White
    Write-Host "  - Anchor date picker with preview of calculated dates" -ForegroundColor White
    Write-Host "  - Tasks auto-linked with precedes/follows relations" -ForegroundColor White
} else {
    Write-Host " BUILD FAILED" -ForegroundColor Red
}
