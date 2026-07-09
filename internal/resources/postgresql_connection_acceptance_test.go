package resources

import (
	"os"
	"testing"
)

func TestAccPostgreSQLConnectionCreate(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to run acceptance tests")
	}

	t.Skip("Acceptance test skeleton requires Databricks workspace credentials and isolated connection names.")
}

func TestAccPostgreSQLConnectionUpdate(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to run acceptance tests")
	}

	t.Skip("Acceptance test skeleton requires Databricks workspace credentials and isolated connection names.")
}

func TestAccPostgreSQLConnectionReplacementPlan(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to run acceptance tests")
	}

	t.Skip("Acceptance test skeleton requires Databricks workspace credentials and isolated connection names.")
}

func TestAccPostgreSQLConnectionPasswordRotation(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to run acceptance tests")
	}

	t.Skip("Acceptance test skeleton requires Databricks workspace credentials and isolated connection names.")
}
