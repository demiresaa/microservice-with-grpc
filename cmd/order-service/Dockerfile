# ============================================================================
# ORDER SERVICE - MULTI-STAGE DOCKERFILE
# ============================================================================
#
# MULTI-STAGE BUILD NEDİR?
# ============================================================================
#
# Docker image'ı oluştururken birden fazla aşama (stage) kullanma tekniğidir.
#
# NEDEN MULTI-STAGE?
#   - Go derlemesi için gereken araçlar (compiler, modules, kaynak kod) ~1GB+
#   - Ama ÇALIŞAN binary sadece ~10-20MB
#   - Multi-stage ile sadece binary'yi küçük bir image'a kopyalarız
#   - Sonuç: güvenli + küçük + hızlı başlayan container
#
# AŞAMALAR:
#   Stage 1 (builder): Go compiler ile binary oluştur
#   Stage 2 (runtime): Sadece binary'yi kopyala, minimal image
#
# IMAGE BOYUT KARŞILAŞTIRMASI:
#   golang:1.25-alpine     ~300MB  (builder - atılır)
#   distroless/static      ~2MB    (runtime - kalır)
#   Toplam sonuç           ~15-20MB (binary + runtime)
#
# ÖNEMLI PARAMETRELER:
#   CGO_ENABLED=0 → C bağımlılığı olmadan statik binary üretir
#                    → distroless/scratch image'da çalışabilir
#                    → Alpine'da musl libc gerekmez
#   GOOS=linux    → Hedef işletim sistemi (container = linux)
#   GOAMD64=v1    → Tüm x86_64 CPU'larda uyumlu
#
# DISTROLESS NEDİR?
#   Google'ın ürettiği minimal container image'larıdır.
#   - Shell yok (sh, bash yok) → attacker shell alamaz
#   - Paket yöneticisi yok
#   - Sadece uygulama binary'si + minimal runtime
#   → Güvenlik açısından en iyi seçenek
#
# ============================================================================

# ---- STAGE 1: BUILDER ----
# Go compiler, modüller ve kaynak kod ile binary üretiyoruz
FROM golang:1.25-alpine AS builder

# Çalışma dizini
WORKDIR /app

# ÖNCE go.mod ve go.sum'u kopyala → Docker layer cache'den faydalanır
# Eğer sadece kod değişirse, modüller tekrar indirilmez (cache hit)
COPY go.mod go.sum ./
RUN go mod download

# Sonra tüm kaynak kodu kopyala
COPY . .

# Binary oluştur
# CGO_ENABLED=0 → statik binary (C bağımlılığı yok)
# -ldflags="-w -s" → debug sembollerini ve DWARF bilgisini sil → ~30% küçülme
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /order-service ./cmd/order-service

# ---- STAGE 2: RUNTIME ----
# Sadece binary'yi kopyala → çok küçük image
FROM gcr.io/distroless/static:nonroot

# nonroot tag'i → uygulama root kullanıcısı yerine nonroot kullanıcısıyla çalışır
# Güvenlik best-practice: container'da root ile çalışma!

# Builder stage'den sadece binary'yi kopyala
COPY --from=builder /order-service /order-service

# Binary çalıştırGO_ENABLED=0 GOOS=linux go bui
CMD ["/order-service"]
