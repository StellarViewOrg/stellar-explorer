package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/miguelnietoa/stellar-explorer/tui/internal/cache"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
)

func importLabelsFromFile(ctx context.Context, store *cache.Store, cfg config.Config, labelsPath string) error {
	if store == nil {
		return nil
	}

	path, err := config.ResolveLabelsPath(labelsPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("stat labels file: %w", err)
	}

	attachments, err := config.LoadLabelsFile(path)
	if err != nil {
		return err
	}

	profileID := cfg.DefaultProfile
	for _, attachment := range attachments {
		for _, name := range attachment.Names {
			labelID := importedLabelID(profileID, name)
			if err := store.UpsertLabel(ctx, cache.Label{
				ID:        labelID,
				ProfileID: profileID,
				Name:      name,
			}); err != nil {
				return err
			}

			targetID := importedLabelTargetID(profileID, labelID, attachment.Kind, attachment.Target)
			if err := store.UpsertLabelTarget(ctx, cache.LabelTarget{
				ID:        targetID,
				LabelID:   labelID,
				ProfileID: profileID,
				Kind:      attachment.Kind,
				Target:    attachment.Target,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}

func importedLabelID(profileID, name string) string {
	return "imported-" + shortHash(profileID+":"+strings.ToLower(strings.TrimSpace(name)))
}

func importedLabelTargetID(profileID, labelID, kind, target string) string {
	return "imported-target-" + shortHash(profileID+":"+labelID+":"+kind+":"+target)
}

func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:6])
}
