# Phase 1 API Contracts

## System Stabilization APIs

### Test Status API
```yaml
GET /api/test/status
Response: 200 OK
{
  "phase": "stabilization",
  "backend_tests": {
    "total": 0,
    "passing": 0,
    "failing": 0,
    "pass_rate": 0.0
  },
  "target_pass_rate": 1.0
}
```

### Task Query API  
```yaml
GET /api/v1/tasks/{id}
Response: 200 OK
{
  "id": 0,
  "title": "string",
  "description": "string", 
  "related_tasks": {
    "subtask": [],
    "parenttask": [],
    "related": []
  },
  "labels": [],
  "attachments": [],
  "assignees": []
}
```

### Label Creation API
```yaml
POST /api/v1/labels
Request:
{
  "title": "string",
  "hex_color": "string",
  "description": "string"
}
Response: 201 Created
{
  "id": 0,
  "title": "string", 
  "hex_color": "string",
  "description": "string",
  "created_by_id": 0
}
```

## Refactor Progress APIs

### Feature Refactor Status API
```yaml
GET /api/refactor/features
Response: 200 OK
{
  "total_features": 18,
  "refactored": 0,
  "remaining": 18,
  "current_phase": "stabilization",
  "features": [
    {
      "name": "string",
      "complexity": "high|medium|low",
      "status": "pending|in_progress|complete",
      "dependencies": []
    }
  ]
}
```

### Service Layer Health API
```yaml
GET /api/health/services
Response: 200 OK
{
  "total_services": 0,
  "healthy": 0,
  "test_coverage": {
    "service_layer": 0.0,
    "target": 0.9
  },
  "architecture_compliance": true
}
```

## Validation APIs

### Test Parity Analysis API
```yaml
GET /api/validation/test-parity
Response: 200 OK
{
  "original_tests": 0,
  "refactored_tests": 0,
  "missing_tests": [],
  "extra_tests": [],
  "parity_score": 0.0
}
```

### Functional Parity API
```yaml
GET /api/validation/functional-parity
Response: 200 OK
{
  "workflows_tested": 0,
  "workflows_passing": 0,
  "discrepancies": [],
  "parity_score": 0.0
}
```

### Architectural Review API
```yaml
GET /api/validation/architecture
Response: 200 OK
{
  "ai_analysis": {
    "status": "complete|pending",
    "violations": [],
    "score": 0.0
  },
  "human_approval": {
    "status": "approved|pending|rejected",
    "reviewer": "string",
    "comments": "string"
  }
}
```

## Performance Monitoring APIs

### Response Time API
```yaml
GET /api/performance/response-times
Response: 200 OK
{
  "p50": 0,
  "p95": 0,
  "p99": 0,
  "target_p95": 200,
  "compliant": true
}
```

### Coverage Metrics API
```yaml
GET /api/metrics/coverage
Response: 200 OK
{
  "service_layer": 0.0,
  "backend_overall": 0.0, 
  "frontend": 0.0,
  "targets": {
    "service_layer": 0.9,
    "backend_overall": 0.8,
    "frontend": 0.7
  }
}
```