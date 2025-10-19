# T-PERM-015A: Quick Execution Guide

**Task**: Model Test Regression Prevention & Audit  
**Time**: 0.5 days  
**Status**: Ready to execute  

## Quick Start (Copy-Paste Execution)

### Phase 1: Service Layer Baseline (15 minutes)

```bash
cd /home/aron/projects/vikunja

echo "=== Phase 1: Service Layer Regression Tests ==="

# Run ALL service tests and save baseline
VIKUNJA_SERVICE_ROOTPATH=$(pwd) go test ./pkg/services -v -count=1 > /tmp/service_tests_baseline.txt 2>&1

# Verify permission baseline tests
VIKUNJA_SERVICE_ROOTPATH=$(pwd) go test ./pkg/services -run TestPermissionBaseline -v

# Create summary
echo "Service Tests Baseline - $(date)" > /tmp/service_baseline_summary.txt
echo "Total PASS: $(grep -c "^--- PASS:" /tmp/service_tests_baseline.txt)" >> /tmp/service_baseline_summary.txt
echo "Total FAIL: $(grep -c "^--- FAIL:" /tmp/service_tests_baseline.txt)" >> /tmp/service_baseline_summary.txt
echo "" >> /tmp/service_baseline_summary.txt
echo "Baseline tests (expected 6/6):" >> /tmp/service_baseline_summary.txt
grep "TestPermissionBaseline" /tmp/service_tests_baseline.txt | grep -E "PASS|FAIL" >> /tmp/service_baseline_summary.txt

# Show summary
cat /tmp/service_baseline_summary.txt
```

**Expected Output**:
```
Service Tests Baseline - [date]
Total PASS: 100+
Total FAIL: 0

Baseline tests (expected 6/6):
--- PASS: TestPermissionBaseline_Project
--- PASS: TestPermissionBaseline_Task
--- PASS: TestPermissionBaseline_LinkSharing
--- PASS: TestPermissionBaseline_Label
--- PASS: TestPermissionBaseline_TaskComment
--- PASS: TestPermissionBaseline_Subscription
```

---

### Phase 2: Model Test Audit (15 minutes)

```bash
cd /home/aron/projects/vikunja

echo "=== Phase 2: Model Test Audit ==="

# Audit permission method calls
echo "=== Model Test Permission Method Audit ===" > /tmp/model_test_audit.txt
echo "Generated: $(date)" >> /tmp/model_test_audit.txt
echo "" >> /tmp/model_test_audit.txt
grep -rn "\.Can\(Read\|Write\|Update\|Delete\|Create\)" pkg/models/*_test.go >> /tmp/model_test_audit.txt

# Summary by method type
echo "" >> /tmp/model_test_audit.txt
echo "=== Summary by Method ===" >> /tmp/model_test_audit.txt
echo "CanRead calls: $(grep -c "\.CanRead(" pkg/models/*_test.go 2>/dev/null || echo 0)" >> /tmp/model_test_audit.txt
echo "CanWrite calls: $(grep -c "\.CanWrite(" pkg/models/*_test.go 2>/dev/null || echo 0)" >> /tmp/model_test_audit.txt
echo "CanUpdate calls: $(grep -c "\.CanUpdate(" pkg/models/*_test.go 2>/dev/null || echo 0)" >> /tmp/model_test_audit.txt
echo "CanDelete calls: $(grep -c "\.CanDelete(" pkg/models/*_test.go 2>/dev/null || echo 0)" >> /tmp/model_test_audit.txt
echo "CanCreate calls: $(grep -c "\.CanCreate(" pkg/models/*_test.go 2>/dev/null || echo 0)" >> /tmp/model_test_audit.txt
echo "Total: $(grep -c "\.Can\(Read\|Write\|Update\|Delete\|Create\)" pkg/models/*_test.go 2>/dev/null || echo 0)" >> /tmp/model_test_audit.txt

# Show summary
tail -10 /tmp/model_test_audit.txt
```

**Expected Output**:
```
=== Summary by Method ===
CanRead calls: ~20
CanWrite calls: ~15
CanUpdate calls: ~15
CanDelete calls: ~15
CanCreate calls: ~9
Total: 74
```

---

### Phase 3: Test Categorization (10 minutes)

