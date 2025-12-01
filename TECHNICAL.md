# Техническая документация Heleket Go SDK

## Аутентификация и подпись запросов

### Алгоритм подписи

SDK использует MD5 хеширование для подписи всех запросов к API. Это соответствует документации Heleket:

```php
// Пример из документации PHP
$data = json_encode($data);
$sign = md5(base64_encode($data) . $API_KEY);
```

В Go SDK это реализовано следующим образом:

```go
// internal/transport/auth.go
func (a *APIKeyAuthProvider) Headers(body []byte, isPayment bool) (http.Header, error) {
    // 1. Кодируем тело запроса в base64
    encoded := base64.StdEncoding.EncodeToString(body)
    
    // 2. Выбираем правильный API ключ
    key := a.paymentKey
    if !isPayment {
        key = a.payoffKey
    }
    
    // 3. Создаем строку для подписи: base64(data) + apiKey
    signString := encoded + key
    
    // 4. Вычисляем MD5 хеш
    hash := md5.Sum([]byte(signString))
    
    // 5. Преобразуем в hex строку
    sign := fmt.Sprintf("%x", hash)
    
    // 6. Добавляем заголовки
    h := http.Header{}
    h.Set("sign", sign)
    h.Set("merchant", a.merchantId)
    
    return h, nil
}
```

### Заголовки запроса

Каждый запрос к API содержит следующие заголовки:

| Заголовок | Описание | Пример |
|-----------|----------|--------|
| `Content-Type` | Тип контента | `application/json` |
| `sign` | MD5 подпись запроса | `a1b2c3d4e5f6...` |
| `merchant` | ID мерчанта | `4d372aa8-93cd-4d4b-a7d1-04ce1834a7fd` |

### Разные ключи для разных операций

SDK использует два разных API ключа:

1. **Payment API Key** - для всех платежных операций:
   - CreateInvoice
   - CreateStaticWallet
   - PaymentInformation
   - PaymentHistory
   - GenerateQr
   - Refund
   - RefundBlocked
   - BlockStaticWallet
   - ResendWebhook
   - TestingWebhook
   - ListOfServices

2. **Withdrawal API Key** - для операций вывода:
   - CreateWithdrawal
   - WithdrawalInformation
   - WithdrawalHistory
   - CalculateWithdrawalFee

### Проверка подписи вебхуков

При получении вебхука от Heleket необходимо проверить его подпись:

```go
func verifyWebhookSignature(body []byte, receivedSign string, apiKey string) bool {
    // 1. Кодируем тело в base64
    encoded := base64.StdEncoding.EncodeToString(body)
    
    // 2. Создаем строку для подписи
    signString := encoded + apiKey
    
    // 3. Вычисляем MD5
    hash := md5.Sum([]byte(signString))
    expectedSign := fmt.Sprintf("%x", hash)
    
    // 4. Сравниваем подписи
    return receivedSign == expectedSign
}
```

## Архитектура SDK

### Структура проекта

```
heleket/
├── cmd/
│   └── main.go                    # Пример использования
├── internal/
│   ├── facade/
│   │   └── facade.go              # Основной клиент SDK
│   ├── transport/
│   │   ├── auth.go                # Реализация аутентификации
│   │   └── json_adapter.go        # JSON маршаллинг
│   └── test/
│       └── mocks/                 # Моки для тестирования
├── pkg/
│   ├── constants/
│   │   └── endpoints.go           # URL эндпоинты API
│   ├── models/
│   │   ├── payment/               # Модели для платежей
│   │   └── withdrawal/            # Модели для выводов
│   └── transport/
│       └── transport.go           # Интерфейсы
└── go.mod
```

### Слои абстракции

#### 1. Transport Layer (pkg/transport)

Интерфейсы для работы с данными:

```go
type RequestAdapter interface {
    Marshal(v interface{}) ([]byte, error)
}

type ResponseAdapter interface {
    Unmarshal(data []byte, v interface{}) error
}

type AuthProvider interface {
    Headers(body []byte, isPayment bool) (http.Header, error)
}
```

#### 2. Implementation Layer (internal/transport)

Реализации интерфейсов:

- `JSONAdapter` - сериализация/десериализация JSON
- `APIKeyAuthProvider` - аутентификация с API ключами

#### 3. Facade Layer (internal/facade)

Главный класс `Facade` объединяет все компоненты:

```go
type Facade struct {
    BaseURL         string
    Client          *http.Client
    requestAdapter  RequestAdapter
    responseAdapter ResponseAdapter
    auth            AuthProvider
}
```

#### 4. Models Layer (pkg/models)

Структуры данных для запросов и ответов, разделенные по доменам:
- `payment/*` - модели платежей
- `withdrawal/*` - модели выводов

### Паттерны проектирования

