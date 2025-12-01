# Быстрый старт с Heleket Go SDK

## 1. Установка

```bash
go get github.com/heleket/go-sdk
```

## 2. Получите API ключи

Зарегистрируйтесь на [Heleket](https://heleket.com) и получите:
- `PAYMENT_API_KEY` - для платежных операций
- `MERCHANT_ID` - ваш идентификатор мерчанта
- `WITHDRAWAL_API_KEY` - для операций вывода

## 3. Настройте переменные окружения

```bash
export HELEKET_PAYMENT_KEY="ваш_payment_api_key"
export HELEKET_MERCHANT_ID="ваш_merchant_id"
export HELEKET_WITHDRAWAL_KEY="ваш_withdrawal_api_key"
```

Или создайте файл `.env`:

```env
HELEKET_PAYMENT_KEY=ваш_payment_api_key
HELEKET_MERCHANT_ID=ваш_merchant_id
HELEKET_WITHDRAWAL_KEY=ваш_withdrawal_api_key
```

## 4. Базовое использование

### Создание клиента

```go
package main

import (
    "log"
    "os"
    
    "github.com/heleket/go-sdk/internal/facade"
    "github.com/heleket/go-sdk/internal/transport"
)

func main() {
    // Инициализация клиента
    client := facade.NewFacade("https://api.heleket.com")
    
    // Настройка аутентификации
    auth, err := transport.NewAPIKeyAuth(
        os.Getenv("HELEKET_PAYMENT_KEY"),
        os.Getenv("HELEKET_MERCHANT_ID"),
        os.Getenv("HELEKET_WITHDRAWAL_KEY"),
    )
    if err != nil {
        log.Fatal(err)
    }
    
    client.SetAuthProvider(auth)
    
    // Теперь можно использовать клиент
}
```

## 5. Создайте первый счет

```go
package main

import (
    "log"
    "os"
    
    "github.com/heleket/go-sdk/internal/facade"
    "github.com/heleket/go-sdk/internal/transport"
    "github.com/heleket/go-sdk/pkg/models/payment"
)

func main() {
    // Инициализация
    client := facade.NewFacade("https://api.heleket.com")
    auth, _ := transport.NewAPIKeyAuth(
        os.Getenv("HELEKET_PAYMENT_KEY"),
        os.Getenv("HELEKET_MERCHANT_ID"),
        os.Getenv("HELEKET_WITHDRAWAL_KEY"),
    )
    client.SetAuthProvider(auth)
    
    // Создание счета
    invoice, err := client.CreateInvoice(&payment.CreateInvoiceRequest{
        Amount:            "100.00",      // Сумма
        Currency:          "USD",          // Валюта счета
        ToCurrency:        "USDT",         // Валюта оплаты
        Network:           "tron",         // Сеть блокчейна
        OrderID:           "ORDER_001",    // Ваш уникальный ID заказа
        UrlCallback:       "https://yoursite.com/webhook",
        UrlSuccess:        "https://yoursite.com/success",
        UrlReturn:         "https://yoursite.com/cancel",
        Lifetime:          3600,           // Время жизни в секундах
        IsRefresh:         false,
        IsPaymentMultiple: false,
        Subtract:          0,
    })
    
    if err != nil {
        log.Fatalf("Ошибка: %v", err)
    }
    
    if invoice.State != 1 {
        log.Fatalf("API ошибка: %s", invoice.Message)
    }
    
    // Получаем ссылку для оплаты
    log.Printf("✅ Счет создан!")
    log.Printf("Ссылка для оплаты: %s", invoice.Result.Url)
    log.Printf("UUID платежа: %s", invoice.Result.Uuid)
}
```

## 6. Обработка вебхуков

Создайте обработчик для получения уведомлений о платежах:

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
    Uuid          string `json:"uuid"`
    OrderId       string `json:"order_id"`
    Amount        string `json:"amount"`
    Currency      string `json:"currency"`
    PaymentStatus string `json:"payment_status"`
    IsFinal       bool   `json:"is_final"`
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
    // Читаем тело запроса
    body, _ := io.ReadAll(r.Body)
    defer r.Body.Close()
    
    // Проверяем подпись
    receivedSign := r.Header.Get("sign")
    apiKey := os.Getenv("HELEKET_PAYMENT_KEY")
    
    encoded := base64.StdEncoding.EncodeToString(body)
    hash := md5.Sum([]byte(encoded + apiKey))
    expectedSign := fmt.Sprintf("%x", hash)
    
    if receivedSign != expectedSign {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }
    
    // Парсим данные
    var webhook PaymentWebhook
    json.Unmarshal(body, &webhook)
    
    // Обрабатываем платеж
    if webhook.IsFinal && webhook.PaymentStatus == "success" {
        log.Printf("✅ Платеж успешен: %s (заказ %s)", 
            webhook.Uuid, webhook.OrderId)
        // TODO: Обновите статус заказа в вашей БД
    }
    
    w.WriteHeader(http.StatusOK)
}

