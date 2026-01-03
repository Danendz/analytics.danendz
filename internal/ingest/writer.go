package ingest

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"analytics-svc/internal/models"

	"gorm.io/gorm"
)

type Writer struct {
	db      *gorm.DB
	ch      chan models.AnalyticsEvent
	maxSize int
	flushIn time.Duration
}

func NewWriter(db *gorm.DB) *Writer {
	bs := envInt("BATCH_SIZE", 200)
	fms := envInt("BATCH_FLUSH_MS", 500)

	return &Writer{
		db:      db,
		ch:      make(chan models.AnalyticsEvent, bs*10),
		maxSize: bs,
		flushIn: time.Duration(fms) * time.Millisecond,
	}
}

func (w *Writer) Enqueue(e models.AnalyticsEvent) bool {
	select {
	case w.ch <- e:
		return true
	default:
		return false
	}
}

func (w *Writer) Run(ctx context.Context) {
	ticker := time.NewTicker(w.flushIn)
	defer ticker.Stop()

	batch := make([]models.AnalyticsEvent, 0, w.maxSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := w.db.CreateInBatches(batch, w.maxSize).Error; err != nil {
			log.Printf("batch insert failed: %v", err)
		}
		batch = batch[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return

		case e := <-w.ch:
			batch = append(batch, e)
			if len(batch) >= w.maxSize {
				flush()
			}

		case <-ticker.C:
			flush()
		}
	}
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil || i <= 0 {
		return def
	}
	return i
}
