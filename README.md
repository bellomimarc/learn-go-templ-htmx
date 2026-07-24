# SaaS Gestionale PoC - Minimal Stack

Un proof-of-concept di applicazione SaaS minimale costruita in Go 1.22+ con il più basso numero di dipendenze esterne possibili, ottimizzato per **compliance regolate** e **velocità di sviluppo early-stage**.

## 🎯 Obiettivi Architetturali

- ✅ **Minimo di dipendenze**: Solo Chi v5 + Templ (no ORM, no full frameworks)
- ✅ **Type-safety**: Templ compila template in Go code verificato dal compiler
- ✅ **Interattività senza JS-heavy**: HTMX via CDN per UX reattiva
- ✅ **REST API pura**: `encoding/json` della stdlib, nessun framework aggiuntivo
- ✅ **Compliance-ready**: Niente database in questo PoC, focus su architettura web minimale
- ✅ **Pragmatico**: Codice leggibile e focalizzato sulla didattica

## 📦 Stack Tecnologico

| Layer | Tecnologia | Uso |
|-------|-----------|-----|
| **Router HTTP** | `github.com/go-chi/chi/v5` | Routing leggero + middleware |
| **Template Engine** | `github.com/a-h/templ` | HTML type-safe compilato in Go |
| **Frontend** | HTMX (CDN) | Interattività senza framework JS |
| **REST API** | `encoding/json` (stdlib) | Serialization JSON |
| **HTTP Server** | `net/http` (stdlib) | Server HTTP nativo |
| **Database** | — | Escludere volutamente da questo PoC |

## 📁 Struttura Progetto

```
learn-new-go-stack/
├── main.go                    # Entry point + shared middleware/server setup
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── features/
│   ├── website/
│   │   ├── handlers.go        # Rotte + handler della landing pubblica
│   │   └── views/             # Templ per il dominio website
│   ├── dashboard/
│   │   ├── handlers.go        # Rotte + handler dashboard/htmx
│   │   └── views/             # Templ per il dominio dashboard
│   └── system/
│       └── handlers.go        # Endpoint trasversali JSON (/api/info, /health)
└── README.md                  # Questo file
```

## 🚀 Quick Start

### 1. Setup Iniziale

```bash
cd /home/marcello/personal-repos/learn-new-go-stack
go mod download
```

### 2. Generare Codice da Templ

Templ compila i file `.templ` in codice Go. Per installarlo come tool locale del progetto:

```bash
go get -tool github.com/a-h/templ/cmd/templ@latest
```

Genera il codice:

```bash
go tool templ generate
```

Questo crea file `.templ.go` vicino ai rispettivi `.templ` dentro ogni feature
(`features/*/views/`).

### 3. Avviare il Server

```bash
go run main.go
```

Output atteso:

```
╔════════════════════════════════════════════════════════╗
║       SaaS Gestionale PoC - Minimal Stack 2026        ║
╚════════════════════════════════════════════════════════╝

📡 Server running on http://localhost:8080
```

### 4. Testare l'Applicazione

#### Landing Page (HTML via Templ)
```
GET http://localhost:8080/
```
Renderizza la landing pubblica del dominio website.

#### Dashboard Completa (HTML via Templ)
```
GET http://localhost:8080/dashboard
```
Renderizza la dashboard con layout dedicato e interazione HTMX.

#### Endpoint HTMX (Componente HTML)
```
GET http://localhost:8080/api/status
```
Restituisce un frammento HTML (`StatusComponent`) che HTMX inserisce nel DOM.

#### REST API (JSON)
```
GET http://localhost:8080/api/info
```
Restituisce JSON strutturato con info sull'applicazione.

#### Health Check
```
GET http://localhost:8080/health
```
Endpoint di monitoraggio che restituisce JSON con status e uptime.

## 🔄 Flusso di Comunicazione

### HTMX + Templ Flow (Dashboard Interattiva)

```
Browser                         Go Server
  │                                  │
  ├──── GET /dashboard ────────────>│
  │                                  │
  │                         [Templ renders]
  │                         Dashboard()
  │                            ↓
  │                         <html>...</html>
  │                                  │
  │<──── Full HTML (200) ───────────┤
  │                                  │
  [User clicks "Carica Status"]      │
  │                                  │
  ├──── HTMX: GET /api/status ────>│
  │    (hx-get, hx-target)           │
  │                                  │
  │                         [Templ renders]
  │                         StatusComponent()
  │                            ↓
  │                         <div>✅ Status...</div>
  │                                  │
  │<──── HTML Fragment (200) ───────┤
  │                                  │
  [HTMX inserts into #status-container]
  [No full page reload!]

```

### REST API Flow (JSON Endpoint)

```
Browser                         Go Server
  │                                  │
  ├──── GET /api/info ─────────────>│
  │                                  │
  │                         [stdlib encoding/json]
  │                         marshals StatusResponse
  │                            ↓
  │                         {"application": "...", ...}
  │                                  │
  │<──── JSON (200) ────────────────┤
  │                                  │
  [Client app parses JSON normally]

```

## 📚 File Sorgente: Spiegazione

### `main.go`

- **Bootstrap del server**: setup Chi, middleware globali, configurazione `http.Server`.
- **Composizione per feature**: registra rotte tramite `website.RegisterRoutes`, `dashboard.RegisterRoutes`, `system.RegisterRoutes`.

