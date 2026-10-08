//go:build e2e

package e2e

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

func iOpenTheLaadpalenUIAs(ctx context.Context, gebruiker string) error {
	s := scenarioFrom(ctx)

	if err := s.do(s.frontpage, chromedp.Navigate(s.env.laadpalenURL)); err != nil {
		return fmt.Errorf("open the Laadpalen UI at %s: %w", s.env.laadpalenURL, err)
	}

	zaaksysteem, err := s.openTab(s.frontpage, "Open het Zaaksysteem Laadpalen")
	if err != nil {
		return err
	}

	s.zaaksysteem = zaaksysteem
	s.gebruiker = gebruiker

	// The options read "Rick (Secretariaat)", the name with the afdeling, so they are matched on their start.
	user := gebruiker + " ("

	if err := s.do(zaaksysteem,
		chromedp.WaitEnabled("", byRole("combobox", "Gebruiker")),
		chromedp.Click("", byRole("combobox", "Gebruiker"), chromedp.NodeVisible),
		chromedp.Click(xpath(`//*[@role="option"][starts-with(normalize-space(.), %s)]`, user), chromedp.NodeVisible),
		chromedp.WaitVisible(xpath(`//button[@aria-label="Gebruiker"][starts-with(normalize-space(.), %s)]`, user)),
	); err != nil {
		return fmt.Errorf("choose the user %s: %w", gebruiker, err)
	}

	return nil
}

func iSubmitAnAanvraag(ctx context.Context, postcode, huisnummer string) error {
	s := scenarioFrom(ctx)
	// The Laadpalen application asks the PDP about the resource postcode-huisnummer.
	s.aanvraag = postcode + "-" + huisnummer

	if err := s.do(s.zaaksysteem, chromedp.Click("", byRole("button", "Nieuwe Aanvraag"), chromedp.NodeVisible)); err != nil {
		return fmt.Errorf("open the form for a new aanvraag: %w", err)
	}

	if err := s.do(s.zaaksysteem, chromedp.SendKeys("", postcode, byRole("textbox", "Postcode"), chromedp.NodeVisible)); err != nil {
		return fmt.Errorf("fill in the postcode: %w", err)
	}

	if err := s.do(s.zaaksysteem, chromedp.SendKeys("", huisnummer, byRole("textbox", "Huisnummer"), chromedp.NodeVisible)); err != nil {
		return fmt.Errorf("fill in the huisnummer: %w", err)
	}

	if err := s.do(s.zaaksysteem, chromedp.Click("", byRole("button", "Aanvraag Indienen"), chromedp.NodeVisible)); err != nil {
		return fmt.Errorf("submit the aanvraag: %w", err)
	}

	return nil
}

func iSee(ctx context.Context, melding string) error {
	s := scenarioFrom(ctx)

	text, err := run(s, s.zaaksysteem, chromedp.Text("", byRole("alert", ""), chromedp.NodeVisible))
	if err != nil {
		return fmt.Errorf("the Zaaksysteem shows no alert: %w", err)
	}

	if strings.TrimSpace(text) != melding {
		return fmt.Errorf("the Zaaksysteem alerts %q, not %q", strings.TrimSpace(text), melding)
	}

	return nil
}

// traceID reads the trace ID that the Zaaksysteem shows with its authorization check. The PDP logs its
// decision under that trace ID, so it finds the decision of this scenario among those of anyone else who uses
// the demo at the same time.
func (s *scenario) traceID() (string, error) {
	text, err := run(s, s.zaaksysteem, chromedp.Text(xpath(`//p[starts-with(normalize-space(.), %s)]`, "Trace ID:")))
	if err != nil {
		return "", fmt.Errorf("the Zaaksysteem shows no trace ID: %w", err)
	}

	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(text), "Trace ID:")), nil
}

// The regions of the checks of an aanvraag, named after their heading.
var checkRegions = map[string]string{
	"BRP":      "BRP Controle",
	"RDW":      "RDW/BRV Controle",
	"Laadpaal": "Laadpaal Check",
}

// checkRegion returns the region of the check. Each check has a button named Uitvoeren, so a step finds the one of
// its check within the region. Like byRole, it needs the Zaaksysteem to be the active tab.
func (s *scenario) checkRegion(check string) (*chromedp.Node, error) {
	if err := s.bringToFront(s.zaaksysteem); err != nil {
		return nil, fmt.Errorf("switch to the Zaaksysteem: %w", err)
	}

	nodes, err := run(s, s.zaaksysteem, chromedp.Nodes("", byRole("region", checkRegions[check]), chromedp.NodeVisible))
	if err != nil {
		return nil, fmt.Errorf("the aanvraag has no %s controle: %w", check, err)
	}

	return nodes[0], nil
}

// iRunTheCheck presses the button of the check once it is enabled: the RDW controle waits for the BRP controle. It
// presses Enter instead of clicking, because a check that finishes adds results above the buttons below it, and a
// click lands where the button was a moment before.
func iRunTheCheck(ctx context.Context, check string) error {
	s := scenarioFrom(ctx)

	region, err := s.checkRegion(check)
	if err != nil {
		return err
	}

	if err := s.do(s.zaaksysteem,
		chromedp.WaitEnabled("", byRole("button", "Uitvoeren"), chromedp.FromNode(region)),
		chromedp.SendKeys("", kb.Enter, byRole("button", "Uitvoeren"), chromedp.FromNode(region)),
	); err != nil {
		return fmt.Errorf("run the %s controle: %w", check, err)
	}

	return nil
}

func theAanvraagShows(ctx context.Context, text string) error {
	s := scenarioFrom(ctx)

	if err := s.do(s.zaaksysteem, chromedp.WaitVisible(xpath(`//*[normalize-space(.)=%s]`, text))); err != nil {
		return fmt.Errorf("the aanvraag doesn't show %q: %w", text, err)
	}

	return nil
}

func theAanvraagCanBeApproved(ctx context.Context) error {
	s := scenarioFrom(ctx)

	if err := theAanvraagShows(ctx, "Aanvraag kan worden goedgekeurd"); err != nil {
		return err
	}

	if err := s.do(s.zaaksysteem, chromedp.WaitEnabled("", byRole("button", "Goedkeuren"))); err != nil {
		return fmt.Errorf("the aanvraag can't be approved: %w", err)
	}

	return nil
}

func theAanvraagCanOnlyBeRejected(ctx context.Context) error {
	s := scenarioFrom(ctx)

	if err := theAanvraagShows(ctx, "Aanvraag kan worden afgewezen"); err != nil {
		return err
	}

	disabled, err := run(s, s.zaaksysteem, chromedp.AttributeValue("", "disabled", byRole("button", "Goedkeuren")))
	if err != nil {
		return err
	}

	if !disabled.Exists {
		return errors.New("the aanvraag can be approved")
	}

	return nil
}

// checkTraceID reads the trace ID that the Zaaksysteem shows with the check, the only link in its region.
func (s *scenario) checkTraceID(check string) (string, error) {
	region, err := s.checkRegion(check)
	if err != nil {
		return "", err
	}

	traceID, err := run(s, s.zaaksysteem, chromedp.Text("", byRole("link", ""), chromedp.FromNode(region)))
	if err != nil {
		return "", fmt.Errorf("the %s controle shows no trace ID: %w", check, err)
	}

	return strings.TrimSpace(traceID), nil
}
