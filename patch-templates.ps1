$ErrorActionPreference = "Stop"
$ROOT = "C:\Users\antho\Downloads\vikunja-task-duplicate"
$NEW_FILES = "$ROOT\new-files"

Write-Host ""
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host " Vikunja Task Templates v3 - Patch & Build" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path "$ROOT\frontend\src")) {
    Write-Host "[!] Vikunja source not found at $ROOT" -ForegroundColor Red; exit 1
}

Write-Host "[1/6] Copying NEW backend files..." -ForegroundColor Green
Copy-Item "$NEW_FILES\task_template.go" "$ROOT\pkg\models\task_template.go" -Force
Write-Host "  + pkg/models/task_template.go"
Copy-Item "$NEW_FILES\task_from_template.go" "$ROOT\pkg\models\task_from_template.go" -Force
Write-Host "  + pkg/models/task_from_template.go"
Copy-Item "$NEW_FILES\20260223120000.go" "$ROOT\pkg\migration\20260223120000.go" -Force
Write-Host "  + pkg/migration/20260223120000.go"

Write-Host ""
Write-Host "[2/6] Copying NEW frontend TypeScript files..." -ForegroundColor Green
@("ITaskTemplate.ts","ITaskFromTemplate.ts") | ForEach-Object {
    Copy-Item "$NEW_FILES\$_" "$ROOT\frontend\src\modelTypes\$_" -Force; Write-Host "  + modelTypes/$_"
}
@("taskTemplate.ts","taskFromTemplate.ts") | ForEach-Object {
    Copy-Item "$NEW_FILES\$_" "$ROOT\frontend\src\models\$_" -Force; Write-Host "  + models/$_"
}
@("taskTemplateService.ts","taskFromTemplateService.ts") | ForEach-Object {
    Copy-Item "$NEW_FILES\$_" "$ROOT\frontend\src\services\$_" -Force; Write-Host "  + services/$_"
}

Write-Host ""
Write-Host "[3/6] Copying NEW Vue components..." -ForegroundColor Green
@("CreateFromTemplateModal.vue","SaveAsTemplateModal.vue") | ForEach-Object {
    Copy-Item "$NEW_FILES\$_" "$ROOT\frontend\src\components\tasks\partials\$_" -Force; Write-Host "  + components/tasks/partials/$_"
}
if (-not (Test-Path "$ROOT\frontend\src\views\templates")) {
    New-Item -ItemType Directory -Path "$ROOT\frontend\src\views\templates" -Force | Out-Null
}
Copy-Item "$NEW_FILES\views-templates\ListTemplates.vue" "$ROOT\frontend\src\views\templates\ListTemplates.vue" -Force
Write-Host "  + views/templates/ListTemplates.vue"

Write-Host ""
Write-Host "[4/6] Replacing MODIFIED files..." -ForegroundColor Green
Copy-Item "$NEW_FILES\routes.go" "$ROOT\pkg\routes\routes.go" -Force
Write-Host "  ~ pkg/routes/routes.go"
Copy-Item "$NEW_FILES\KanbanCard.vue" "$ROOT\frontend\src\components\tasks\partials\KanbanCard.vue" -Force
Write-Host "  ~ KanbanCard.vue"
Copy-Item "$NEW_FILES\ProjectKanban.vue" "$ROOT\frontend\src\components\project\views\ProjectKanban.vue" -Force
Write-Host "  ~ ProjectKanban.vue"
Copy-Item "$NEW_FILES\ProjectList.vue" "$ROOT\frontend\src\components\project\views\ProjectList.vue" -Force
Write-Host "  ~ ProjectList.vue"
Copy-Item "$NEW_FILES\ProjectTable.vue" "$ROOT\frontend\src\components\project\views\ProjectTable.vue" -Force
Write-Host "  ~ ProjectTable.vue"
Copy-Item "$NEW_FILES\TaskDetailView.vue" "$ROOT\frontend\src\views\tasks\TaskDetailView.vue" -Force
Write-Host "  ~ TaskDetailView.vue"
Copy-Item "$NEW_FILES\en.json" "$ROOT\frontend\src\i18n\lang\en.json" -Force
Write-Host "  ~ en.json"
Copy-Item "$NEW_FILES\index.ts" "$ROOT\frontend\src\router\index.ts" -Force
Write-Host "  ~ router/index.ts"
Copy-Item "$NEW_FILES\Navigation.vue" "$ROOT\frontend\src\components\home\Navigation.vue" -Force
Write-Host "  ~ Navigation.vue"

Write-Host ""
Write-Host "[5/6] All files patched!" -ForegroundColor Green
Write-Host ""
Write-Host "[6/6] Building Docker image..." -ForegroundColor Green
Write-Host "       This will take 5-10 minutes." -ForegroundColor DarkGray

Set-Location $ROOT
docker buildx build --tag vikunja-custom:latest --load .

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host " BUILD SUCCESSFUL!" -ForegroundColor Green
    Write-Host "  docker save vikunja-custom:latest -o vikunja-custom.tar" -ForegroundColor Yellow
    Write-Host "  scp vikunja-custom.tar superuser@mail:/tmp/" -ForegroundColor Yellow
    Write-Host ""
} else {
    Write-Host " BUILD FAILED" -ForegroundColor Red
}
