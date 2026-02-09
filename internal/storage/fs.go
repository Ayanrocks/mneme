package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"mneme/internal/config"
	"mneme/internal/constants"
	"mneme/internal/core"
	"mneme/internal/core/pb"
	"mneme/internal/logger"
	"mneme/internal/platform"
	"mneme/internal/utils"
	"mneme/internal/version"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
)

// ErrNoSegments is returned when there are no segments available to search
var ErrNoSegments = errors.New("no segments found: please run 'mneme index' first to create an index")

// CreateDir Create directory function
func CreateDir(path string) error {
	expandedPath, err := utils.ExpandFilePath(path)
	if err != nil {
		logger.Errorf("Error Expanding Path: %+v", err.Error())
		return err
	}
	return os.MkdirAll(expandedPath, os.ModePerm)
}

// CreateFile creates a file at the given path, expanding ~ to home directory
func CreateFile(path string) (*os.File, error) {
	expandedPath, err := utils.ExpandFilePath(path)
	if err != nil {
		logger.Errorf("Error expanding path: %+v", err.Error())
		return nil, err
	}
	return os.Create(expandedPath)
}

// FileExists checks if a file exists at the given path
func FileExists(path string) (bool, error) {
	expandedPath, err := utils.ExpandFilePath(path)
	if err != nil {
		logger.Errorf("Error expanding path: %+v", err.Error())
		return false, err
	}

	info, err := os.Stat(expandedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Debugf("File does not exist: %s", expandedPath)
			return false, nil
		}
		logger.Errorf("Error stating file %s: %+v", expandedPath, err)
		return false, fmt.Errorf("failed to check file: %w", err)
	}

	if info.IsDir() {
		logger.Debugf("Path is a directory, not a file: %s", expandedPath)
		return false, nil
	}

	logger.Debugf("File exists: %s", expandedPath)
	return true, nil
}

// ReadVersionFile reads and returns the contents of the VERSION file
func ReadVersionFile() (string, error) {
	return readVersionFileInternal(constants.DirPath)
}

func readVersionFileInternal(dirPath string) (string, error) {
	versionPath := filepath.Join(dirPath, "VERSION")
	expandedPath, err := utils.ExpandFilePath(versionPath)
	if err != nil {
		logger.Errorf("Error expanding version path: %+v", err.Error())
		return "", err
	}

	file, err := os.Open(expandedPath)
	if err != nil {
		logger.Errorf("Error opening VERSION file: %+v", err)
		return "", fmt.Errorf("failed to open VERSION file: %w", err)
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			logger.Errorf("Error closing VERSION file: %+v", err)
		}
	}(file)

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		logger.Errorf("Error reading VERSION file: %+v", err)
		return "", fmt.Errorf("failed to read VERSION file: %w", err)
	}

	return content.String(), nil
}

// ParseVersionFile parses the VERSION file content and extracts version information
// Returns storageVersion, cliVersion, platformStr, error
func ParseVersionFile(content string) (storageVersion string, cliVersion string, platformStr string, err error) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "STORAGE_VERSION:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				storageVersion = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "MNEME_CLI_VERSION:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				cliVersion = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "PLATFORM:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				platformStr = strings.TrimSpace(parts[1])
			}
		}
	}

	if storageVersion == "" {
		return "", "", "", fmt.Errorf("STORAGE_VERSION not found in VERSION file")
	}

	return storageVersion, cliVersion, platformStr, nil
}

// IsVersionCompatible checks if the existing storage version is compatible with current version
func IsVersionCompatible() (bool, error) {
	return isVersionCompatibleInternal(constants.DirPath)
}

func isVersionCompatibleInternal(dirPath string) (bool, error) {
	exists, err := FileExists(filepath.Join(dirPath, "VERSION"))
	if err != nil {
		return false, err
	}
	if !exists {
		logger.Debug("VERSION file does not exist, initialization needed")
		return false, nil
	}

	content, err := readVersionFileInternal(dirPath)
	if err != nil {
		return false, err
	}

	existingStorageVersion, existingCliVersion, _, err := ParseVersionFile(content)
	if err != nil {
		logger.Errorf("Error parsing VERSION file: %+v", err)
		return false, err
	}

	currentStorageVersion := version.MnemeStorageEngineVersion
	currentCliVersion := version.MnemeVersion

	logger.Debugf("Existing storage version: %s, Current: %s", existingStorageVersion, currentStorageVersion)
	logger.Debugf("Existing CLI version: %s, Current: %s", existingCliVersion, currentCliVersion)

	// Check if versions match exactly
	if existingStorageVersion == currentStorageVersion && existingCliVersion == currentCliVersion {
		logger.Info("Storage is already initialized with compatible version")
		return true, nil
	}

	// For now, we consider it incompatible if versions don't match
	// In a real implementation, you might want to support version upgrades
	logger.Warnf("Storage version mismatch: existing=%s, current=%s", existingStorageVersion, currentStorageVersion)
	return false, nil
}

