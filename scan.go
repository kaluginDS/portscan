package portscan

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"syscall"
	"time"
)

// State представляет текущий статус проверки конкретного TCP-порта.
type State string

const (
	// StateOpen указывает, что TCP-соединение установлено
	StateOpen State = "open"

	// StateClosed указывает, что соединение явно отклонено
	StateClosed State = "close"

	// StateTimeout указывает, что подключение не завершилось за установленное время
	StateTimeout State = "timeout"

	// StateUnreachable указывает, что узел или сеть недоступны
	StateUnreachable State = "unreachable"

	// StateCanceled указывает, что проверка отменена
	StateCanceled State = "canceled"

	// StateError указывает, что произошла другая ошибка
	StateError State = "error"
)

// Result содержит полную информацию о результате сканирования одной комбинации хост:порт.
type Result struct {
	Host     string        // Имя хоста или IP-адрес, переданный для проверки
	Ip       net.IP        // Разрешенный (resolved) IP-адрес узла
	Port     int           // Номер сканируемого TCP-порта
	State    State         // Статус порта по результатам проверки
	Duration time.Duration // Время, затраченное на попытку подключения
	Err      error         // Исходная ошибка, если State отличный от StateOpen
}

type task struct {
	host string
	port int
}

// Scan асинхронно сканирует все уникальные комбинации переданных hosts и ports.
// Учитывает отмену контекста. Если контекст истечет или будет отменен, сканирование прервется.
// Возвращает канал <-chan Result, в который по мере готовности отправляются результаты проверок.
// Возвращает error, если количество hosts или ports равно 0.
func (s *Scanner) Scan(ctx context.Context, hosts []string, ports Ports) (<-chan Result, error) {
	if len(hosts) == 0 {
		return nil, errors.New("hosts list cannot be empty")
	}
	if len(ports.values) == 0 {
		return nil, errors.New("ports list cannot be empty")
	}

	uniqueHosts := uniqueStrings(hosts)
	uniquePorts := uniqueInts(ports.values)

	tasks := make(chan task, s.concurrency)
	results := make(chan Result, s.concurrency)

	var wg sync.WaitGroup

	for i := 0; i < s.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range tasks {
				results <- s.scanPort(ctx, t.host, t.port)
			}
		}()
	}

	go func() {
		defer close(tasks)
		for _, host := range uniqueHosts {
			for _, port := range uniquePorts {
				select {
				case <-ctx.Done():
					return
				default:
					tasks <- task{host: host, port: port}
				}
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	return results, nil
}

func (s *Scanner) scanPort(ctx context.Context, host string, port int) Result {
	result := Result{
		Host:  host,
		Port:  port,
		State: StateError,
	}

	if ctx.Err() != nil {
		result.State = StateCanceled
		result.Err = ctx.Err()
		return result
	}

	if host == "" {
		result.State = StateError
		result.Err = errors.New("empty host")
		return result
	}

	address := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	dialer := net.Dialer{Timeout: s.connectTimeout}

	startTime := time.Now()

	conn, err := dialer.DialContext(ctx, "tcp", address)

	duration := time.Since(startTime)
	result.Duration = duration

	if err != nil {
		result.State = StateError
		result.Err = err

		var opErr *net.OpError
		if errors.As(err, &opErr) {
			if opErr.Timeout() {
				result.State = StateTimeout
				return result
			}

			var sysErr *os.SyscallError
			if errors.As(opErr.Err, &sysErr) {

				var errno syscall.Errno
				if errors.As(sysErr.Err, &errno) {
					switch errno {
					case syscall.EHOSTUNREACH, syscall.ENETUNREACH:
						result.State = StateUnreachable
					case syscall.ECONNREFUSED:
						result.State = StateClosed
					}
				}
			}

			var dnsErr *net.DNSError
			if errors.As(opErr.Err, &dnsErr) {
				if dnsErr.IsNotFound {
					result.State = StateUnreachable
				}
			}
		}

		return result
	}
	_ = conn.Close()

	if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		result.Ip = tcpAddr.IP
	}

	result.State = StateOpen
	return result
}
