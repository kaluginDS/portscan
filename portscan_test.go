package portscan

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestPortscan(t *testing.T) {
	// 1. Инициализация сканера с настройками конкурентности и таймаута подключения
	scanner := New(
		WithConcurrency(50),
		WithConnectTimeout(500*time.Millisecond),
	)

	// 2. Создание контекста с общим таймаутом на всё сканирование
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

	// 3. Подготовка целевых хостов (поддерживаются доменные имена, IPv4 и IPv6)
	hosts := []string{
		"127.0.0.1",
		"localhost",
		"google.com",
		"invalid.hostname.test",
		"2001:4860:4860::8888",
	}

	// 4. Формирование списка портов для проверки.
	ports := Ports{}.List(80).Range(1, 10)

	// 5. Запуск процесса сканирования.
	results, err := scanner.Scan(ctx, hosts, ports)
	if err != nil {
		fmt.Println("Error: ", err)
	}

	// 6. Чтение результатов из канала.
	for result := range results {
		if result.Err == nil {
			fmt.Printf("✅ %s:%d -> %s\n", result.Host, result.Port, result.State)
			fmt.Printf("Ip: %s\n", result.Ip.String())
			fmt.Printf("Duration: %s\n\n", result.Duration.String())
		} else {
			fmt.Printf("❌ %s:%d -> %s\n", result.Host, result.Port, result.State)
			fmt.Printf("Error: %v\n\n", result.Err)
		}
	}
}