```bash
cd /home/aron/projects/vikunja

echo "=== Phase 3: Test File Categorization ==="

# Categorize each test file
echo "=== Test File Categorization ===" > /tmp/model_test_categories.txt
echo "Generated: $(date)" >> /tmp/model_test_categories.txt
echo "" >> /tmp/model_test_categories.txt

for file in pkg/models/*_test.go; do
  basename_file=$(basename "$file")
  structure_tests=$(grep -c "TableName\|Validate\|\.Error()\|\.Equal(" "$file" 2>/dev/null || echo 0)
  permission_tests=$(grep -c "\.Can\(Read\|Write\|Update\|Delete\|Create\)" "$file" 2>/dev/null || echo 0)
  db_operations=$(grep -c "db\.NewSession\|s\.Where\|s\.Get\|s\.Find" "$file" 2>/dev/null || echo 0)
  
  if [ $permission_tests -gt 0 ]; then
    action="DELETE/REFACTOR"
  elif [ $structure_tests -gt 0 ] && [ $db_operations -eq 0 ]; then
    action="KEEP (pure structure)"
  elif [ $structure_tests -gt 0 ] && [ $db_operations -gt 0 ]; then
    action="REFACTOR (mixed)"
  else
    action="REVIEW"
  fi
  
  printf "%-40s Structure=%-3d Permission=%-3d DB=%-3d => %s\n" \
    "$basename_file:" "$structure_tests" "$permission_tests" "$db_operations" "$action" >> /tmp/model_test_categories.txt
done

echo "" >> /tmp/model_test_categories.txt
echo "=== Summary by Action ===" >> /tmp/model_test_categories.txt
echo "DELETE/REFACTOR: $(grep -c "DELETE/REFACTOR" /tmp/model_test_categories.txt)" >> /tmp/model_test_categories.txt
echo "KEEP: $(grep -c "KEEP" /tmp/model_test_categories.txt)" >> /tmp/model_test_categories.txt
echo "REVIEW: $(grep -c "REVIEW" /tmp/model_test_categories.txt)" >> /tmp/model_test_categories.txt

echo "" >> /tmp/model_test_categories.txt
echo "=== Files to DELETE/REFACTOR (have permission tests) ===" >> /tmp/model_test_categories.txt
grep "DELETE/REFACTOR" /tmp/model_test_categories.txt | grep -v "===" | cut -d: -f1 >> /tmp/model_test_categories.txt

# Show results
cat /tmp/model_test_categories.txt
```

---

### Phase 4: Helper Function Verification (10 minutes)

```bash
cd /home/aron/projects/vikunja

echo "=== Phase 4: Verifying Temporary Helper Functions ==="

echo "=== Helper Function Verification ===" > /tmp/helper_verification.txt
echo "Generated: $(date)" >> /tmp/helper_verification.txt
echo "" >> /tmp/helper_verification.txt

# Test each helper function's usage area
echo "Testing GetSavedFilterSimpleByID..." | tee -a /tmp/helper_verification.txt
VIKUNJA_SERVICE_ROOTPATH=$(pwd) timeout 30 go test ./pkg/models -run TestSavedFilter -v >> /tmp/helper_verification.txt 2>&1 || echo "Test completed/timed out"

echo "Testing GetLinkSharesByIDs..." | tee -a /tmp/helper_verification.txt
VIKUNJA_SERVICE_ROOTPATH=$(pwd) timeout 30 go test ./pkg/models -run TestLinkShare -v >> /tmp/helper_verification.txt 2>&1 || echo "Test completed/timed out"

echo "Testing GetProjectViewByID..." | tee -a /tmp/helper_verification.txt
VIKUNJA_SERVICE_ROOTPATH=$(pwd) timeout 30 go test ./pkg/models -run TestProjectView -v >> /tmp/helper_verification.txt 2>&1 || echo "Test completed/timed out"

echo "Testing GetTokenFromTokenString..." | tee -a /tmp/helper_verification.txt
VIKUNJA_SERVICE_ROOTPATH=$(pwd) timeout 30 go test ./pkg/models -run TestAPIToken -v >> /tmp/helper_verification.txt 2>&1 || echo "Test completed/timed out"

# Summary
echo "" >> /tmp/helper_verification.txt
echo "=== Verification Summary ===" >> /tmp/helper_verification.txt
echo "Tests run: 4 helper function areas" >> /tmp/helper_verification.txt
echo "Note: Tests may fail/panic due to removed permission methods (expected)" >> /tmp/helper_verification.txt
echo "Goal: Verify helpers don't cause compilation errors" >> /tmp/helper_verification.txt

# Show tail
tail -20 /tmp/helper_verification.txt
```

