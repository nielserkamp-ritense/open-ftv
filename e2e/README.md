# End-to-end tests

The end-to-end tests walk through the [scenarios of the Laadpalen demo](https://vng-realisatie.github.io/ftv/toepassen/demonstratie/)
in a browser, starting from the frontpage of the Laadpalen application. Each scenario is a feature file in
`features`. They run against the [local setup](../README.md#local-setup) by default:

```shell
go test -tags=e2e -count=1 ./e2e/
```

They need Chrome or Chromium. On macOS they find it only at `/Applications/Google Chrome.app` or
`/Applications/Chromium.app`. For another browser, such as Chrome Canary or Brave, set `E2E_BROWSER_PATH` to the
program inside the app:

```shell
E2E_BROWSER_PATH="/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary" go test -tags=e2e -count=1 ./e2e/
```

Environment variables point them at another environment:

| Variable                       | Default                                     | Purpose                                                                            |
|--------------------------------|---------------------------------------------|------------------------------------------------------------------------------------|
| `E2E_LAADPALEN_URL`            | `http://localhost:8080`                     | The frontpage of the Laadpalen application.                                        |
| `E2E_VLIERDAM_BEHEER_USERNAME` | `auditor-user`                              | The account for the management interface of Gemeente Vlierdam.                     |
| `E2E_VLIERDAM_BEHEER_PASSWORD` | `password`                                  | Its password.                                                                      |
| `E2E_BROWSER_PATH`             | Chrome or Chromium in the standard location | The browser to run, for example Chrome Canary or Brave.                            |
| `E2E_TIMEOUT`                  | `30s`                                       | How long a step waits for the page.                                                |
| `E2E_TAGS`                     |                                             | Runs only the scenarios with these [tags](https://github.com/cucumber/godog#tags). |

The browser runs headless; `CHROMEDP_VISIBLEWINDOW=1` shows it. A failed scenario leaves a screenshot of each of its
tabs in `screenshots`, and `output/report.xml` has the results as JUnit XML.
