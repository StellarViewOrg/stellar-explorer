package main

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/app"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/config"
	"github.com/miguelnietoa/stellar-explorer/tui/internal/livestream"
)

type liveStreamUpdateMsg struct {
	Update livestream.Update
}

type liveStreamErrorMsg struct {
	Err error
}

type liveStreamListener struct {
	updates <-chan livestream.Update
	errs    <-chan error
}

func newLiveStreamListener(cfg config.Config) *liveStreamListener {
	profile, ok := cfg.Profile(cfg.DefaultProfile)
	if !ok || strings.TrimSpace(profile.RedisURL) == "" {
		return nil
	}

	consumer, err := livestream.NewConsumer(profile.RedisURL)
	if err != nil {
		return &liveStreamListener{
			errs: asyncErrorChannel(err),
		}
	}

	updates, errs := consumer.Subscribe(context.Background())
	return &liveStreamListener{
		updates: updates,
		errs:    errs,
	}
}

func asyncErrorChannel(err error) <-chan error {
	ch := make(chan error, 1)
	ch <- err
	close(ch)
	return ch
}

func (l *liveStreamListener) wait() tea.Cmd {
	if l == nil {
		return nil
	}

	return func() tea.Msg {
		select {
		case update, ok := <-l.updates:
			if !ok {
				if l.errs == nil {
					return liveStreamErrorMsg{Err: context.Canceled}
				}
				select {
				case err, ok := <-l.errs:
					if ok && err != nil {
						return liveStreamErrorMsg{Err: err}
					}
				default:
				}
				return liveStreamErrorMsg{Err: context.Canceled}
			}
			return liveStreamUpdateMsg{Update: update}
		case err, ok := <-l.errs:
			if !ok {
				return liveStreamErrorMsg{Err: context.Canceled}
			}
			return liveStreamErrorMsg{Err: err}
		}
	}
}

func applyLiveStreamUpdate(model *app.Model, update livestream.Update) {
	if model == nil {
		return
	}
	model.ApplyLiveFeedStreamUpdate(app.LiveFeedStreamUpdate{
		Ledger:       update.Ledger,
		Transactions: update.Transactions,
	})
}

func handleLiveStreamError(model *app.Model, err error) {
	if model == nil || err == nil {
		return
	}
	if model.Snapshot().LiveFeed.SourceMode == app.LiveFeedSourceStream {
		model.SetLiveFeedSourceMode(app.LiveFeedSourceDegraded)
		model.SetWarningStatus("Live stream degraded; falling back to polling.")
		return
	}
	model.SetLiveFeedSourceMode(app.LiveFeedSourcePoll)
}
