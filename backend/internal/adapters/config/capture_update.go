package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// CaptureConfigUpdate is a preview of one explicitly requested configuration edit.
// Preparing it performs no writes and does not establish operational readiness.
type CaptureConfigUpdate struct {
	Original         []byte
	Data             []byte
	PreviousStateDir string
	StateDir         string
	Changed          bool
}

// PrepareCaptureConfigUpdate fills an absent state path, or applies an explicit
// replacement. Callers must coordinate relocation before replacing populated state.
func PrepareCaptureConfigUpdate(filePath, replacement string) (CaptureConfigUpdate, error) {
	var result CaptureConfigUpdate
	info, err := os.Lstat(filePath)
	if err != nil {
		return result, fmt.Errorf("reading configuration to update: %w", err)
	}
	if !info.Mode().IsRegular() {
		return result, errors.New("configuration update requires an existing regular file, not a symlink or directory")
	}
	result.Original, err = os.ReadFile(filePath)
	if err != nil {
		return result, err
	}
	previous, present, valueErr := captureStateValue(result.Original)
	if valueErr != nil && replacement == "" {
		return result, valueErr
	}
	result.PreviousStateDir = previous
	result.StateDir = previous
	if replacement == "" && present {
		if strings.TrimSpace(previous) == "" {
			return result, errors.New("telemetry.capture.state_dir is present but invalid; supply an explicit correction")
		}
		result.Data = append([]byte(nil), result.Original...)
		return result, nil
	}
	if strings.TrimSpace(replacement) == "" {
		return result, errors.New("an explicit telemetry.capture.state_dir replacement is required; no state path is inferred")
	}
	if _, err := ResolveCaptureStatePath(filepath.Dir(filePath), replacement); err != nil {
		return result, err
	}
	result.StateDir = replacement
	if present && previous == replacement && valueErr == nil {
		result.Data = append([]byte(nil), result.Original...)
		return result, nil
	}
	if ext := strings.ToLower(filepath.Ext(filePath)); ext != ".yaml" && ext != ".yml" {
		return result, errors.New("configuration must have a .yaml or .yml suffix")
	}
	result.Data, err = updateCaptureYAML(result.Original, replacement)
	if err != nil {
		return result, err
	}
	result.Changed = !bytes.Equal(result.Original, result.Data)
	return result, nil
}

func updateCaptureYAML(data []byte, replacement string) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	if len(doc.Content) != 0 {
		if err := validateCaptureDocument(doc.Content[0]); err != nil {
			return nil, err
		}
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		return nil, errors.New("configuration update requires a YAML mapping")
	}
	node := doc.Content[0]
	for _, key := range []string{"telemetry", "capture"} {
		var next *yaml.Node
		for i := 0; i+1 < len(node.Content); i += 2 {
			if node.Content[i].Value == key {
				if next != nil {
					return nil, fmt.Errorf("duplicate configuration key %s", key)
				}
				next = node.Content[i+1]
			}
		}
		if next == nil {
			next = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, next)
		}
		if next.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("configuration %s must be a mapping before updating capture", key)
		}
		node = next
	}
	index := -1
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == "state_dir" {
			if index >= 0 {
				return nil, errors.New("duplicate telemetry.capture.state_dir")
			}
			index = i + 1
		}
	}
	value := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: replacement}
	if index < 0 {
		node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "state_dir"}, value)
	} else {
		value.HeadComment = node.Content[index].HeadComment
		value.LineComment = node.Content[index].LineComment
		node.Content[index] = value
	}
	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(&doc); err != nil {
		return nil, errors.Join(err, encoder.Close())
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

// WriteCaptureConfigUpdate commits a preview under crash-released OS coordination for
// the existing configuration. Activation replaces the path in one operation, so interruption
// leaves either the valid prior configuration or its valid replacement.
func WriteCaptureConfigUpdate(filePath string, update CaptureConfigUpdate) error {
	return writeCaptureConfigUpdate(filePath, update, nil)
}

type configUpdateStage uint8

const (
	configUpdateStaged configUpdateStage = iota
	configUpdateActivated
)

// checkpoint supports real-file interruption tests at the activation boundary.
// Ordinary callers use WriteCaptureConfigUpdate without a checkpoint.
func writeCaptureConfigUpdate(filePath string, update CaptureConfigUpdate, checkpoint func(configUpdateStage) error) (returnErr error) {
	if !update.Changed {
		return nil
	}
	lock, err := lockCaptureConfig(filePath)
	if err != nil {
		return fmt.Errorf("configuration update lock: %w", err)
	}
	defer func() { returnErr = errors.Join(returnErr, lock.Close()) }()
	return writeLockedCaptureConfigUpdate(filePath, update, lock, checkpoint)
}

func writeLockedCaptureConfigUpdate(filePath string, update CaptureConfigUpdate, lock *captureConfigLock, checkpoint func(configUpdateStage) error) (returnErr error) {
	info, err := lock.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("configuration update target is no longer a regular file")
	}
	matches := func() error {
		latest, err := os.Lstat(filePath)
		if err != nil {
			return err
		}
		if !latest.Mode().IsRegular() || !os.SameFile(info, latest) {
			return errors.New("configuration changed after preview; retry without overwriting the newer file")
		}
		current := make([]byte, len(update.Original)+1)
		n, err := lock.ReadAt(current, 0)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		if !bytes.Equal(current[:n], update.Original) {
			return errors.New("configuration changed after preview; retry without overwriting the newer file")
		}
		return nil
	}
	if err = matches(); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(filePath), ".agenticlens-config-update-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer func() {
		if err := os.Remove(tempPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			returnErr = errors.Join(returnErr, err)
		}
	}()
	if err = temp.Chmod(info.Mode().Perm()); err != nil {
		return errors.Join(err, temp.Close())
	}
	if _, err = temp.Write(update.Data); err != nil {
		return errors.Join(err, temp.Close())
	}
	if err = temp.Sync(); err != nil {
		return errors.Join(err, temp.Close())
	}
	if err = temp.Close(); err != nil {
		return err
	}
	if checkpoint != nil {
		if err = checkpoint(configUpdateStaged); err != nil {
			return err
		}
	}
	if err = matches(); err != nil {
		return err
	}
	if err = lock.prepareReplace(); err != nil {
		return err
	}
	if err = replaceCaptureConfig(tempPath, filePath); err != nil {
		return fmt.Errorf("activating configuration update: %w", err)
	}
	if checkpoint != nil {
		if err = checkpoint(configUpdateActivated); err != nil {
			return fmt.Errorf("configuration updated before interruption: %w", err)
		}
	}
	return nil
}