// CheckPlatformCompatibility reads the VERSION file and compares the stored platform
// with the current platform. Returns whether they're compatible and the stored platform string.
// This is used to warn users when running on a different OS than where the index was created.
func CheckPlatformCompatibility() (isCompatible bool, storedPlatform string, err error) {
	content, err := ReadVersionFile()
	if err != nil {
		// If VERSION file doesn't exist or can't be read, assume compatible (new install)
		return true, "", nil
	}

	_, _, storedPlatform, err = ParseVersionFile(content)
	if err != nil {
		return true, "", nil
	}

	// If no platform was stored (older VERSION file), assume compatible
	if storedPlatform == "" {
		return true, "", nil
	}

	currentPlatform := platform.Current()
	isCompatible = platform.IsPlatformCompatible(storedPlatform, currentPlatform)

	return isCompatible, storedPlatform, nil
}

// ShouldInitialize determines if storage initialization is needed
func ShouldInitialize() (bool, error) {
	// Check if storage directory exists
	exists, err := DirExists(constants.DirPath)
	if err != nil {
		return false, err
	}

	if !exists {
		logger.Debug("Storage directory does not exist, initialization needed")
		return true, nil
	}

	// Check if VERSION file exists and is compatible
	compatible, err := IsVersionCompatible()
	if err != nil {
		return false, err
	}

	if !compatible {
		logger.Debug("Storage version is older")
		return false, nil
	}

	logger.Debug("Storage is already initialized and compatible, skipping initialization")
	return false, nil
}

// ShouldInitializeConfig determines if config initialization is needed
func ShouldInitializeConfig() (bool, error) {
	// Check if config file exists
	exists, err := FileExists(constants.ConfigPath)
	if err != nil {
		return false, err
	}

	if !exists {
		logger.Debug("Config file does not exist, initialization needed")
		return true, nil
	}

	logger.Debug("Config file already exists, skipping initialization")
	return false, nil
}

func InitMnemeStorage() error {
	logger.Info("Initializing mneme storage...")

	// Check if initialization is needed
	shouldInit, err := ShouldInitialize()
	if err != nil {
		logger.Errorf("Error checking if initialization is needed: %+v", err)
		return err
	}

	if !shouldInit {
		logger.Info("Storage already initialized, skipping...")
		return nil
	}

	// Fetch the default directory and check if it exists
	exists, err := DirExists(filepath.Dir(constants.ConfigPath))
	if err != nil {
		logger.Errorf("Error checking if config directory exists: %+v", err)
		return err
	}

	if !exists {
		logger.Infof("Creating storage directory: %s", constants.DirPath)
		if err := CreateDir(constants.DirPath); err != nil {
			logger.Errorf("Error creating directory %s: %+v", constants.DirPath, err)
			return err
		}
	}

	// Now if dir exists, let's initialize the folder structure for mneme
	// Create the mneme directory first
	logger.Infof("Creating app directory: %s", constants.DirPath)
	err = CreateDir(constants.DirPath)
	if err != nil {
		logger.Errorf("Error creating app directory %s: %+v", constants.AppName, err)
		return err
	}

	// create internal directories
	logger.Info("Creating internal directories...")
	err = CreateDir(path.Join(constants.DirPath, "meta"))
	if err != nil {
		logger.Errorf("Error creating meta directory: %+v", err)
		return err
	}

	err = CreateDir(path.Join(constants.DirPath, "segments"))
	if err != nil {
		logger.Errorf("Error creating segments directory: %+v", err)
		return err
	}

	err = CreateDir(path.Join(constants.DirPath, "tombstones"))
	if err != nil {
		logger.Errorf("Error creating tombstones directory: %+v", err)
		return err
	}

	// Create an empty version file with the current version of the mneme that created the folder
	logger.Info("Creating version file...")
	file, err := CreateFile(path.Join(constants.DirPath, "VERSION"))
	if err != nil {
		logger.Errorf("Error creating VERSION file: %+v", err)
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			logger.Errorf("Error closing VERSION file: %+v", err)
		}
	}()

	_, err = file.WriteString(getVersionFileContents())
	if err != nil {
		logger.Errorf("Error writing to VERSION file: %+v", err)
		return err
	}

	logger.Info("Storage initialization completed successfully!")
	return nil
}

