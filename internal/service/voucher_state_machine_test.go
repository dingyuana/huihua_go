package service

import (
	"testing"

	"huihua/finance/internal/model"
)

// ─── ValidateTransition ───

func TestValidateTransition_Draft_Submit(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(0, model.VoucherActionSubmit); err != nil {
		t.Errorf("draft -> submit should be valid: %v", err)
	}
}

func TestValidateTransition_Draft_Cancel(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(0, model.VoucherActionCancel); err != nil {
		t.Errorf("draft -> cancel should be valid: %v", err)
	}
}

func TestValidateTransition_Draft_Approve_Invalid(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(0, model.VoucherActionApprove); err == nil {
		t.Error("draft -> approve should be invalid")
	}
}

func TestValidateTransition_Draft_Reverse_Invalid(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(0, model.VoucherActionReverse); err == nil {
		t.Error("draft -> reverse should be invalid")
	}
}

func TestValidateTransition_Draft_Reject_Invalid(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(0, model.VoucherActionReject); err == nil {
		t.Error("draft -> reject should be invalid")
	}
}

func TestValidateTransition_Posted_Approve(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(1, model.VoucherActionApprove); err != nil {
		t.Errorf("posted -> approve should be valid: %v", err)
	}
}

func TestValidateTransition_Posted_Reject(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(1, model.VoucherActionReject); err != nil {
		t.Errorf("posted -> reject should be valid: %v", err)
	}
}

func TestValidateTransition_Posted_Reverse(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(1, model.VoucherActionReverse); err != nil {
		t.Errorf("posted -> reverse should be valid: %v", err)
	}
}

func TestValidateTransition_Posted_Submit_Invalid(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(1, model.VoucherActionSubmit); err == nil {
		t.Error("posted -> submit should be invalid")
	}
}

func TestValidateTransition_Posted_Cancel_Invalid(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(1, model.VoucherActionCancel); err == nil {
		t.Error("posted -> cancel should be invalid")
	}
}

func TestValidateTransition_Verified_Reverse(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(2, model.VoucherActionReverse); err != nil {
		t.Errorf("verified -> reverse should be valid: %v", err)
	}
}

func TestValidateTransition_Verified_Submit_Invalid(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(2, model.VoucherActionSubmit); err == nil {
		t.Error("verified -> submit should be invalid")
	}
}

func TestValidateTransition_Verified_Approve_Invalid(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(2, model.VoucherActionApprove); err == nil {
		t.Error("verified -> approve should be invalid")
	}
}

func TestValidateTransition_Verified_Reject_Invalid(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(2, model.VoucherActionReject); err == nil {
		t.Error("verified -> reject should be invalid")
	}
}

func TestValidateTransition_Cancelled_AnyAction_Invalid(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	actions := []model.VoucherAction{
		model.VoucherActionSubmit,
		model.VoucherActionApprove,
		model.VoucherActionReject,
		model.VoucherActionCancel,
		model.VoucherActionReverse,
	}
	for _, action := range actions {
		if err := sm.ValidateTransition(3, action); err == nil {
			t.Errorf("cancelled -> %s should be invalid", action)
		}
	}
}

func TestValidateTransition_UnknownStatus(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransition(99, model.VoucherActionSubmit); err == nil {
		t.Error("unknown status should produce error")
	}
}

func TestValidateTransition_AllDraftActions(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	tests := []struct {
		action model.VoucherAction
		valid  bool
	}{
		{model.VoucherActionSubmit, true},
		{model.VoucherActionCancel, true},
		{model.VoucherActionApprove, false},
		{model.VoucherActionReject, false},
		{model.VoucherActionReverse, false},
	}
	for _, tt := range tests {
		err := sm.ValidateTransition(0, tt.action)
		if tt.valid && err != nil {
			t.Errorf("draft -> %s should be valid, got: %v", tt.action, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("draft -> %s should be invalid", tt.action)
		}
	}
}

func TestValidateTransition_AllPostedActions(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	tests := []struct {
		action model.VoucherAction
		valid  bool
	}{
		{model.VoucherActionSubmit, false},
		{model.VoucherActionCancel, false},
		{model.VoucherActionApprove, true},
		{model.VoucherActionReject, true},
		{model.VoucherActionReverse, true},
	}
	for _, tt := range tests {
		err := sm.ValidateTransition(1, tt.action)
		if tt.valid && err != nil {
			t.Errorf("posted -> %s should be valid, got: %v", tt.action, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("posted -> %s should be invalid", tt.action)
		}
	}
}

func TestValidateTransition_AllVerifiedActions(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	tests := []struct {
		action model.VoucherAction
		valid  bool
	}{
		{model.VoucherActionSubmit, false},
		{model.VoucherActionCancel, false},
		{model.VoucherActionApprove, false},
		{model.VoucherActionReject, false},
		{model.VoucherActionReverse, true},
	}
	for _, tt := range tests {
		err := sm.ValidateTransition(2, tt.action)
		if tt.valid && err != nil {
			t.Errorf("verified -> %s should be valid, got: %v", tt.action, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("verified -> %s should be invalid", tt.action)
		}
	}
}

// ─── ValidateTransitionStatusString ───

func TestValidateTransitionStatusString_Draft(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransitionStatusString("draft", model.VoucherActionSubmit); err != nil {
		t.Errorf("draft -> submit should be valid: %v", err)
	}
}

func TestValidateTransitionStatusString_Unknown(t *testing.T) {
	sm := NewVoucherStateMachine(nil, nil, nil)
	if err := sm.ValidateTransitionStatusString("unknown_status", model.VoucherActionSubmit); err == nil {
		t.Error("unknown status string should produce error")
	}
}
