package portscan

// Ports - структура для генерации портов.
type Ports struct {
	values []int
}

// List - добавляет конкретные порты в список.
// Возвращает измененный объект Ports, метод поддерживает вызов по цепочке.
func (p Ports) List(ports ...int) Ports {
	p.values = append(p.values, ports...)
	return p
}

// Range - добавляет в список портов диапазон от start до end включительно.
// Функция валидирует входные параметры. Диапазон не будет добавлен, если:
//   - start больше end
//   - start меньше минимально допустимого порта (1)
//   - end больше максимально допустимого порта (65535)
//
// Возвращает измененный объект Ports, метод поддерживает вызов по цепочке.
func (p Ports) Range(start int, end int) Ports {
	if start > end || start < 1 || end > 65535 {
		return p
	}

	for i := start; i <= end; i++ {
		p.values = append(p.values, i)
	}

	return p
}
