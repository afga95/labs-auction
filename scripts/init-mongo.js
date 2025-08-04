// Script de inicialização do MongoDB para o projeto Labs Auction

// Conectar ao banco de dados auctions
db = db.getSiblingDB('auctions');

// Criar coleções e índices
db.createCollection('auctions');
db.createCollection('bids');
db.createCollection('users');

// Índices para performance
db.auctions.createIndex({ "status": 1 });
db.auctions.createIndex({ "category": 1 });
db.auctions.createIndex({ "timestamp": 1 });
db.auctions.createIndex({ "product_name": "text" });

db.bids.createIndex({ "auction_id": 1 });
db.bids.createIndex({ "user_id": 1 });
db.bids.createIndex({ "amount": -1 });
db.bids.createIndex({ "timestamp": 1 });

db.users.createIndex({ "_id": 1 });

// Inserir dados de exemplo para testes
db.users.insertMany([
  {
    "_id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "João Silva"
  },
  {
    "_id": "550e8400-e29b-41d4-a716-446655440002", 
    "name": "Maria Santos"
  },
  {
    "_id": "550e8400-e29b-41d4-a716-446655440003",
    "name": "Pedro Oliveira"
  }
]);

// Leilão de exemplo (ativo)
db.auctions.insertOne({
  "_id": "auction-550e8400-e29b-41d4-a716-446655440001",
  "product_name": "iPhone 13 Pro",
  "category": "Electronics",
  "description": "iPhone 13 Pro 256GB in excellent condition with original box and accessories",
  "condition": 1,
  "status": 0,
  "timestamp": Math.floor(Date.now() / 1000)
});

// Alguns lances de exemplo
db.bids.insertMany([
  {
    "_id": "bid-550e8400-e29b-41d4-a716-446655440001",
    "user_id": "550e8400-e29b-41d4-a716-446655440001",
    "auction_id": "auction-550e8400-e29b-41d4-a716-446655440001",
    "amount": 800.00,
    "timestamp": Math.floor(Date.now() / 1000)
  },
  {
    "_id": "bid-550e8400-e29b-41d4-a716-446655440002",
    "user_id": "550e8400-e29b-41d4-a716-446655440002", 
    "auction_id": "auction-550e8400-e29b-41d4-a716-446655440001",
    "amount": 850.00,
    "timestamp": Math.floor(Date.now() / 1000)
  }
]);

print("MongoDB initialized successfully with sample data for Labs Auction!");
print("Collections created: auctions, bids, users");
print("Sample users, auctions, and bids inserted");
print("Indexes created for optimal performance");