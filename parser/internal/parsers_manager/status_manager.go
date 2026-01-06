// status_manager.go
package parsers_manager

import "parser/internal/interfaces"

// метод для обновления статуса одного парсера в мэнеджере состояний парсеров
func (pm *ParsersManager) updateParserStatus(name string, success bool, err error) {
	// проверяем, что был создан менеджер сосстояний парсеров
	if pm.parsersStatusManager != nil {
		pm.parsersStatusManager.UpdateStatus(name, success, err)
	}
}

// метод для обновления статуса всех парсеров в мэнеджере состояний парсеров
func (pm *ParsersManager) updateAllParsersStatus(success bool) {
	// проверяем, что был создан менеджер сосстояний парсеров
	if pm.parsersStatusManager != nil {
		for _, name := range pm.getAllParsersNames() {
			pm.parsersStatusManager.UpdateStatus(name, success, nil)
		}
	}
}

// метод парсер-мэнеджера получения списка парсеров из мэнеджера СОСТОЯНИЯ парсеров
func (pm *ParsersManager) getHealthyParsers() []string {
	// проверяем, что был создан менеджер сосстояний парсеров
	if pm.parsersStatusManager != nil {
		return pm.parsersStatusManager.GetHealthyParsers()
	}
	return []string{}
}

// метод для получения слайса имён, парсеров, которые зарегестрированы в парсер-менеджере
func (pm *ParsersManager) getAllParsersNames() []string {
	names := make([]string, len(pm.parsers))
	for i, parser := range pm.parsers {
		names[i] = parser.GetName()
	}
	return names
}

// метод поиска парсера по имени
func (pm *ParsersManager) findParserByName(name string) interfaces.Parser {
	for _, parser := range pm.parsers {
		if parser.GetName() == name {
			return parser
		}
	}
	return nil
}
