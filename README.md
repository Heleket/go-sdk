# Heleket Go SDK

Go SDK для работы с API платежного сервиса [Heleket](https://doc.heleket.com/ru/).

## Установка

```bash
go get github.com/heleket/go-sdk
```

## Быстрый старт

```go
package main

import (
    "log"
    
    "github.com/heleket/go-sdk/internal/facade"
    "github.com/heleket/go-sdk/internal/transport"
    "github.com/heleket/go-sdk/pkg/models/payment"
)

func main() {
    // Создание клиента
    client := facade.NewFacade("https://api.heleket.com")
    
    // Настройка аутентификации
    auth, err := transport.NewAPIKeyAuth(
        "your_payment_api_key",    // API ключ для платежей
        "your_merchant_id",         // ID мерчанта
        "your_withdrawal_api_key",  // API ключ для выводов
    )
    if err != nil {
        log.Fatal(err)
    }
    
    client.SetAuthProvider(auth)
    
    // Теперь можно использовать SDK
    services, err := client.ListOfServices()
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Доступные сервисы: %+v", services)
}
```

## Инициализация

### Базовая инициализация

```go
client := facade.NewFacade("https://api.heleket.com")
```

### Инициализация с кастомными настройками

```go
import (
    "net/http"
    "time"
)

httpClient := &http.Client{
    Timeout: 30 * time.Second,
}

client := facade.NewFacadeWith(
    "https://api.heleket.com",
    httpClient,
    nil,  // RequestAdapter (nil = использовать JSON по умолчанию)
    nil,  // ResponseAdapter (nil = использовать JSON по умолчанию)
    nil,  // AuthProvider (можно установить позже)
)
```

## Аутентификация

SDK использует подпись запросов по алгоритму:
```
sign = md5(base64(json_body) + api_key)
```

Создание провайдера аутентификации:

```go
auth, err := transport.NewAPIKeyAuth(
    "payment_api_key",      // Ключ для платежных операций
    "merchant_id",          // ID мерчанта
    "withdrawal_api_key",   // Ключ для операций вывода
)
if err != nil {
    log.Fatal(err)
}

client.SetAuthProvider(auth)
```

**Важно:** Для платежных операций используется `payment_api_key`, для операций вывода - `withdrawal_api_key`.

## Платежные операции

### Создание статического кошелька

Создает постоянный адрес для приема платежей:

```go
req := &payment.CreateStaticWalletRequest{
    Currency:         "USDT",
    Network:          "tron",
    OrderID:          "order_12345",
    UrlCallback:      "https://yoursite.com/webhook",
    FromReferralCode: "", // опционально
}

resp, err := client.CreateStaticWallet(req)
if err != nil {
    log.Fatal(err)
}

log.Printf("Адрес кошелька: %s", resp.Result.Address)
log.Printf("Wallet UUID: %s", resp.Result.WalletUuid)
log.Printf("URL для оплаты: %s", resp.Result.Url)
```

### Создание счета на оплату

Создает одноразовый счет с ограниченным временем жизни:

```go
req := &payment.CreateInvoiceRequest{
    Amount:                 "100.50",
    Currency:               "USD",
    ToCurrency:             "USDT",
    Network:                "tron",
    OrderID:                "invoice_12345",
    UrlCallback:            "https://yoursite.com/webhook",
    UrlSuccess:             "https://yoursite.com/success",
    UrlReturn:              "https://yoursite.com/return",
    Lifetime:               3600, // время жизни в секундах
    IsRefresh:              false,
    IsPaymentMultiple:      false,
    Subtract:               0,
    AccuracyPaymentPercent: 0,
    ExpectCurrencies:       []string{"USDT", "BTC"},
    Currencies:             []string{"USDT"},
    DiscountPercent:        0,
    CourseSource:           "",
    FromReferralCode:       "",
    AdditionalData:         map[string]interface{}{
        "user_id": "123",
        "product": "Premium subscription",
    },
}

resp, err := client.CreateInvoice(req)
if err != nil {
    log.Fatal(err)
}

log.Printf("UUID счета: %s", resp.Result.Uuid)
log.Printf("URL для оплаты: %s", resp.Result.Url)
log.Printf("Истекает: %d", resp.Result.ExpiredAt)
```

### Генерация QR-кода

Создает QR-код для платежа:

```go
req := &payment.GenerateQrRequest{
    Uuid: "payment_uuid_from_invoice",
    Size: 300, // размер QR-кода в пикселях
}

resp, err := client.GenerateQr(req)
if err != nil {
    log.Fatal(err)
}

// resp.Result.QrCode содержит base64 изображение
log.Printf("QR код: %s", resp.Result.QrCode)
```

### Получение списка доступных сервисов

```go
resp, err := client.ListOfServices()
if err != nil {
    log.Fatal(err)
}

for _, service := range resp.Result {
    log.Printf("Валюта: %s, Сеть: %s, Комиссия: %s", 
        service.Currency, service.Network, service.Fee)
}
```

### История платежей

Получение истории платежей с фильтрацией:

```go
req := &payment.HistoryRequest{
    Limit:  50,
    Offset: 0,
    Status: "success", // опционально: success, pending, failed
    From:   "2024-01-01 00:00:00", // опционально
    To:     "2024-12-31 23:59:59", // опционально
}

resp, err := client.PaymentHistory(req)
if err != nil {
    log.Fatal(err)
}

for _, payment := range resp.Result.Data {
    log.Printf("Платеж %s: %s %s", payment.Uuid, payment.Amount, payment.Currency)
}
```

### Информация о платеже

Получение детальной информации о конкретном платеже:

```go
req := &payment.InformationRequest{
    Uuid: "payment_uuid",
}

resp, err := client.PaymentInformation(req)
if err != nil {
    log.Fatal(err)
}

log.Printf("Статус: %s", resp.Result.PaymentStatus)
log.Printf("Сумма: %s %s", resp.Result.Amount, resp.Result.Currency)
log.Printf("Адрес: %s", resp.Result.Address)
```

### Возврат средств

Возврат средств по платежу:

```go
req := &payment.RefundRequest{
    Uuid:    "payment_uuid",
    Address: "recipient_address",
    Network: "tron",
}

resp, err := client.Refund(req)
if err != nil {
    log.Fatal(err)
}

log.Printf("Возврат создан: %s", resp.Result.Status)
```

### Возврат заблокированных средств

Возврат средств, заблокированных за нарушение лимитов:

```go
req := &payment.RefundBlockedRequest{
    Uuid:    "payment_uuid",
    Address: "recipient_address",
}

resp, err := client.RefundBlocked(req)
if err != nil {
    log.Fatal(err)
}

log.Printf("Статус возврата: %s", resp.Result.Status)
```

### Блокировка статического кошелька

```go
req := &payment.BlockStaticWalletRequest{
    WalletUuid: "wallet_uuid",
}

resp, err := client.BlockStaticWallet(req)
if err != nil {
    log.Fatal(err)
}

log.Printf("Кошелек заблокирован: %v", resp.Result)
```

### Повторная отправка вебхука

```go
req := &payment.ResendWebhookRequest{
    Uuid: "payment_uuid",
}

resp, err := client.ResendWebhook(req)
if err != nil {
    log.Fatal(err)
}

log.Printf("Вебхук отправлен: %v", resp.Result)
```

### Тестирование вебхука

```go
req := &payment.TestingWebhookRequest{
    Url: "https://yoursite.com/webhook",
}

resp, err := client.TestingWebhook(req)
if err != nil {
    log.Fatal(err)
}

log.Printf("Результат теста: %v", resp.Result)
```

## Операции вывода средств

### Создание заявки на вывод

```go
req := &withdrawal.CreateWithdrawalRequest{
    Amount:      100.50,
    Currency:    "USDT",
    Network:     "tron",
    ToAddress:   "TRX_ADDRESS_HERE",
    OrderID:     "withdrawal_12345",
    UrlCallback: "https://yoursite.com/webhook",
    Metadata: map[string]interface{}{
        "user_id": "123",
        "note":    "Monthly payout",
    },
}

resp, err := client.CreateWithdrawal(req)
if err != nil {
    log.Fatal(err)
}

log.Printf("Вывод создан: %s", resp.Result.Uuid)
log.Printf("Статус: %s", resp.Result.Status)
```

### Информация о выводе

```go
req := &withdrawal.InformationRequest{
    Uuid: "withdrawal_uuid",
}

resp, err := client.WithdrawalInformation(req)
if err != nil {
    log.Fatal(err)
}

log.Printf("Статус вывода: %s", resp.Result.Status)
log.Printf("Сумма: %s %s", resp.Result.Amount, resp.Result.Currency)
```

### История выводов

```go
req := &withdrawal.HistoryRequest{
    Limit:  50,
    Offset: 0,
    Status: "completed", // опционально
    From:   "2024-01-01", // опционально
    To:     "2024-12-31", // опционально
}

resp, err := client.WithdrawalHistory(req)
if err != nil {
    log.Fatal(err)
}

for _, w := range resp.Result.Data {
    log.Printf("Вывод %s: %s %s", w.Uuid, w.Amount, w.Currency)
}
```

### Расчет комиссии за вывод

```go
req := &withdrawal.CalculateRequest{
    Amount:   100.00,
    Currency: "USDT",
    Network:  "tron",
}

resp, err := client.CalculateWithdrawalFee(req)
if err != nil {
    log.Fatal(err)
}

log.Printf("Комиссия: %v", resp.Result)
```

## Обработка вебхуков

SDK автоматически проверяет подпись вебхуков от Heleket. Пример обработчика:

```go
package main

import (
    "crypto/md5"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
)

type WebhookPayload struct {
    Uuid          string `json:"uuid"`
    OrderId       string `json:"order_id"`
    Amount        string `json:"amount"`
    Currency      string `json:"currency"`
    PaymentStatus string `json:"payment_status"`
    // ... другие поля
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    // Читаем тело запроса
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Получаем подпись из заголовка
    receivedSign := r.Header.Get("sign")
    merchantId := r.Header.Get("merchant")
    
    // Проверяем подпись
    apiKey := "your_payment_api_key" // или withdrawal_api_key в зависимости от типа webhook
    encoded := base64.StdEncoding.EncodeToString(body)
    signString := encoded + apiKey
    hash := md5.Sum([]byte(signString))
    expectedSign := fmt.Sprintf("%x", hash)
    
    if receivedSign != expectedSign {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // Парсим данные
    var payload WebhookPayload
    if err := json.Unmarshal(body, &payload); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Обрабатываем вебхук
    log.Printf("Получен платеж: %s, статус: %s", payload.Uuid, payload.PaymentStatus)
    
    // Возвращаем успешный ответ
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func main() {
    http.HandleFunc("/webhook", webhookHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Обработка ошибок

SDK возвращает детализированные ошибки:

```go
resp, err := client.CreateInvoice(req)
if err != nil {
    log.Printf("Ошибка при создании счета: %v", err)
    return
}

// Проверка статуса ответа
if resp.State != 1 {
    log.Printf("API вернул ошибку: %s", resp.Message)
    log.Printf("Детали ошибок: %+v", resp.Errors)
    return
}

// Успешная обработка
log.Printf("Счет создан: %s", resp.Result.Uuid)
```

## Примеры использования

### Полный пример приема платежа

```go
package main

import (
    "log"
    
    "github.com/heleket/go-sdk/internal/facade"
    "github.com/heleket/go-sdk/internal/transport"
    "github.com/heleket/go-sdk/pkg/models/payment"
)

func main() {
    // Инициализация
    client := facade.NewFacade("https://api.heleket.com")
    auth, err := transport.NewAPIKeyAuth(
        "payment_key",
        "merchant_id",
        "withdrawal_key",
    )
    if err != nil {
        log.Fatal(err)
    }
    client.SetAuthProvider(auth)
    
    // Создание счета
    invoice, err := client.CreateInvoice(&payment.CreateInvoiceRequest{
        Amount:            "100",
        Currency:          "USD",
        ToCurrency:        "USDT",
        Network:           "tron",
        OrderID:           "ORDER123",
        UrlCallback:       "https://mysite.com/webhook",
        UrlSuccess:        "https://mysite.com/success",
        UrlReturn:         "https://mysite.com/cancel",
        Lifetime:          3600,
        IsRefresh:         false,
        IsPaymentMultiple: false,
        Subtract:          0,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Счет создан!")
    log.Printf("UUID: %s", invoice.Result.Uuid)
    log.Printf("Направьте клиента на: %s", invoice.Result.Url)
    
    // Генерация QR-кода
    qr, err := client.GenerateQr(&payment.GenerateQrRequest{
        Uuid: invoice.Result.Uuid,
        Size: 300,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("QR код: %s", qr.Result.QrCode)
}
```

### Полный пример вывода средств

```go
package main

import (
    "log"
    
    "github.com/heleket/go-sdk/internal/facade"
    "github.com/heleket/go-sdk/internal/transport"
    "github.com/heleket/go-sdk/pkg/models/withdrawal"
)

func main() {
    // Инициализация
    client := facade.NewFacade("https://api.heleket.com")
    auth, err := transport.NewAPIKeyAuth(
        "payment_key",
        "merchant_id",
        "withdrawal_key",
    )
    if err != nil {
        log.Fatal(err)
    }
    client.SetAuthProvider(auth)
    
    // Сначала рассчитываем комиссию
    fee, err := client.CalculateWithdrawalFee(&withdrawal.CalculateRequest{
        Amount:   100.0,
        Currency: "USDT",
        Network:  "tron",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Комиссия за вывод: %+v", fee.Result)
    
    // Создаем заявку на вывод
    w, err := client.CreateWithdrawal(&withdrawal.CreateWithdrawalRequest{
        Amount:      100.0,
        Currency:    "USDT",
        Network:     "tron",
        ToAddress:   "TRX_ADDRESS",
        OrderID:     "PAYOUT123",
        UrlCallback: "https://mysite.com/webhook",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Вывод создан!")
    log.Printf("UUID: %s", w.Result.Uuid)
    log.Printf("Статус: %s", w.Result.Status)
}
```

## Тестирование

Для тестирования можно использовать тестовое окружение Heleket.

## Безопасность

- **Никогда не коммитьте API ключи** в систему контроля версий
- Используйте переменные окружения для хранения ключей:

```go
import "os"

auth, err := transport.NewAPIKeyAuth(
    os.Getenv("HELEKET_PAYMENT_KEY"),
    os.Getenv("HELEKET_MERCHANT_ID"),
    os.Getenv("HELEKET_WITHDRAWAL_KEY"),
)
```

- Всегда проверяйте подпись вебхуков
- Используйте HTTPS для вебхук URL

## Поддержка

- Документация API: https://doc.heleket.com/ru/
- GitHub Issues: https://github.com/heleket/go-sdk/issues

## Лицензия

MIT License

