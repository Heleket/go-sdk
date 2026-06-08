# 01 — Installation

## Requirements

| Component | Minimum version | Notes |
|---|---|---|
| Go | 1.22 | `log/slog` in stdlib (1.21+), range-over-int (1.22) |
| OS | any | The SDK uses only stdlib net/http and crypto/md5 |

The SDK has **no runtime third-party dependencies**.

## Install

```bash
go get github.com/heleket/go-sdk
```

Then import:

```go
import heleket "github.com/heleket/go-sdk"
```

For webhook handling only:

```go
import "github.com/heleket/go-sdk/webhook"
```

## Verifying the install

A one-liner that signs an empty body — does not hit the network:

```go
package main

import (
    "fmt"

    heleket "github.com/heleket/go-sdk"
)

func main() {
    fmt.Println(heleket.Sign([]byte{}, "your-api-key"))
}
```

If you see a 32-character lowercase hex string, the SDK is wired up correctly.

## Getting API credentials

You need three pieces of information from <https://dash.heleket.com>:

1. **Merchant UUID** — Settings → API. Looks like `8b03432e-385b-4670-8d06-064591096795`.
2. **Payment API key** — Business → Domain → API key. Used for invoices, balance, exchange rates, and **payment webhook verification**.
3. **Payout API key** — Settings → API → Payout. Used for withdrawals and **payout webhook verification**. Generating a new key locks withdrawals for 24h.

Store them outside source control (use `.env`, secrets manager, or your hosting platform's environment-variable UI). The SDK never logs them in plaintext.

## Next

→ [02 — Configuration](02-configuration.md)
