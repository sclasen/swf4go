package swf

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func TestMigrateDomains(t *testing.T) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Printf("WARNING: NO AWS CREDS SPECIFIED, SKIPPING MIGRATIONS TEST")
		return
	}

	client := NewClient(MustGetenv("AWS_ACCESS_KEY_ID"), MustGetenv("AWS_SECRET_ACCESS_KEY"), USEast1)

	domain := fmt.Sprintf("test-domain-%d", time.Now().UnixNano())

	req := RegisterDomain{
		Name:                                   domain,
		Description:                            "test domain",
		WorkflowExecutionRetentionPeriodInDays: "30",
	}

	d := DomainMigrator{
		RegisteredDomains: []RegisterDomain{req},
		Client:            client,
	}

	d.Migrate()
	d.Migrate()

	dep := DeprecateDomain{
		Name: domain,
	}

	dd := DomainMigrator{
		DeprecatedDomains: []DeprecateDomain{dep},
		Client:            client,
	}

	dd.Migrate()
	dd.Migrate()

}

func TestMigrateWorkflowTypes(t *testing.T) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Printf("WARNING: NO AWS CREDS SPECIFIED, SKIPPING MIGRATIONS TEST")
		return
	}

	client := NewClient(MustGetenv("AWS_ACCESS_KEY_ID"), MustGetenv("AWS_SECRET_ACCESS_KEY"), USEast1)

	workflow := fmt.Sprintf("test-workflow-%d", time.Now().UnixNano())
	version := fmt.Sprintf("test-workflow-version-%d", time.Now().UnixNano())

	req := RegisterWorkflowType{
		Name:        workflow,
		Description: "test workflow migration",
		Version:     version,
		Domain:      "test-domain",
	}

	w := WorkflowTypeMigrator{
		RegisteredWorkflowTypes: []RegisterWorkflowType{req},
		Client:                  client,
	}

	w.Migrate()
	w.Migrate()

	dep := DeprecateWorkflowType{
		WorkflowType: WorkflowType{
			Name:    workflow,
			Version: version,
		},
		Domain: "test-domain",
	}

	wd := WorkflowTypeMigrator{
		DeprecatedWorkflowTypes: []DeprecateWorkflowType{dep},
		Client:                  client,
	}

	wd.Migrate()
	wd.Migrate()

}

func TestMigrateActivityTypes(t *testing.T) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Printf("WARNING: NO AWS CREDS SPECIFIED, SKIPPING MIGRATIONS TEST")
		return
	}

	client := NewClient(MustGetenv("AWS_ACCESS_KEY_ID"), MustGetenv("AWS_SECRET_ACCESS_KEY"), USEast1)

	activity := fmt.Sprintf("test-activity-%d", time.Now().UnixNano())
	version := fmt.Sprintf("test-activity-version-%d", time.Now().UnixNano())

	req := RegisterActivityType{
		Name:        activity,
		Description: "test activity migration",
		Version:     version,
		Domain:      "test-domain",
	}

	a := ActivityTypeMigrator{
		RegisteredActivityTypes: []RegisterActivityType{req},
		Client:                  client,
	}

	a.Migrate()
	a.Migrate()

	dep := DeprecateActivityType{
		ActivityType: ActivityType{
			Name:    activity,
			Version: version,
		},
		Domain: "test-domain",
	}

	ad := ActivityTypeMigrator{
		DeprecatedActivityTypes: []DeprecateActivityType{dep},
		Client:                  client,
	}

	ad.Migrate()
	ad.Migrate()

}

func TestMigrateStreams(t *testing.T) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_SECRET_ACCESS_KEY") == "" {
		log.Printf("WARNING: NO AWS CREDS SPECIFIED, SKIPPING MIGRATIONS TEST")
		return
	}

	client := NewClient(MustGetenv("AWS_ACCESS_KEY_ID"), MustGetenv("AWS_SECRET_ACCESS_KEY"), USEast1)

	sm := StreamMigrator{
		Streams: []CreateStream{
			CreateStream{
				StreamName: "test",
				ShardCount: 1,
			},
		},
		Client: client,
	}

	sm.Migrate()
	sm.Migrate()

}
