# Примеры использования Heleket Go SDK

## Содержание

- [Базовая настройка](#базовая-настройка)
- [Платежи](#платежи)
- [Выводы](#выводы)
- [Вебхуки](#вебхуки)
- [Продвинутые примеры](#продвинутые-примеры)

## Базовая настройка

### Простая инициализация

```go
package main

import (
    "log"
    "os"
    
    "github.com/heleket/go-sdk/internal/facade"
    "github.com/heleket/go-sdk/internal/transport"
)

func initClient() *facade.Facade {
    client := facade.NewFacade("https://api.heleket.com")
    
    auth, err := transport.NewAPIKeyAuth(
        os.Getenv("HELEKET_PAYMENT_KEY"),
        os.Getenv("HELEKET_MERCHANT_ID"),
        os.Getenv("HELEKET_WITHDRAWAL_KEY"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    client.SetAuthProvider(auth)
    return client
}

func main() {
    client := initClient()
    
    // Проверяем доступные сервисы
    services, err := client.ListOfServices()
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("SDK инициализирован. Доступно сервисов: %d", len(services.Result))
}
```

### Инициализация с кастомным HTTP клиентом

```go
package main

import (
    "log"
    "net/http"
    "time"
    
    "github.com/heleket/go-sdk/internal/facade"
    "github.com/heleket/go-sdk/internal/transport"
)

func main() {
    // Кастомный HTTP клиент с таймаутами
    httpClient := &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            MaxIdleConns:        100,
            MaxIdleConnsPerHost: 100,
            IdleConnTimeout:     90 * time.Second,
        },
    }
    
    client := facade.NewFacadeWith(
        "https://api.heleket.com",
        httpClient,
        nil, // используем JSON адаптер по умолчанию
        nil,
        nil,
    )
    
    auth, _ := transport.NewAPIKeyAuth(
        "payment_key",
        "merchant_id",
        "withdrawal_key",
    )
    
    client.SetAuthProvider(auth)
}
```

## Платежи

### Пример 1: Создание счета с минимальными параметрами

```go
package main

import (
    "log"
    
    "github.com/heleket/go-sdk/pkg/models/payment"
)

func createSimpleInvoice(client *facade.Facade) {
    req := &payment.CreateInvoiceRequest{
        Amount:            "50.00",
        Currency:          "USD",
        ToCurrency:        "USDT",
        Network:           "tron",
        OrderID:           "ORDER_001",
        UrlCallback:       "https://myshop.com/webhook",
        UrlSuccess:        "https://myshop.com/success",
        UrlReturn:         "https://myshop.com/cart",
        IsRefresh:         false,
        IsPaymentMultiple: false,
        Subtract:          0,
    }
    
    resp, err := client.CreateInvoice(req)
    if err != nil {
        log.Fatalf("Ошибка создания счета: %v", err)
    }
    
    if resp.State != 1 {
        log.Fatalf("API ошибка: %s, детали: %+v", resp.Message, resp.Errors)
    }
    
    log.Printf("✅ Счет создан!")
    log.Printf("   UUID: %s", resp.Result.Uuid)
    log.Printf("   Ссылка для оплаты: %s", resp.Result.Url)
    log.Printf("   Срок действия до: %d", resp.Result.ExpiredAt)
}
```

### Пример 2: Счет с дополнительными параметрами

```go
func createAdvancedInvoice(client *facade.Facade, userID string, productName string) {
    req := &payment.CreateInvoiceRequest{
        Amount:                 "100.00",
        Currency:               "USD",
        ToCurrency:             "USDT",
        Network:                "tron",
        OrderID:                fmt.Sprintf("ORDER_%d", time.Now().Unix()),
        UrlCallback:            "https://myshop.com/webhook",
        UrlSuccess:             "https://myshop.com/success",
        UrlReturn:              "https://myshop.com/cart",
        Lifetime:               1800, // 30 минут
        IsRefresh:              true,
        IsPaymentMultiple:      false,
        Subtract:               0,
        AccuracyPaymentPercent: 5, // допустимое отклонение 5%
        DiscountPercent:        10, // скидка 10%
        ExpectCurrencies:       []string{"USDT", "BTC", "ETH"},
        Currencies:             []string{"USDT"},
        CourseSource:           "binance",
        FromReferralCode:       "REF123",
        AdditionalData: map[string]interface{}{
            "user_id":      userID,
            "product_name": productName,
            "quantity":     1,
            "source":       "web",
        },
    }
    
    resp, err := client.CreateInvoice(req)
    if err != nil {
        log.Fatalf("Ошибка: %v", err)
    }
    
    if resp.State == 1 {
        log.Printf("Счет создан: %s", resp.Result.Uuid)
        log.Printf("Скидка применена: %s", resp.Result.Discount)
    }
}
```

### Пример 3: Статический кошелек для регулярных платежей

```go
func createStaticWallet(client *facade.Facade, userID string) string {
    req := &payment.CreateStaticWalletRequest{
        Currency:         "USDT",
        Network:          "tron",
        OrderID:          fmt.Sprintf("USER_%s_WALLET", userID),
        UrlCallback:      "https://myshop.com/webhook/static",
        FromReferralCode: "",
    }
    
    resp, err := client.CreateStaticWallet(req)
    if err != nil {
        log.Fatalf("Ошибка создания кошелька: %v", err)
    }
    
    if resp.State != 1 {
        log.Fatalf("API ошибка: %s", resp.Message)
    }
    
    log.Printf("✅ Статический кошелек создан!")
    log.Printf("   Адрес: %s", resp.Result.Address)
    log.Printf("   Сеть: %s", resp.Result.Network)
    log.Printf("   UUID кошелька: %s", resp.Result.WalletUuid)
    log.Printf("   Ссылка: %s", resp.Result.Url)
    
    // Сохраните wallet_uuid для последующей блокировки
    return resp.Result.WalletUuid
}
```

### Пример 4: Генерация QR-кода для мобильного приложения

```go
import (
    "encoding/base64"
    "os"
)

func generateAndSaveQR(client *facade.Facade, paymentUUID string) {
    req := &payment.GenerateQrRequest{
        Uuid: paymentUUID,
        Size: 500, // 500x500 пикселей
    }
    
    resp, err := client.GenerateQr(req)
    if err != nil {
        log.Fatalf("Ошибка генерации QR: %v", err)
    }
    
    if resp.State != 1 {
        log.Fatalf("API ошибка: %s", resp.Message)
    }
    
    // QR код в base64, можно сохранить или отправить клиенту
    qrData := resp.Result.QrCode
    
    // Декодируем и сохраняем в файл
    decoded, err := base64.StdEncoding.DecodeString(qrData)
    if err != nil {
        log.Fatalf("Ошибка декодирования: %v", err)
    }
    
    err = os.WriteFile("payment_qr.png", decoded, 0644)
    if err != nil {
        log.Fatalf("Ошибка сохранения: %v", err)
    }
    
    log.Printf("✅ QR код сохранен в payment_qr.png")
}
```

### Пример 5: Проверка статуса платежа

```go
func checkPaymentStatus(client *facade.Facade, uuid string) {
    req := &payment.InformationRequest{
        Uuid: uuid,
    }
    
    resp, err := client.PaymentInformation(req)
    if err != nil {
        log.Fatalf("Ошибка: %v", err)
    }
    
    if resp.State != 1 {
        log.Fatalf("API ошибка: %s", resp.Message)
    }
    
    info := resp.Result
    
    log.Printf("Информация о платеже %s:", uuid)
    log.Printf("   Статус: %s", info.PaymentStatus)
    log.Printf("   Сумма: %s %s", info.Amount, info.Currency)
    log.Printf("   Адрес: %s", info.Address)
    log.Printf("   Сеть: %s", info.Network)
    log.Printf("   TXID: %v", info.Txid)
    log.Printf("   Создан: %s", info.CreatedAt)
    log.Printf("   Обновлен: %s", info.UpdatedAt)
    
    if info.IsFinal {
        log.Printf("   ✅ Платеж завершен")
    } else {
        log.Printf("   ⏳ Платеж в обработке")
    }
}
```

### Пример 6: Получение истории платежей

```go
func getPaymentHistory(client *facade.Facade) {
    req := &payment.HistoryRequest{
        Limit:  100,
        Offset: 0,
        Status: "success",
        From:   "2024-01-01 00:00:00",
        To:     "2024-12-31 23:59:59",
    }
    
    resp, err := client.PaymentHistory(req)
    if err != nil {
        log.Fatalf("Ошибка: %v", err)
    }
    
    if resp.State != 1 {
        log.Fatalf("API ошибка: %s", resp.Message)
    }
    
    log.Printf("Найдено платежей: %d из %d", 
        len(resp.Result.Data), resp.Result.Total)
    
    for i, p := range resp.Result.Data {
        log.Printf("%d. %s - %s %s [%s]", 
            i+1, p.Uuid, p.Amount, p.Currency, p.PaymentStatus)
    }
}
```

### Пример 7: Возврат средств

```go
func refundPayment(client *facade.Facade, paymentUUID, address, network string) {
    req := &payment.RefundRequest{
        Uuid:    paymentUUID,
        Address: address,
        Network: network,
    }
    
    resp, err := client.Refund(req)
    if err != nil {
        log.Fatalf("Ошибка возврата: %v", err)
    }
    
    if resp.State != 1 {
        log.Fatalf("API ошибка: %s, детали: %+v", resp.Message, resp.Errors)
    }
    
    log.Printf("✅ Возврат инициирован")
    log.Printf("   Статус: %s", resp.Result.Status)
}
```

## Выводы

### Пример 8: Создание вывода с проверкой комиссии

```go
func createWithdrawalWithFeeCheck(client *facade.Facade) {
    amount := 100.0
    currency := "USDT"
    network := "tron"
    address := "TYourAddressHere"
    
    // Сначала проверяем комиссию
    feeReq := &withdrawal.CalculateRequest{
        Amount:   amount,
        Currency: currency,
        Network:  network,
    }
    
    feeResp, err := client.CalculateWithdrawalFee(feeReq)
    if err != nil {
        log.Fatalf("Ошибка расчета комиссии: %v", err)
    }
    
    if feeResp.State != 1 {
        log.Fatalf("Ошибка API: %s", feeResp.Message)
    }
    
    log.Printf("Комиссия за вывод: %+v", feeResp.Result)
    
    // Создаем вывод
    withdrawReq := &withdrawal.CreateWithdrawalRequest{
        Amount:      amount,
        Currency:    currency,
        Network:     network,
        ToAddress:   address,
        OrderID:     fmt.Sprintf("WD_%d", time.Now().Unix()),
        UrlCallback: "https://myshop.com/webhook/withdrawal",
        Metadata: map[string]interface{}{
            "type":   "user_payout",
            "reason": "monthly_earnings",
        },
    }
    
    withdrawResp, err := client.CreateWithdrawal(withdrawReq)
    if err != nil {
        log.Fatalf("Ошибка создания вывода: %v", err)
    }
    
    if withdrawResp.State != 1 {
        log.Fatalf("API ошибка: %s, детали: %+v", 
            withdrawResp.Message, withdrawResp.Errors)
    }
    
    log.Printf("✅ Вывод создан!")
    log.Printf("   UUID: %s", withdrawResp.Result.Uuid)
    log.Printf("   Статус: %s", withdrawResp.Result.Status)
    log.Printf("   Сумма: %s %s", withdrawResp.Result.Amount, withdrawResp.Result.Currency)
    log.Printf("   Адрес: %s", withdrawResp.Result.ToAddress)
}
```

### Пример 9: Проверка статуса вывода

```go
func checkWithdrawalStatus(client *facade.Facade, uuid string) {
    req := &withdrawal.InformationRequest{
        Uuid: uuid,
    }
    
    resp, err := client.WithdrawalInformation(req)
    if err != nil {
        log.Fatalf("Ошибка: %v", err)
    }
    
    if resp.State != 1 {
        log.Fatalf("API ошибка: %s", resp.Message)
    }
    
    w := resp.Result
    log.Printf("Статус вывода %s:", uuid)
    log.Printf("   Статус: %s", w.Status)
    log.Printf("   Сумма: %s %s", w.Amount, w.Currency)
    log.Printf("   Адрес: %s", w.ToAddress)
    log.Printf("   Сеть: %s", w.Network)
}
```

### Пример 10: История выводов

```go
func getWithdrawalHistory(client *facade.Facade) {
    req := &withdrawal.HistoryRequest{
        Limit:  50,
        Offset: 0,
        Status: "completed",
        From:   "2024-01-01",
        To:     "2024-12-31",
    }
    
    resp, err := client.WithdrawalHistory(req)
    if err != nil {
        log.Fatalf("Ошибка: %v", err)
    }
    
    if resp.State != 1 {
        log.Fatalf("API ошибка: %s", resp.Message)
    }
    
    log.Printf("История выводов: %d записей", resp.Result.Total)
    
    for i, w := range resp.Result.Data {
        log.Printf("%d. %s - %s %s [%s]", 
            i+1, w.Uuid, w.Amount, w.Currency, w.Status)
    }
}
```

## Вебхуки

### Пример 11: Обработчик вебхуков для платежей

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
    "os"
)

type PaymentWebhook struct {
    Uuid          string      `json:"uuid"`
    OrderId       string      `json:"order_id"`
    Amount        string      `json:"amount"`
    Currency      string      `json:"currency"`
    Network       string      `json:"network"`
    Address       string      `json:"address"`
    Txid          string      `json:"txid"`
    PaymentStatus string      `json:"payment_status"`
    IsFinal       bool        `json:"is_final"`
    AdditionalData interface{} `json:"additional_data"`
}

func verifyWebhookSignature(body []byte, receivedSign string) bool {
    apiKey := os.Getenv("HELEKET_PAYMENT_KEY")
    encoded := base64.StdEncoding.EncodeToString(body)
    signString := encoded + apiKey
    hash := md5.Sum([]byte(signString))
    expectedSign := fmt.Sprintf("%x", hash)
    
    return receivedSign == expectedSign
}

func paymentWebhookHandler(w http.ResponseWriter, r *http.Request) {
    // Читаем тело запроса
    body, err := io.ReadAll(r.Body)
    if err != nil {
        log.Printf("Ошибка чтения тела: %v", err)
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()
    
    // Проверяем подпись
    receivedSign := r.Header.Get("sign")
    merchantId := r.Header.Get("merchant")
    
    if !verifyWebhookSignature(body, receivedSign) {
        log.Printf("Неверная подпись вебхука от merchant: %s", merchantId)
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // Парсим данные
    var webhook PaymentWebhook
    if err := json.Unmarshal(body, &webhook); err != nil {
        log.Printf("Ошибка парсинга JSON: %v", err)
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    // Обрабатываем вебхук
    log.Printf("Получен вебхук платежа:")
    log.Printf("   UUID: %s", webhook.Uuid)
    log.Printf("   Order ID: %s", webhook.OrderId)
    log.Printf("   Статус: %s", webhook.PaymentStatus)
    log.Printf("   Сумма: %s %s", webhook.Amount, webhook.Currency)
    log.Printf("   TXID: %s", webhook.Txid)
    log.Printf("   Final: %v", webhook.IsFinal)
    
    // Обрабатываем в зависимости от статуса
    switch webhook.PaymentStatus {
    case "success":
        if webhook.IsFinal {
            // Платеж успешно завершен
            log.Printf("✅ Платеж %s успешно завершен", webhook.OrderId)
            // TODO: Обновите заказ в вашей БД
            // TODO: Отправьте уведомление пользователю
        }
    case "pending":
        log.Printf("⏳ Платеж %s в обработке", webhook.OrderId)
    case "failed":
        log.Printf("❌ Платеж %s не удался", webhook.OrderId)
        // TODO: Обработайте неудачный платеж
    }
    
    // Возвращаем успешный ответ
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func main() {
    http.HandleFunc("/webhook/payment", paymentWebhookHandler)
    
    log.Println("Сервер вебхуков запущен на :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Пример 12: Обработчик вебхуков для выводов

```go
type WithdrawalWebhook struct {
    Uuid      string                 `json:"uuid"`
    OrderId   string                 `json:"order_id"`
    Amount    string                 `json:"amount"`
    Currency  string                 `json:"currency"`
    Network   string                 `json:"network"`
    ToAddress string                 `json:"to_address"`
    Txid      string                 `json:"txid"`
    Status    string                 `json:"status"`
    Metadata  map[string]interface{} `json:"metadata"`
}

func withdrawalWebhookHandler(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()
    
    // Проверяем подпись (используем withdrawal api key)
    receivedSign := r.Header.Get("sign")
    
    // Для вебхуков вывода используется другой ключ
    apiKey := os.Getenv("HELEKET_WITHDRAWAL_KEY")
    encoded := base64.StdEncoding.EncodeToString(body)
    signString := encoded + apiKey
    hash := md5.Sum([]byte(signString))
    expectedSign := fmt.Sprintf("%x", hash)
    
    if receivedSign != expectedSign {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    var webhook WithdrawalWebhook
    if err := json.Unmarshal(body, &webhook); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    
    log.Printf("Вебхук вывода:")
    log.Printf("   UUID: %s", webhook.Uuid)
    log.Printf("   Order ID: %s", webhook.OrderId)
    log.Printf("   Статус: %s", webhook.Status)
    log.Printf("   Сумма: %s %s", webhook.Amount, webhook.Currency)
    log.Printf("   TXID: %s", webhook.Txid)
    
    switch webhook.Status {
    case "completed":
        log.Printf("✅ Вывод %s успешно выполнен", webhook.OrderId)
        // TODO: Обновите баланс пользователя
    case "processing":
        log.Printf("⏳ Вывод %s в обработке", webhook.OrderId)
    case "failed":
        log.Printf("❌ Вывод %s не удался", webhook.OrderId)
        // TODO: Верните средства на баланс пользователя
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

## Продвинутые примеры

### Пример 13: Полноценный сервис приема платежей

```go
package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "os"
    
    "github.com/heleket/go-sdk/internal/facade"
    "github.com/heleket/go-sdk/internal/transport"
    "github.com/heleket/go-sdk/pkg/models/payment"
)

type PaymentService struct {
    client *facade.Facade
    db     *sql.DB
}

func NewPaymentService(db *sql.DB) *PaymentService {
    client := facade.NewFacade("https://api.heleket.com")
    
    auth, err := transport.NewAPIKeyAuth(
        os.Getenv("HELEKET_PAYMENT_KEY"),
        os.Getenv("HELEKET_MERCHANT_ID"),
        os.Getenv("HELEKET_WITHDRAWAL_KEY"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    client.SetAuthProvider(auth)
    
    return &PaymentService{
        client: client,
        db:     db,
    }
}

func (s *PaymentService) CreateOrder(orderID string, amount string, userID string) (string, error) {
    // Создаем счет
    req := &payment.CreateInvoiceRequest{
        Amount:            amount,
        Currency:          "USD",
        ToCurrency:        "USDT",
        Network:           "tron",
        OrderID:           orderID,
        UrlCallback:       "https://myshop.com/webhook",
        UrlSuccess:        "https://myshop.com/success",
        UrlReturn:         "https://myshop.com/cart",
        Lifetime:          3600,
        IsRefresh:         false,
        IsPaymentMultiple: false,
        Subtract:          0,
        AdditionalData: map[string]interface{}{
            "user_id": userID,
        },
    }
    
    resp, err := s.client.CreateInvoice(req)
    if err != nil {
        return "", err
    }
    
    if resp.State != 1 {
        return "", fmt.Errorf("API error: %s", resp.Message)
    }
    
    // Сохраняем в БД
    _, err = s.db.Exec(`
        INSERT INTO payments (uuid, order_id, amount, currency, status, payment_url, user_id)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `, resp.Result.Uuid, orderID, amount, "USD", "pending", resp.Result.Url, userID)
    
    if err != nil {
        return "", err
    }
    
    return resp.Result.Url, nil
}

func (s *PaymentService) HandleWebhook(w http.ResponseWriter, r *http.Request) {
    // ... проверка подписи ...
    
    var webhook PaymentWebhook
    // ... парсинг ...
    
    if webhook.IsFinal && webhook.PaymentStatus == "success" {
        // Обновляем статус в БД
        _, err := s.db.Exec(`
            UPDATE payments 
            SET status = ?, txid = ?, updated_at = NOW()
            WHERE uuid = ?
        `, webhook.PaymentStatus, webhook.Txid, webhook.Uuid)
        
        if err != nil {
            log.Printf("Ошибка обновления БД: %v", err)
            http.Error(w, "Internal error", http.StatusInternalServerError)
            return
        }
        
        // Активируем заказ
        // TODO: Ваша бизнес-логика
        
        log.Printf("Платеж %s успешно обработан", webhook.Uuid)
    }
    
    w.WriteHeader(http.StatusOK)
}
```

### Пример 14: Мониторинг платежей

```go
func monitorPayments(client *facade.Facade) {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        // Получаем платежи за последние 10 минут
        from := time.Now().Add(-10 * time.Minute).Format("2006-01-02 15:04:05")
        to := time.Now().Format("2006-01-02 15:04:05")
        
        req := &payment.HistoryRequest{
            Limit:  100,
            Offset: 0,
            From:   from,
            To:     to,
        }
        
        resp, err := client.PaymentHistory(req)
        if err != nil {
            log.Printf("Ошибка получения истории: %v", err)
            continue
        }
        
        if resp.State != 1 {
            log.Printf("API ошибка: %s", resp.Message)
            continue
        }
        
        // Анализируем статусы
        pending := 0
        success := 0
        failed := 0
        
        for _, p := range resp.Result.Data {
            switch p.PaymentStatus {
            case "pending":
                pending++
            case "success":
                success++
            case "failed":
                failed++
            }
        }
        
        log.Printf("📊 Статистика платежей:")
        log.Printf("   Успешных: %d", success)
        log.Printf("   В обработке: %d", pending)
        log.Printf("   Неудачных: %d", failed)
        log.Printf("   Всего: %d", resp.Result.Total)
    }
}
```

### Пример 15: Повторная отправка неудачных вебхуков

```go
func resendFailedWebhooks(client *facade.Facade, paymentUUIDs []string) {
    for _, uuid := range paymentUUIDs {
        req := &payment.ResendWebhookRequest{
            Uuid: uuid,
        }
        
        resp, err := client.ResendWebhook(req)
        if err != nil {
            log.Printf("❌ Ошибка отправки вебхука для %s: %v", uuid, err)
            continue
        }
        
        if resp.State == 1 {
            log.Printf("✅ Вебхук для %s отправлен повторно", uuid)
        } else {
            log.Printf("⚠️  Не удалось отправить вебхук для %s: %s", uuid, resp.Message)
        }
        
        // Задержка между запросами
        time.Sleep(1 * time.Second)
    }
}
```

## Обработка ошибок

### Пример 16: Правильная обработка ошибок

```go
func safeCreateInvoice(client *facade.Facade, req *payment.CreateInvoiceRequest) (*payment.CreateInvoiceResponse, error) {
    resp, err := client.CreateInvoice(req)
    
    // Сетевая или другая критическая ошибка
    if err != nil {
        return nil, fmt.Errorf("network error: %w", err)
    }
    
    // Проверка кода состояния API
    if resp.State != 1 {
        // Логируем детали ошибки
        if resp.Errors != nil {
            for field, errMsg := range resp.Errors {
                log.Printf("Ошибка в поле %s: %v", field, errMsg)
            }
        }
        
        return nil, fmt.Errorf("API error: %s (state: %d)", resp.Message, resp.State)
    }
    
    return resp, nil
}
```

## Заключение

Эти примеры покрывают большинство сценариев использования Heleket Go SDK. Для получения дополнительной информации обратитесь к [официальной документации API](https://doc.heleket.com/ru/).

