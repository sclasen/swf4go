package swf

import (
	"log"
)

type TypesMigrator struct {
	DomainMigrator       *DomainMigrator
	WorkflowTypeMigrator *WorkflowTypeMigrator
	ActivityTypeMigrator *ActivityTypeMigrator
}

func (t *TypesMigrator) Migrate() {
	if t.ActivityTypeMigrator == nil {
		t.ActivityTypeMigrator = new(ActivityTypeMigrator)
	}
	if t.DomainMigrator == nil {
		t.DomainMigrator = new(DomainMigrator)
	}
	if t.WorkflowTypeMigrator == nil {
		t.WorkflowTypeMigrator = new(WorkflowTypeMigrator)
	}
	t.DomainMigrator.Migrate()
	t.WorkflowTypeMigrator.Migrate()
	t.ActivityTypeMigrator.Migrate()
}

type DomainMigrator struct {
	RegisteredDomains []RegisterDomain
	DeprecatedDomains []DeprecateDomain
	Client            *Client
}

func (d *DomainMigrator) Migrate() {
	for _, dd := range d.DeprecatedDomains {
		if d.isDeprecated(dd.Name) {
			log.Printf("action=migrate at=deprecate-domain domain=%s status=previously-deprecated", dd.Name)
		} else {
			d.deprecate(dd)
			log.Printf("action=migrate at=deprecate-domain domain=%s status=deprecated", dd.Name)
		}
	}
	for _, r := range d.RegisteredDomains {
		if d.isRegisteredNotDeprecated(r) {
			log.Printf("action=migrate at=register-domain domain=%s status=previously-registered", r.Name)
		} else {
			d.register(r)
			log.Printf("action=migrate at=register-domain domain=%s status=registered", r.Name)
		}
	}
}

func (d *DomainMigrator) isRegisteredNotDeprecated(rd RegisterDomain) bool {
	desc, err := d.describe(rd.Name)
	if err != nil {
		if err.Type == ErrorTypeUnknownResourceFault {
			return false
		} else {
			panic(err)
		}
	}

	return desc.DomainInfo.Status == StatusRegistered
}

func (d *DomainMigrator) register(rd RegisterDomain) {
	err := d.Client.RegisterDomain(rd)
	if err != nil {
		panic(err)
	}
}

func (d *DomainMigrator) isDeprecated(domain string) bool {
	desc, err := d.describe(domain)
	if err != nil {
		log.Printf("action=migrate at=is-dep domain=%s error=%s", err.Error())
		return false
	}

	return desc.DomainInfo.Status == StatusDeprecated
}

func (d *DomainMigrator) deprecate(dd DeprecateDomain) {
	err := d.Client.DeprecateDomain(dd)
	if err != nil {
		panic(err)
	}
}

func (d *DomainMigrator) describe(domain string) (*DescribeDomainResponse, *ErrorResponse) {
	resp, err := d.Client.DescribeDomain(DescribeDomainRequest{Name: domain})
	if err != nil {
		return nil, err.(*ErrorResponse)
	}
	return resp, nil
}

type WorkflowTypeMigrator struct {
	RegisteredWorkflowTypes []RegisterWorkflowType
	DeprecatedWorkflowTypes []DeprecateWorkflowType
	Client                  *Client
}

func (w *WorkflowTypeMigrator) Migrate() {
	for _, dd := range w.DeprecatedWorkflowTypes {
		if w.isDeprecated(dd.Domain, dd.WorkflowType.Name, dd.WorkflowType.Version) {
			log.Printf("action=migrate at=deprecate-workflow domain=%s workflow=%s version=%s status=previously-deprecated", dd.Domain, dd.WorkflowType.Name, dd.WorkflowType.Version)
		} else {
			w.deprecate(dd)
			log.Printf("action=migrate at=deprecate-workflow domain=%s  workflow=%s version=%s status=deprecate", dd.Domain, dd.WorkflowType.Name, dd.WorkflowType.Version)
		}
	}
	for _, r := range w.RegisteredWorkflowTypes {
		if w.isRegisteredNotDeprecated(r) {
			log.Printf("action=migrate at=register-workflow domain=%s workflow=%s version=%s status=previously-registered", r.Domain, r.Name, r.Version)
		} else {
			w.register(r)
			log.Printf("action=migrate at=register-workflow domain=%s  workflow=%s version=%s status=registered", r.Domain, r.Name, r.Version)
		}
	}
}

