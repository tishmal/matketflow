package output

import "marketflow/internal/domain/models"

type PricePublisher interface {
	Publish(update models.PriceUpdate) error
}