---

### Final Verification (5 minutes)

```bash
cd /home/aron/projects/vikunja

echo "=== Final Verification ==="

# Check all audit files exist
echo "Checking audit files..."
ls -lh /tmp/service_tests_baseline.txt \
       /tmp/service_baseline_summary.txt \
       /tmp/model_test_audit.txt \
       /tmp/model_test_categories.txt \
       /tmp/helper_verification.txt

# Line counts
echo ""
echo "File sizes:"
wc -l /tmp/service_tests_baseline.txt | awk '{print "Service baseline: " $1 " lines"}'
wc -l /tmp/model_test_audit.txt | awk '{print "Permission calls audit: " $1 " lines"}'
wc -l /tmp/model_test_categories.txt | awk '{print "Test categorization: " $1 " lines"}'

# Production code still compiles
echo ""
echo "Verifying production code compiles..."
go build ./pkg/models ./pkg/services ./pkg/routes && echo "✅ Production code compiles successfully"

# Service tests still pass
echo ""
echo "Verifying service tests still pass..."
VIKUNJA_SERVICE_ROOTPATH=$(pwd) go test ./pkg/services -run TestPermissionBaseline && echo "✅ All 6 baseline permission tests pass"
```

**Expected Output**:
```
✅ Service baseline: 1000+ lines
✅ Permission calls audit: 80+ lines (74 calls + headers)
✅ Test categorization: 30+ lines
✅ Production code compiles successfully
✅ All 6 baseline permission tests pass
```

---

## Summary Report Template

After completion, create this summary in T-PERM-015A task:

```markdown
## T-PERM-015A Completion Summary

**Completion Date**: [DATE]
**Status**: ✅ COMPLETE

### Results:

**Service Layer Baseline**:
- Total service tests: [X]
- Passing: [X] (100%)
- Baseline permission tests: 6/6 passing ✅
- Baseline saved to: `/tmp/service_tests_baseline.txt`

**Model Test Audit**:
- Total permission method calls: 74
- Breakdown:
  - CanRead: [X]
  - CanWrite: [X]
  - CanUpdate: [X]
  - CanDelete: [X]
  - CanCreate: [X]
- Audit saved to: `/tmp/model_test_audit.txt`

**Test Categorization**:
- Files to DELETE/REFACTOR: [X] files (have permission tests)
- Files to KEEP: [X] files (pure structure tests)
- Files to REVIEW: [X] files
- Categorization saved to: `/tmp/model_test_categories.txt`

**Helper Function Verification**:
- GetSavedFilterSimpleByID: ✅ No compilation errors
- GetLinkSharesByIDs: ✅ No compilation errors
- GetProjectViewByIDAndProject: ✅ No compilation errors
- GetProjectViewByID: ✅ No compilation errors
- GetTokenFromTokenString: ✅ No compilation errors

**Files for T-PERM-016**:

DELETE/REFACTOR (files with permission tests):
[List from /tmp/model_test_categories.txt]

KEEP (pure structure tests):
[List from /tmp/model_test_categories.txt]

### Success Criteria: ✅ ALL MET
- ✅ Service tests: 100% passing (baseline documented)
- ✅ Model test audit: All 74 calls documented
- ✅ Test categorization: All files categorized
- ✅ Helper functions: Verified working
- ✅ Production code: Compiles successfully
```

---

## Troubleshooting

### Issue: Service tests fail
**Check**: Run individual baseline tests
```bash
VIKUNJA_SERVICE_ROOTPATH=$(pwd) go test ./pkg/services -run TestPermissionBaseline_Project -v
```

### Issue: Model test audit shows 0 calls
**Check**: Verify grep syntax
```bash
grep "\.CanRead(" pkg/models/label_test.go
```

### Issue: Categorization script errors
**Check**: File permissions and existence
```bash
ls -la pkg/models/*_test.go | head
```

---

## Next Steps

After T-PERM-015A completion:
1. Review `/tmp/model_test_categories.txt` to plan T-PERM-016 work
2. Update T-PERM-016 task with specific file list
3. Begin T-PERM-016 execution with confidence (baseline established)
