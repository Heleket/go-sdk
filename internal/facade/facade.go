package facade

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	pkgtransport "github.com/heleket/go-sdk/pkg/transport"

	internaltransport "github.com/heleket/go-sdk/internal/transport"
	"github.com/heleket/go-sdk/pkg/constants"
	"github.com/heleket/go-sdk/pkg/models/payment"
	"github.com/heleket/go-sdk/pkg/models/withdrawal"
)

type Facade struct {
	BaseURL         string
	Client          *http.Client
	requestAdapter  pkgtransport.RequestAdapter
	responseAdapter pkgtransport.ResponseAdapter
	auth            pkgtransport.AuthProvider
}

func NewFacade(baseURL string) *Facade {
	return &Facade{
		BaseURL:         baseURL,
		Client:          &http.Client{},
		requestAdapter:  internaltransport.JSONAdapter{},
		responseAdapter: internaltransport.JSONAdapter{},
	}
}

// Доп. конструктор с внедрением зависимостей
func NewFacadeWith(baseURL string, client *http.Client, req pkgtransport.RequestAdapter, resp pkgtransport.ResponseAdapter, auth pkgtransport.AuthProvider) *Facade {
	if client == nil {
		client = &http.Client{}
	}
	if req == nil {
		req = internaltransport.JSONAdapter{}
	}
	if resp == nil {
		resp = internaltransport.JSONAdapter{}
	}
	return &Facade{
		BaseURL:         baseURL,
		Client:          client,
		requestAdapter:  req,
		responseAdapter: resp,
		auth:            auth,
	}
}

func (f *Facade) SetRequestAdapter(adapter pkgtransport.RequestAdapter) {
	f.requestAdapter = adapter
}

func (f *Facade) SetResponseAdapter(adapter pkgtransport.ResponseAdapter) {
	f.responseAdapter = adapter
}

func (f *Facade) SetAuthProvider(auth pkgtransport.AuthProvider) {
	f.auth = auth
}

func (f *Facade) doPost(path string, reqBody interface{}, out interface{}, isPayment bool) error {
	data, err := f.requestAdapter.Marshal(reqBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, f.BaseURL+path, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if f.auth != nil {
		if hdrs, err := f.auth.Headers(data, isPayment); err != nil {
			return err
		} else if hdrs != nil {
			for k, vs := range hdrs {
				for _, v := range vs {
					req.Header.Add(k, v)
				}
			}
		}
	}
	resp, err := f.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	return f.responseAdapter.Unmarshal(body, out)
}

func (f *Facade) doGet(path string, out interface{}, isPayment bool) error {
	req, err := http.NewRequest(http.MethodGet, f.BaseURL+path, nil)
	if err != nil {
		return err
	}
	if f.auth != nil {
		if hdrs, err := f.auth.Headers(nil, isPayment); err != nil {
			return err
		} else if hdrs != nil {
			for k, vs := range hdrs {
				for _, v := range vs {
					req.Header.Add(k, v)
				}
			}
		}
	}
	resp, err := f.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}
	return f.responseAdapter.Unmarshal(body, out)
}

func (f *Facade) CreateStaticWallet(req *payment.CreateStaticWalletRequest) (*payment.CreateStaticWalletResponse, error) {
	var result payment.CreateStaticWalletResponse
	if err := f.doPost(constants.URLCreateStaticWallet, req, &result, true); err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *Facade) CreateInvoice(req *payment.CreateInvoiceRequest) (*payment.CreateInvoiceResponse, error) {
	var result payment.CreateInvoiceResponse
	if err := f.doPost(constants.URLCreateInvoice, req, &result, true); err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *Facade) GenerateQr(req *payment.GenerateQrRequest) (*payment.GenerateQrResponse, error) {
	var result payment.GenerateQrResponse
	if err := f.doPost(constants.URLGenerateQr, req, &result, true); err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *Facade) ListOfServices() (*payment.ListOfServicesResponse, error) {
	var result payment.ListOfServicesResponse

	if err := f.doPost(constants.URLListOfServices, struct{}{}, &result, true); err != nil {
		return nil, err
	}

	return &result, nil
}

func (f *Facade) PaymentHistory(req *payment.HistoryRequest) (*payment.HistoryResponse, error) {
	var result payment.HistoryResponse
	if err := f.doPost(constants.URLPaymentHistory, req, &result, true); err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *Facade) PaymentInformation(req *payment.InformationRequest) (*payment.InformationResponse, error) {
	var result payment.InformationResponse
	if err := f.doPost(constants.URLPaymentInformation, req, &result, true); err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *Facade) RefundBlocked(req *payment.RefundBlockedRequest) (*payment.RefundBlockedResponse, error) {
	var result payment.RefundBlockedResponse
	if err := f.doPost(constants.URLRefundBlocked, req, &result, true); err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *Facade) Refund(req *payment.RefundRequest) (*payment.RefundResponse, error) {
	var result payment.RefundResponse
	if err := f.doPost(constants.URLRefund, req, &result, true); err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *Facade) ResendWebhook(req *payment.ResendWebhookRequest) (*payment.ResendWebhookResponse, error) {
	var result payment.ResendWebhookResponse
	if err := f.doPost(constants.URLResendWebhook, req, &result, true); err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *Facade) TestingWebhook(req *payment.TestingWebhookRequest) (*payment.TestingWebhookResponse, error) {
	var result payment.TestingWebhookResponse
	if err := f.doPost(constants.URLTestingWebhook, req, &result, true); err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *Facade) BlockStaticWallet(req *payment.BlockStaticWalletRequest) (*payment.BlockStaticWalletResponse, error) {
	var result payment.BlockStaticWalletResponse
	if err := f.doPost(constants.URLBlockStaticWallet, req, &result, true); err != nil {
		return nil, err
	}
	return &result, nil
}

// Withdrawals
func (f *Facade) CreateWithdrawal(req *withdrawal.CreateWithdrawalRequest) (*withdrawal.CreateWithdrawalResponse, error) {
	var result withdrawal.CreateWithdrawalResponse
	if err := f.doPost(constants.URLCreateWithdrawal, req, &result, false); err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *Facade) WithdrawalInformation(req *withdrawal.InformationRequest) (*withdrawal.InformationResponse, error) {
	var result withdrawal.InformationResponse
	if err := f.doPost(constants.URLWithdrawalInformation, req, &result, false); err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *Facade) WithdrawalHistory(req *withdrawal.HistoryRequest) (*withdrawal.HistoryResponse, error) {
	var result withdrawal.HistoryResponse
	if err := f.doPost(constants.URLWithdrawalHistory, req, &result, false); err != nil {
		return nil, err
	}
	return &result, nil
}

func (f *Facade) CalculateWithdrawalFee(req *withdrawal.CalculateRequest) (*withdrawal.CancelWithdrawalResponse, error) {
	var result withdrawal.CancelWithdrawalResponse
	if err := f.doPost(constants.URLCalculateWithdraw, req, &result, false); err != nil {
		return nil, err
	}
	return &result, nil
}
