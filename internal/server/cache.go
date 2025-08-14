package server

import (
	"context"
	"math"
	"time"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/dhcmrlchtdj/godns/internal/client"
	"github.com/dhcmrlchtdj/godns/internal/util"
)

type cachedAnswer struct {
	expired time.Time
	answer  []dns.RR
}
type deferredAnswer = util.Deferred[*cachedAnswer]

///

func (s *DnsServer) cacheGet(ctx context.Context, key string) ([]dns.RR, error) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "server.cache.get").
		Str("key", key).
		Logger()

	deferred, found := s.cache.Get(key)
	if !found {
		logger.Trace().Msg("missed")
		return nil, nil
	}

	cached, err := deferred.Wait()
	if err != nil {
		var errRcode client.DnsResponseError
		if errors.As(err, errRcode) {
			s.cache.Delete(key)
			logger.Trace().Str("rcode", dns.RcodeToString[errRcode.Rcode]).Msg("rcode")
		}
		return nil, err // nolint: wrapcheck
	}

	sec := math.Ceil(time.Until(cached.expired).Seconds())
	if sec <= 0 {
		s.cache.Delete(key)
		logger.Trace().Msg("expired")

		// If the upstream return a empty answer, the ttl will be set to 0,
		// the answer will be considered expired when it added to cache.
		// Here we set sec=1 to reuse the expired cache.
		sec = 1
	}
	ttl := uint32(sec)

	for idx := range cached.answer {
		cached.answer[idx].Header().Ttl = ttl
	}

	logger.Debug().Uint32("TTL", ttl).Msg("hit")

	return cached.answer, nil
}

func (s *DnsServer) cacheSet(ctx context.Context, key string, deferred *deferredAnswer) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "server.cache.set").
		Str("key", key).
		Logger()

	s.cache.Mutate(key, func(old *deferredAnswer, found bool) (*deferredAnswer, bool) {
		if found {
			logger.Trace().Msg("existed")
			return old, true
		} else {
			logger.Trace().Msg("added")
			return deferred, true
		}
	})
}

func (s *DnsServer) cacheResolve(ctx context.Context, key string, answer []dns.RR) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "server.cache.resolve").
		Str("key", key).
		Logger()

	deferred, found := s.cache.Get(key)
	if !found {
		logger.Trace().Msg("missed")
		return
	}

	// limit the max ttl to 1 hour
	maxTtl := uint32(60 * 60)
	ttl := maxTtl
	// set ttl to the minimum ttl among all answers
	for _, ans := range answer {
		currTtl := ans.Header().Ttl
		if currTtl < ttl {
			ttl = currTtl
		}
	}
	if len(answer) == 0 {
		ttl = 0
	}

	ans := cachedAnswer{
		answer:  answer,
		expired: time.Now().Add(time.Duration(ttl) * time.Second),
	}
	deferred.Resolve(&ans)

	logger.Trace().Uint32("TTL", ttl).Msg("resolved")
}

func (s *DnsServer) cacheReject(ctx context.Context, key string, rcode int) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "server.cache.reject").
		Str("key", key).
		Logger()

	deferred, found := s.cache.Delete(key)
	if !found {
		logger.Trace().Msg("missed")
		return
	}

	deferred.Reject(client.DnsResponseError{Rcode: rcode})
	logger.Trace().Msg("rejected")

	logger.Trace().Str("rcode", dns.RcodeToString[rcode]).Msg("rcode")
}

///

func (s *DnsServer) cleanupExpiredCache() {
	logger := zerolog.Ctx(s.ctx).
		With().
		Str("module", "server.cache.cleanup").
		Logger()

	ticker := time.NewTicker(time.Minute * 10)
	for {
		select {
		case <-ticker.C:
			logger.Trace().Msg("cleaning")
			start := time.Now()
			var allKey []string
			s.cache.Range(func(key string, val *deferredAnswer) bool {
				allKey = append(allKey, key)
				return true
			})
			for _, key := range allKey {
				deferred, found := s.cache.Get(key)
				if !found {
					continue
				}

				cached, err := deferred.Wait()
				if err != nil {
					var errRcode client.DnsResponseError
					if errors.As(err, errRcode) {
						s.cache.Delete(key)
						logger.Trace().Str("rcode", dns.RcodeToString[errRcode.Rcode]).Msg("rcode")
					}
				}

				sec := math.Ceil(time.Until(cached.expired).Seconds())
				if sec <= 0 {
					s.cache.Delete(key)
					logger.Trace().Msg("expired")
				}
			}
			used := time.Since(start)
			logger.Trace().Dur("duration", used).Msg("cleaned")
		case <-s.ctx.Done():
			ticker.Stop()
			return
		}
	}
}
