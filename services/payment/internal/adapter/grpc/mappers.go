package grpc

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"
	neobankv1 "github.com/iho/neobank/pkg/gen/neobank/v1"
	"github.com/iho/neobank/services/payment/internal/gen/api"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func timeRFC3339(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func transferToProto(t api.Transfer) *neobankv1.Transfer {
	out := &neobankv1.Transfer{
		Id:              t.Id.String(),
		Status:          t.Status,
		SenderUserId:    t.SenderUserId.String(),
		RecipientUserId: t.RecipientUserId.String(),
		Amount:          t.Amount,
		Currency:        t.Currency,
	}
	if t.LedgerTransferId != nil {
		out.LedgerTransferId = *t.LedgerTransferId
	}
	if t.FailureReason != nil {
		out.FailureReason = *t.FailureReason
	}
	if t.Memo != nil {
		out.Memo = *t.Memo
	}
	if t.CreatedAt != nil {
		out.CreatedAt = timeRFC3339(t.CreatedAt)
	}
	if t.CompletedAt != nil {
		out.CompletedAt = timeRFC3339(t.CompletedAt)
	}
	return out
}

func limitsToProto(l api.LimitsResponse) *neobankv1.TransferLimits {
	return &neobankv1.TransferLimits{
		HourlyTransferCount: &neobankv1.LimitGauge{
			Limit:     l.P2p.HourlyTransferCount.Limit,
			Used:      l.P2p.HourlyTransferCount.Used,
			Remaining: l.P2p.HourlyTransferCount.Remaining,
		},
		DailyTransferAmount: &neobankv1.LimitGauge{
			Limit:     l.P2p.DailyTransferAmount.Limit,
			Used:      l.P2p.DailyTransferAmount.Used,
			Remaining: l.P2p.DailyTransferAmount.Remaining,
		},
		SingleTransferMax: l.P2p.SingleTransferMax,
	}
}

func mapCreateP2PTransferResponse(resp api.CreateP2PTransferResponseObject) (*neobankv1.TransferResponse, error) {
	switch r := resp.(type) {
	case api.CreateP2PTransfer200JSONResponse:
		return &neobankv1.TransferResponse{
			Transfer:   transferToProto(api.Transfer(r)),
			HttpStatus: 200,
		}, nil
	case api.CreateP2PTransfer201JSONResponse:
		return &neobankv1.TransferResponse{
			Transfer:   transferToProto(api.Transfer(r)),
			HttpStatus: 201,
		}, nil
	case api.CreateP2PTransfer400JSONResponse:
		return &neobankv1.TransferResponse{HttpStatus: 400, Error: r.Error}, nil
	case api.CreateP2PTransfer422JSONResponse:
		return &neobankv1.TransferResponse{
			Transfer:   transferToProto(api.Transfer(r)),
			HttpStatus: 422,
		}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected create transfer response")
	}
}

func mapGetTransferResponse(resp api.GetTransferResponseObject) (*neobankv1.TransferResponse, error) {
	switch r := resp.(type) {
	case api.GetTransfer200JSONResponse:
		return &neobankv1.TransferResponse{
			Transfer:   transferToProto(api.Transfer(r)),
			HttpStatus: 200,
		}, nil
	case api.GetTransfer404JSONResponse:
		return &neobankv1.TransferResponse{HttpStatus: 404, Error: r.Error}, nil
	case api.GetTransfer401JSONResponse:
		return &neobankv1.TransferResponse{HttpStatus: 401, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected get transfer response")
	}
}

func mapListTransfersResponse(resp api.ListTransfersResponseObject) (*neobankv1.ListTransfersResponse, error) {
	switch r := resp.(type) {
	case api.ListTransfers200JSONResponse:
		transfers := make([]*neobankv1.Transfer, 0, len(r.Transfers))
		for _, t := range r.Transfers {
			transfers = append(transfers, transferToProto(t))
		}
		out := &neobankv1.ListTransfersResponse{
			Transfers:  transfers,
			HttpStatus: 200,
		}
		if r.NextCursor != nil {
			out.NextCursor = *r.NextCursor
		}
		return out, nil
	case api.ListTransfers401JSONResponse:
		return &neobankv1.ListTransfersResponse{HttpStatus: 401, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected list transfers response")
	}
}

func mapGetLimitsResponse(resp api.GetLimitsResponseObject) (*neobankv1.LimitsResponse, error) {
	switch r := resp.(type) {
	case api.GetLimits200JSONResponse:
		return &neobankv1.LimitsResponse{
			P2P:        limitsToProto(api.LimitsResponse(r)),
			HttpStatus: 200,
		}, nil
	case api.GetLimits401JSONResponse:
		return &neobankv1.LimitsResponse{HttpStatus: 401, Error: r.Error}, nil
	default:
		return nil, status.Error(codes.Internal, "unexpected get limits response")
	}
}

func parseUserID(userID string) (openapi_types.UUID, error) {
	return parseUUID(userID)
}