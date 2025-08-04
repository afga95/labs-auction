package auction

import (
	"context"
	"os"
	"sync"
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

	// Campos para controle de fechamento automático
	auctionInterval     time.Duration
	activeAuctions      map[string]*AuctionTimer
	activeAuctionsMutex *sync.RWMutex
	stopChan            chan bool
	once                sync.Once
}

type AuctionTimer struct {
	AuctionId string
	EndTime   time.Time
	Timer     *time.Timer
	Cancel    chan bool
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	repo := &AuctionRepository{
		Collection:          database.Collection("auctions"),
		auctionInterval:     getAuctionInterval(),
		activeAuctions:      make(map[string]*AuctionTimer),
		activeAuctionsMutex: &sync.RWMutex{},
		stopChan:            make(chan bool),
	}

	// Inicializa a rotina de monitoramento apenas uma vez
	repo.once.Do(func() {
		go repo.startAuctionMonitoring(context.Background())
	})

	return repo
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {

	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}

	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	// Agenda o fechamento automático do leilão
	ar.scheduleAuctionClose(auctionEntity.Id, auctionEntity.Timestamp)

	return nil
}

// scheduleAuctionClose agenda o fechamento automático de um leilão
func (ar *AuctionRepository) scheduleAuctionClose(auctionId string, startTime time.Time) {
	endTime := startTime.Add(ar.auctionInterval)
	duration := time.Until(endTime)

	// Se o leilão já deveria ter terminado, fecha imediatamente
	if duration <= 0 {
		go ar.closeAuction(context.Background(), auctionId)
		return
	}

	// Cria um timer para fechar o leilão
	timer := time.NewTimer(duration)
	cancelChan := make(chan bool, 1)

	auctionTimer := &AuctionTimer{
		AuctionId: auctionId,
		EndTime:   endTime,
		Timer:     timer,
		Cancel:    cancelChan,
	}

	// Adiciona o timer aos leilões ativos
	ar.activeAuctionsMutex.Lock()
	ar.activeAuctions[auctionId] = auctionTimer
	ar.activeAuctionsMutex.Unlock()

	// Goroutine para aguardar o fechamento
	go func() {
		select {
		case <-timer.C:
			// Timer disparou - fechar leilão
			ar.closeAuction(context.Background(), auctionId)
		case <-cancelChan:
			// Cancelado externamente
			timer.Stop()
		}

		// Remove da lista de leilões ativos
		ar.activeAuctionsMutex.Lock()
		delete(ar.activeAuctions, auctionId)
		ar.activeAuctionsMutex.Unlock()
	}()
}

// closeAuction fecha um leilão específico
func (ar *AuctionRepository) closeAuction(ctx context.Context, auctionId string) {
	filter := bson.M{"_id": auctionId, "status": auction_entity.Active}
	update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}

	result, err := ar.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Error trying to close auction", err)
		return
	}

	if result.ModifiedCount > 0 {
		logger.Info("Auction closed automatically")
	}
}

// startAuctionMonitoring inicia a rotina de monitoramento de leilões
func (ar *AuctionRepository) startAuctionMonitoring(ctx context.Context) {
	// Carrega leilões ativos existentes no banco
	ar.loadActiveAuctions(ctx)

	// Rotina de verificação periódica
	ticker := time.NewTicker(1 * time.Minute) // Verifica a cada minuto
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ar.checkExpiredAuctions(ctx)
		case <-ar.stopChan:
			return
		}
	}
}

// loadActiveAuctions carrega leilões ativos do banco e agenda seu fechamento
func (ar *AuctionRepository) loadActiveAuctions(ctx context.Context) {
	filter := bson.M{"status": auction_entity.Active}
	cursor, err := ar.Collection.Find(ctx, filter)
	if err != nil {
		logger.Error("Error loading active auctions", err)
		return
	}
	defer cursor.Close(ctx)

	var auctions []AuctionEntityMongo
	if err := cursor.All(ctx, &auctions); err != nil {
		logger.Error("Error decoding active auctions", err)
		return
	}

	for _, auction := range auctions {
		startTime := time.Unix(auction.Timestamp, 0)
		endTime := startTime.Add(ar.auctionInterval)

		// Se já expirou, fecha imediatamente
		if time.Now().After(endTime) {
			go ar.closeAuction(ctx, auction.Id)
			continue
		}

		// Agenda o fechamento
		ar.scheduleAuctionClose(auction.Id, startTime)
	}
}

// checkExpiredAuctions verifica e fecha leilões que expiraram
func (ar *AuctionRepository) checkExpiredAuctions(ctx context.Context) {
	currentTime := time.Now()

	ar.activeAuctionsMutex.RLock()
	expiredAuctions := make([]string, 0)

	for auctionId, auctionTimer := range ar.activeAuctions {
		if currentTime.After(auctionTimer.EndTime) {
			expiredAuctions = append(expiredAuctions, auctionId)
		}
	}
	ar.activeAuctionsMutex.RUnlock()

	// Fecha leilões expirados
	for _, auctionId := range expiredAuctions {
		go ar.closeAuction(ctx, auctionId)

		// Remove e cancela o timer
		ar.activeAuctionsMutex.Lock()
		if auctionTimer, exists := ar.activeAuctions[auctionId]; exists {
			close(auctionTimer.Cancel)
			delete(ar.activeAuctions, auctionId)
		}
		ar.activeAuctionsMutex.Unlock()
	}
}

// CancelAuctionTimer cancela o timer de fechamento de um leilão específico
func (ar *AuctionRepository) CancelAuctionTimer(auctionId string) {
	ar.activeAuctionsMutex.Lock()
	defer ar.activeAuctionsMutex.Unlock()

	if auctionTimer, exists := ar.activeAuctions[auctionId]; exists {
		close(auctionTimer.Cancel)
		delete(ar.activeAuctions, auctionId)
	}
}

// GetActiveAuctionsCount retorna o número de leilões ativos sendo monitorados
func (ar *AuctionRepository) GetActiveAuctionsCount() int {
	ar.activeAuctionsMutex.RLock()
	defer ar.activeAuctionsMutex.RUnlock()
	return len(ar.activeAuctions)
}

// Stop para a rotina de monitoramento
func (ar *AuctionRepository) Stop() {
	close(ar.stopChan)

	// Cancela todos os timers ativos
	ar.activeAuctionsMutex.Lock()
	for _, auctionTimer := range ar.activeAuctions {
		close(auctionTimer.Cancel)
	}
	ar.activeAuctions = make(map[string]*AuctionTimer)
	ar.activeAuctionsMutex.Unlock()
}

// getAuctionInterval obtém o intervalo de duração do leilão das variáveis de ambiente
func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5 // Padrão de 5 minutos
	}
	return duration
}
