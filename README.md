# Oryn — Discover Beyond the Screen

[![CI Pipeline](https://github.com/Amirh2006/Oryn/actions/workflows/ci.yml/badge.svg)](https://github.com/Amirh2006/Oryn/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-18+-61DAFB?style=flat&logo=react)](https://reactjs.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Oryn** is a production-grade, AI-powered Movie & TV Series Discovery Web Platform. It enables users to discover, search, compare, and receive personalized media recommendations through both classic multi-criteria exploration tools and a conversational AI Agent.

---

## Author & Project Metadata

- **Author:** Amirhossein Porsharifi
- **Email:** amir.h.porsharifi@gmail.com
- **Repository:** [https://github.com/Amirh2006/Oryn.git](https://github.com/Amirh2006/Oryn.git)

---

## Core Pillars & Capabilities

1. **Movie & Series Discovery:** Structured search, multi-criteria filtering (genres, release year, rating, runtime, country, language), sorting, pagination, and comprehensive media details.
2. **AI Assistant (Custom Agent Harness):** Natural language recommendations, plot/theme similarity matching, multi-title comparison, constraint-based suggestions, and real-time Server-Sent Events (SSE) streaming with live tool-execution feedback.
3. **Personalization & Memory:** Short-term conversational context, transparent long-term taste profiling (`user_memory`), watchlists, favorites, and custom collections.

---

## Architectural Stack

### Backend
- **Language & Framework:** Go + Gin.
- **Architecture:** Clean / Hexagonal Architecture (`domain` -> `usecase` -> `repository` / `adapter` -> `transport/http`).
- **Database:** PostgreSQL with versioned SQL migrations using `golang-migrate`.
- **Caching & Rate Limiting:** Redis for media response caching, distributed rate limiting, and token revocation.
- **Media Provider:** TMDB API behind a decoupled `MediaProvider` adapter interface with exponential backoff retry, circuit breaker, and optional outbound proxy support.
- **Authentication:** Argon2id password hashing, rotating JWT access tokens & refresh tokens stored in `HttpOnly`, `Secure`, `SameSite` cookies (IP-independent sessions).
- **Observability:** Structured JSON logging via `log/slog`, Request ID propagation, unified error response envelopes, graceful shutdown, and container health probes (`/healthz`, `/readyz`).

### AI Engine
- **Provider:** Cloudflare Workers AI via direct Go REST client.
- **Model Router:** Multi-model routing (Main Reasoning Model + Lightweight Intent/Summarization Model).
- **Agent Harness:** Custom ReAct loop (capped at 3 iterations to conserve daily Neurons quota), deterministic tool execution, and anti-hallucination factual verification.

### Frontend
- **Framework:** React + Vite + Tailwind CSS.
- **State & Networking:** TanStack Query, Axios (request ID, token management, error unwrap).
- **Bilingual & Directional Support:** First-class Persian (`fa`, RTL) and English (`en`, LTR) with zero external font CDN dependencies (self-hosted `Vazirmatn` WOFF2 fonts bundled locally).
- **Dual-Theme Design:** High-contrast cinema dark mode and clean, crisp light mode with zero flash of unstyled content (FOUC).

---

## Getting Started

### Prerequisites
- Docker & Docker Compose
- Go 1.26+ (for local development)
- Node.js 20+ & npm (for local frontend development)

### Quick Start with Docker Compose
```bash
# Clone the repository
git clone https://github.com/Amirh2006/Oryn.git
cd Oryn

# Copy environment example
cp .env.example .env

# Run services
docker compose up -d --build
```

### Development with Makefile
```bash
make help        # Show all available commands
make dev         # Run backend and frontend locally in development mode
make test        # Run Go unit tests with race detector and frontend tests
make lint        # Run linters on backend and frontend
make up          # Start all containers via Docker Compose
make down        # Stop all containers
```

---

## License

This project is licensed under the MIT License.
