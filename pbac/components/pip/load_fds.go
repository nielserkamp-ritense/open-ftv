package pip

import (
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func (p *pip) loadFDS(path string) {
	err := filepath.WalkDir(path, func(path2 string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasPrefix(path2, ".") {
			f, err2 := os.Open(path2)
			if err2 != nil {
				return err2
			}
			defer f.Close()

			l := llConfig{}
			if err2 = yaml.NewDecoder(f).Decode(&l); err2 != nil {
				return err2
			}

			u, err3 := url.Parse(l.Endpoint)
			if err3 != nil {
				return err3
			}

			// for now we support only a single FDS!
			// TODO: support multiple.
			p.fds = newFDS(p.ctx, p.logger, u, l.TTL)
		}

		return nil
	})

	if err != nil {
		p.logger.Error("fds: error iterating folders", "path", path, "err", err)
	} else {
		p.logger.Info("fds: configuration loaded successfully", "path", path)
	}
}

type llConfig struct {
	Endpoint string        `json:"endpoint,omitempty"`
	TTL      time.Duration `json:"ttl,omitempty"`
}
