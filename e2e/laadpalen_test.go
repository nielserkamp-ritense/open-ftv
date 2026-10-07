//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"strings"

	"github.com/chromedp/chromedp"
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
