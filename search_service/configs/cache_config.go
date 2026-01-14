package configs

import (
	"time"
)

// структура конфига для инмемори шардированных кэшэй с TTL
type CachesConfig struct {
	NumOfShards               int // количество шардов
	SearchCacheConfig         SearchCacheConfig
	VacancyCacheConfig        VacancyCacheConfig
	VacancyDetailsCacheConfig VacancyDetailsCacheConfig
	MaxMemoryUsageMB          int
}

// структура конфига для кэша поиска
type SearchCacheConfig struct {
	SearchCacheTTL     time.Duration // время жизни элементов кэша поиска
	SearchCacheCleanUp time.Duration // интервал самоочистки для инмэмори кэша поиска
}

// структура конфига для кэша обратных индексов (вакансий)
type VacancyCacheConfig struct {
	VacancyCacheTTL     time.Duration // время жизни элементов кэша индекса
	VacancyCacheCleanUp time.Duration // интервал самоочистки для инмэмори кэша индекса
}

type VacancyDetailsCacheConfig struct {
	VacDetCacheTTL     time.Duration // время жизни элементов кэша деталей вакансии
	VacDetCacheCleanUp time.Duration // интервал самоочистки для инмэмори кэша деталей вакансии
}

// функция, которая возвращает указатель на дэфолтный конфиг для кэшэй
func DefaultCacheConfig() *CachesConfig {
	return &CachesConfig{
		NumOfShards: 7,
		SearchCacheConfig: SearchCacheConfig{
			SearchCacheTTL:     60 * time.Second,
			SearchCacheCleanUp: 30 * time.Second,
		},
		VacancyCacheConfig: VacancyCacheConfig{
			VacancyCacheTTL:     60 * time.Second,
			VacancyCacheCleanUp: 30 * time.Second,
		},
		VacancyDetailsCacheConfig: VacancyDetailsCacheConfig{
			VacDetCacheTTL:     60 * time.Second,
			VacDetCacheCleanUp: 30 * time.Second,
		},
	}
}