func InitMnemeConfigStorage() error {
	logger.Info("Initializing mneme configuration storage...")

	shouldInit, err := ShouldInitializeConfig()
	if err != nil {
		logger.Errorf("Error checking if initialization is needed: %+v", err)
		return err
	}

	if !shouldInit {
		logger.Info("Config already initialized, skipping...")
		return nil
	}

	exists, err := DirExists(filepath.Dir(constants.ConfigPath))
	if err != nil {
		logger.Errorf("Error checking if config directory exists: %+v", err)
		return err
	}

	if !exists {
		logger.Infof("Creating storage directory: %s", filepath.Dir(constants.ConfigPath))
		if err := CreateDir(filepath.Dir(constants.ConfigPath)); err != nil {
			logger.Errorf("Error creating config directory: %+v", err)
			return err
		}
	}

	logger.Info("Creating default config")

	// Get default config as TOML string
	configContent, err := config.DefaultConfigWriter()
	if err != nil {
		logger.Errorf("Error generating default config: %+v", err)
		return err
	}

	// Create config file
	file, err := CreateFile(constants.ConfigPath)
	if err != nil {
		logger.Errorf("Error creating config file: %+v", err)
		return err
	}
	defer func() {
		err = file.Close()
		if err != nil {
			logger.Errorf("Error closing VERSION file: %+v", err)
		}
	}()

	// Write config to file
	_, err = file.WriteString(configContent)
	if err != nil {
		logger.Errorf("Error writing to config file: %+v", err)
		return err
	}

	logger.Info("Config initialization completed successfully!")
	return nil
}

// DirExists checks if a directory exists at the given path
func DirExists(path string) (bool, error) {
	// Expand ~ to home directory
	expandedPath, err := utils.ExpandFilePath(path)
	if err != nil {
		logger.Errorf("Error Expanding Path: %+v", err.Error())
		return false, err
	}

	info, err := os.Stat(expandedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Debugf("Directory does not exist: %s", expandedPath)
			return false, nil
		}
		logger.Errorf("Error stating directory %s: %+v", expandedPath, err)
		return false, fmt.Errorf("failed to check directory: %w", err)
	}

	if info.IsDir() {
		logger.Debugf("Directory exists: %s", expandedPath)
	}
	return info.IsDir(), nil
}

func getVersionFileContents() string {
	return fmt.Sprintf(`STORAGE_VERSION: %s
MNEME_CLI_VERSION: %s
PLATFORM: %s
`, version.MnemeStorageEngineVersion, version.MnemeVersion, platform.Current())
}

// ReadFileContents reads and returns the contents of the file at the given path
func ReadFileContents(path string) ([]string, error) {
	expandedPath, err := utils.ExpandFilePath(path)
	if err != nil {
		logger.Errorf("Error expanding path: %+v", err)
		return nil, err
	}

	file, err := os.Open(expandedPath)
	if err != nil {
		logger.Errorf("Error opening file %s: %+v", expandedPath, err)
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		logger.Errorf("Error reading file %s: %+v", expandedPath, err)
		return nil, err
	}

	return lines, nil
}

// SaveSegmentIndex saves the segment index using the binary protobuf format (default)
func SaveSegmentIndex(segmentIndex *core.Segment) error {
	return SaveSegmentIndexBinary(segmentIndex)
}

