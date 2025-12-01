package constants

const (
	URLCreateStaticWallet = "/v1/wallet"
	URLCreateInvoice      = "/v1/payment"
	URLGenerateQr         = "/v1/wallet/qr"
	URLListOfServices     = "/v1/payment/services"
	URLPaymentHistory     = "/v1/payment/list"
	URLPaymentInformation = "/v1/payment/info"
	URLRefund             = "/v1/payment/refund"
	URLRefundBlocked      = "/v1/wallet/blocked-address-refund"
	URLResendWebhook      = "/v1/payment/resend"
	URLTestingWebhook     = "/v1/test-webhook/payment"
	URLBlockStaticWallet  = "/v1/wallet/block-address"

	// withdrawals
	URLCreateWithdrawal      = "/v1/payout"
	URLWithdrawalInformation = "/v1/payout/info"
	URLCalculateWithdraw     = "v1/payout/calc"
	URLWithdrawalHistory     = "/v1/payout/list"
)
