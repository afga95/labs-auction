# Labs Auction - Sistema de Leilões

Sistema de leilões em Go com fechamento automático, criado com arquitetura limpa e MongoDB.

## 🚀 **Início Rápido**

### **Pré-requisitos**
- Docker e Docker Compose
- Go 1.20+ (para desenvolvimento)

### **1. Clone o projeto**
```bash
git clone <repository-url>
cd labs-auction
```

### **2. Execute com Docker Compose**
```bash
docker-compose up --build
```

A aplicação estará disponível em: `http://localhost:8080`

## 📋 **Funcionalidades**

- ✅ Criação de leilões
- ✅ Sistema de lances (bids)
- ✅ **Fechamento automático de leilões** (Nova funcionalidade)
- ✅ Consulta de leilões e lances
- ✅ Identificação de lances vencedores

## 🔧 **Configuração**

### **Variáveis de Ambiente**
As configurações estão no arquivo `cmd/auction/.env`:

```env
# Intervalo para fechamento automático dos leilões
AUCTION_INTERVAL=20s

# Configurações de batch para lances
BATCH_INSERT_INTERVAL=20s
MAX_BATCH_SIZE=4

# MongoDB
MONGODB_URL=mongodb://admin:admin@mongodb:27017/auctions?authSource=admin
MONGODB_DB=auctions
MONGO_INITDB_ROOT_USERNAME=admin
MONGO_INITDB_ROOT_PASSWORD=admin
```

## 🛠 **Comandos Docker**

### **Executar aplicação completa:**
```bash
docker-compose up --build
```

### **Executar apenas MongoDB:**
```bash
docker-compose up mongodb
```

### **Parar todos os serviços:**
```bash
docker-compose down
```

### **Ver logs:**
```bash
docker-compose logs -f app
```

## 🧪 **Executando Testes**

### **1. Testes com Docker**
```bash
# Subir apenas o MongoDB para testes
docker-compose up -d mongodb

# Executar testes (certifique-se de ter Go instalado)
go test -v ./internal/infra/database/auction
```

### **2. Testes de fechamento automático**
```bash
# Teste específico da nova funcionalidade
go test -v ./internal/infra/database/auction -run TestAuctionAutoClose

# Teste com intervalo personalizado (mais rápido)
AUCTION_INTERVAL=2s go test -v ./internal/infra/database/auction -run TestAuctionAutoClose
```

### **3. Todos os testes do projeto**
```bash
go test ./...
```

## 📡 **Endpoints da API**

### **Leilões**
- `GET /auction` - Listar leilões
- `GET /auction/:auctionId` - Buscar leilão por ID
- `POST /auction` - Criar novo leilão
- `GET /auction/winner/:auctionId` - Buscar lance vencedor

### **Lances**
- `POST /bid` - Criar novo lance
- `GET /bid/:auctionId` - Buscar lances por leilão

### **Usuários**
- `GET /user/:userId` - Buscar usuário por ID

## 📝 **Exemplos de Uso**

### **1. Criar um leilão**
```bash
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "iPhone 13",
    "category": "Electronics",
    "description": "iPhone 13 in excellent condition",
    "condition": 1
  }'
```

### **2. Fazer um lance**
```bash
curl -X POST http://localhost:8080/bid \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid-here",
    "auction_id": "auction-uuid-here",
    "amount": 1000.00
  }'
```

### **3. Listar leilões ativos**
```bash
curl "http://localhost:8080/auction?status=0"
```

## 🔄 **Fechamento Automático de Leilões**

### **Como funciona:**
1. Ao criar um leilão, um timer é automaticamente agendado
2. Depois de `AUCTION_INTERVAL`, o leilão é fechado automaticamente
3. Novos lances em leilões fechados são rejeitados
4. Sistema carrega leilões ativos ao reiniciar e agenda fechamentos

### **Testando o fechamento automático:**
```bash
# 1. Configure intervalo curto
export AUCTION_INTERVAL=30s

# 2. Crie um leilão
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "Test Product",
    "category": "Electronics", 
    "description": "Testing auto close functionality",
    "condition": 1
  }'

# 3. Aguarde 30s e verifique que o status mudou para "completed" (1)
curl "http://localhost:8080/auction?status=1"
```

## 🗂 **Estrutura do Projeto**

```
labs-auction/
├── cmd/auction/           # Ponto de entrada da aplicação
├── internal/
│   ├── entity/           # Entidades de domínio
│   ├── usecase/          # Casos de uso (regras de negócio)
│   └── infra/
│       ├── api/          # Controllers HTTP
│       └── database/     # Repositórios e acesso a dados
├── configuration/        # Configurações (logger, database, etc)
├── docker-compose.yml    # Orquestração de containers
└── Dockerfile           # Container da aplicação
```

## 🐛 **Solução de Problemas**

### **MongoDB não conecta:**
```bash
# Verificar se container está rodando
docker ps | grep mongodb

# Reiniciar MongoDB
docker-compose restart mongodb
```

### **Aplicação não inicia:**
```bash
# Ver logs detalhados
docker-compose logs app

# Reconstruir imagem
docker-compose up --build --force-recreate
```

### **Testes falham:**
```bash
# Certificar que MongoDB está rodando
docker-compose up -d mongodb

# Verificar conectividade
docker exec -it mongodb mongo --eval "db.runCommand('ismaster')"
```

## ⚡ **Performance**

### **Configurações recomendadas:**
- **Desenvolvimento**: `AUCTION_INTERVAL=1m`
- **Testes**: `AUCTION_INTERVAL=10s`  
- **Produção**: `AUCTION_INTERVAL=24h` (ou conforme necessário)

### **Monitoramento:**
- Logs estruturados em JSON
- Métricas de leilões ativos disponíveis
- Cleanup automático de recursos

## 🔒 **Considerações de Segurança**

- Validação de UUIDs em todos os endpoints
- Sanitização de inputs
- Tratamento adequado de erros
- Logs sem informações sensíveis

---

## 🏗 **Desenvolvimento**

### **Executar em modo desenvolvimento:**
```bash
# Apenas MongoDB
docker-compose up -d mongodb

# Aplicação local
go run cmd/auction/main.go
```

### **Rebuild completo:**
```bash
docker-compose down
docker-compose up --build --force-recreate
```

### **Executar benchmark:**
```bash
go test -v -bench=. ./internal/infra/database/auction
```

---

**Projeto desenvolvido com ❤️ em Go**