// SaveSegmentIndexJSON saves the segment index in JSON format (legacy)
func SaveSegmentIndexJSON(segmentIndex *core.Segment) error {
	logger.Info("Saving segment index (JSON)...")

	// Define the path for the segment JSON file
	segmentPath := filepath.Join(constants.DirPath, "segments", "segment.json")

	// Expand the file path (handles ~ expansion)
	expandedPath, err := utils.ExpandFilePath(segmentPath)
	if err != nil {
		logger.Errorf("Error expanding segment path: %+v", err)
		return fmt.Errorf("failed to expand segment path: %w", err)
	}

	// Marshal the segment index to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(segmentIndex, "", "  ")
	if err != nil {
		logger.Errorf("Error marshaling segment index to JSON: %+v", err)
		return fmt.Errorf("failed to marshal segment index: %w", err)
	}

	// Write the JSON data to the file
	err = os.WriteFile(expandedPath, jsonData, 0644)
	if err != nil {
		logger.Errorf("Error writing segment index to file %s: %+v", expandedPath, err)
		return fmt.Errorf("failed to write segment index: %w", err)
	}

	logger.Infof("Segment index saved successfully to %s", expandedPath)
	return nil
}

// LoadSegmentIndex loads the segment index, auto-detecting the format
// Priority: 1) manifest.json (chunk-based), 2) segment.idx (binary), 3) segment.json (legacy)
func LoadSegmentIndex() (*core.Segment, error) {
	logger.Info("Loading segment index...")

	// First, check if manifest exists (new chunk-based format)
	manifestPath := filepath.Join(constants.DirPath, "segments", "manifest.json")
	expandedManifestPath, err := utils.ExpandFilePath(manifestPath)
	if err != nil {
		logger.Errorf("Error expanding manifest path: %+v", err)
		return nil, fmt.Errorf("failed to expand manifest path: %w", err)
	}

	if _, err := os.Stat(expandedManifestPath); err == nil {
		logger.Debug("Found manifest, loading chunks...")
		return LoadAllChunks()
	}

	// Fall back to legacy binary format (segment.idx)
	binaryPath := filepath.Join(constants.DirPath, "segments", "segment.idx")
	expandedBinaryPath, err := utils.ExpandFilePath(binaryPath)
	if err != nil {
		logger.Errorf("Error expanding binary segment path: %+v", err)
		return nil, fmt.Errorf("failed to expand segment path: %w", err)
	}

	if _, err := os.Stat(expandedBinaryPath); err == nil {
		logger.Debug("Found binary segment file, loading...")
		return LoadSegmentIndexBinary()
	}

	// Fall back to JSON format
	logger.Debug("Binary segment not found, trying JSON format...")
	return LoadSegmentIndexJSON()
}

// LoadSegmentIndexJSON loads the segment index from JSON format (legacy)
func LoadSegmentIndexJSON() (*core.Segment, error) {
	logger.Info("Loading segment index (JSON)...")

	segmentPath := filepath.Join(constants.DirPath, "segments", "segment.json")
	expandedPath, err := utils.ExpandFilePath(segmentPath)
	if err != nil {
		logger.Errorf("Error expanding segment path: %+v", err)
		return nil, fmt.Errorf("failed to expand segment path: %w", err)
	}

	jsonData, err := os.ReadFile(expandedPath)
	if err != nil {
		logger.Errorf("Error reading segment index from file %s: %+v", expandedPath, err)
		return nil, fmt.Errorf("failed to read segment index: %w", err)
	}

	var segmentIndex core.Segment
	err = json.Unmarshal(jsonData, &segmentIndex)
	if err != nil {
		logger.Errorf("Error unmarshaling segment index from JSON: %+v", err)
		return nil, fmt.Errorf("failed to unmarshal segment index: %w", err)
	}

	logger.Infof("Segment index loaded successfully from %s", expandedPath)
	return &segmentIndex, nil
}

// SaveSegmentIndexBinary saves the segment index in binary protobuf format (.idx file)
func SaveSegmentIndexBinary(segmentIndex *core.Segment) error {
	logger.Info("Saving segment index in binary format...")

	// Define the path for the segment binary file
	segmentPath := filepath.Join(constants.DirPath, "segments", "segment.idx")

	// Expand the file path (handles ~ expansion)
	expandedPath, err := utils.ExpandFilePath(segmentPath)
	if err != nil {
		logger.Errorf("Error expanding segment path: %+v", err)
		return fmt.Errorf("failed to expand segment path: %w", err)
	}

	// Convert to protobuf format
	pbSegment := segmentIndex.ToPB()

	// Marshal to binary protobuf
	binaryData, err := proto.Marshal(pbSegment)
	if err != nil {
		logger.Errorf("Error marshaling segment index to protobuf: %+v", err)
		return fmt.Errorf("failed to marshal segment index: %w", err)
	}

	// Write the binary data to the file
	err = os.WriteFile(expandedPath, binaryData, 0644)
	if err != nil {
		logger.Errorf("Error writing segment index to file %s: %+v", expandedPath, err)
		return fmt.Errorf("failed to write segment index: %w", err)
	}

	logger.Infof("Segment index saved successfully to %s (binary, %d bytes)", expandedPath, len(binaryData))
	return nil
}

