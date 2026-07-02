package sqlcrepo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/iho/neobank/pkg/pgutil"
	"github.com/iho/neobank/pkg/piicrypto"
	"github.com/iho/neobank/services/user/internal/domain"
	"github.com/iho/neobank/services/user/internal/gen/sqlc"
	"github.com/iho/neobank/services/user/internal/port"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	q   sqlc.Querier
	pii piicrypto.Protector
}

func NewUserRepository(q sqlc.Querier, pii piicrypto.Protector) *UserRepository {
	if pii == nil {
		pii = piicrypto.NewNoop()
	}
	return &UserRepository{q: q, pii: pii}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	id, err := pgutil.ParseUUID(user.ID)
	if err != nil {
		return err
	}
	phoneStored, err := piicrypto.Store(ctx, r.pii, user.Phone)
	if err != nil {
		return err
	}
	phoneLookup, err := r.pii.PhoneLookup(ctx, user.Phone)
	if err != nil {
		return err
	}
	return r.q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           id,
		Email:        user.Email,
		Phone:        phoneStored,
		PhoneLookup:  textOrNil(phoneLookup),
		PasswordHash: user.PasswordHash,
		Status:       string(user.Status),
	})
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return r.rowToUser(ctx, row.ID, row.Email, row.Phone, row.PasswordHash, row.Status)
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	lookup, err := r.pii.PhoneLookup(ctx, phone)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetUserByPhone(ctx, pgutil.Text(lookup))
	if err != nil {
		return nil, err
	}
	return r.rowToUser(ctx, row.ID, row.Email, row.Phone, row.PasswordHash, row.Status)
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
	return r.rowToUser(ctx, row.ID, row.Email, row.Phone, row.PasswordHash, row.Status)
}

func (r *UserRepository) UpdatePasswordHash(ctx context.Context, userID, passwordHash string) error {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return err
	}
	return r.q.UpdatePasswordHash(ctx, sqlc.UpdatePasswordHashParams{
		ID:           uid,
		PasswordHash: passwordHash,
	})
}

func (r *UserRepository) GetProfile(ctx context.Context, userID string) (*domain.Profile, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	row, err := r.q.GetUserProfile(ctx, uid)
	if err != nil {
		return nil, err
	}
	phone, err := piicrypto.Read(ctx, r.pii, row.Phone)
	if err != nil {
		return nil, err
	}
	out := &domain.Profile{
		UserID:      row.ID.String(),
		Email:       row.Email,
		Phone:       phone,
		Status:      row.Status,
		FullName:    row.FullName,
		CountryCode: row.CountryCode,
		KYCStatus:   row.KycStatus,
	}
	if row.CreatedAt.Valid {
		out.CreatedAt = row.CreatedAt.Time
	}
	dob, err := r.readDOB(ctx, row.DateOfBirth, row.DateOfBirthEncrypted)
	if err != nil {
		return nil, err
	}
	out.DateOfBirth = dob
	return out, nil
}

func (r *UserRepository) readDOB(ctx context.Context, plain pgtype.Date, encrypted string) (string, error) {
	if encrypted != "" {
		return piicrypto.Read(ctx, r.pii, encrypted)
	}
	if plain.Valid {
		return plain.Time.Format(time.DateOnly), nil
	}
	return "", nil
}

func (r *UserRepository) rowToUser(ctx context.Context, id uuid.UUID, email, phone, passwordHash, status string) (*domain.User, error) {
	plainPhone, err := piicrypto.Read(ctx, r.pii, phone)
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:           id.String(),
		Email:        email,
		Phone:        plainPhone,
		PasswordHash: passwordHash,
		Status:       domain.UserStatus(status),
	}, nil
}

type WalletRepository struct {
	q sqlc.Querier
}

func NewWalletRepository(q sqlc.Querier) *WalletRepository {
	return &WalletRepository{q: q}
}

func (r *WalletRepository) WithTx(tx pgx.Tx) port.WalletRepository {
	return &WalletRepository{q: withTx(r.q, tx)}
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

func (r *WalletRepository) DeleteByID(ctx context.Context, walletID string) error {
	id, err := pgutil.ParseUUID(walletID)
	if err != nil {
		return err
	}
	return r.q.DeleteWalletByID(ctx, id)
}

func (r *WalletRepository) ListByUser(ctx context.Context, userID string) ([]domain.Wallet, error) {
	uid, err := pgutil.ParseUUID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListWalletsByUserID(ctx, uid)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Wallet, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.Wallet{
			ID:              row.ID.String(),
			UserID:          row.UserID.String(),
			Currency:        row.Currency,
			LedgerAccountID: row.LedgerAccountID,
			Status:          row.Status,
		})
	}
	return out, nil
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