package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestShow_ExternalRef(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	// Build bd binary
	tmpBin := filepath.Join(t.TempDir(), "bd")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, "./")
	buildCmd.Dir = "."
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build bd: %v\n%s", err, out)
	}

	// Create temp directory for test database
	tmpDir := t.TempDir()

	// Initialize beads
	initCmd := exec.Command(tmpBin, "init", "--prefix", "test", "--quiet")
	initCmd.Dir = tmpDir
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\n%s", err, out)
	}

	// Create issue with external ref
	createCmd := exec.Command(tmpBin, "--no-daemon", "create", "External ref test", "-p", "1",
		"--external-ref", "https://example.com/spec.md", "--json")
	createCmd.Dir = tmpDir
	createOut, err := createCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("create failed: %v\n%s", err, createOut)
	}

	var issue map[string]interface{}
	if err := json.Unmarshal(createOut, &issue); err != nil {
		t.Fatalf("failed to parse create output: %v, output: %s", err, createOut)
	}
	id := issue["id"].(string)

	// Show the issue and verify external ref is displayed
	showCmd := exec.Command(tmpBin, "--no-daemon", "show", id)
	showCmd.Dir = tmpDir
	showOut, err := showCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("show failed: %v\n%s", err, showOut)
	}

	out := string(showOut)
	if !strings.Contains(out, "External Ref:") {
		t.Errorf("expected 'External Ref:' in output, got: %s", out)
	}
	if !strings.Contains(out, "https://example.com/spec.md") {
		t.Errorf("expected external ref URL in output, got: %s", out)
	}
}

func TestShow_NoExternalRef(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	// Build bd binary
	tmpBin := filepath.Join(t.TempDir(), "bd")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, "./")
	buildCmd.Dir = "."
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build bd: %v\n%s", err, out)
	}

	tmpDir := t.TempDir()

	// Initialize beads
	initCmd := exec.Command(tmpBin, "init", "--prefix", "test", "--quiet")
	initCmd.Dir = tmpDir
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\n%s", err, out)
	}

	// Create issue WITHOUT external ref
	createCmd := exec.Command(tmpBin, "--no-daemon", "create", "No ref test", "-p", "1", "--json")
	createCmd.Dir = tmpDir
	createOut, err := createCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("create failed: %v\n%s", err, createOut)
	}

	var issue map[string]interface{}
	if err := json.Unmarshal(createOut, &issue); err != nil {
		t.Fatalf("failed to parse create output: %v, output: %s", err, createOut)
	}
	id := issue["id"].(string)

	// Show the issue - should NOT contain External Ref line
	showCmd := exec.Command(tmpBin, "--no-daemon", "show", id)
	showCmd.Dir = tmpDir
	showOut, err := showCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("show failed: %v\n%s", err, showOut)
	}

	out := string(showOut)
	if strings.Contains(out, "External Ref:") {
		t.Errorf("expected no 'External Ref:' line for issue without external ref, got: %s", out)
	}
}

