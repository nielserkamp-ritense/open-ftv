package x509

import (
	"crypto/rsa"
	"crypto/x509"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Fix(t *testing.T) {
	RootCA()

	now := time.Now().UTC().Truncate(time.Second)
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	testCases := []struct {
		name       string
		ca         *x509.Certificate
		caKey      *rsa.PrivateKey
		from       time.Time
		to         time.Time
		expire     time.Duration
		wantCA     *x509.Certificate
		wantKey    *rsa.PrivateKey
		wantFrom   time.Time
		wantTo     time.Time
		wantExpire time.Duration
	}{
		{
			name:       "all default",
			wantCA:     RootCA(),
			wantKey:    RootKey(),
			wantFrom:   now,
			wantTo:     now.Add(5 * time.Second),
			wantExpire: 5 * time.Second,
		},
		{
			name:       "CA",
			ca:         RootCA(),
			caKey:      RootKey(),
			wantCA:     RootCA(),
			wantKey:    RootKey(),
			wantFrom:   now,
			wantTo:     now.Add(5 * time.Second),
			wantExpire: 5 * time.Second,
		},
		{
			name:       "from",
			from:       yesterday,
			wantCA:     RootCA(),
			wantKey:    RootKey(),
			wantFrom:   yesterday,
			wantTo:     yesterday.Add(5 * time.Second),
			wantExpire: 5 * time.Second,
		},
		{
			name:       "to",
			to:         tomorrow,
			wantCA:     RootCA(),
			wantKey:    RootKey(),
			wantFrom:   now,
			wantTo:     tomorrow,
			wantExpire: 5 * time.Second,
		},
		{
			name:       "expires",
			expire:     time.Hour,
			wantCA:     RootCA(),
			wantKey:    RootKey(),
			wantFrom:   now,
			wantTo:     now.Add(time.Hour),
			wantExpire: time.Hour,
		},
		{
			name:       "all but expires",
			ca:         RootCA(),
			caKey:      RootKey(),
			from:       yesterday,
			to:         tomorrow,
			wantCA:     RootCA(),
			wantKey:    RootKey(),
			wantFrom:   yesterday,
			wantTo:     tomorrow,
			wantExpire: 5 * time.Second,
		},
		{
			name:       "all but to",
			ca:         RootCA(),
			caKey:      RootKey(),
			from:       yesterday,
			expire:     48 * time.Hour,
			wantCA:     RootCA(),
			wantKey:    rootKey,
			wantFrom:   yesterday,
			wantTo:     yesterday.Add(48 * time.Hour),
			wantExpire: 48 * time.Hour,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{
				CA:           tc.ca,
				CAKey:        tc.caKey,
				ValidFrom:    tc.from,
				ValidTo:      tc.to,
				ExpiresAfter: tc.expire,
			}

			cfg.fix()

			assert.Equal(t, tc.wantCA, cfg.CA)
			assert.Equal(t, tc.wantKey, cfg.CAKey)
			assert.Equal(t, tc.wantExpire, cfg.ExpiresAfter)

			switch {
			case tc.wantFrom.IsZero() && tc.wantTo.IsZero():
				assert.GreaterOrEqual(t, cfg.ValidFrom, now.Add(-time.Second))
				assert.LessOrEqual(t, cfg.ValidFrom, now.Add(time.Second))
				assert.GreaterOrEqual(t, cfg.ValidTo, now.Add(cfg.ExpiresAfter-time.Second))
				assert.LessOrEqual(t, cfg.ValidTo, now.Add(cfg.ExpiresAfter+time.Second))

			case tc.wantFrom.IsZero():
				assert.GreaterOrEqual(t, cfg.ValidFrom, now.Add(-time.Second))
				assert.LessOrEqual(t, cfg.ValidFrom, now.Add(time.Second))
				assert.Equal(t, tc.to, cfg.ValidTo)

			case tc.wantTo.IsZero():
				assert.Equal(t, tc.wantFrom, cfg.ValidFrom)
				assert.GreaterOrEqual(t, cfg.ValidTo, now.Add(cfg.ExpiresAfter-time.Second))
				assert.LessOrEqual(t, cfg.ValidTo, now.Add(cfg.ExpiresAfter+time.Second))

			default:
				assert.Equal(t, tc.wantFrom, cfg.ValidFrom)
				assert.Equal(t, tc.wantTo, cfg.ValidTo)
			}
		})
	}
}