func (w *WorkflowTypeMigrator) isRegisteredNotDeprecated(rd RegisterWorkflowType) bool {
	desc, err := w.describe(rd.Domain, rd.Name, rd.Version)
	if err != nil {
		if err.Type == ErrorTypeUnknownResourceFault {
			return false
		} else {
			panic(err)
		}
	}

	return desc.TypeInfo.Status == StatusRegistered
}

func (w *WorkflowTypeMigrator) register(rd RegisterWorkflowType) {
	err := w.Client.RegisterWorkflowType(rd)
	if err != nil {
		panic(err)
	}
}

func (w *WorkflowTypeMigrator) isDeprecated(domain string, name string, version string) bool {
	desc, err := w.describe(domain, name, version)
	if err != nil {
		log.Printf("action=migrate at=is-dep domain=%s workflow=%s version=%s error=%s", domain, name, version, err.Error())
		return false
	}

	return desc.TypeInfo.Status == StatusDeprecated
}

func (w *WorkflowTypeMigrator) deprecate(dd DeprecateWorkflowType) {
	err := w.Client.DeprecateWorkflowType(dd)
	if err != nil {
		panic(err)
	}
}

func (w *WorkflowTypeMigrator) describe(domain string, name string, version string) (*DescribeWorkflowTypeResponse, *ErrorResponse) {
	resp, err := w.Client.DescribeWorkflowType(DescribeWorkflowTypeRequest{Domain: domain, WorkflowType: WorkflowType{Name: name, Version: version}})
	if err != nil {
		return nil, err.(*ErrorResponse)
	}
	return resp, nil
}

type ActivityTypeMigrator struct {
	RegisteredActivityTypes []RegisterActivityType
	DeprecatedActivityTypes []DeprecateActivityType
	Client                  *Client
}

func (a *ActivityTypeMigrator) Migrate() {
	for _, d := range a.DeprecatedActivityTypes {
		if a.isDeprecated(d.Domain, d.ActivityType.Name, d.ActivityType.Version) {
			log.Printf("action=migrate at=deprecate-activity domain=%s activity=%s version=%s status=previously-deprecated", d.Domain, d.ActivityType.Name, d.ActivityType.Version)
		} else {
			a.deprecate(d)
			log.Printf("action=migrate at=depreacate-activity domain=%s activity=%s version=%s status=deprecated", d.Domain, d.ActivityType.Name, d.ActivityType.Version)
		}
	}
	for _, r := range a.RegisteredActivityTypes {
		if a.isRegisteredNotDeprecated(r) {
			log.Printf("action=migrate at=register-activity domain=%s activity=%s version=%s status=previously-registered", r.Domain, r.Name, r.Version)
		} else {
			a.register(r)
			log.Printf("action=migrate at=register-activity domain=%s activity=%s version=%s status=registered", r.Domain, r.Name, r.Version)
		}
	}
}

func (a *ActivityTypeMigrator) isRegisteredNotDeprecated(rd RegisterActivityType) bool {
	desc, err := a.describe(rd.Domain, rd.Name, rd.Version)
	if err != nil {
		if err.Type == ErrorTypeUnknownResourceFault {
			return false
		} else {
			panic(err)
		}
	}

	return desc.TypeInfo.Status == StatusRegistered
}

func (a *ActivityTypeMigrator) register(rd RegisterActivityType) {
	err := a.Client.RegisterActivityType(rd)
	if err != nil {
		panic(err)
	}
}

func (a *ActivityTypeMigrator) isDeprecated(domain string, name string, version string) bool {
	desc, err := a.describe(domain, name, version)
	if err != nil {
		log.Printf("action=migrate at=is-dep domain=%s activity=%s version=%s error=%s", domain, name, version, err.Error())
		return false
	}

	return desc.TypeInfo.Status == StatusDeprecated
}

func (a *ActivityTypeMigrator) deprecate(dd DeprecateActivityType) {
	err := a.Client.DeprecateActivityType(dd)
	if err != nil {
		panic(err)
	}
}

func (a *ActivityTypeMigrator) describe(domain string, name string, version string) (*DescribeActivityTypeResponse, *ErrorResponse) {
	resp, err := a.Client.DescribeActivityType(DescribeActivityTypeRequest{Domain: domain, ActivityType: ActivityType{Name: name, Version: version}})
	if err != nil {
		return nil, err.(*ErrorResponse)
	}
	return resp, nil
}
