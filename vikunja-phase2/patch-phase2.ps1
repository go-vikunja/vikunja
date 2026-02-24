$ErrorActionPreference = "Stop"
$ROOT = "C:\Users\antho\Downloads\vikunja-task-duplicate"
$PATCH = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "`n========== Phase 2b: Cascade + Arrows + Tooltips + Attachments (Echo v5 fix) ==========`n" -ForegroundColor Cyan

if (-not (Test-Path "$ROOT\frontend\src")) {
    Write-Host "[!] Source not found at $ROOT" -ForegroundColor Red; exit 1
}

Write-Host "[1/8] Backend: chain model + bidirectional relations + cumulative offsets..." -ForegroundColor Green
Copy-Item "$PATCH\task_chain.go"      "$ROOT\pkg\models\task_chain.go" -Force
Copy-Item "$PATCH\task_from_chain.go" "$ROOT\pkg\models\task_from_chain.go" -Force

Write-Host "[2/8] Backend: step attachment handler (Echo v5 fixed)..." -ForegroundColor Green
New-Item -ItemType Directory -Path "$ROOT\pkg\routes\api\v1" -Force | Out-Null
Copy-Item "$PATCH\chain_step_attachment.go" "$ROOT\pkg\routes\api\v1\chain_step_attachment.go" -Force

Write-Host "[3/8] Migration: step attachments table..." -ForegroundColor Green
Copy-Item "$PATCH\20260224050000.go" "$ROOT\pkg\migration\20260224050000.go" -Force

Write-Host "[4/8] Routes: attachment endpoints..." -ForegroundColor Green
Copy-Item "$PATCH\routes.go" "$ROOT\pkg\routes\routes.go" -Force

Write-Host "[5/8] Gantt: dependency arrows + bar tooltips..." -ForegroundColor Green
Copy-Item "$PATCH\GanttDependencyArrows.vue" "$ROOT\frontend\src\components\gantt\GanttDependencyArrows.vue" -Force
Copy-Item "$PATCH\GanttChart.vue"            "$ROOT\frontend\src\components\gantt\GanttChart.vue" -Force
Copy-Item "$PATCH\GanttRowBars.vue"          "$ROOT\frontend\src\components\gantt\GanttRowBars.vue" -Force

Write-Host "[6/8] Cascade + chain editor + attachments..." -ForegroundColor Green
Copy-Item "$PATCH\useGanttTaskList.ts"      "$ROOT\frontend\src\views\project\helpers\useGanttTaskList.ts" -Force
Copy-Item "$PATCH\tasks.ts"                 "$ROOT\frontend\src\stores\tasks.ts" -Force
Copy-Item "$PATCH\taskChainApi.ts"          "$ROOT\frontend\src\services\taskChainApi.ts" -Force
Copy-Item "$PATCH\ChainEditor.vue"          "$ROOT\frontend\src\components\tasks\partials\ChainEditor.vue" -Force
Copy-Item "$PATCH\CreateFromChainModal.vue" "$ROOT\frontend\src\components\tasks\partials\CreateFromChainModal.vue" -Force

Write-Host "[7/8] i18n + subproject filter fix..." -ForegroundColor Green
Copy-Item "$PATCH\en.json"              "$ROOT\frontend\src\i18n\lang\en.json" -Force
Copy-Item "$PATCH\SubprojectFilter.vue" "$ROOT\frontend\src\components\project\partials\SubprojectFilter.vue" -Force

Write-Host "[8/8] Building..." -ForegroundColor Green
Set-Location $ROOT
docker buildx build --tag vikunja-custom:latest --load .

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n BUILD OK`n" -ForegroundColor Green
    Write-Host "========== Next Steps: Export & Deploy ==========" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "  1) Export the image to a tar file:" -ForegroundColor Yellow
    Write-Host "     docker save vikunja-custom:latest -o vikunja-custom.tar" -ForegroundColor White
    Write-Host ""
    Write-Host "  2) Transfer to your server:" -ForegroundColor Yellow
    Write-Host "     scp vikunja-custom.tar root@<SERVER_IP>:/tmp/" -ForegroundColor White
    Write-Host ""
    Write-Host "  3) SSH into the server and load the image:" -ForegroundColor Yellow
    Write-Host "     ssh root@<SERVER_IP>" -ForegroundColor White
    Write-Host "     docker load -i /tmp/vikunja-custom.tar" -ForegroundColor White
    Write-Host ""
    Write-Host "  4) Restart the stack:" -ForegroundColor Yellow
    Write-Host "     cd /path/to/vikunja-compose && docker compose down && docker compose up -d" -ForegroundColor White
    Write-Host ""
    Write-Host "  5) Clean up:" -ForegroundColor Yellow
    Write-Host "     rm /tmp/vikunja-custom.tar" -ForegroundColor White
    Write-Host ""
} else {
    Write-Host "`n BUILD FAILED" -ForegroundColor Red
}
