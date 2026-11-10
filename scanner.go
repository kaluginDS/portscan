// Package portscan предоставляет инструменты для конкурентного сканирования портов.
package portscan

import (
	"time"
)

// Scanner представляет собой конфигурацию сканера портов.
type Scanner struct {
	concurrency    int
	connectTimeout time.Duration
}

type option func(*Scanner)

const (
	defaultConcurrency    = 10
	defaultConnectTimeout = 1 * time.Second
)

// WithConcurrency возвращает option, которая устанавливает максимальное количество одновременно выполняемых запросов.
func WithConcurrency(n int) option {
	return func(s *Scanner) {
		if n > 0 {
			s.concurrency = n
		}
	}
}

// WithConnectTimeout возвращает option, которая устанавливает таймаут для каждого устанавливаемого соединения.
func WithConnectTimeout(d time.Duration) option {
	return func(s *Scanner) {
		if d > 0 {
			s.connectTimeout = d
		}
	}
}

// New создает и возвращает новый экземпляр Scanner.
// По умолчанию инициализируется с 10 воркерами и таймаутом в 1 секунду.
// Принимает опциональные параметры WithConcurrency и WithConnectTimeout для кастомизации.
func New(opts ...option) *Scanner {
	s := &Scanner{
		concurrency:    defaultConcurrency,
		connectTimeout: defaultConnectTimeout,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}
