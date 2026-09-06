package dto

import "github.com/mudflap-autobotz/payment-transaction-service/internal/domain"

type MerchantSummaryResponse struct {
	Balance              string `json:"balance"`
	Reserved             string `json:"reserved"`
	TodayDepositTotal    string `json:"today_deposit_total"`
	TodayDepositCount    int64  `json:"today_deposit_count"`
	TodayWithdrawalTotal string `json:"today_withdrawal_total"`
	TodayWithdrawalCount int64  `json:"today_withdrawal_count"`
	PendingDepositCount  int64  `json:"pending_deposit_count"`
}

func ToMerchantSummaryResponse(summary *domain.MerchantSummary) MerchantSummaryResponse {
	return MerchantSummaryResponse{
		Balance:              summary.Balance,
		Reserved:             summary.Reserved,
		TodayDepositTotal:    summary.TodayDepositTotal,
		TodayDepositCount:    summary.TodayDepositCount,
		TodayWithdrawalTotal: summary.TodayWithdrawalTotal,
		TodayWithdrawalCount: summary.TodayWithdrawalCount,
		PendingDepositCount:  summary.PendingDepositCount,
	}
}