func TestShow_ParentSeparation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI test in short mode")
	}

	// Build bd binary
	tmpBin := filepath.Join(t.TempDir(), "bd")
	buildCmd := exec.Command("go", "build", "-o", tmpBin, "./")
	buildCmd.Dir = "."
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build bd: %v\n%s", err, out)
	}

	tmpDir := t.TempDir()

	// Initialize beads
	initCmd := exec.Command(tmpBin, "init", "--prefix", "test", "--quiet")
	initCmd.Dir = tmpDir
	if out, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\n%s", err, out)
	}

	// Helper to create issue and get ID
	createIssue := func(title, issueType string) string {
		createCmd := exec.Command(tmpBin, "--no-daemon", "create", title, "-t", issueType, "-p", "2", "--json")
		createCmd.Dir = tmpDir
		createOut, err := createCmd.Output() // Use Output() to get only stdout, not stderr warnings
		if err != nil {
			t.Fatalf("create failed: %v\n%s", err, createOut)
		}
		var issue map[string]interface{}
		if err := json.Unmarshal(createOut, &issue); err != nil {
			t.Fatalf("failed to parse create output: %v, output: %s", err, createOut)
		}
		return issue["id"].(string)
	}

	// Helper to add dependency
	addDep := func(child, parent, depType string) {
		depCmd := exec.Command(tmpBin, "--no-daemon", "dep", "add", child, parent, "--type", depType)
		depCmd.Dir = tmpDir
		if out, err := depCmd.CombinedOutput(); err != nil {
			t.Fatalf("dep add failed: %v\n%s", err, out)
		}
	}

	t.Run("single_parent", func(t *testing.T) {
		epic := createIssue("Epic 1", "epic")
		task := createIssue("Task with single parent", "task")
		addDep(task, epic, "parent-child")

		// Text output should show "Parent:" (singular)
		showCmd := exec.Command(tmpBin, "--no-daemon", "show", task)
		showCmd.Dir = tmpDir
		showOut, err := showCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("show failed: %v\n%s", err, showOut)
		}

		out := string(showOut)
		if !strings.Contains(out, "Parent:\n") {
			t.Errorf("expected 'Parent:' (singular) in output, got: %s", out)
		}
		if !strings.Contains(out, "↑ "+epic) {
			t.Errorf("expected parent with ↑ symbol, got: %s", out)
		}
		if strings.Contains(out, "Parents (") {
			t.Errorf("should not show 'Parents (' for single parent, got: %s", out)
		}
	})

	t.Run("multiple_parents", func(t *testing.T) {
		epic1 := createIssue("Epic 1", "epic")
		epic2 := createIssue("Epic 2", "epic")
		task := createIssue("Task with multiple parents", "task")
		addDep(task, epic1, "parent-child")
		addDep(task, epic2, "parent-child")

		// Text output should show "Parents (2):" (plural with count)
		showCmd := exec.Command(tmpBin, "--no-daemon", "show", task)
		showCmd.Dir = tmpDir
		showOut, err := showCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("show failed: %v\n%s", err, showOut)
		}

		out := string(showOut)
		if !strings.Contains(out, "Parents (2):\n") {
			t.Errorf("expected 'Parents (2):' in output, got: %s", out)
		}
		if !strings.Contains(out, "↑ "+epic1) || !strings.Contains(out, "↑ "+epic2) {
			t.Errorf("expected both parents with ↑ symbols, got: %s", out)
		}
	})

	t.Run("parent_and_blocking", func(t *testing.T) {
		epic := createIssue("Epic", "epic")
		blocker := createIssue("Blocker", "task")
		task := createIssue("Task with parent and blocker", "task")
		addDep(task, epic, "parent-child")
		addDep(task, blocker, "blocks")

		// Text output should show parent separately from blocking dep
		showCmd := exec.Command(tmpBin, "--no-daemon", "show", task)
		showCmd.Dir = tmpDir
		showOut, err := showCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("show failed: %v\n%s", err, showOut)
		}

		out := string(showOut)
		if !strings.Contains(out, "Parent:\n") {
			t.Errorf("expected 'Parent:' section, got: %s", out)
		}
		if !strings.Contains(out, "Depends on (1):\n") {
			t.Errorf("expected 'Depends on (1):' section, got: %s", out)
		}
		if !strings.Contains(out, "↑ "+epic) {
			t.Errorf("expected parent with ↑ symbol, got: %s", out)
		}
		if !strings.Contains(out, "→ "+blocker) {
			t.Errorf("expected blocker with → symbol, got: %s", out)
		}
	})

	t.Run("json_output_unchanged", func(t *testing.T) {
		epic := createIssue("Epic JSON", "epic")
		blocker := createIssue("Blocker JSON", "task")
		task := createIssue("Task JSON", "task")
		addDep(task, epic, "parent-child")
		addDep(task, blocker, "blocks")

		// JSON output should include all deps with dependency_type field
		showCmd := exec.Command(tmpBin, "--no-daemon", "show", task, "--json")
		showCmd.Dir = tmpDir
		showOut, err := showCmd.CombinedOutput()
		if err != nil {
			t.Fatalf("show --json failed: %v\n%s", err, showOut)
		}

		var result []map[string]interface{}
		if err := json.Unmarshal(showOut, &result); err != nil {
			t.Fatalf("failed to parse JSON output: %v, output: %s", err, showOut)
		}
		if len(result) == 0 {
			t.Fatal("expected at least one result")
		}

		issue := result[0]
		deps, ok := issue["dependencies"].([]interface{})
		if !ok {
			t.Fatalf("expected dependencies array in JSON output, got: %v", issue)
		}
		if len(deps) != 2 {
			t.Errorf("expected 2 dependencies in JSON, got %d", len(deps))
		}

		// Verify both deps have dependency_type field
		for _, dep := range deps {
			depMap := dep.(map[string]interface{})
			if _, ok := depMap["dependency_type"]; !ok {
				t.Errorf("expected dependency_type field in JSON, got: %v", depMap)
			}
		}

		// Verify parent field exists (convenience field)
		if _, ok := issue["parent"]; !ok {
			t.Error("expected parent convenience field in JSON output")
		}
	})
}
