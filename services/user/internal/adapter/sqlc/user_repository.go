package sqlcrepo

import (
	"context"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
)

type UserRepository struct {
	q sqlc.Querier
}

func NewUserRepository(q sqlc.Querier) *UserRepository {
	return &UserRepository{q: q}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	id, err := pgutil.ParseUUID(user.ID)
	if err != nil {
		return err
	}
	return r.q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           id,
		Email:        user.Email,
		Phone:        user.Phone,
		PasswordHash: user.PasswordHash,
		Status:       string(user.Status),
	})
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return rowToUser(row.ID, row.Email, row.Phone, row.PasswordHash, row.Status), nil
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	row, err := r.q.GetUserByPhone(ctx, pgutil.Text(phone))
	if err != nil {
		return nil, err
	}
	return rowToUser(row.ID, row.Email, row.Phone, row.PasswordHash, row.Status), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	uid, err := pgutil.ParseUUID(id)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetUserByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return rowToUser(row.ID, row.Email, row.Phone, row.PasswordHash, row.Status), nil
}

func rowToUser(id uuid.UUID, email, phone, passwordHash, status string) *domain.User {
	return &domain.User{
		ID:           id.String(),
		Email:        email,
		Phone:        phone,
		PasswordHash: passwordHash,
		Status:       domain.UserStatus(status),
	}
}

type WalletRepository struct {
	q sqlc.Querier
}

func NewWalletRepository(q sqlc.Querier) *WalletRepository {
	return &WalletRepository{q: q}
}

func (r *WalletRepository) Create(ctx context.Context, wallet domain.Wallet) error {
	id, err := pgutil.ParseUUID(wallet.ID)
	if err != nil {
		return err
	}
	userID, err := pgutil.ParseUUID(wallet.UserID)
	if err != nil {
		return err
	}
	return r.q.CreateWallet(ctx, sqlc.CreateWalletParams{
		ID:              id,
		UserID:          userID,
		Currency:        wallet.Currency,
		LedgerAccountID: wallet.LedgerAccountID,
		Status:          wallet.Status,
	})
}

func (r *WalletRepository) GetByUserAndCurrency(ctx context.Context, userID, currency string) (*domain.Wallet, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetWalletByUserAndCurrency(ctx, sqlc.GetWalletByUserAndCurrencyParams{
		UserID:   uid,
		Currency: currency,
	})
	if err != nil {
		return nil, err
	}
	return &domain.Wallet{
		ID:              row.ID.String(),
		UserID:          row.UserID.String(),
		Currency:        row.Currency,
		LedgerAccountID: row.LedgerAccountID,
		Status:          row.Status,
	}, nil
}

