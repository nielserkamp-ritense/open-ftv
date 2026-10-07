//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

func theLogboekShowsMyAanvraagAsDenied(ctx context.Context, organization, reden string) error {
	s := scenarioFrom(ctx)

	beheer, err := s.openTab(s.frontpage, organization)
	if err != nil {
		return err
	}

	if err := s.logIn(beheer); err != nil {
		return err
	}

	traceID, err := s.traceID()
	if err != nil {
		return err
	}

	row, err := s.findLogboekEntry(beheer, traceID)
	if err != nil {
		return err
	}

	if err := s.checkLogboekRow(beheer, row, reden); err != nil {
		return err
	}

	if err := s.do(beheer, chromedp.Click("", byRole("gridcell", traceID), chromedp.NodeVisible)); err != nil {
		return fmt.Errorf("open the Logboek entry: %w", err)
	}

	return s.checkLogboekDetails(beheer, reden)
}

// logIn logs in to Keycloak when the beheeromgeving sends the browser there.
func (s *scenario) logIn(beheer context.Context) error {
	page, err := s.keycloakOrBeheer(beheer)
	if err != nil {
		return fmt.Errorf("the beheeromgeving shows neither Keycloak nor its menu: %w", err)
	}

	if page == "beheer" {
		return nil
	}

	if err := s.do(beheer,
		chromedp.SendKeys(chromedp.ID("username"), s.env.vlierdamUsername),
		chromedp.SendKeys(chromedp.ID("password"), s.env.vlierdamPassword),
		chromedp.Click(chromedp.ID("kc-login")),
		chromedp.WaitVisible("", byRole("link", "Logboek")),
	); err != nil {
		return fmt.Errorf("log in to the beheeromgeving as %s: %w", s.env.vlierdamUsername, err)
	}

	return nil
}

// keycloakOrBeheer waits for the Keycloak login or the menu of the beheeromgeving. The beheeromgeving redirects
// to Keycloak, and each redirect ends the check that runs, so it starts again until the timeout. Chrome renews the
// document of the Keycloak login twice while it loads, and each time the elements that a step found are gone, so
// the login only counts once it has loaded.
func (s *scenario) keycloakOrBeheer(beheer context.Context) (string, error) {
	deadline := time.Now().Add(s.env.timeout)

	for {
		page, err := run(s, beheer, chromedp.Poll[string](`
			document.readyState === "complete" && document.querySelector("#kc-login") ? "keycloak" :
			document.querySelector('a[href="/logboek"]') ? "beheer" :
			null`))
		if err == nil || time.Now().After(deadline) {
			return page, err
		}
	}
}

// findLogboekEntry filters the Logboek on the trace ID and returns the row of its decision. The PDP writes its
// decisions to the log in the background, so the page is reloaded until the row is there; the URL keeps the filter.
// The Logboek first shows the recent decisions, which may hold this row too, and the grid replaces all its rows
// when the filter applies. So the row only counts once the grid shows no other trace ID, or the step would hold
// on to a row that is about to be replaced.
func (s *scenario) findLogboekEntry(beheer context.Context, traceID string) (chromedp.Search, error) {
	row := xpath(`//div[@role="row"][.//div[@role="gridcell"][normalize-space(.)=%s]]`+
		`[not(ancestor::*[@role="grid" or @role="treegrid"]//div[@col-id="traceId"][@role="gridcell"][normalize-space(.)!=%s])]`,
		traceID, traceID)

	if err := s.do(beheer,
		chromedp.Click("", byRole("link", "Logboek"), chromedp.NodeVisible),
		// Like a keyboard user, it opens the filter of the Trace ID column with the shortcut the column header announces.
		chromedp.Focus("", byRole("columnheader", "Trace ID"), chromedp.NodeVisible),
		chromedp.KeyEvent(kb.Enter, chromedp.KeyModifiers(chromedp.ModifierCtrl)),
		chromedp.SendKeys("", traceID+"\n", byRole("textbox", "Filter Value"), chromedp.NodeVisible),
	); err != nil {
		return "", fmt.Errorf("filter the Logboek on trace ID %s: %w", traceID, err)
	}

	deadline := time.Now().Add(s.env.timeout)

	for {
		attempt, cancel := context.WithTimeout(beheer, 2*time.Second)
		err := chromedp.Do(attempt, chromedp.WaitVisible(row))

		cancel()

		if err == nil {
			return row, nil
		}

		if !errors.Is(err, context.DeadlineExceeded) || time.Now().After(deadline) {
			return "", fmt.Errorf("the Logboek has no entry with trace ID %s: %w", traceID, err)
		}

		if err := s.do(beheer, chromedp.Reload()); err != nil {
			return "", err
		}
	}
}

// checkLogboekRow waits for the row to show the denial, because the grid fills in the cells of a row after it
// adds the row.
func (s *scenario) checkLogboekRow(beheer context.Context, row chromedp.Search, reden string) error {
	denied := row + xpath(`[.//div[@role="gridcell"][normalize-space(.)=%s]][.//div[@role="gridcell"][normalize-space(.)=%s]]`,
		"Geweigerd", reden)

	if err := s.do(beheer, chromedp.WaitVisible(denied)); err != nil {
		text, _ := run(s, beheer, chromedp.Text(row))
		return fmt.Errorf("the Logboek row shows no denial because of %q: %q: %w", reden, text, err)
	}

	return nil
}

func (s *scenario) checkLogboekDetails(beheer context.Context, reden string) error {
	var request struct {
		Subject  struct{ ID string } `json:"subject"`
		Resource struct{ ID string } `json:"resource"`
	}

	if err := s.readJSON(beheer, "Request", &request); err != nil {
		return err
	}

	if request.Subject.ID != s.gebruiker || request.Resource.ID != s.aanvraag {
		return fmt.Errorf("the Logboek entry is about %q asking for %q, not %q asking for %q",
			request.Subject.ID, request.Resource.ID, s.gebruiker, s.aanvraag)
	}

	var response struct {
		Decision *bool `json:"decision"`
		Context  struct {
			Reason string `json:"reason"`
		} `json:"context"`
	}

	if err := s.readJSON(beheer, "Response", &response); err != nil {
		return err
	}

	if response.Decision == nil || *response.Decision || response.Context.Reason != reden {
		return fmt.Errorf("the Logboek entry has decision %v with reason %q, not a denial because of %q",
			response.Decision, response.Context.Reason, reden)
	}

	return nil
}

func (s *scenario) readJSON(beheer context.Context, textarea string, v any) error {
	value, err := run(s, beheer, chromedp.Value("", byRole("textbox", textarea)))
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(value), v); err != nil {
		return fmt.Errorf("read the %s of the Logboek entry: %w", textarea, err)
	}

	return nil
}
