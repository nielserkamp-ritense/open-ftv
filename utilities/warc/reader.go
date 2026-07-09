package warc

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Read parses all WARC records from r.
//
// It is the inverse of Writer/Record serialization and is primarily intended for
// verification and tooling; it is not a hardened parser for arbitrary WARC files.
func Read(r io.Reader) ([]*Record, error) {
	br := bufio.NewReader(r)
	var out []*Record

	for {
		rec, err := readRecord(br)
		if err == io.EOF {
			break
		}
		if err != nil {
			return out, err
		}
		if rec == nil {
			break
		}
		out = append(out, rec)
	}
	return out, nil
}

func readRecord(br *bufio.Reader) (*Record, error) {
	// Skip any blank lines between records.
	var version string
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF && line == "" {
			return nil, io.EOF
		}
		if err != nil && err != io.EOF {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if err == io.EOF {
				return nil, io.EOF
			}
			continue
		}
		version = line
		break
	}

	if !strings.HasPrefix(version, "WARC/") {
		return nil, fmt.Errorf("warc: expected version line, got %q", version)
	}

	rec := &Record{Custom: map[string]string{}}
	var contentLen int

	// Read headers until the blank line.
	for {
		line, err := br.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}

		key, val, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("warc: malformed header %q", line)
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		switch key {
		case "WARC-Type":
			rec.Type = val
		case "WARC-Record-ID":
			rec.RecordID = strings.Trim(val, "<>")
		case "WARC-Date":
			if t, e := time.Parse(time.RFC3339Nano, val); e == nil {
				rec.Date = t
			}
		case "WARC-Target-URI":
			rec.TargetURI = val
		case "WARC-Concurrent-To":
			rec.ConcurrentTo = strings.Trim(val, "<>")
		case "Content-Type":
			rec.ContentType = val
		case "Content-Length":
			contentLen, _ = strconv.Atoi(val)
		default:
			rec.Custom[key] = val
		}
	}

	// Read the exact content block.
	rec.Block = make([]byte, contentLen)
	if _, err := io.ReadFull(br, rec.Block); err != nil {
		return nil, fmt.Errorf("warc: failed to read block: %w", err)
	}

	// Consume the trailing record separator (\r\n\r\n).
	_, _ = br.Discard(4)

	return rec, nil
}