// LoadSegmentIndexBinary loads the segment index from binary protobuf format (.idx file)
func LoadSegmentIndexBinary() (*core.Segment, error) {
	logger.Info("Loading segment index from binary format...")

	segmentPath := filepath.Join(constants.DirPath, "segments", "segment.idx")
	expandedPath, err := utils.ExpandFilePath(segmentPath)
	if err != nil {
		logger.Errorf("Error expanding segment path: %+v", err)
		return nil, fmt.Errorf("failed to expand segment path: %w", err)
	}

	binaryData, err := os.ReadFile(expandedPath)
	if err != nil {
		logger.Errorf("Error reading segment index from file %s: %+v", expandedPath, err)
		return nil, fmt.Errorf("failed to read segment index: %w", err)
	}

	var pbSegment pb.Segment
	err = proto.Unmarshal(binaryData, &pbSegment)
	if err != nil {
		logger.Errorf("Error unmarshaling segment index from protobuf: %+v", err)
		return nil, fmt.Errorf("failed to unmarshal segment index: %w", err)
	}

	segment := core.SegmentFromPB(&pbSegment)

	logger.Infof("Segment index loaded successfully from %s (binary, %d bytes)", expandedPath, len(binaryData))
	return segment, nil
}

// ============================================================================
// Chunk-based storage functions for LSM-style batch indexing
// ============================================================================

// SaveChunk saves a segment chunk as a numbered file (e.g., 001.idx, 002.idx)
func SaveChunk(chunk *core.Segment, chunkID int) error {
	logger.Infof("Saving chunk %03d...", chunkID)

	// Format chunk filename with zero-padding
	chunkFilename := fmt.Sprintf("%03d.idx", chunkID)
	chunkPath := filepath.Join(constants.DirPath, "segments", chunkFilename)

	expandedPath, err := utils.ExpandFilePath(chunkPath)
	if err != nil {
		logger.Errorf("Error expanding chunk path: %+v", err)
		return fmt.Errorf("failed to expand chunk path: %w", err)
	}

	// Convert to protobuf and marshal
	pbSegment := chunk.ToPB()
	binaryData, err := proto.Marshal(pbSegment)
	if err != nil {
		logger.Errorf("Error marshaling chunk to protobuf: %+v", err)
		return fmt.Errorf("failed to marshal chunk: %w", err)
	}

	// Write to file
	err = os.WriteFile(expandedPath, binaryData, 0644)
	if err != nil {
		logger.Errorf("Error writing chunk to file %s: %+v", expandedPath, err)
		return fmt.Errorf("failed to write chunk: %w", err)
	}

	logger.Infof("Chunk %03d saved successfully (%d bytes)", chunkID, len(binaryData))
	return nil
}

// LoadChunk loads a specific chunk by ID
func LoadChunk(chunkID int) (*core.Segment, error) {
	logger.Debugf("Loading chunk %03d...", chunkID)

	chunkFilename := fmt.Sprintf("%03d.idx", chunkID)
	chunkPath := filepath.Join(constants.DirPath, "segments", chunkFilename)

	expandedPath, err := utils.ExpandFilePath(chunkPath)
	if err != nil {
		logger.Errorf("Error expanding chunk path: %+v", err)
		return nil, fmt.Errorf("failed to expand chunk path: %w", err)
	}

	binaryData, err := os.ReadFile(expandedPath)
	if err != nil {
		logger.Errorf("Error reading chunk from file %s: %+v", expandedPath, err)
		return nil, fmt.Errorf("failed to read chunk: %w", err)
	}

	var pbSegment pb.Segment
	err = proto.Unmarshal(binaryData, &pbSegment)
	if err != nil {
		logger.Errorf("Error unmarshaling chunk from protobuf: %+v", err)
		return nil, fmt.Errorf("failed to unmarshal chunk: %w", err)
	}

	segment := core.SegmentFromPB(&pbSegment)
	logger.Debugf("Chunk %03d loaded successfully (%d bytes)", chunkID, len(binaryData))
	return segment, nil
}

