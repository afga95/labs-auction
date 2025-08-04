package auction

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/afga95/labs-auction/internal/entity/auction_entity"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAuctionAutoClose(t *testing.T) {
	// Setup do banco de dados de teste
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	assert.NoError(t, err)

	db := client.Database("test_auctions")
	defer func() {
		db.Drop(context.Background())
		client.Disconnect(context.Background())
	}()

	// Define um intervalo curto para teste (2 segundos)
	os.Setenv("AUCTION_INTERVAL", "2s")
	defer os.Unsetenv("AUCTION_INTERVAL")

	// Cria o repositório
	repo := NewAuctionRepository(db)
	defer repo.Stop()

	t.Run("Should close auction automatically after interval", func(t *testing.T) {
		// Cria um leilão de teste
		auction, err := auction_entity.CreateAuction(
			"Test Product",
			"Electronics",
			"Test Description for auction",
			auction_entity.New,
		)
		assert.NoError(t, err)

		// Cria o leilão no banco
		createErr := repo.CreateAuction(context.Background(), auction)
		assert.NoError(t, createErr)

		// Verifica que o leilão está ativo
		filter := bson.M{"_id": auction.Id}
		var auctionMongo AuctionEntityMongo
		decodeErr := repo.Collection.FindOne(context.Background(), filter).Decode(&auctionMongo)
		assert.NoError(t, decodeErr)
		assert.Equal(t, auction_entity.Active, auctionMongo.Status)

		// Verifica que o leilão está sendo monitorado
		assert.Equal(t, 1, repo.GetActiveAuctionsCount())

		// Aguarda o fechamento automático (2s + margem)
		time.Sleep(3 * time.Second)

		// Verifica que o leilão foi fechado
		decodeErr = repo.Collection.FindOne(context.Background(), filter).Decode(&auctionMongo)
		assert.NoError(t, decodeErr)
		assert.Equal(t, auction_entity.Completed, auctionMongo.Status)

		// Verifica que o leilão não está mais sendo monitorado
		assert.Equal(t, 0, repo.GetActiveAuctionsCount())
	})

	t.Run("Should not close auction if already completed", func(t *testing.T) {
		// Cria um leilão de teste
		auction, err := auction_entity.CreateAuction(
			"Test Product 2",
			"Electronics",
			"Test Description for auction 2",
			auction_entity.Used,
		)
		assert.NoError(t, err)

		// Cria o leilão no banco
		createErr := repo.CreateAuction(context.Background(), auction)
		assert.NoError(t, createErr)

		// Marca o leilão como completado manualmente
		filter := bson.M{"_id": auction.Id}
		update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}
		_, updateErr := repo.Collection.UpdateOne(context.Background(), filter, update)
		assert.NoError(t, updateErr)

		// Aguarda um pouco
		time.Sleep(1 * time.Second)

		// Verifica que o status permanece como completado
		var auctionMongo AuctionEntityMongo
		decodeErr := repo.Collection.FindOne(context.Background(), filter).Decode(&auctionMongo)
		assert.NoError(t, decodeErr)
		assert.Equal(t, auction_entity.Completed, auctionMongo.Status)
	})

	t.Run("Should load and close expired auctions on startup", func(t *testing.T) {
		// Cria um novo repositório para simular restart
		repo2 := NewAuctionRepository(db)
		defer repo2.Stop()

		// Cria um leilão já expirado diretamente no banco
		expiredAuction := &AuctionEntityMongo{
			Id:          "expired-auction-id",
			ProductName: "Expired Product",
			Category:    "Electronics",
			Description: "Expired auction description",
			Condition:   auction_entity.New,
			Status:      auction_entity.Active,
			Timestamp:   time.Now().Add(-10 * time.Minute).Unix(), // Expirado há 10 minutos
		}

		_, insertErr := repo2.Collection.InsertOne(context.Background(), expiredAuction)
		assert.NoError(t, insertErr)

		// Aguarda um pouco para o repositório processar
		time.Sleep(2 * time.Second)

		// Verifica que o leilão foi fechado
		filter := bson.M{"_id": "expired-auction-id"}
		var auctionMongo AuctionEntityMongo
		decodeErr := repo2.Collection.FindOne(context.Background(), filter).Decode(&auctionMongo)
		assert.NoError(t, decodeErr)
		assert.Equal(t, auction_entity.Completed, auctionMongo.Status)
	})

	t.Run("Should handle multiple concurrent auctions", func(t *testing.T) {
		repo3 := NewAuctionRepository(db)
		defer repo3.Stop()

		// Cria múltiplos leilões
		numAuctions := 5
		auctions := make([]*auction_entity.Auction, numAuctions)

		for i := 0; i < numAuctions; i++ {
			auction, err := auction_entity.CreateAuction(
				"Concurrent Product",
				"Electronics",
				"Concurrent auction description",
				auction_entity.New,
			)
			assert.NoError(t, err)
			auctions[i] = auction

			createErr := repo3.CreateAuction(context.Background(), auction)
			assert.NoError(t, createErr)
		}

		// Verifica que todos estão sendo monitorados
		assert.Equal(t, numAuctions, repo3.GetActiveAuctionsCount())

		// Aguarda o fechamento
		time.Sleep(3 * time.Second)

		// Verifica que todos foram fechados
		for _, auction := range auctions {
			filter := bson.M{"_id": auction.Id}
			var auctionMongo AuctionEntityMongo
			decodeErr := repo3.Collection.FindOne(context.Background(), filter).Decode(&auctionMongo)
			assert.NoError(t, decodeErr)
			assert.Equal(t, auction_entity.Completed, auctionMongo.Status)
		}

		// Verifica que nenhum está mais sendo monitorado
		assert.Equal(t, 0, repo3.GetActiveAuctionsCount())
	})
}

func TestAuctionInterval(t *testing.T) {
	t.Run("Should use default interval when env var is invalid", func(t *testing.T) {
		os.Setenv("AUCTION_INTERVAL", "invalid")
		defer os.Unsetenv("AUCTION_INTERVAL")

		interval := getAuctionInterval()
		assert.Equal(t, 5*time.Minute, interval)
	})

	t.Run("Should use env var when valid", func(t *testing.T) {
		os.Setenv("AUCTION_INTERVAL", "10m")
		defer os.Unsetenv("AUCTION_INTERVAL")

		interval := getAuctionInterval()
		assert.Equal(t, 10*time.Minute, interval)
	})

	t.Run("Should use default when env var is not set", func(t *testing.T) {
		os.Unsetenv("AUCTION_INTERVAL")

		interval := getAuctionInterval()
		assert.Equal(t, 5*time.Minute, interval)
	})
}

// Teste de benchmark para verificar performance com muitos leilões
func BenchmarkAuctionAutoClose(b *testing.B) {
	client, _ := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	db := client.Database("benchmark_auctions")
	defer func() {
		db.Drop(context.Background())
		client.Disconnect(context.Background())
	}()

	os.Setenv("AUCTION_INTERVAL", "1h")
	defer os.Unsetenv("AUCTION_INTERVAL")

	repo := NewAuctionRepository(db)
	defer repo.Stop()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		auction, _ := auction_entity.CreateAuction(
			"Benchmark Product",
			"Electronics",
			"Benchmark auction description",
			auction_entity.New,
		)

		repo.CreateAuction(context.Background(), auction)
	}
}
