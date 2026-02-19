package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const counterFilename = ".apeople-counter.json"

// CounterData represents the on-disk counter file.
type CounterData struct {
	NextIndexID int    `json:"next_index_id"`
	SpecVersion string `json:"spec_version"`
}

// IDCounter manages sequential IDs for contacts.
type IDCounter struct {
	CounterData
	mu       sync.Mutex
	filePath string
}

var (
	globalCounter     *IDCounter
	globalCounterOnce sync.Once
)

// GetIDCounter returns the singleton ID counter for the given directory.
func GetIDCounter(dir string) (*IDCounter, error) {
	var err error
	globalCounterOnce.Do(func() {
		globalCounter, err = loadOrCreateCounter(dir)
	})
	return globalCounter, err
}

func loadOrCreateCounter(dir string) (*IDCounter, error) {
	counterFile := filepath.Join(dir, counterFilename)

	data, err := os.ReadFile(counterFile)
	if err != nil {
		if os.IsNotExist(err) {
			maxID := findMaxIndexID(dir)
			counter := &IDCounter{
				CounterData: CounterData{
					NextIndexID: maxID + 1,
					SpecVersion: "0.1.0",
				},
				filePath: counterFile,
			}
			if err := counter.save(); err != nil {
				return nil, fmt.Errorf("failed to save initial counter: %w", err)
			}
			return counter, nil
		}
		return nil, fmt.Errorf("failed to read counter file: %w", err)
	}

	var counterData CounterData
	if err := json.Unmarshal(data, &counterData); err != nil {
		return nil, fmt.Errorf("failed to parse counter file: %w", err)
	}

	counter := &IDCounter{
		CounterData: counterData,
		filePath:    counterFile,
	}

	if counter.SpecVersion == "" {
		counter.SpecVersion = "0.1.0"
	}

	return counter, nil
}

func findMaxIndexID(dir string) int {
	maxID := 0

	pattern := filepath.Join(dir, "*__contact*.md")
	files, _ := filepath.Glob(pattern)

	for _, file := range files {
		contact, err := ParseContactFile(file)
		if err != nil {
			continue
		}
		if contact.IndexID > maxID {
			maxID = contact.IndexID
		}
	}

	return maxID
}

// NextID returns the next index ID and increments the counter.
func (c *IDCounter) NextID() (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.CounterData.NextIndexID
	c.CounterData.NextIndexID++

	if err := c.save(); err != nil {
		c.CounterData.NextIndexID--
		return 0, fmt.Errorf("failed to save counter: %w", err)
	}

	return id, nil
}

func (c *IDCounter) save() error {
	data, err := json.MarshalIndent(c.CounterData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal counter: %w", err)
	}

	tempFile := c.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempFile, c.filePath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename counter file: %w", err)
	}

	return nil
}

// ResetSingleton resets the global counter (for testing).
func ResetSingleton() {
	globalCounterOnce = sync.Once{}
	globalCounter = nil
}