#### Dependency Injection

SDK использует внедрение зависимостей для гибкости:

```go
// Базовый конструктор с дефолтными зависимостями
func NewFacade(baseURL string) *Facade

// Конструктор с кастомными зависимостями
func NewFacadeWith(
    baseURL string,
    client *http.Client,
    req RequestAdapter,
    resp ResponseAdapter,
    auth AuthProvider,
) *Facade
```

#### Adapter Pattern

Адаптеры для работы с разными форматами данных:

```go
type JSONAdapter struct{}

func (j JSONAdapter) Marshal(v interface{}) ([]byte, error) {
    return json.Marshal(v)
}

func (j JSONAdapter) Unmarshal(data []byte, v interface{}) error {
    return json.Unmarshal(data, v)
}
```

#### Strategy Pattern

Разные стратегии аутентификации:

```go
// API Key аутентификация
type APIKeyAuthProvider struct {
    paymentKey string
    merchantId string
    payoffKey  string
}

// Bearer token аутентификация (для будущего использования)
type BearerAuthProvider struct {
    token string
}
```

## HTTP клиент

### Методы запросов

#### POST запросы

```go
func (f *Facade) doPost(path string, reqBody interface{}, out interface{}, isPayment bool) error {
    // 1. Маршаллинг тела запроса
    data, err := f.requestAdapter.Marshal(reqBody)
    
    // 2. Создание HTTP запроса
    req, err := http.NewRequest(http.MethodPost, f.BaseURL+path, bytes.NewBuffer(data))
    
    // 3. Установка заголовков
    req.Header.Set("Content-Type", "application/json")
    
    // 4. Добавление заголовков аутентификации
    if f.auth != nil {
        hdrs, _ := f.auth.Headers(data, isPayment)
        for k, vs := range hdrs {
            for _, v := range vs {
                req.Header.Add(k, v)
            }
        }
    }
    
    // 5. Выполнение запроса
    resp, err := f.Client.Do(req)
    
    // 6. Чтение и десериализация ответа
    body, _ := io.ReadAll(resp.Body)
    return f.responseAdapter.Unmarshal(body, out)
}
```

#### GET запросы

```go
func (f *Facade) doGet(path string, out interface{}, isPayment bool) error {
    req, err := http.NewRequest(http.MethodGet, f.BaseURL+path, nil)
    
    // Добавление заголовков аутентификации
    if f.auth != nil {
        hdrs, _ := f.auth.Headers(nil, isPayment)
        // ...
    }
    
    // Выполнение и обработка
    // ...
}
```

### Кастомизация HTTP клиента

Можно настроить таймауты, retry политику и т.д.:

```go
import (
    "net/http"
    "time"
)

httpClient := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:          100,
        MaxIdleConnsPerHost:   100,
        MaxConnsPerHost:       100,
        IdleConnTimeout:       90 * time.Second,
        TLSHandshakeTimeout:   10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
    },
}

client := facade.NewFacadeWith("https://api.heleket.com", httpClient, nil, nil, nil)
```

## Обработка ошибок

### Типы ошибок

1. **Сетевые ошибки** - проблемы с соединением:
```go
resp, err := client.CreateInvoice(req)
if err != nil {
    // Сетевая ошибка, timeout, DNS и т.д.
    log.Printf("Network error: %v", err)
    return err
}
```

2. **API ошибки** - ошибки бизнес-логики:
```go
if resp.State != 1 {
    // API вернул ошибку
    log.Printf("API error: %s", resp.Message)
    log.Printf("Details: %+v", resp.Errors)
    return fmt.Errorf("api error: %s", resp.Message)
}
```

3. **HTTP ошибки** - неожиданные статус коды:
```go
if resp.StatusCode < 200 || resp.StatusCode >= 300 {
    return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
}
```

### Структура ответов

Все ответы имеют единый формат:

```go
type Response struct {
    State   int                    `json:"state"`   // 1 = успех, другое = ошибка
    Result  interface{}            `json:"result"`  // Данные при успехе
    Errors  map[string]interface{} `json:"errors"`  // Детали ошибок
    Message string                 `json:"message"` // Сообщение об ошибке
}
```

## Эндпоинты API

### Константы URL

```go
// pkg/constants/endpoints.go
const (
    // Payment endpoints
    URLCreateInvoice      = "/api/payment/create-invoice"
    URLCreateStaticWallet = "/api/payment/create-static-wallet"
    URLGenerateQr         = "/api/payment/generate-qr"
    URLPaymentHistory     = "/api/payment/history"
    URLPaymentInformation = "/api/payment/information"
    URLRefund             = "/api/payment/refund"
    URLRefundBlocked      = "/api/payment/refund-blocked"
    URLBlockStaticWallet  = "/api/payment/block-static-wallet"
    URLResendWebhook      = "/api/payment/resend-webhook"
    URLTestingWebhook     = "/api/payment/testing-webhook"
    URLListOfServices     = "/api/payment/list-of-services"
    
    // Withdrawal endpoints
    URLCreateWithdrawal      = "/api/withdrawal/create"
    URLWithdrawalInformation = "/api/withdrawal/information"
    URLWithdrawalHistory     = "/api/withdrawal/history"
    URLCalculateWithdraw     = "/api/withdrawal/calculate"
)
```