// SaveManifest saves the manifest as JSON in the segments directory
func SaveManifest(manifest *core.Manifest) error {
	logger.Info("Saving manifest...")

	manifestPath := filepath.Join(constants.DirPath, "segments", "manifest.json")
	expandedPath, err := utils.ExpandFilePath(manifestPath)
	if err != nil {
		logger.Errorf("Error expanding manifest path: %+v", err)
		return fmt.Errorf("failed to expand manifest path: %w", err)
	}

	jsonData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		logger.Errorf("Error marshaling manifest to JSON: %+v", err)
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	err = os.WriteFile(expandedPath, jsonData, 0644)
	if err != nil {
		logger.Errorf("Error writing manifest to file %s: %+v", expandedPath, err)
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	logger.Infof("Manifest saved successfully to %s", expandedPath)
	return nil
}

// LoadManifest loads the manifest from the segments directory
func LoadManifest() (*core.Manifest, error) {
	logger.Info("Loading manifest...")

	manifestPath := filepath.Join(constants.DirPath, "segments", "manifest.json")
	expandedPath, err := utils.ExpandFilePath(manifestPath)
	if err != nil {
		logger.Errorf("Error expanding manifest path: %+v", err)
		return nil, fmt.Errorf("failed to expand manifest path: %w", err)
	}

	jsonData, err := os.ReadFile(expandedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Debug("Manifest does not exist")
			return nil, nil // Return nil manifest if not found (not an error)
		}
		logger.Errorf("Error reading manifest from file %s: %+v", expandedPath, err)
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest core.Manifest
	err = json.Unmarshal(jsonData, &manifest)
	if err != nil {
		logger.Errorf("Error unmarshaling manifest from JSON: %+v", err)
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	logger.Debugf("Manifest loaded successfully with %d chunks", len(manifest.Chunks))
	return &manifest, nil
}

// LoadAllChunks loads all complete chunks and merges them into a single segment
func LoadAllChunks() (*core.Segment, error) {
	logger.Info("Loading all chunks...")

	// First try to load manifest
	manifest, err := LoadManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	// If no manifest exists, fall back to legacy single segment
	if manifest == nil {
		logger.Debug("No manifest found, trying legacy segment format...")
		return LoadSegmentIndexBinary()
	}

	// Get only complete chunks
	completeChunks := manifest.GetCompleteChunks()
	if len(completeChunks) == 0 {
		logger.Warn("No complete chunks found in manifest")
		return nil, ErrNoSegments
	}

	// Merge all chunks into a single segment
	mergedDocs := make([]core.Document, 0)
	mergedIndex := make(map[string][]core.Posting)

	for _, chunkInfo := range completeChunks {
		chunk, err := LoadChunk(chunkInfo.ID)
		if err != nil {
			logger.Errorf("Error loading chunk %d: %+v", chunkInfo.ID, err)
			return nil, fmt.Errorf("failed to load chunk %d: %w", chunkInfo.ID, err)
		}

		// Merge documents
		mergedDocs = append(mergedDocs, chunk.Docs...)

		// Merge inverted index
		for term, postings := range chunk.InvertedIndex {
			mergedIndex[term] = append(mergedIndex[term], postings...)
		}
	}

	mergedSegment := &core.Segment{
		Docs:          mergedDocs,
		InvertedIndex: mergedIndex,
		TotalDocs:     manifest.TotalDocs,
		TotalTokens:   manifest.TotalTokens,
		AvgDocLen:     manifest.AvgDocLen,
	}

	logger.Debugf("Merged %d chunks into single segment (%d docs, %d tokens)",
		len(completeChunks), len(mergedDocs), len(mergedIndex))
	return mergedSegment, nil
}

// MoveSegmentsToTombstones moves all segment files to the tombstones directory
// instead of deleting them. Files are prefixed with a timestamp to prevent naming conflicts.
func MoveSegmentsToTombstones() error {
	logger.Info("Moving segments to tombstones...")

	segmentsPath := filepath.Join(constants.DirPath, "segments")
	expandedSegmentsPath, err := utils.ExpandFilePath(segmentsPath)
	if err != nil {
		logger.Errorf("Error expanding segments path: %+v", err)
		return fmt.Errorf("failed to expand segments path: %w", err)
	}

	tombstonesPath := filepath.Join(constants.DirPath, "tombstones")
	expandedTombstonesPath, err := utils.ExpandFilePath(tombstonesPath)
	if err != nil {
		logger.Errorf("Error expanding tombstones path: %+v", err)
		return fmt.Errorf("failed to expand tombstones path: %w", err)
	}

	// Ensure tombstones directory exists
	if err := os.MkdirAll(expandedTombstonesPath, os.ModePerm); err != nil {
		logger.Errorf("Error creating tombstones directory: %+v", err)
		return fmt.Errorf("failed to create tombstones directory: %w", err)
	}

	// Read segments directory contents
	entries, err := os.ReadDir(expandedSegmentsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Debug("Segments directory does not exist, nothing to move")
			return nil
		}
		logger.Errorf("Error reading segments directory: %+v", err)
		return fmt.Errorf("failed to read segments directory: %w", err)
	}

	// Generate timestamp prefix for this batch
	timestamp := time.Now().Format("2006-01-02T15-04-05")
	movedCount := 0

	// Move all files to tombstones
	for _, entry := range entries {
		if entry.IsDir() {
			continue // Skip subdirectories
		}

		srcPath := filepath.Join(expandedSegmentsPath, entry.Name())
		// Add timestamp prefix to prevent naming conflicts
		destPath := filepath.Join(expandedTombstonesPath, fmt.Sprintf("%s_%s", timestamp, entry.Name()))

		err = os.Rename(srcPath, destPath)
		if err != nil {
			logger.Errorf("Error moving file %s to tombstones: %+v", entry.Name(), err)
			return fmt.Errorf("failed to move file %s: %w", entry.Name(), err)
		}
		logger.Debugf("Moved to tombstones: %s -> %s", entry.Name(), filepath.Base(destPath))
		movedCount++
	}

	logger.Infof("Moved %d files to tombstones", movedCount)
	return nil
}

// GetTombstonesSize calculates the total size of files in the tombstones directory
func GetTombstonesSize() (int64, error) {
	tombstonesPath := filepath.Join(constants.DirPath, "tombstones")
	expandedPath, err := utils.ExpandFilePath(tombstonesPath)
	if err != nil {
		logger.Errorf("Error expanding tombstones path: %+v", err)
		return 0, fmt.Errorf("failed to expand tombstones path: %w", err)
	}

	var totalSize int64

	entries, err := os.ReadDir(expandedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil // Tombstones directory doesn't exist, size is 0
		}
		logger.Errorf("Error reading tombstones directory: %+v", err)
		return 0, fmt.Errorf("failed to read tombstones directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			logger.Warnf("Error getting info for %s: %+v", entry.Name(), err)
			continue
		}
		totalSize += info.Size()
	}

	return totalSize, nil
}

