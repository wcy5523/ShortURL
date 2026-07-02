package service

import (
	"log"
	"shorturl/config"
	"shorturl/dao"
	"shorturl/model"
	"sync"
	"time"
)

type VisitEvent struct {
	ShortCode string
	IP        string
	UserAgent string
	Referer   string
}

type StatsService struct {
	visitLogDAO *dao.VisitLogDAO
	shortURLDAO *dao.ShortURLDAO
	visitCh     chan VisitEvent
	wg          sync.WaitGroup
	batchSize   int
	workerCount int
}

func NewStatsService() *StatsService {
	return &StatsService{
		visitLogDAO: dao.NewVisitLogDAO(),
		shortURLDAO: dao.NewShortURLDAO(),
		visitCh:     make(chan VisitEvent, config.AppConfig.Stats.ChannelSize),
		batchSize:   config.AppConfig.Stats.BatchSize,
		workerCount: config.AppConfig.Stats.WorkerCount,
	}
}

func (s *StatsService) Start() {
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker()
	}
}

func (s *StatsService) Stop() {
	close(s.visitCh)
	s.wg.Wait()
}

func (s *StatsService) RecordVisit(event VisitEvent) {
	select {
	case s.visitCh <- event:
	default:
		log.Printf("stats channel full, dropping event for %s", event.ShortCode)
	}
}

func (s *StatsService) worker() {
	defer s.wg.Done()

	batch := make([]model.VisitLog, 0, s.batchSize)
	codes := make([]string, 0, s.batchSize)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if err := s.visitLogDAO.BatchInsert(batch); err != nil {
			log.Printf("batch insert visit logs failed: %v", err)
		}

		if err := s.shortURLDAO.IncrementVisitBatch(codes); err != nil {
			log.Printf("batch increment visit count failed: %v", err)
		}

		batch = batch[:0]
		codes = codes[:0]
	}

	for {
		select {
		case event, ok := <-s.visitCh:
			if !ok {
				flush()
				return
			}
			visitLog := model.VisitLog{
				ShortCode: event.ShortCode,
				IP:        event.IP,
				UserAgent: event.UserAgent,
				Referer:   event.Referer,
				VisitedAt: time.Now(),
			}
			batch = append(batch, visitLog)
			codes = append(codes, event.ShortCode)

			if len(batch) >= s.batchSize {
				flush()
			}

		case <-ticker.C:
			flush()
		}
	}
}
