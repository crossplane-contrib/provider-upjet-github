package repository

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// When a repository is archived on GitHub but the desired state does not set
// archived=true, Terraform computes a spurious archived: true->false diff.
// GitHub does not support unarchiving via this path, and worse: on delete this
// diff makes the terraform delete function re-PATCH the already-archived
// (read-only) repository, which returns HTTP 403 and wedges the managed
// resource forever. The diff must be dropped.
func TestRepositoryCustomDiff_StripsUnarchiveFlip(t *testing.T) {
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"archived": {Old: "true", New: "false"},
		},
	}

	out, err := repositoryCustomDiff(diff, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := out.Attributes["archived"]; ok {
		t.Fatalf("expected archived true->false diff to be stripped, but it is still present: %#v", out.Attributes["archived"])
	}
}

// A genuine request to archive (false->true) must be preserved so users can
// still archive repositories through the resource.
func TestRepositoryCustomDiff_PreservesArchiveRequest(t *testing.T) {
	diff := &terraform.InstanceDiff{
		Attributes: map[string]*terraform.ResourceAttrDiff{
			"archived": {Old: "false", New: "true"},
		},
	}

	out, err := repositoryCustomDiff(diff, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	d, ok := out.Attributes["archived"]
	if !ok {
		t.Fatal("expected archived false->true diff to be preserved, but it was removed")
	}
	if d.Old != "false" || d.New != "true" {
		t.Fatalf("archived diff altered: got Old=%q New=%q", d.Old, d.New)
	}
}
