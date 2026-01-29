package index

import (
	"mneme/internal/logger"
	"mneme/internal/storage"
)

func IndexBuilder(paths []string) {
	logger.Info("Starting IndexBuilder")

	docId := uint(0)
	// Loop through the path to get the contents first
	for _, path := range paths {
		// Get a new instance of the crawler for the specified path
		crawlPaths, err := storage.Crawler(path, storage.DefaultCrawlerOptions())
		if err != nil {
			logger.Errorf("Error crawling path %s: %+v", path, err)
			continue
		}

		// Tokenize the file
		tokens, err := TokenizeFile(path)
		if err != nil {
			logger.Errorf("Error tokenizing file %s: %+v", path, err)
			continue
		}

	}

}

func TokenizeFile(path string) ([]string, error) {

}
