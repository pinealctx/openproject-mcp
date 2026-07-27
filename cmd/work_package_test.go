package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	external "github.com/pinealctx/openproject"
	"github.com/spf13/cobra"
)

func TestOptionalIntFlagDistinguishesOmittedAndExplicitValues(t *testing.T) {
	command := &cobra.Command{Use: "test"}
	value := 0
	command.Flags().IntVar(&value, "person", 0, "person ID")

	if got := optionalIntFlag(command, "person", value); got != nil {
		t.Fatalf("omitted flag returned %#v, want nil", got)
	}
	if err := command.Flags().Set("person", "7"); err != nil {
		t.Fatalf("set person flag: %v", err)
	}
	if got := optionalIntFlag(command, "person", value); got == nil || *got != 7 {
		t.Fatalf("explicit flag returned %#v, want pointer to 7", got)
	}
}

func TestWorkPackageOutputDistinguishesAssigneeAndAccountable(t *testing.T) {
	var workPackage external.WorkPackageModel
	if err := json.Unmarshal([]byte(`{
        "_type":"WorkPackage",
        "id":42,
        "subject":"Review release",
        "_links":{
          "type":{"href":"/api/v3/types/1","title":"Task"},
          "status":{"href":"/api/v3/statuses/1","title":"In progress"},
          "priority":{"href":"/api/v3/priorities/8","title":"Normal"},
          "project":{"href":"/api/v3/projects/3","title":"Team"},
          "author":{"href":"/api/v3/users/1","title":"Author"},
          "schema":{"href":"/api/v3/work_packages/schemas/3-1"},
          "self":{"href":"/api/v3/work_packages/42"},
          "assignee":{"href":"/api/v3/users/5","title":"Current Worker"},
          "responsible":{"href":"/api/v3/users/8","title":"Delivery Owner"}
        }
      }`), &workPackage); err != nil {
		t.Fatalf("decode work package fixture: %v", err)
	}

	var buffer bytes.Buffer
	previousWriter := outputWriter
	previousFormat := flagOutput
	outputWriter = &buffer
	flagOutput = "text"
	t.Cleanup(func() {
		outputWriter = previousWriter
		flagOutput = previousFormat
	})

	if err := output(&workPackage); err != nil {
		t.Fatalf("output work package: %v", err)
	}
	for _, expected := range []string{"Assignee: Current Worker", "Accountable: Delivery Owner"} {
		if !strings.Contains(buffer.String(), expected) {
			t.Fatalf("work package output missing %q:\n%s", expected, buffer.String())
		}
	}
}
