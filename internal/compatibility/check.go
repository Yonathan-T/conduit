package compatibility

import (
	"fmt"
	"strings"

	"conduit/internal/model"
)

type Severity string

const (
	SeverityBreaking Severity = "BREAKING"
	SeverityWarning  Severity = "WARNING"
)

type Change struct {
	Severity    Severity
	Target      string
	Description string
}

type Report struct {
	Changes []Change
}

func (r *Report) HasBreaking() bool {
	for _, c := range r.Changes {
		if c.Severity == SeverityBreaking {
			return true
		}
	}
	return false
}

func (r *Report) HasWarnings() bool {
	for _, c := range r.Changes {
		if c.Severity == SeverityWarning {
			return true
		}
	}
	return false
}

func (r *Report) String() string {
	if len(r.Changes) == 0 {
		return "Schema compatibility check passed with no warnings or breaking changes."
	}

	var b strings.Builder
	for _, c := range r.Changes {
		fmt.Fprintf(&b, "%s: %s\n  %s\n", c.Severity, c.Target, c.Description)
	}
	return strings.TrimSpace(b.String())
}

func (r *Report) add(sev Severity, target string, format string, args ...any) {
	r.Changes = append(r.Changes, Change{
		Severity:    sev,
		Target:      target,
		Description: fmt.Sprintf(format, args...),
	})
}

// Check compares an old baseline API with a new API and detects breaking wire changes.
func Check(oldAPI, newAPI *model.API) *Report {
	report := &Report{Changes: make([]Change, 0)}

	if oldAPI == nil {
		// No previous schema to compare against — brand new API is always compatible
		return report
	}

	if newAPI == nil {
		report.add(SeverityBreaking, "API", "entire API definition was removed")
		return report
	}

	newServices := make(map[string]model.Service)
	for _, s := range newAPI.Services {
		newServices[s.Name] = s
	}

	for _, oldSvc := range oldAPI.Services {
		newSvc, exists := newServices[oldSvc.Name]
		if !exists {
			report.add(SeverityBreaking, oldSvc.Name, "service was removed")
			continue
		}

		checkService(report, oldSvc, newSvc)
	}

	newTypes := make(map[string]model.Message)
	for _, t := range newAPI.Types {
		newTypes[t.Name] = t
	}

	for _, oldType := range oldAPI.Types {
		newType, exists := newTypes[oldType.Name]
		if !exists {
			report.add(SeverityBreaking, oldType.Name, "type was removed")
			continue
		}
		checkMessage(report, oldType.Name, oldType, newType)
	}

	return report
}

func checkService(report *Report, oldSvc, newSvc model.Service) {
	newMethods := make(map[string]model.Method)
	for _, m := range newSvc.Methods {
		newMethods[m.Name] = m
	}

	for _, oldMethod := range oldSvc.Methods {
		newMethod, exists := newMethods[oldMethod.Name]
		if !exists {
			report.add(SeverityBreaking, fmt.Sprintf("%s.%s", oldSvc.Name, oldMethod.Name), "RPC method was removed")
			continue
		}

		reqTarget := fmt.Sprintf("%s.%s", oldSvc.Name, oldMethod.Request.Name)
		checkMessage(report, reqTarget, oldMethod.Request, newMethod.Request)

		respTarget := fmt.Sprintf("%s.%s", oldSvc.Name, oldMethod.Response.Name)
		checkMessage(report, respTarget, oldMethod.Response, newMethod.Response)
	}
}

func checkMessage(report *Report, target string, oldMsg, newMsg model.Message) {
	oldFieldsByName := make(map[string]model.Field)
	oldFieldsByTag := make(map[int]model.Field)
	for _, f := range oldMsg.Fields {
		oldFieldsByName[f.Name] = f
		oldFieldsByTag[f.Tag] = f
	}

	newFieldsByName := make(map[string]model.Field)
	newFieldsByTag := make(map[int]model.Field)
	for _, f := range newMsg.Fields {
		newFieldsByName[f.Name] = f
		newFieldsByTag[f.Tag] = f
	}

	for _, oldField := range oldMsg.Fields {
		newField, foundByName := newFieldsByName[oldField.Name]

		if foundByName {
			if oldField.Tag != newField.Tag {
				report.add(SeverityBreaking, target, "field %q changed tag number %d -> %d", oldField.Name, oldField.Tag, newField.Tag)
			}
			if oldField.Type != newField.Type {
				report.add(SeverityBreaking, target, "field %q changed type %q -> %q", oldField.Name, oldField.Type, newField.Type)
			}
		} else {
			if reusedField, tagInUse := newFieldsByTag[oldField.Tag]; tagInUse {
				if reusedField.Type == oldField.Type {
					report.add(SeverityWarning, target, "field %q was renamed to %q (tag %d and type %q retained)", oldField.Name, reusedField.Name, oldField.Tag, oldField.Type)
				} else {
					report.add(SeverityBreaking, target, "deleted tag %d (field %q) was reused by field %q with different type %q", oldField.Tag, oldField.Name, reusedField.Name, reusedField.Type)
				}
			} else {
				report.add(SeverityBreaking, target, "field %q (tag %d) was removed", oldField.Name, oldField.Tag)
			}
		}
	}
}
