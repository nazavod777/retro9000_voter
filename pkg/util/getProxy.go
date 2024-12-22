package util

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"regexp"
	"strings"
	"sync"
)

type ProxyCycler struct {
	proxies []string
	index   int
	mu      sync.Mutex
}

var Proxies []string
var ProxiesCycler *ProxyCycler

func parseProxy(proxy string) (string, error) {
	patterns := []struct {
		regex    *regexp.Regexp
		template string
	}{
		// ip:port
		{
			regexp.MustCompile(`^([^:@]+):(\d+)$`),
			"%s://%s:%s",
		},
		// scheme://ip:port
		{
			regexp.MustCompile(`^((?:http|https|socks4|socks5)://)([^:@]+):(\d+)$`),
			"%s%s:%s",
		},
		// scheme://user:pass@ip:port
		{
			regexp.MustCompile(`^((?:http|https|socks4|socks5)://)?([^:@]+):([^:@]+)@([^:@]+):(\d+)$`),
			"%s://%s:%s@%s:%s",
		},
		// scheme://user:pass:ip:port
		{
			regexp.MustCompile(`^((?:http|https|socks4|socks5)://)?([^:@]+):([^:@]+):([^:@]+):(\d+)$`),
			"%s://%s:%s@%s:%s",
		},
		// scheme://ip:port@user:pass
		{
			regexp.MustCompile(`^((?:http|https|socks4|socks5)://)?([^:@]+):(\d+)@([^:@]+):([^:@]+)$`),
			"%s://%s:%s@%s:%s",
		},
		// scheme://ip:port:user:pass
		{
			regexp.MustCompile(`^((?:http|https|socks4|socks5)://)?([^:@]+):(\d+):([^:@]+):([^:@]+)$`),
			"%s://%s:%s@%s:%s",
		},
	}

	for _, pattern := range patterns {
		matches := pattern.regex.FindStringSubmatch(proxy)
		if matches == nil {
			continue
		}

		switch len(matches) {
		case 3: // Простой формат ip:port
			return fmt.Sprintf(pattern.template, "http", matches[1], matches[2]), nil
		case 4: // Формат scheme://ip:port
			return fmt.Sprintf(pattern.template, matches[1], matches[2], matches[3]), nil
		case 6: // Форматы с user:pass
			scheme := matches[1]
			if scheme == "" {
				scheme = "http://"
			}
			scheme = strings.TrimSuffix(scheme, "://")

			if strings.Contains(pattern.template, "@") {
				if isPort(matches[3]) {
					return fmt.Sprintf(pattern.template, scheme, matches[4], matches[5], matches[2], matches[3]), nil
				}
				return fmt.Sprintf(pattern.template, scheme, matches[2], matches[3], matches[4], matches[5]), nil
			}
		}
	}

	return "", fmt.Errorf("invalid proxy format: %s", proxy)
}

func isPort(s string) bool {
	match, _ := regexp.MatchString(`^\d+$`, s)
	return match
}

func (pc *ProxyCycler) Next() string {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if len(pc.proxies) == 0 {
		return ""
	}

	proxy := pc.proxies[pc.index]
	pc.index = (pc.index + 1) % len(pc.proxies)

	return proxy
}

func InitProxies(proxyPath string) error {
	proxiesFile, err := ReadFileByRows(proxyPath)

	if err != nil {
		return fmt.Errorf("error When Reading Proxy: %v", err)
	}

	for _, proxy := range proxiesFile {
		parsedProxy, err := parseProxy(proxy)

		if err != nil {
			log.Printf("Error When Parsing Proxy %s: %s", proxy, err)
			continue
		}

		Proxies = append(Proxies, parsedProxy)
	}

	ProxiesCycler = &ProxyCycler{proxies: Proxies, index: 0}

	return nil
}

func GetProxy() string {
	if len(Proxies) == 0 {
		return ""
	}
	return Proxies[rand.Intn(len(Proxies))]
}
