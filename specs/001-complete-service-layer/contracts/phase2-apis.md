# Phase 2 Refactor APIs

## Service Layer Management APIs

### Service Registration API
```yaml
POST /api/services/register
Request:
{
  "service_name": "string",
  "model_files": ["string"],
  "complexity": "high|medium|low",
  "dependencies": ["string"]
}
Response: 201 Created
{
  "service_id": "string", 
  "status": "registered",
  "priority": 0
}
```

### Dependency Graph API
```yaml
GET /api/services/dependencies
Response: 200 OK
{
  "graph": {
    "nodes": [
      {
        "name": "string",
        "complexity": "high|medium|low",
        "status": "pending|in_progress|complete"
      }
    ],
    "edges": [
      {
        "from": "string",
        "to": "string", 
        "type": "dependency"
      }
    ]
  }
}
```

## Feature-Specific Service APIs

### Projects Service API
```yaml
GET /api/v1/projects/{id}/service-status
Response: 200 OK
{
  "service_implemented": false,
  "business_logic_location": "model|service", 
  "handler_uses_wrapper": false,
  "test_coverage": 0.0
}
```

### Tasks Service API  
```yaml
GET /api/v1/tasks/service-status
Response: 200 OK
{
  "service_implemented": false,
  "query_methods_complete": false,
  "related_data_population": false,
  "test_coverage": 0.0
}
```

### Labels Service API
```yaml
POST /api/v1/labels/migrate-service
Request:
{
  "force": false,
  "test_first": true
}
Response: 200 OK
{
  "migration_id": "string",
  "status": "started",
  "estimated_duration": "5m"
}
```

## Handler Wrapper APIs

### Declarative Routing API
```yaml
GET /api/handlers/routes
Response: 200 OK
{
  "total_routes": 0,
  "using_wrappers": 0,
  "using_declarative": 0,
  "coverage": 0.0
}
```

### Handler Wrapper Health API
```yaml
GET /api/handlers/wrapper-health
Response: 200 OK
{
  "wrappers": [
    {
      "name": "WithDBAndUser",
      "usage_count": 0,
      "error_rate": 0.0
    }
  ]
}
```

## Backward Compatibility APIs

### Dependency Inversion Status API
```yaml
GET /api/compatibility/dependency-inversion
Response: 200 OK
{
  "total_model_functions": 0,
  "using_inversion": 0,
  "deprecated_functions": 0,
  "compatibility_score": 0.0
}
```

### Model Deprecation API
```yaml
POST /api/compatibility/deprecate-model-method
Request:
{
  "model": "string",
  "method": "string",
  "service_replacement": "string"
}
Response: 200 OK
{
  "deprecation_id": "string",
  "warning_added": true,
  "inversion_setup": true
}
```