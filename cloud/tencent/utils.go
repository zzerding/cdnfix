package tencent

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// readCacheFile reads task IDs from the cache file.
// If the cache file does not exist, it returns an empty list.
// It returns an error if there is a problem reading the cache file.
//
// Parameters:
// - cacheFile: The path to the cache file.
//
// Returns:
// - []string: The list of task IDs read from the cache file.
// - error: An error if there was a problem reading the cache file.
func readCacheFile(cacheFile string) ([]string, error) {
	// Attempt to read the content of the cache file
	data, err := os.ReadFile(cacheFile)

	// If there was an error reading the cache file, check if it's because the file does not exist
	if err != nil {
		// If the cache file does not exist, return an empty list
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		// If there was an error other than the cache file not existing, return the error
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	// Split the content of the cache file by newline characters
	tasks := strings.Split(string(data), "\n")

	// Filter out empty tasks
	var validTasks []string
	for _, task := range tasks {
		// If the task is not empty, add it to the list of valid tasks
		if task != "" {
			validTasks = append(validTasks, task)
		}
	}

	// Return the list of valid task IDs
	return validTasks, nil
}

// updateCacheFile updates the task IDs in the cache file.
// It uses a buffered writer for efficient writing to the cache file.
//
// Parameters:
// - cacheFile string: The path to the cache file.
// - taskID string: The task ID to be added or removed from the cache file.
// - completed bool: A boolean indicating whether the task is completed.
//
// Returns:
// - error: An error if there was a problem updating the cache file.
func updateCacheFile(cacheFile string, taskID string, completed bool) error {
	file, err := os.OpenFile(cacheFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	tasks := strings.Split(string(data), "\n")

	if completed {
		tasks = removeTaskID(tasks, taskID)
	} else {
		tasks = append(tasks, taskID)
	}

	writer := bufio.NewWriter(file)
	for _, task := range tasks {
		_, err := writer.WriteString(task + "\n")
		if err != nil {
			return fmt.Errorf("failed to write to cache file: %w", err)
		}
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush buffered writer: %w", err)
	}

	return nil
}

func removeTaskID(tasks []string, taskID string) []string {
	var updatedTasks []string
	for _, task := range tasks {
		if task != "" && task != taskID {
			updatedTasks = append(updatedTasks, task)
		}
	}
	return updatedTasks
}
