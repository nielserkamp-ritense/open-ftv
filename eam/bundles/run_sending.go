package bundles

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"gitlab.com/digilab.overheid.nl/ecosystem/ftv/open-ftv/oas/bundles"
)

func (r *runner) sending() {
	r.debug("send stage started")

	r.gatherTargets("send")

	ok := r.sendBundles()
	if ok && r.advance("send") {
		r.info("send stage statistics", "bundle targets", r.targetCount, "bundles sent", r.sendCount)
	}
}

func (r *runner) gatherTargets(stage string) {
	if r.bundledCount == 0 {
		r.createBundles(stage)
	}

	r.targets = make(map[string]*sendJob, r.targetCount)

	for k := range r.m.bundles {
		cfg := r.m.bundles[k]

		for i := range cfg.Targets {
			key := fmt.Sprintf("%s:%s", cfg.ID, cfg.Targets[i].URI)
			r.targets[key] = &sendJob{
				key:     key,
				target:  cfg.Targets[i],
				bundle:  r.bundles[k],
				timeout: r.bundleTimeout,
				client:  r.client,
			}
		}
	}
}

func (r *runner) sendBundles() bool {
	if len(r.targets) == 0 {
		return true
	}

	jobs := make(sendJobCH, len(r.targets))
	results := make(sendResultCH, len(r.targets))

	// we use a local context to block this function until either the context is canceled,
	// or all targets have been serviced (indicated by the wg.Done).
	ctx, cancel := context.WithCancel(r.ctx)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(r.targets))
	go func() {
		wg.Wait()
		cancel()
	}()

	// run as many workers as needed, up to the configured maximum, but at least one.
	workers := max(min(r.m.workers, len(r.targets)), 1)
	for range workers {
		go processQueue(ctx, jobs, results, &wg)
	}

	// pass all targets to the worker queue.
	for key := range r.targets {
		jobs <- r.targets[key]
	}
	close(jobs)

	select {
	case <-ctx.Done():
	}

	var failed bool

	// check all worker responses.
	for range len(r.targets) {
		result := <-results
		target := r.targets[result.key]
		if result.success {
			r.sendCount++

			// report a PDP that was more than 1 version behind!
			if result.previous < int(r.d.version)-1 {
				r.info("updated lagging PDP", "target", target.target.URI, "old version", result.previous, "new version", r.d.version)
			}
		} else {
			failed = true
			r.error("failed to send bundle", "target", target.target.URI, "error", result.err)
		}
	}

	if failed {
		msg := "failed to send all bundles"
		r.error(msg, "targets", r.targetCount, "success", r.sendCount, "failed", r.targetCount-r.sendCount)

		var err error
		if r.d, err = r.handler.Fail(msg); err != nil {
			r.error("send stage: unable to set fail status", "error", err)
		}

		r.cancel()
	}

	return !failed
}

type sendJob struct {
	key     string
	target  *Target
	bundle  *Bundle
	timeout time.Duration
	client  *http.Client
}

type sendResult struct {
	success  bool
	previous int
	key      string
	err      error
}

type (
	sendJobCH    chan *sendJob
	sendResultCH chan *sendResult
)

func processQueue(ctx context.Context, jobs sendJobCH, results sendResultCH, wg *sync.WaitGroup) {
	for {
		select {
		case <-ctx.Done():
			return

		case job, ok := <-jobs:
			if !ok {
				return
			}

			if previous, err := sendBundle(ctx, job); err != nil {
				results <- &sendResult{key: job.key, err: err}
			} else {
				results <- &sendResult{success: true, previous: previous, key: job.key}
			}
			wg.Done()
		}
	}
}

func sendBundle(ctx context.Context, job *sendJob) (int, error) {
	ctx2, cancel := context.WithTimeout(ctx, job.timeout)
	defer cancel()

	ct := CompressionTypeFromString(job.target.Compress)

	// TODO: re-use compressed bundles when possible.
	var b bytes.Buffer
	if err := job.bundle.Compress(ct, &b); err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx2, "POST", job.target.URI, &b)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", ct.String())

	if job.target.APIKey != "" {
		req.Header.Set("Api-Key", job.target.APIKey)
	}

	resp, err2 := job.client.Do(req)
	if err2 != nil {
		return 0, err2
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result bundles.BundleActivated
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return result.PreviousVersion, nil
}