### `features/website/handlers.go`

- Rotta e handler per la landing pubblica (`GET /`).
- Rendering della landing tramite template Templ locali del dominio website.

### `features/dashboard/handlers.go`

- Rotte e handler per dashboard (`GET /dashboard`) e frammento HTMX (`GET /api/status`).
- Rendering dei componenti dashboard tramite template locali del dominio dashboard.

### `features/system/handlers.go`

- Endpoint JSON cross-domain:
  - `GET /api/info`
  - `GET /health`
- Include i modelli `StatusResponse` e `InfoResponse` vicino ai relativi handler.

### `features/dashboard/views/layout.templ`

- Wrapper HTML base con:
  - CDN HTMX caricato globalmente
  - Styling inline (gradiente, card, button, spinner animation)
  - Sezione `{ children... }` dove ogni pagina inietta contenuto

### `features/dashboard/views/dashboard.templ`

- Pagina completa che usa `@Layout("Dashboard")`
- Contiene:
  - Card didattica: intro al PoC
  - Sezione HTMX con bottone che chiama `/api/status`
  - Target div `#status-container` dove HTMX inserisce il componente
  - Link all'endpoint `/api/info`
  - Note architetturali sulla type-safety e separazione concerns

### `features/dashboard/views/components.templ`

- Frammenti HTML riusabili:
  - `StatusComponent(status string, timestamp time.Time)`: Componente status con timestamp formattato
  - `AlertComponent(message string, severity string)`: Alert generico per error/warning/info
  - Helper `getSeverityStyle()`: Dimostra Go logic in template (CSS inline basato su severity)
- **Type-safety**: Templ verifica i parametri e il template al compile-time

### `go.mod` e `go.sum`

- Dipendenze:
  - `github.com/a-h/templ v0.2.543`: Template engine type-safe
  - `github.com/go-chi/chi/v5 v5.0.11`: Router HTTP
  - Zero altri dipendenze (solo la stdlib per il resto)

## 🛡️ Considerazioni di Compliance & Security

### Type-Safe HTML (XSS Prevention)

Templ compila i template in Go code verificato dal compiler:

```go
templ StatusComponent(status string, timestamp time.Time) {
    <p>{ status }</p>  // Status è type-safe, auto-escaped per HTML
}
```

Se provi a iniettare HTML in `status`, Templ lo escapa automaticamente. Non c'è risk di XSS se usi le features giuste di Templ.

### Minimal External Dependencies

Solo 2 dipendenze go.sum (non contiamo stdlib):
- Chi per routing
- Templ per template

**Vantaggio**: Minore superficie di attacco, compliance più facile, audit delle dipendenze semplice.

### Secrets & Environment

In questo PoC non abbiamo secrets hardcoded. In produzione:

```go
// DO NOT DO THIS:
// const DBPassword = "supersecret"

// DO THIS:
dbPassword := os.Getenv("DB_PASSWORD")
if dbPassword == "" {
    log.Fatal("DB_PASSWORD not set")
}
```

## 🔧 Development Workflow

### 1. Modifica Templ

Se modifichi un file `.templ`, rigenerano il codice:

```bash
go tool templ generate
```

### 2. Rebuild & Run

```bash
go build -o saas-poc
./saas-poc
```

O con watch (usa `air` o simile):

```bash
go get -tool github.com/air-verse/air@latest
go tool air  # watch + auto-reload
```

### 3. Testing

Aggiungi test file:

```go
// main_test.go
func TestHandleInfo(t *testing.T) {
    req := httptest.NewRequest("GET", "/api/info", nil)
    w := httptest.NewRecorder()
    handleInfo(w, req)
    if w.Code != 200 { t.Fail() }
}
```

## 📖 Concetti Chiave

### HTMX: Server-Side HTML Generation

- **Paradigma**: Client invia richiesta → Server renderizza HTML → Browser aggiorna DOM
- **Vs SPA (React/Vue)**: Niente JavaScript Framework client-side, logica nel server
- **Performance**: Ridotto payload JS, rendering server-side elimina waterfall UI

### Templ: Type-Safe Templates

- **Compilazione**: `go tool templ generate` trasforma `.templ` in Go code
- **Errori al compile-time**: Typo in HTML o variable? Il compiler Go lo rileva prima della runtime
- **Performance**: Template compilato = zero overhead di parsing runtime

### Chi: Lightweight Router

- **Middleware**: Stack-based, facile compose
- **Type-safe**: Signature handlers standardizzate `http.HandlerFunc`
- **Nested Routes**: Supporta subrouter per logica complessa

## 🎓 Prossimi Passi per Estendere il PoC

1. **Aggiungere Database**: Integra `database/sql` + driver Postgres (compliance-ready)
2. **Autenticazione**: JWT token + middleware custom Chi
3. **Error Handling**: Wrap error con context, ritorna JSON/HTML basato su Accept header
4. **Logging Strutturato**: Usa `log/slog` (Go 1.21+) per JSON logs
5. **Testing**: Unit test handlers + integration test HTMX flow
6. **Docker**: Containeriza con Dockerfile minimale (no bloat)

## 📝 Licenza

Questo PoC è educational. Usa liberamente come base per progetti personali/aziendali.

---

**Fatto con ❤️ per Go developers minimalist-focused.**