func main() {
    http.HandleFunc("/webhook", webhookHandler)
    log.Println("Сервер запущен на :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## 7. Создание вывода средств

```go
package main

import (
    "log"
    "os"
    
    "github.com/heleket/go-sdk/internal/facade"
    "github.com/heleket/go-sdk/internal/transport"
    "github.com/heleket/go-sdk/pkg/models/withdrawal"
)

func main() {
    // Инициализация
    client := facade.NewFacade("https://api.heleket.com")
    auth, _ := transport.NewAPIKeyAuth(
        os.Getenv("HELEKET_PAYMENT_KEY"),
        os.Getenv("HELEKET_MERCHANT_ID"),
        os.Getenv("HELEKET_WITHDRAWAL_KEY"),
    )
    client.SetAuthProvider(auth)
    
    // Создание вывода
    withdrawal, err := client.CreateWithdrawal(&withdrawal.CreateWithdrawalRequest{
        Amount:      100.0,
        Currency:    "USDT",
        Network:     "tron",
        ToAddress:   "TYourTronAddressHere",
        OrderID:     "PAYOUT_001",
        UrlCallback: "https://yoursite.com/webhook/withdrawal",
    })
    
    if err != nil {
        log.Fatalf("Ошибка: %v", err)
    }
    
    if withdrawal.State != 1 {
        log.Fatalf("API ошибка: %s", withdrawal.Message)
    }
    
    log.Printf("✅ Вывод создан!")
    log.Printf("UUID: %s", withdrawal.Result.Uuid)
    log.Printf("Статус: %s", withdrawal.Result.Status)
}
```

## 8. Проверка статуса платежа

```go
// Получение информации о платеже
info, err := client.PaymentInformation(&payment.InformationRequest{
    Uuid: "payment_uuid_здесь",
})

if err != nil {
    log.Fatal(err)
}

log.Printf("Статус: %s", info.Result.PaymentStatus)
log.Printf("Сумма: %s %s", info.Result.Amount, info.Result.Currency)
```

## 9. Получение списка доступных криптовалют

```go
services, err := client.ListOfServices()
if err != nil {
    log.Fatal(err)
}

for _, service := range services.Result {
    log.Printf("%s (%s) - комиссия: %s", 
        service.Currency, service.Network, service.Fee)
}
```

## Полезные ссылки

- 📖 [Полная документация](README.md)
- 💡 [Примеры кода](EXAMPLES.md)
- 🌐 [API документация](https://doc.heleket.com/ru/)
- 🔧 [Репозиторий на GitHub](https://github.com/heleket/go-sdk)

## Основные методы

### Платежи
- `CreateInvoice()` - Создание счета
- `CreateStaticWallet()` - Создание постоянного кошелька
- `GenerateQr()` - Генерация QR-кода
- `PaymentInformation()` - Информация о платеже
- `PaymentHistory()` - История платежей
- `Refund()` - Возврат средств
- `ListOfServices()` - Доступные криптовалюты

### Выводы
- `CreateWithdrawal()` - Создание вывода
- `WithdrawalInformation()` - Информация о выводе
- `WithdrawalHistory()` - История выводов
- `CalculateWithdrawalFee()` - Расчет комиссии

## Поддержка

При возникновении вопросов:
1. Проверьте [примеры](EXAMPLES.md)
2. Изучите [документацию API](https://doc.heleket.com/ru/)
3. Создайте issue на GitHub

## Важные замечания

⚠️ **Безопасность:**
- Никогда не коммитьте API ключи в репозиторий
- Используйте переменные окружения
- Всегда проверяйте подпись вебхуков

✅ **Лучшие практики:**
- Обрабатывайте ошибки API
- Логируйте все операции
- Используйте таймауты для HTTP клиента
- Сохраняйте UUID платежей в БД

🔐 **Подпись запросов:**
```
sign = md5(base64(json_body) + api_key)
```

Для платежей используется `PAYMENT_API_KEY`, для выводов - `WITHDRAWAL_API_KEY`.

