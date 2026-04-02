# Étape 1 : Build de l'application
FROM golang:1.23-alpine AS builder

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers de dépendances et les télécharger
COPY go.mod go.sum ./
RUN go mod download

# Copier le reste du code source
COPY . .

# Compiler l'application Go statiquement
RUN CGO_ENABLED=0 GOOS=linux go build -o ymmo-server ./cmd/server/main.go

# Étape 2 : Image finale légère
FROM alpine:latest

WORKDIR /app

# Copier le binaire depuis l'étape de build
COPY --from=builder /app/ymmo-server .

# Copier les dossiers contenant les fichiers statiques et les templates
# (Nécessaire car main.go sert les fichiers depuis ./web/static)
COPY --from=builder /app/web ./web

# Exposer le port par défaut
EXPOSE 8080

# Lancer l'application
CMD ["./ymmo-server"]