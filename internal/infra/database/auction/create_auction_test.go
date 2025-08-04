package auction_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/afga95/labs-auction/internal/entity/auction_entity"
	"github.com/afga95/labs-auction/internal/infra/database/auction"
	"github.com/stretchr/testify/assert"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAuctionAutoClose(t *testing.T) {

	/*godotenv.Load("cmd/auction/.env")
	log.Println(os.Getenv("AUCTION_INTERVAL"))
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://admin:admin@localhost:27017/auctions?authSource=admin"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.TODO())*/

	// Conecta ao MongoDB
	clientOpts := options.Client().ApplyURI("mongodb://admin:admin@localhost:27017/auctions?authSource=admin")
	client, err := mongo.Connect(context.TODO(), clientOpts)
	assert.NoError(t, err, "Erro ao conectar no MongoDB")
	defer client.Disconnect(context.TODO())

	// Define um intervalo curto para teste (2 segundos)
	os.Setenv("AUCTION_INTERVAL", "2s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	// Cria o repositório
	db := client.Database("mongodb")
	repo := auction.NewAuctionRepository(db)

	// Cria uma auction de teste
	auctionEntity := &auction_entity.Auction{
		Id:          "test-auto-close-id",
		ProductName: "Produto Teste",
		Category:    "Categoria Teste",
		Description: "Descrição Teste",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   time.Now(),
	}

	// Insere no banco
	errObj := repo.CreateAuction(context.TODO(), auctionEntity)
	assert.Nil(t, errObj, "Erro ao criar leilão")

	// Espera que o status mude para Completed
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var completed bool
	for !completed {
		select {
		case <-timeout:
			t.Fatal("Timeout: o status do leilão não foi atualizado para Completed")
		case <-ticker.C:
			var result auction.AuctionEntityMongo
			err := repo.Collection.FindOne(context.TODO(), bson.M{"_id": auctionEntity.Id}).Decode(&result)
			if err != nil {
				t.Fatalf("Erro ao buscar leilão: %v", err)
			}
			if result.Status == auction_entity.Completed {
				completed = true
			}
		}
	}

}
