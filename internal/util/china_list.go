package util

import (
	"bufio"
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type ChinaList struct {
	httpClient *http.Client
}

const CHINA_LIST_URL = "https://raw.githubusercontent.com/felixonmars/dnsmasq-china-list/master/accelerated-domains.china.conf"

func MakeChinaList(proxy string) *ChinaList {
	chinaList := ChinaList{
		httpClient: new(http.Client),
	}

	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			panic(err)
		}
		chinaList.httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}

	return &chinaList
}

func (c *ChinaList) Fetch(ctx context.Context) ([]string, error) {
	logger := zerolog.Ctx(ctx).
		With().
		Str("module", "china_list").
		Logger()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, CHINA_LIST_URL, http.NoBody)
	if err != nil {
		err = errors.WithStack(err)
		logger.Error().Err(err).Msg("failed to create request")
		return nil, err
	}

	logger.Trace().Msg("fetching china list")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		err = errors.WithStack(err)
		logger.Error().Err(err).Msg("failed to send request")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err = errors.WithStack(err)
		logger.Error().Err(err).Int("StatusCode", resp.StatusCode).Msg("StatusCode")
		return nil, err
	}

	chinalist := make([]string, 0, 64500)

	buf := bufio.NewScanner(resp.Body)
	for buf.Scan() {
		line := buf.Text()
		if line[0] == '#' {
			continue
		}
		e := strings.LastIndexByte(line, '/')
		chinalist = append(chinalist, line[8:e])
	}

	return chinalist, nil
}
