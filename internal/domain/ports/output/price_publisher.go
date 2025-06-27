package output

import "marketflow/internal/domain/models"

type PricePublisher interface {
	PublishRedis(key, value string, update models.PriceUpdate) error
}
