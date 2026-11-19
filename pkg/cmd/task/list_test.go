package task

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/triptechtravel/clickup-cli/internal/config"
	"github.com/triptechtravel/clickup-cli/internal/testutil"
)

var sampleTasksJSON = `{
	"tasks": [
		{
			"id": "t1",
			"name": "Fix bug",
			"status": {"status": "open", "color": "#999"},
			"priority": {"priority": "high", "color": "#f00"},
			"assignees": [],
			"tags": []
		}
	]
}`

func TestTaskList_FallsBackToConfiguredList(t *testing.T) {
	tf := testutil.NewTestFactory(t)

	tf.Factory.SetConfig(&config.Config{
		Workspace: "12345",
		Space:     "67890",
		List:      "configured-list",
	})

	tf.HandleFunc("list/configured-list/task", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "99")
		w.Write([]byte(sampleTasksJSON))
	})

	cmd := NewCmdList(tf.Factory)
	err := testutil.RunCommand(t, cmd)
	require.NoError(t, err)

	out := tf.OutBuf.String()
	assert.Contains(t, out, "t1")
	assert.Contains(t, out, "Fix bug")
}

func TestTaskList_IncludeClosed(t *testing.T) {
	tf := testutil.NewTestFactory(t)

	var capturedQuery string
	tf.HandleFunc("list/mylist/task", func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "99")
		w.Write([]byte(sampleTasksJSON))
	})

	cmd := NewCmdList(tf.Factory)
	err := testutil.RunCommand(t, cmd, "--list-id", "mylist", "--include-closed")
	require.NoError(t, err)

	assert.True(t, strings.Contains(capturedQuery, "include_closed=true"),
		"expected include_closed=true in query, got: %s", capturedQuery)
}

func TestTaskList_NoListError(t *testing.T) {
	tf := testutil.NewTestFactory(t)

	tf.Factory.SetConfig(&config.Config{
		Workspace: "12345",
		Space:     "67890",
		// No List configured
	})

	cmd := NewCmdList(tf.Factory)
	err := testutil.RunCommand(t, cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no list specified")
}

func TestTaskList_Subtasks(t *testing.T) {
	tf := testutil.NewTestFactory(t)

	var capturedQuery string
	tf.HandleFunc("list/mylist/task", func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "99")
		w.Write([]byte(`{"tasks": [
			{"id": "t1", "name": "Parent task", "status": {"status": "open"}, "priority": {}, "assignees": [], "tags": []},
			{"id": "t2", "name": "Child task", "parent": "t1", "status": {"status": "open"}, "priority": {}, "assignees": [], "tags": []}
		]}`))
	})

	cmd := NewCmdList(tf.Factory)
	err := testutil.RunCommand(t, cmd, "--list-id", "mylist", "--subtasks", "--status", "scoping")
	require.NoError(t, err)

	assert.True(t, strings.Contains(capturedQuery, "subtasks=true"),
		"expected subtasks=true in query, got: %s", capturedQuery)
	assert.True(t, strings.Contains(capturedQuery, "statuses%5B%5D=scoping"),
		"expected status filter to be kept, got: %s", capturedQuery)

	out := tf.OutBuf.String()
	assert.Contains(t, out, "Parent task")
	assert.Contains(t, out, "Child task")
}

func TestTaskList_OmitsSubtasksByDefault(t *testing.T) {
	tf := testutil.NewTestFactory(t)

	var capturedQuery string
	tf.HandleFunc("list/mylist/task", func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-RateLimit-Remaining", "99")
		w.Write([]byte(sampleTasksJSON))
	})

	cmd := NewCmdList(tf.Factory)
	err := testutil.RunCommand(t, cmd, "--list-id", "mylist")
	require.NoError(t, err)

	assert.False(t, strings.Contains(capturedQuery, "subtasks"),
		"expected no subtasks parameter by default, got: %s", capturedQuery)
}
