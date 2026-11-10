# portscan

Package portscan предоставляет инструменты для конкурентного сканирования портов.

## Example

```bash
go get github.com/kaluginDS/portscan@latest
```

```go
import "portscan"

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
```

</p>
</details>

## Index

- [type Ports](<#Ports>)
  - [func \(p Ports\) List\(ports ...int\) Ports](<#Ports.List>)
  - [func \(p Ports\) Range\(start int, end int\) Ports](<#Ports.Range>)
- [type Result](<#Result>)
- [type Scanner](<#Scanner>)
  - [func New\(opts ...option\) \*Scanner](<#New>)
  - [func \(s \*Scanner\) Scan\(ctx context.Context, hosts \[\]string, ports Ports\) \(\<\-chan Result, error\)](<#Scanner.Scan>)
- [type State](<#State>)


<a name="Ports"></a>
## type Ports

Ports \- структура для генерации портов.

```go
type Ports struct {
    values []int
}
```

<a name="Ports.List"></a>
### func \(Ports\) List

```go
func (p Ports) List(ports ...int) Ports
```

List \- добавляет конкретные порты в список. Возвращает измененный объект Ports, метод поддерживает вызов по цепочке.

<a name="Ports.Range"></a>
### func \(Ports\) Range

```go
func (p Ports) Range(start int, end int) Ports
```

Range \- добавляет в список портов диапазон от start до end включительно. Функция валидирует входные параметры. Диапазон не будет добавлен, если:

- start больше end
- start меньше минимально допустимого порта \(1\)
- end больше максимально допустимого порта \(65535\)

Возвращает измененный объект Ports, метод поддерживает вызов по цепочке.

<a name="Result"></a>
## type Result

Result содержит полную информацию о результате сканирования одной комбинации хост:порт.

```go
type Result struct {
    Host     string        // Имя хоста или IP-адрес, переданный для проверки
    Ip       net.IP        // Разрешенный (resolved) IP-адрес узла
    Port     int           // Номер сканируемого TCP-порта
    State    State         // Статус порта по результатам проверки
    Duration time.Duration // Время, затраченное на попытку подключения
    Err      error         // Исходная ошибка, если State отличный от StateOpen
}
```

<a name="Scanner"></a>
## type Scanner

Scanner представляет собой конфигурацию сканера портов.

```go
type Scanner struct {
    // contains filtered or unexported fields
}
```

<a name="New"></a>
### func New

```go
func New(opts ...option) *Scanner
```

New создает и возвращает новый экземпляр Scanner. По умолчанию инициализируется с 10 воркерами и таймаутом в 1 секунду. Принимает опциональные параметры WithConcurrency и WithConnectTimeout для кастомизации.

<a name="Scanner.Scan"></a>
### func \(\*Scanner\) Scan

```go
func (s *Scanner) Scan(ctx context.Context, hosts []string, ports Ports) (<-chan Result, error)
```

Scan асинхронно сканирует все уникальные комбинации переданных hosts и ports. Учитывает отмену контекста. Если контекст истечет или будет отменен, сканирование прервется. Возвращает канал \<\-chan Result, в который по мере готовности отправляются результаты проверок. Возвращает error, если количество hosts или ports равно 0.

<a name="State"></a>
## type State

State представляет текущий статус проверки конкретного TCP\-порта.

```go
type State string
```

<a name="StateOpen"></a>

```go
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
```