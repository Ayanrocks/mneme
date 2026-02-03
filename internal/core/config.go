package core

type Config struct {
	Version uint8         `toml:"version"`
	Index   IndexConfig   `toml:"index"`
	Sources SourcesConfig `toml:"sources"`
	Watcher WatcherConfig `toml:"watcher"`
	Search  SearchConfig  `toml:"search"`
	Ranking RankingConfig `toml:"ranking"`
	Logging LoggingConfig `toml:"logging"`
}

type IndexConfig struct {
	SegmentSize          int  `toml:"segment_size"`
	MaxTokensPerDocument int  `toml:"max_tokens_per_document"`
	ReindexOnModify      bool `toml:"reindex_on_modify"`
	SkipBinaryFiles      bool `toml:"skip_binary_files"`
}

type SourcesConfig struct {
	Paths             []string `toml:"paths"`
	IncludeExtensions []string `toml:"include_extensions"`
	ExcludeExtensions []string `toml:"exclude_extensions"`
	Ignore            []string `toml:"ignore"`
}

type WatcherConfig struct {
	Enabled    bool `toml:"enabled"`
	DebounceMS int  `toml:"debounce_ms"`
}

type SearchConfig struct {
	DefaultLimit int    `toml:"default_limit"`
	UseStopwords bool   `toml:"use_stopwords"`
	Language     string `toml:"language"`
}

type RankingConfig struct {
	BM25Weight          float64 `toml:"bm25_weight"`
	VSMWeight           float64 `toml:"vsm_weight"`
	RecencyHalfLifeDays int     `toml:"recency_half_life_days"`
}

type LoggingConfig struct {
	Level string `toml:"level"`
	JSON  bool   `toml:"json"`
}
