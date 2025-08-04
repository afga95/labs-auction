package auction

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/afga95/labs-auction/configuration/logger"
	"github.com/afga95/labs-auction/internal/entity/auction_entity"
	"github.com/afga95/labs-auction/internal/internal_error"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection *mongo.Collection
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction,
) *internal_error.InternalError {
	log.Println("GOROUTINE DISPARADA")
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	err := ar.Collection.Database().Client().Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Não conseguiu pingar o MongoDB: %v", err)
	}
	deadline, ok := ctx.Deadline()
	if ok {
		log.Println("Deadline do contexto:", deadline)
	} else {
		log.Println("Contexto sem deadline (OK)")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}
	log.Println("GOROUTINE DISPARADA2")
	go func() {
		log.Println("GOROUTINE COMEÇOU")
		time.Sleep(2 * time.Second)
		log.Println("GOROUTINE VAI ATUALIZAR")
		time.Sleep(getAuctionDuration())
		update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}
		filter := bson.M{"_id": auctionEntityMongo.Id}
		_, err := ar.Collection.UpdateOne(
			context.Background(),
			filter,
			update,
		)
		if err != nil {
			logger.Error("Error trying to update auction status", err)
			return
		}
		logger.Info("Auction finished")
	}()

	return nil
}

func getAuctionDuration() time.Duration {
	durationStr := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 5 * time.Minute
	}
	return duration
}
