# Labs Auction - Makefile

.PHONY: help build up down logs test test-auto-close clean dev

# Variáveis
COMPOSE_FILE = docker-compose.yml
APP_NAME = labs-auction

help: ## Mostra esta ajuda
	@echo "Labs Auction - Comandos disponíveis:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Constrói as imagens Docker
	docker-compose build

up: ## Inicia todos os serviços
	docker-compose up -d
	@echo "✅ Aplicação rodando em http://localhost:8080"
	@echo "📊 MongoDB rodando em localhost:27017"

down: ## Para todos os serviços
	docker-compose down

logs: ## Mostra logs da aplicação
	docker-compose logs -f app

logs-all: ## Mostra logs de todos os serviços
	docker-compose logs -f

mongo-logs: ## Mostra logs do MongoDB
	docker-compose logs -f mongodb

restart: ## Reinicia todos os serviços
	docker-compose restart

restart-app: ## Reinicia apenas a aplicação
	docker-compose restart app

# === TESTES ===

test: ## Executa todos os testes (requer Go instalado)
	@echo "🧪 Iniciando MongoDB para testes..."
	docker-compose up -d mongodb
	@echo "⏳ Aguardando MongoDB ficar pronto..."
	@sleep 5
	@echo "🚀 Executando testes..."
	go test -v ./...

test-auto-close: ## Executa testes de fechamento automático
	@echo "🧪 Iniciando MongoDB para testes..."
	docker-compose up -d mongodb
	@echo "⏳ Aguardando MongoDB ficar pronto..."
	@sleep 5
	@echo "🎯 Executando testes de fechamento automático..."
	AUCTION_INTERVAL=2s go test -v ./internal/infra/database/auction -run TestAuctionAutoClose

test-docker: ## Executa testes usando Docker
	docker-compose --profile testing up --build test-runner

test-coverage: ## Executa testes com coverage
	@echo "🧪 Iniciando MongoDB para testes..."
	docker-compose up -d mongodb
	@echo "⏳ Aguardando MongoDB ficar pronto..."
	@sleep 5
	@echo "📊 Executando testes com coverage..."
	go test -v -cover ./internal/infra/database/auction

benchmark: ## Executa benchmark dos testes
	@echo "🧪 Iniciando MongoDB para benchmark..."
	docker-compose up -d mongodb  
	@echo "⏳ Aguardando MongoDB ficar pronto..."
	@sleep 5
	@echo "⚡ Executando benchmark..."
	go test -v -bench=. ./internal/infra/database/auction

# === DESENVOLVIMENTO ===

dev: ## Inicia ambiente de desenvolvimento (só MongoDB)
	docker-compose up -d mongodb
	@echo "🔧 MongoDB iniciado para desenvolvimento"
	@echo "💡 Execute: go run cmd/auction/main.go"

dev-run: ## Inicia desenvolvimento completo
	docker-compose up -d mongodb
	@echo "⏳ Aguardando MongoDB ficar pronto..."
	@sleep 5
	@echo "🚀 Iniciando aplicação em modo desenvolvimento..."
	go run cmd/auction/main.go

# === LIMPEZA ===

clean: ## Remove containers, volumes e imagens
	docker-compose down -v --rmi all --remove-orphans
	docker system prune -f

clean-data: ## Remove apenas os dados (volumes)
	docker-compose down -v

# === UTILITÁRIOS ===

status: ## Mostra status dos containers
	docker-compose ps

mongo-shell: ## Conecta ao shell do MongoDB
	docker exec -it mongodb-auction mongo -u admin -p admin --authenticationDatabase admin auctions

mongo-status: ## Verifica status do MongoDB
	docker exec mongodb-auction mongo --eval "db.runCommand('ismaster')" --quiet

api-health: ## Verifica se a API está respondendo
	@echo "🏥 Verificando saúde da API..."
	@curl -f http://localhost:8080/auction?status=0 && echo "✅ API está saudável" || echo "❌ API não está respondendo"

sample-auction: ## Cria um leilão de exemplo
	@echo "📦 Criando leilão de exemplo..."
	@curl -X POST http://localhost:8080/auction \
		-H "Content-Type: application/json" \
		-d '{"product_name": "MacBook Pro M2", "category": "Electronics", "description": "MacBook Pro M2 16GB RAM 512GB SSD in excellent condition", "condition": 1}' \
		&& echo "\n✅ Leilão criado com sucesso!"

sample-bid: ## Cria um lance de exemplo (requer auction_id)
	@echo "💰 Para criar um lance, use:"
	@echo "curl -X POST http://localhost:8080/bid -H 'Content-Type: application/json' -d '{\"user_id\": \"550e8400-e29b-41d4-a716-446655440001\", \"auction_id\": \"SEU_AUCTION_ID\", \"amount\": 1200.00}'"

# === INFORMAÇÕES ===

info: ## Mostra informações do projeto
	@echo "📋 Labs Auction - Informações do Projeto"
	@echo "----------------------------------------"
	@echo "🌐 API: http://localhost:8080"
	@echo "🗄️  MongoDB: localhost:27017"
	@echo "👤 MongoDB User: admin"
	@echo "🔑 MongoDB Password: admin"
	@echo "📊 Database: auctions"
	@echo ""
	@echo "📡 Endpoints principais:"
	@echo "  GET  /auction              - Listar leilões"
	@echo "  POST /auction              - Criar leilão"
	@echo "  GET  /auction/:id          - Buscar leilão"
	@echo "  POST /bid                  - Criar lance"
	@echo "  GET  /bid/:auctionId       - Buscar lances"
	@echo "  GET  /user/:userId         - Buscar usuário"

# Comando padrão
.DEFAULT_GOAL := help