// ClearTombstones permanently deletes all files in the tombstones directory
func ClearTombstones() (int64, int, error) {
	logger.Info("Clearing tombstones directory...")

	tombstonesPath := filepath.Join(constants.DirPath, "tombstones")
	expandedPath, err := utils.ExpandFilePath(tombstonesPath)
	if err != nil {
		logger.Errorf("Error expanding tombstones path: %+v", err)
		return 0, 0, fmt.Errorf("failed to expand tombstones path: %w", err)
	}

	entries, err := os.ReadDir(expandedPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			logger.Debug("Tombstones directory does not exist, nothing to clear")
			return 0, 0, nil
		}
		logger.Errorf("Error reading tombstones directory: %+v", err)
		return 0, 0, fmt.Errorf("failed to read tombstones directory: %w", err)
	}

	var freedBytes int64
	deletedCount := 0

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(expandedPath, entry.Name())

		// Get file size before deleting
		info, err := entry.Info()
		if err == nil {
			freedBytes += info.Size()
		}

		err = os.Remove(filePath)
		if err != nil {
			logger.Errorf("Error removing file %s: %+v", filePath, err)
			return freedBytes, deletedCount, fmt.Errorf("failed to remove file %s: %w", entry.Name(), err)
		}
		deletedCount++
		logger.Debugf("Deleted: %s", entry.Name())
	}

	logger.Infof("Tombstones cleared: %d files, %d bytes freed", deletedCount, freedBytes)
	return freedBytes, deletedCount, nil
}

// FormatBytes formats bytes into a human-readable string (KB, MB, GB)
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
