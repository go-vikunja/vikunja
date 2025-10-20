# Phase 3 Validation APIs

## Test Parity Analysis APIs

### Test Suite Comparison API
```yaml
POST /api/validation/compare-test-suites
Request:
{
  "original_path": "/path/to/vikunja_original_main",
  "refactored_path": "/path/to/vikunja",
  "analyze_coverage": true
}
Response: 200 OK
{
  "comparison_id": "string",
  "status": "started",
  "estimated_duration": "10m"
}
```

### Missing Test Cases API
```yaml
GET /api/validation/missing-tests/{comparison_id}
Response: 200 OK
{
  "missing_tests": [
    {
      "test_name": "string",
      "file_path": "string", 
      "test_type": "unit|integration|e2e",
      "severity": "critical|high|medium|low"
    }
  ],
  "total_missing": 0,
  "critical_missing": 0
}
```

### Test Case Migration API
```yaml
POST /api/validation/migrate-test-case
Request:
{
  "test_name": "string",
  "source_file": "string",
  "target_file": "string",
  "adapt_for_services": true
}
Response: 201 Created
{
  "migration_id": "string",
  "status": "complete",
  "adapted_test": "string"
}
```

## Functional Parity Validation APIs

### Workflow Checklist API
```yaml
GET /api/validation/workflow-checklist
Response: 200 OK
{
  "workflows": [
    {
      "name": "Create Project and Add Task",
      "steps": ["string"],
      "status": "pending|passed|failed",
      "discrepancies": ["string"]
    }
  ],
  "total_workflows": 0,
  "passed_workflows": 0
}
```

### Execute Workflow API
```yaml
POST /api/validation/execute-workflow
Request:
{
  "workflow_name": "string",
  "system": "original|refactored",
  "record_interactions": true
}
Response: 200 OK
{
  "execution_id": "string",
  "status": "running",
  "steps_completed": 0,
  "total_steps": 0
}
```

### Compare Workflow Results API
```yaml
POST /api/validation/compare-workflow-results
Request:
{
  "original_execution_id": "string",
  "refactored_execution_id": "string"
}
Response: 200 OK
{
  "identical": false,
  "differences": [
    {
      "step": "string",
      "original_result": {},
      "refactored_result": {},
      "impact": "critical|high|medium|low"
    }
  ],
  "resolution": "prefer_original"
}
```

## Architectural Review APIs

### AI Analysis API
```yaml
POST /api/validation/ai-analysis
Request:
{
  "scope": "full|incremental",
  "focus_areas": ["architecture", "patterns", "compliance"]
}
Response: 200 OK
{
  "analysis_id": "string",
  "status": "started",
  "estimated_duration": "15m"
}
```

### AI Analysis Results API
```yaml
GET /api/validation/ai-analysis/{analysis_id}
Response: 200 OK
{
  "status": "complete",
  "architectural_violations": [
    {
      "violation": "string",
      "severity": "critical|high|medium|low",
      "location": "string",
      "recommendation": "string"
    }
  ],
  "pattern_compliance": {
    "chef_waiter_pantry": true,
    "declarative_routing": true,
    "dependency_inversion": true,
    "tdd_coverage": true
  },
  "overall_score": 0.95
}
```

### Human Approval API
```yaml
POST /api/validation/human-approval
Request:
{
  "ai_analysis_id": "string",
  "reviewer": "string",
  "approved": true,
  "comments": "string",
  "conditions": ["string"]
}
Response: 200 OK
{
  "approval_id": "string",
  "status": "approved|rejected|conditional",
  "final_validation": true
}
```

## Final Quality Gates APIs

### Coverage Validation API
```yaml
GET /api/validation/final-coverage
Response: 200 OK
{
  "service_layer_coverage": 0.92,
  "backend_coverage": 0.85,
  "frontend_coverage": 0.75,
  "all_targets_met": true,
  "gaps": []
}
```

### Performance Validation API
```yaml
GET /api/validation/final-performance
Response: 200 OK
{
  "api_response_times": {
    "p95": 180,
    "target": 200,
    "compliant": true
  },
  "load_test_results": {
    "concurrent_users": 100,
    "success_rate": 1.0,
    "error_rate": 0.0
  }
}
```

### Completion Certificate API
```yaml
GET /api/validation/completion-certificate
Response: 200 OK
{
  "refactor_complete": true,
  "all_phases_passed": true,
  "quality_gates_met": true,
  "functional_parity_confirmed": true,
  "architectural_approval": "approved",
  "certificate_id": "string",
  "issued_date": "2025-09-25T00:00:00Z"
}
```