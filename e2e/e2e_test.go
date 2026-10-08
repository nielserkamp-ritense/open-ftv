//go:build e2e

package e2e

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	env, err := environmentFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll("output", 0o755); err != nil {
		t.Fatal(err)
	}

	// The screenshots of an earlier run would pass for those of a scenario that fails in this one.
	if err := os.RemoveAll("screenshots"); err != nil {
		t.Fatal(err)
	}

	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			initializeScenario(sc, &env)
		},
		Options: &godog.Options{
			Format:      "pretty,junit:output/report.xml",
			Paths:       []string{"features"},
			Tags:        env.tags,
			Concurrency: env.concurrency,
			Strict:      true,
			TestingT:    t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("the end-to-end tests failed")
	}
}

func initializeScenario(sc *godog.ScenarioContext, env *environment) {
	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		s, err := startScenario(env)
		if err != nil {
			return ctx, err
		}

		return context.WithValue(ctx, scenarioKey{}, s), nil
	})

	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		s, ok := ctx.Value(scenarioKey{}).(*scenario)
		if !ok {
			return ctx, nil
		}
		defer s.close()

		if err != nil {
			return ctx, s.screenshot(sc.Name)
		}

		return ctx, nil
	})

	sc.Given(`^I open the Laadpalen UI as "([^"]*)"$`, iOpenTheLaadpalenUIAs)
	sc.When(`^I submit an aanvraag for postcode "([^"]*)" and huisnummer "([^"]*)"$`, iSubmitAnAanvraag)
	sc.Then(`^I see "([^"]*)"$`, iSee)
	sc.Then(`^the Logboek of "([^"]*)" shows my aanvraag as Geweigerd because of "([^"]*)"$`, theLogboekShowsMyAanvraagAsDenied)
}

type scenarioKey struct{}

// scenario holds the browser of one scenario. The frontpage is its first tab, and the steps open the other
// tabs from its links, as a visitor of the demo does.
type scenario struct {
	env         *environment
	frontpage   context.Context
	zaaksysteem context.Context
	tabs        []context.Context
	cancels     []context.CancelFunc
	gebruiker   string
	aanvraag    string
}

func scenarioFrom(ctx context.Context) *scenario {
	return ctx.Value(scenarioKey{}).(*scenario)
}

func startScenario(env *environment) (*scenario, error) {
	options := []chromedp.ExecAllocatorOption{
		chromedp.WindowSize(1440, 900),
		chromedp.Flag("force-prefers-reduced-motion", true),
	}

	if env.browserPath != "" {
		options = append(options, chromedp.ExecPath(env.browserPath))
	}

	frontpage, cancel := chromedp.NewContext(context.Background(), chromedp.WithAllocatorOptions(options...))
	s := &scenario{
		env:       env,
		frontpage: frontpage,
		tabs:      []context.Context{frontpage},
		cancels:   []context.CancelFunc{cancel},
	}

	// The first run starts the browser and ties it to this context, so it must not run under a timeout.
	if err := chromedp.Do(frontpage); err != nil {
		s.close()
		return nil, fmt.Errorf("start the browser: %w", err)
	}

	return s, nil
}

func (s *scenario) close() {
	for i := len(s.cancels) - 1; i >= 0; i-- {
		s.cancels[i]()
	}
}

func (s *scenario) do(tab context.Context, actions ...chromedp.Action[chromedp.Void]) error {
	ctx, cancel := context.WithTimeout(tab, s.env.timeout)
	defer cancel()

	return chromedp.Do(ctx, actions...)
}

func run[T any](s *scenario, tab context.Context, action chromedp.Action[T]) (T, error) {
	ctx, cancel := context.WithTimeout(tab, s.env.timeout)
	defer cancel()

	return chromedp.Run(ctx, action)
}

// openTab clicks the link with this name, which opens in a new tab, and returns that tab. Like a visitor, it
// first switches to the tab with the link: Chrome answers the accessibility queries of byRole only for the
// active tab.
func (s *scenario) openTab(from context.Context, link string) (context.Context, error) {
	if err := s.bringToFront(from); err != nil {
		return nil, fmt.Errorf("switch to the tab with the link %s: %w", link, err)
	}

	opened := chromedp.WaitNewTarget(from, func(info *target.Info) bool {
		return info.Type == "page"
	})

	if err := s.do(from, chromedp.Click("", byRole("link", link), chromedp.NodeVisible)); err != nil {
		return nil, fmt.Errorf("click the link %s: %w", link, err)
	}

	var id target.ID
	select {
	case id = <-opened:
	case <-time.After(s.env.timeout):
		return nil, fmt.Errorf("the link %s opened no new tab", link)
	}

	tab, cancel := chromedp.NewContext(s.frontpage, chromedp.WithTargetID(id))
	s.tabs = append(s.tabs, tab)
	s.cancels = append(s.cancels, cancel)

	if err := chromedp.Do(tab); err != nil {
		return nil, fmt.Errorf("attach to the tab of the link %s: %w", link, err)
	}

	return tab, nil
}

func (s *scenario) bringToFront(tab context.Context) error {
	ctx, cancel := context.WithTimeout(tab, s.env.timeout)
	defer cancel()

	_, err := chromedp.Call(ctx, page.BringToFront, cdp.Empty{})

	return err
}

var unsafeFileName = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func (s *scenario) screenshot(name string) error {
	if err := os.MkdirAll("screenshots", 0o755); err != nil {
		return err
	}

	var errs []error

	for i, tab := range s.tabs {
		png, err := run(s, tab, chromedp.FullScreenshot(90))
		if err != nil {
			errs = append(errs, err)
			continue
		}

		file := filepath.Join("screenshots", fmt.Sprintf("%s-%d.png", unsafeFileName.ReplaceAllString(name, "-"), i+1))
		errs = append(errs, os.WriteFile(file, png, 0o600))
	}

	return errors.Join(errs...)
}

// xpath builds a Search for an XPath query. The values are quoted, so they must not contain a double quote.
func xpath(format string, values ...string) chromedp.Search {
	quoted := make([]any, len(values))
	for i, v := range values {
		quoted[i] = `"` + v + `"`
	}

	return chromedp.Search(fmt.Sprintf(format, quoted...))
}

// byRole finds the elements with this role and accessible name, as a screen reader knows them, so a step fails
// when the page loses a name or label it relies on. It runs the query of Puppeteer's ARIA selectors. The name
// must match exactly; an empty name matches any name.
func byRole(role, name string) chromedp.QueryOption {
	return chromedp.ByFunc(func(ctx context.Context, t *chromedp.Target, n *chromedp.Node) ([]cdp.NodeID, error) {
		found, err := cdp.Call(ctx, t, accessibility.QueryAXTree, accessibility.QueryAXTreeParams{
			NodeID:         n.NodeID,
			Role:           role,
			AccessibleName: name,
		})
		if err != nil {
			return nil, err
		}

		var ids []cdp.BackendNodeID

		for _, node := range found.Nodes {
			if !node.Ignored {
				ids = append(ids, node.BackendDOMNodeID)
			}
		}

		if len(ids) == 0 {
			return nil, nil
		}

		pushed, err := cdp.Call(ctx, t, dom.PushNodesByBackendIDsToFrontend, dom.PushNodesByBackendIDsToFrontendParams{
			BackendNodeIDs: ids,
		})

		return pushed.NodeIDs, err
	})
}
