package auction_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/afga95/labs-auction/internal/entity/auction_entity"
	"github.com/afga95/labs-auction/internal/infra/database/auction"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAuctionAutoClose(t *testing.T) {
	godotenv.Load("/cmd/auction/.env")
	log.Println(os.Getenv("AUCTION_INTERVAL"))
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(os.Getenv("MONGODB_URL")))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database("mongodb")
	repo := auction.NewAuctionRepository(db)
	auctionEntity := &auction_entity.Auction{
		Id:          "test-auction-id",
		ProductName: "Test Product",
		Category:    "Test Category",
		Description: "Test Description",
		Condition:   auction_entity.New,
		Status:      auction_entity.Active,
		Timestamp:   time.Now(),
	}

	repo.CreateAuction(context.TODO(), auctionEntity)
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var updated bool
	for !updated {
		select {
		case <-timeout:
			t.Fatalf("Timeout esperando o status mudar pra Completed")
		case <-ticker.C:
			var updatedAuction auction.AuctionEntityMongo
			err := repo.Collection.FindOne(context.Background(), bson.M{"_id": auctionEntity.Id}).Decode(&updatedAuction)
			if err != nil {
				t.Fatalf("Erro ao buscar auction: %v", err)
			}
			if updatedAuction.Status == auction_entity.Completed {
				updated = true
			}
		}
	}

}
