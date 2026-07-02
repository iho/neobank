package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

)

type ReferralInvite struct {
	ID            string
	InviterUserID string
	InviteCode    string
	InviteeUserID string
	Status        string
	CreatedAt     time.Time
	AcceptedAt    *time.Time
}

type ReferralInviteRepository interface {
	Create(ctx context.Context, inviterUserID, inviteCode string) (*ReferralInvite, error)
	ListByInviter(ctx context.Context, inviterUserID string, limit int) ([]ReferralInvite, error)
	GetByCode(ctx context.Context, code string) (*ReferralInvite, error)
	Accept(ctx context.Context, code, inviteeUserID string) (*ReferralInvite, error)
}

type CreateReferralInviteUseCase struct {
	repo ReferralInviteRepository
}

func NewCreateReferralInviteUseCase(repo ReferralInviteRepository) *CreateReferralInviteUseCase {
	return &CreateReferralInviteUseCase{repo: repo}
}

func (uc *CreateReferralInviteUseCase) Execute(ctx context.Context, inviterUserID string) (*ReferralInvite, error) {
	if inviterUserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	code, err := generateInviteCode()
	if err != nil {
		return nil, err
	}
	return uc.repo.Create(ctx, inviterUserID, code)
}

type ListReferralInvitesUseCase struct {
	repo ReferralInviteRepository
}

func NewListReferralInvitesUseCase(repo ReferralInviteRepository) *ListReferralInvitesUseCase {
	return &ListReferralInvitesUseCase{repo: repo}
}

func (uc *ListReferralInvitesUseCase) Execute(ctx context.Context, inviterUserID string, limit int) ([]ReferralInvite, error) {
	if inviterUserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if limit <= 0 {
		limit = 20
	}
	return uc.repo.ListByInviter(ctx, inviterUserID, limit)
}

type AcceptReferralInviteUseCase struct {
	repo ReferralInviteRepository
}

func NewAcceptReferralInviteUseCase(repo ReferralInviteRepository) *AcceptReferralInviteUseCase {
	return &AcceptReferralInviteUseCase{repo: repo}
}

func (uc *AcceptReferralInviteUseCase) Execute(ctx context.Context, code, inviteeUserID string) error {
	if code == "" || inviteeUserID == "" {
		return fmt.Errorf("invite_code and user_id are required")
	}
	invite, err := uc.repo.GetByCode(ctx, code)
	if err != nil {
		return fmt.Errorf("invalid invite code")
	}
	if invite.InviterUserID == inviteeUserID {
		return fmt.Errorf("cannot accept your own invite")
	}
	if invite.Status == "accepted" {
		return nil
	}
	_, err = uc.repo.Accept(ctx, code, inviteeUserID)
	return err
}

func generateInviteCode() (string, error) {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}