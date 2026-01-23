# Idempotent Storage Initialization Implementation

## Overview
This document describes the idempotent logic implemented in `internal/storage/fs.go` to prevent overwriting existing files and versions during storage initialization.

## Key Features

### 1. Idempotent Behavior
- Running `mneme init` multiple times with the same version will skip initialization
- Only creates directories and files when necessary
- Prevents accidental overwrites of existing data

### 2. Version Compatibility Checking
- Compares both `STORAGE_VERSION` and `MNEME_CLI_VERSION`
- Exact version match: Initialization is skipped (idempotent)
- Version mismatch: Logs warning and proceeds with initialization
- Missing VERSION file: Triggers initialization

### 3. New Helper Functions

#### `FileExists(path string) (bool, error)`
- Checks if a file exists at the given path
- Similar to `DirExists()` but for files
- Returns false if path is a directory

#### `ReadVersionFile() (string, error)`
- Reads the contents of the VERSION file
- Returns the file content as a string
- Handles file opening and reading errors

#### `ParseVersionFile(content string) (string, string, error)`
- Parses VERSION file content
- Extracts `STORAGE_VERSION` and `MNEME_CLI_VERSION`
- Returns parsed versions or error if format is invalid

#### `IsVersionCompatible() (bool, error)`
- Checks if existing storage version matches current version
- Returns true if versions match exactly
- Returns false if versions differ or file doesn't exist
- Logs appropriate debug/warning messages

#### `ShouldInitialize() (bool, error)`
- Determines if storage initialization is needed
- Checks if storage directory exists
- Checks if VERSION file is compatible
- Returns true if initialization is needed, false otherwise

## Modified Function

### `InitMnemeStorage()`
The main initialization function now follows this flow:

1. **Check if initialization is needed** using `ShouldInitialize()`
2. **If already initialized**: Log message and skip (idempotent)
3. **If initialization needed**: Proceed with creating directories and VERSION file

## Test Scenarios

### Scenario 1: First Initialization
```bash
$ ./mneme init
# Creates all directories and VERSION file
# Output: "Storage initialization completed successfully!"
```

### Scenario 2: Idempotent Behavior (Same Version)
```bash
$ ./mneme init
# Detects existing compatible version
# Output: "Storage already initialized, skipping..."
```

### Scenario 3: Missing VERSION File
```bash
$ rm ~/.local/share/mneme/VERSION
$ ./mneme init
# Detects missing VERSION file
# Recreates VERSION file
# Output: "Storage initialization completed successfully!"
```

### Scenario 4: Version Mismatch
```bash
$ echo "STORAGE_VERSION: 0.9.0" > ~/.local/share/mneme/VERSION
$ ./mneme init
# Detects version mismatch
# Logs warning: "Storage version mismatch: existing=0.9.0, current=1.0.0"
# Proceeds with initialization
# Output: "Storage initialization completed successfully!"
```

## Benefits

1. **Safety**: Prevents accidental overwrites of existing data
2. **Clarity**: Clear logging shows what's happening
3. **Version Tracking**: VERSION file tracks storage format version
4. **Upgrade Path**: Version mismatch detection enables future upgrade logic
5. **Idempotency**: Safe to run multiple times without side effects

## Future Enhancements

- Add version upgrade logic for backward compatibility
- Add migration scripts for version upgrades
- Add checksum validation for VERSION file
- Add backup before overwriting on version mismatch
