package repository

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

//nolint:goconst // Literal Terraform state values keep the test fixtures readable.
func TestPreserveArchivedRepositoryState(t *testing.T) {
	tests := []struct {
		name                 string
		archiveOnDestroy     string
		stateArchived        string
		wantArchivedDiffKept bool
	}{
		{
			name:                 "already archived with archive on destroy",
			archiveOnDestroy:     "true",
			stateArchived:        "true",
			wantArchivedDiffKept: false,
		},
		{
			name:                 "unarchived repository",
			archiveOnDestroy:     "true",
			stateArchived:        "false",
			wantArchivedDiffKept: true,
		},
		{
			name:                 "archive on destroy disabled",
			archiveOnDestroy:     "false",
			stateArchived:        "true",
			wantArchivedDiffKept: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := &terraform.InstanceDiff{Attributes: map[string]*terraform.ResourceAttrDiff{
				"archived": {Old: "true", New: "false"},
			}}
			state := &terraform.InstanceState{Attributes: map[string]string{
				"archive_on_destroy": tt.archiveOnDestroy,
				"archived":           tt.stateArchived,
			}}

			got, err := PreserveArchivedRepositoryState(diff, state, nil)
			if err != nil {
				t.Fatalf("PreserveArchivedRepositoryState returned error: %v", err)
			}
			_, gotArchivedDiffKept := got.Attributes["archived"]
			if gotArchivedDiffKept != tt.wantArchivedDiffKept {
				t.Errorf("archived diff kept = %v, want %v", gotArchivedDiffKept, tt.wantArchivedDiffKept)
			}
		})
	}
}
