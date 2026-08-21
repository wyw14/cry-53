package domain

type ApprovalDenial string

const (
	ApprovalAllowed          ApprovalDenial = ""
	ApprovalRoleMissing      ApprovalDenial = "role_missing"
	ApprovalStateInvalid     ApprovalDenial = "state_invalid"
	ApprovalIssuesPresent    ApprovalDenial = "issues_present"
	ApprovalDutyNotSeparated ApprovalDenial = "duty_not_separated"
)

type ApprovalRequest struct {
	State      BundleState
	Issues     []Issue
	UploadedBy string
	Actor      Actor
}

type ApprovalDecision struct {
	Allowed bool
	Denial  ApprovalDenial
}

func DecideBundleApproval(request ApprovalRequest) ApprovalDecision {
	checks := []func(ApprovalRequest) ApprovalDenial{
		requireApprovalRole,
		requireValidatedState,
		requireCleanValidation,
		requireSeparatedDuties,
	}
	for _, check := range checks {
		if denial := check(request); denial != ApprovalAllowed {
			return ApprovalDecision{Denial: denial}
		}
	}
	return ApprovalDecision{Allowed: true}
}

func requireApprovalRole(request ApprovalRequest) ApprovalDenial {
	if request.Actor.HasRole("approver") {
		return ApprovalAllowed
	}
	return ApprovalRoleMissing
}

func requireValidatedState(request ApprovalRequest) ApprovalDenial {
	if request.State == BundleValidated {
		return ApprovalAllowed
	}
	return ApprovalStateInvalid
}

func requireCleanValidation(request ApprovalRequest) ApprovalDenial {
	for _, issue := range request.Issues {
		if issue.Severity == SeverityError {
			return ApprovalIssuesPresent
		}
	}
	return ApprovalAllowed
}

func requireSeparatedDuties(request ApprovalRequest) ApprovalDenial {
	if request.UploadedBy != request.Actor.ID {
		return ApprovalAllowed
	}
	if request.Actor.HasRole("approver") {
		return ApprovalAllowed
	}
	return ApprovalDutyNotSeparated
}

func (d ApprovalDecision) Error() error {
	switch d.Denial {
	case ApprovalRoleMissing, ApprovalDutyNotSeparated:
		return ErrForbidden
	case ApprovalStateInvalid, ApprovalIssuesPresent:
		return ErrInvalidTransition
	default:
		return nil
	}
}
