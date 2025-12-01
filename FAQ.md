# FAQ - Часто задаваемые вопросы

## Общие вопросы

### Как получить API ключи?

1. Зарегистрируйтесь на [heleket.com](https://heleket.com)
2. Войдите в личный кабинет
3. Перейдите в раздел "API"
4. Скопируйте:
   - Payment API Key (для приема платежей)
   - Merchant ID (ваш идентификатор)
   - Withdrawal API Key (для вывода средств)

### Какие криптовалюты поддерживаются?

Получите актуальный список через API:

```go
services, err := client.ListOfServices()
for _, service := range services.Result {
    fmt.Printf("%s в сети %s\n", service.Currency, service.Network)
}
```

Обычно поддерживаются: USDT, BTC, ETH, TRX, USDC и другие.

### Какие блокчейн сети поддерживаются?

- Tron (TRC-20)
- Ethereum (ERC-20)
- Binance Smart Chain (BEP-20)
- И другие - проверьте через `ListOfServices()`

### В какой валюте указывать суммы?

В методе `CreateInvoice` есть два параметра:
- `Currency` - валюта счета (USD, EUR, RUB и т.д.)
- `ToCurrency` - валюта для оплаты (USDT, BTC и т.д.)

Система автоматически конвертирует по текущему курсу.

## Аутентификация

### Почему два разных API ключа?

Для безопасности Heleket разделяет операции:
- **Payment API Key** - для приема платежей (входящие операции)
- **Withdrawal API Key** - для вывода средств (исходящие операции)

Это позволяет ограничить права доступа разных частей системы.

### Как работает подпись запросов?

Каждый запрос подписывается по алгоритму:

```
sign = md5(base64(json_body) + api_key)
```

SDK автоматически подписывает все запросы. Вам не нужно делать это вручную.

### Как проверить подпись вебхука?

```go
import (
    "crypto/md5"
    "encoding/base64"
    "fmt"
)

func verifyWebhook(body []byte, receivedSign, apiKey string) bool {
    encoded := base64.StdEncoding.EncodeToString(body)
    hash := md5.Sum([]byte(encoded + apiKey))
    expectedSign := fmt.Sprintf("%x", hash)
    return receivedSign == expectedSign
}
```

## Платежи

### Как создать простой счет на оплату?

```go
invoice, err := client.CreateInvoice(&payment.CreateInvoiceRequest{
    Amount:            "100.00",
    Currency:          "USD",
    ToCurrency:        "USDT",
    Network:           "tron",
    OrderID:           "ORDER_123",
    UrlCallback:       "https://yoursite.com/webhook",
    UrlSuccess:        "https://yoursite.com/success",
    UrlReturn:         "https://yoursite.com/cancel",
    Lifetime:          3600,
    IsRefresh:         false,
    IsPaymentMultiple: false,
    Subtract:          0,
})
```

### В чем разница между счетом и статическим кошельком?

**Счет (Invoice)**:
- Одноразовый
- Имеет срок действия (Lifetime)
- Конкретная сумма
- Уникальный адрес для каждого платежа

**Статический кошелек**:
- Постоянный адрес
- Без ограничения времени
- Принимает любые суммы
- Один адрес для многих платежей

### Что такое `IsRefresh`?

Если `IsRefresh = true`, клиент может обновить курс валюты на странице оплаты. Полезно для длинных счетов.

### Что такое `IsPaymentMultiple`?

Если `true`, разрешает несколько платежей на один счет. Обычно используется `false`.

### Что такое `Subtract`?

Определяет, кто платит комиссию:
- `0` - комиссию платит плательщик (сумма добавляется к платежу)
- `1` - комиссию платит получатель (вычитается из суммы)

### Как установить время жизни счета?

Параметр `Lifetime` указывается в секундах:

```go
Lifetime: 3600,  // 1 час
Lifetime: 1800,  // 30 минут
Lifetime: 86400, // 24 часа
```

### Как добавить дополнительные данные к платежу?

Используйте поле `AdditionalData`:

```go
AdditionalData: map[string]interface{}{
    "user_id": "12345",
    "product": "Premium subscription",
    "quantity": 1,
    "notes": "Monthly payment",
}
```

Эти данные вернутся в вебхуке.

### Как применить скидку?

```go
DiscountPercent: 10, // 10% скидка
```

### Что такое `AccuracyPaymentPercent`?

Допустимое отклонение суммы платежа в процентах. Например, если установлено `5`, платеж на 95-105% от суммы будет принят.

## Вебхуки

### Когда приходят вебхуки?

Вебхук отправляется при каждом изменении статуса платежа:
- При создании платежа
- При получении транзакции
- При подтверждении транзакции
- При завершении платежа

### Как понять, что платеж завершен?

Проверяйте два поля:

```go
if webhook.PaymentStatus == "success" && webhook.IsFinal {
    // Платеж успешно завершен
}
```

### Какие статусы платежа существуют?

- `pending` - ожидает оплаты
- `processing` - транзакция найдена, идет подтверждение
- `success` - успешно завершен
- `failed` - неудачный
- `expired` - истек срок действия
- `partially_paid` - частичная оплата

### Что делать, если вебхук не пришел?

1. Проверьте доступность вашего URL
2. Используйте `ResendWebhook()`:

```go
client.ResendWebhook(&payment.ResendWebhookRequest{
    Uuid: "payment_uuid",
})
```

3. Или периодически проверяйте статус через `PaymentInformation()`

### Как протестировать вебхук?

```go
client.TestingWebhook(&payment.TestingWebhookRequest{
    Url: "https://yoursite.com/webhook",
})
```

### Нужно ли отвечать на вебхук?

Да, верните HTTP 200 OK:

```go
func webhookHandler(w http.ResponseWriter, r *http.Request) {
    // Обработка...
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

## Выводы

### Как создать вывод средств?

```go
withdrawal, err := client.CreateWithdrawal(&withdrawal.CreateWithdrawalRequest{
    Amount:    100.0,
    Currency:  "USDT",
    Network:   "tron",
    ToAddress: "TYourAddressHere",
    OrderID:   "PAYOUT_123",
})
```

### Как узнать комиссию за вывод?

```go
fee, err := client.CalculateWithdrawalFee(&withdrawal.CalculateRequest{
    Amount:   100.0,
    Currency: "USDT",
    Network:  "tron",
})
```

### Сколько времени занимает вывод?

Обычно 10-30 минут, в зависимости от сети блокчейна и загрузки.

### Можно ли отменить вывод?

Зависит от статуса:
- `pending` - можно отменить
- `processing` - возможно, обратитесь в поддержку
- `completed` - нельзя отменить

### Какие статусы вывода существуют?

- `pending` - создан, ожидает обработки
- `processing` - в обработке
- `completed` - успешно завершен
- `failed` - неудачный
- `cancelled` - отменен

## Ошибки

### Ошибка "Invalid signature"

1. Проверьте API ключ
2. Убедитесь, что используете правильный ключ:
   - Payment Key для платежей
   - Withdrawal Key для выводов
3. Проверьте, что ключ указан без пробелов

### Ошибка "Merchant not found"

Проверьте правильность `MERCHANT_ID`.

### Ошибка "Insufficient balance"

Недостаточно средств для вывода. Пополните баланс мерчанта.

### Ошибка "Invalid address"

Адрес кошелька некорректен:
- Проверьте формат адреса
- Убедитесь, что адрес соответствует выбранной сети

### Ошибка "Currency not supported"

Валюта или сеть не поддерживается. Проверьте список через `ListOfServices()`.

### Ошибка при разборе JSON

```go
if resp.State != 1 {
    log.Printf("API Error: %s", resp.Message)
    log.Printf("Details: %+v", resp.Errors)
}
```

Проверьте поле `Errors` для деталей.

## Производительность

### Сколько запросов можно делать?

Нет жесткого лимита, но рекомендуется не более 100 запросов в минуту.

### Можно ли использовать один клиент для всех запросов?

Да, рекомендуется создать один клиент и переиспользовать:

```go
var globalClient = initClient()

func makePayment() {
    invoice, err := globalClient.CreateInvoice(...)
}
```

### Как установить таймаут?

```go
httpClient := &http.Client{
    Timeout: 30 * time.Second,
}

client := facade.NewFacadeWith(
    "https://api.heleket.com",
    httpClient,
    nil, nil, nil,
)
```

## Безопасность

### Где хранить API ключи?

**Никогда** не храните ключи в коде! Используйте:

1. Переменные окружения:
```bash
export HELEKET_PAYMENT_KEY="your_key"
```

2. Файл .env (добавьте в .gitignore):
```env
HELEKET_PAYMENT_KEY=your_key
HELEKET_MERCHANT_ID=your_id
HELEKET_WITHDRAWAL_KEY=your_key
```

3. Менеджер секретов (Vault, AWS Secrets Manager и т.д.)

### Безопасно ли использовать MD5 для подписи?

Да, для этой цели MD5 достаточно. Подпись используется только для проверки целостности данных, а не для криптографической защиты.

### Нужно ли проверять SSL сертификат?

Да, всегда используйте HTTPS и проверяйте сертификаты. SDK делает это автоматически.

## Тестирование

### Есть ли тестовое окружение?

Проверьте документацию Heleket на наличие sandbox окружения.

### Как тестировать локально?

Для вебхуков используйте:
- ngrok для туннелирования
- локальный сервер с публичным IP
- webhook.site для тестирования

```bash
# Пример с ngrok
ngrok http 8080
# Используйте полученный URL как UrlCallback
```

### Как написать unit-тесты?

Используйте моки:

```go
mockClient := &http.Client{
    Transport: mockRoundTripper{},
}

client := facade.NewFacadeWith(
    "https://api.test.com",
    mockClient,
    nil, nil,
    &MockAuthProvider{},
)
```

## Интеграция

### Как интегрировать с базой данных?

```go
type Order struct {
    ID              int
    PaymentUUID     string
    Amount          string
    Status          string
    PaymentURL      string
}

func createOrder(db *sql.DB, client *facade.Facade, amount string) error {
    // Создаем счет
    invoice, err := client.CreateInvoice(&payment.CreateInvoiceRequest{
        Amount: amount,
        // ...
    })
    
    // Сохраняем в БД
    _, err = db.Exec(`
        INSERT INTO orders (payment_uuid, amount, status, payment_url)
        VALUES (?, ?, ?, ?)
    `, invoice.Result.Uuid, amount, "pending", invoice.Result.Url)
    
    return err
}
```

### Как обновлять статус заказа?

В обработчике вебхука:

```go
func webhookHandler(w http.ResponseWriter, r *http.Request) {
    var webhook PaymentWebhook
    // ... парсинг и проверка подписи ...
    
    if webhook.IsFinal && webhook.PaymentStatus == "success" {
        db.Exec(`
            UPDATE orders 
            SET status = 'paid', paid_at = NOW() 
            WHERE payment_uuid = ?
        `, webhook.Uuid)
    }
    
    w.WriteHeader(http.StatusOK)
}
```

### Как логировать операции?

```go
import "log"

log.SetFlags(log.LstdFlags | log.Lshortfile)

resp, err := client.CreateInvoice(req)
if err != nil {
    log.Printf("ERROR: Failed to create invoice: %v", err)
    return err
}

log.Printf("INFO: Invoice created: %s", resp.Result.Uuid)
```

## Миграция

### Переход с другого SDK

Если вы переходите с другого SDK:

1. Установите новый SDK
2. Обновите импорты
3. Замените методы создания клиента
4. Проверьте названия полей в моделях
5. Обновите обработчики вебхуков

### Обновление версии SDK

```bash
go get -u github.com/heleket/go-sdk
go mod tidy
```

## Поддержка

### Где получить помощь?

1. Проверьте эту документацию
2. Изучите [примеры](EXAMPLES.md)
3. Прочитайте [техническую документацию](TECHNICAL.md)
4. Обратитесь в поддержку Heleket
5. Создайте issue на GitHub

### Как сообщить об ошибке?

1. Создайте issue на GitHub
2. Укажите:
   - Версию SDK
   - Версию Go
   - Описание проблемы
   - Код для воспроизведения
   - Логи (без API ключей!)

### Где найти примеры кода?

- [EXAMPLES.md](EXAMPLES.md) - подробные примеры
- [QUICKSTART.md](QUICKSTART.md) - быстрый старт
- [cmd/main.go](cmd/main.go) - базовый пример

## Частые сценарии

### Прием платежа за подписку

```go
// 1. Создаем счет
invoice, _ := client.CreateInvoice(&payment.CreateInvoiceRequest{
    Amount:     "9.99",
    Currency:   "USD",
    ToCurrency: "USDT",
    Network:    "tron",
    OrderID:    fmt.Sprintf("SUB_%s_%d", userID, time.Now().Unix()),
    AdditionalData: map[string]interface{}{
        "user_id": userID,
        "type":    "subscription",
        "plan":    "premium",
    },
})

// 2. Показываем ссылку пользователю
showPaymentURL(invoice.Result.Url)

// 3. Ждем вебхук
// 4. Активируем подписку в обработчике вебхука
```

### Выплата партнерских вознаграждений

```go
// 1. Рассчитываем комиссию
fee, _ := client.CalculateWithdrawalFee(&withdrawal.CalculateRequest{
    Amount:   partnerReward,
    Currency: "USDT",
    Network:  "tron",
})

// 2. Создаем вывод
withdrawal, _ := client.CreateWithdrawal(&withdrawal.CreateWithdrawalRequest{
    Amount:    partnerReward,
    Currency:  "USDT",
    Network:   "tron",
    ToAddress: partnerAddress,
    OrderID:   fmt.Sprintf("PAYOUT_%s", partnerID),
    Metadata: map[string]interface{}{
        "partner_id": partnerID,
        "period":     "2024-01",
    },
})

// 3. Отслеживаем статус через вебхук или WithdrawalInformation()
```

### Создание личного кабинета

```go
// История платежей пользователя
func getUserPayments(userID string) {
    // Получаем все платежи
    history, _ := client.PaymentHistory(&payment.HistoryRequest{
        Limit: 100,
    })
    
    // Фильтруем по user_id в AdditionalData
    userPayments := []PaymentInfo{}
    for _, p := range history.Result.Data {
        if data, ok := p.AdditionalData.(map[string]interface{}); ok {
            if uid, ok := data["user_id"].(string); ok && uid == userID {
                userPayments = append(userPayments, p)
            }
        }
    }
    
    return userPayments
}
```