## Модели данных

### Соглашения об именовании

Все модели следуют паттерну:
- `{Operation}Request` - для запросов
- `{Operation}Response` - для ответов
- `{Operation}Result` - для данных в поле `result`

Пример:
```go
type CreateInvoiceRequest struct { ... }
type CreateInvoiceResponse struct {
    State   int
    Result  CreateInvoiceResult
    Errors  map[string]interface{}
    Message string
}
type CreateInvoiceResult struct { ... }
```

### JSON теги

Все поля имеют JSON теги для правильной сериализации:

```go
type CreateInvoiceRequest struct {
    Amount          string   `json:"amount"`
    Currency        string   `json:"currency"`
    OrderID         string   `json:"order_id"`
    UrlCallback     string   `json:"url_callback"`
    DiscountPercent int64    `json:"discount_percent,omitempty"` // omitempty для опциональных
}
```

### Типы данных

- Суммы представлены как `string` для точности
- Даты используют `time.Time`
- Метаданные используют `map[string]interface{}`
- Опциональные поля помечены `omitempty`

## Тестирование

### Моки для тестирования

SDK предоставляет моки для unit-тестов:

```go
// internal/test/mocks/adapters.go
type MockAuthProvider struct {
    HeadersFunc func([]byte, bool) (http.Header, error)
}

func (m *MockAuthProvider) Headers(body []byte, isPayment bool) (http.Header, error) {
    if m.HeadersFunc != nil {
        return m.HeadersFunc(body, isPayment)
    }
    return http.Header{}, nil
}
```

### Пример теста

```go
func TestCreateInvoice(t *testing.T) {
    // Создаем мок HTTP клиента
    mockClient := &http.Client{
        Transport: &mockRoundTripper{
            response: `{"state":1,"result":{"uuid":"test-uuid"}}`,
        },
    }
    
    // Создаем клиент с моками
    client := facade.NewFacadeWith(
        "https://api.test.com",
        mockClient,
        nil,
        nil,
        &MockAuthProvider{},
    )
    
    // Тестируем
    resp, err := client.CreateInvoice(&payment.CreateInvoiceRequest{
        Amount: "100",
        // ...
    })
    
    assert.NoError(t, err)
    assert.Equal(t, 1, resp.State)
}
```

## Производительность

### Рекомендации

1. **Переиспользуйте клиент**:
```go
// ✅ Правильно - один клиент для всех запросов
var globalClient = initClient()

// ❌ Неправильно - создание нового клиента для каждого запроса
func makePayment() {
    client := facade.NewFacade("...")
    // ...
}
```

2. **Настройте connection pooling**:
```go
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 100,
    },
}
```

3. **Используйте контексты для таймаутов**:
```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

req = req.WithContext(ctx)
```

## Безопасность

### Лучшие практики

1. **Храните ключи в переменных окружения**:
```go
auth, err := transport.NewAPIKeyAuth(
    os.Getenv("HELEKET_PAYMENT_KEY"),
    os.Getenv("HELEKET_MERCHANT_ID"),
    os.Getenv("HELEKET_WITHDRAWAL_KEY"),
)
```

2. **Используйте HTTPS**:
```go
client := facade.NewFacade("https://api.heleket.com") // не http://
```

3. **Проверяйте подписи вебхуков**:
```go
if !verifyWebhookSignature(body, receivedSign, apiKey) {
    http.Error(w, "Invalid signature", http.StatusUnauthorized)
    return
}
```

4. **Логируйте без чувствительных данных**:
```go
// ❌ Неправильно
log.Printf("API Key: %s", apiKey)

// ✅ Правильно
log.Printf("Request to %s completed", endpoint)
```

## Версионирование

SDK следует семантическому версионированию (SemVer):

- **MAJOR** версия при несовместимых изменениях API
- **MINOR** версия при добавлении функциональности с обратной совместимостью
- **PATCH** версия при обратно совместимых исправлениях

## Зависимости

SDK использует только стандартную библиотеку Go, без внешних зависимостей:

```go
import (
    "bytes"
    "crypto/md5"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)
```

Это обеспечивает:
- Легкий вес SDK
- Быструю компиляцию
- Минимум конфликтов версий
- Долгосрочную стабильность

