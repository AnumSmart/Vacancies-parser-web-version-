package config

import (
	"os"
	"testing"
)

// Тестовые структуры для проверки
type TestConfig struct {
	Port    int    `yaml:"port"`
	Host    string `yaml:"host"`
	Enabled bool   `yaml:"enabled"`
}

func TestLoadYAMLConfig(t *testing.T) {

	// Создаем временный каталог для тестовых файлов
	tmpDir := t.TempDir()

	t.Run("пустой путь к конфигу", func(t *testing.T) {
		cfg, err := LoadYAMLConfig("", func() *TestConfig {
			return &TestConfig{
				Port:    8080,
				Host:    "localhost",
				Enabled: true,
			}
		})

		// Проверяем, что вернулась ошибка (согласно вашей реализации)
		if err == nil || err.Error() != "configPath == (пустая строка)" {
			t.Errorf("ожидалась ошибка 'configPath == (пустая строка)', получено: %v", err)
		}

		// Проверяем, что конфиг nil (так как возвращаете ошибку)
		if cfg != nil {
			t.Error("ожидался nil конфиг при ошибке")
		}
	})

	t.Run("файл не существует", func(t *testing.T) {
		nonExistentFile := tmpDir + "/nonexistent.yaml"

		cfg, err := LoadYAMLConfig(nonExistentFile, func() *TestConfig {
			return &TestConfig{
				Port:    3000,
				Host:    "127.0.0.1",
				Enabled: false,
			}
		})

		// Проверяем ошибку
		if err == nil || err.Error() != "файл по указанному пути не существует" {
			t.Errorf("ожидалась ошибка 'файл по указанному пути не существует', получено: %v", err)
		}

		if cfg != nil {
			t.Error("ожидался nil конфиг при ошибке")
		}
	})

	t.Run("успешная загрузка конфига", func(t *testing.T) {
		// Создаем тестовый YAML файл
		yamlContent := `
port: 9090
host: "example.com"
enabled: true
`
		configFile := tmpDir + "/test-config.yaml"
		err := os.WriteFile(configFile, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadYAMLConfig(configFile, func() *TestConfig {
			return &TestConfig{
				Port:    8080, // значения по умолчанию
				Host:    "localhost",
				Enabled: false,
			}
		})

		if err != nil {
			t.Errorf("не ожидалась ошибка: %v", err)
		}

		if cfg == nil {
			t.Fatal("конфиг не должен быть nil")
		}

		// Проверяем, что значения из файла перезаписали значения по умолчанию
		if cfg.Port != 9090 {
			t.Errorf("ожидался Port=9090, получено %d", cfg.Port)
		}
		if cfg.Host != "example.com" {
			t.Errorf("ожидался Host=example.com, получено %s", cfg.Host)
		}
		if !cfg.Enabled {
			t.Error("ожидался Enabled=true")
		}
	})

	t.Run("ошибка парсинга YAML", func(t *testing.T) {
		// Создаем некорректный YAML файл
		invalidYaml := `
port: "не число"  # строка вместо числа
host: example.com
`
		configFile := tmpDir + "/invalid-config.yaml"
		err := os.WriteFile(configFile, []byte(invalidYaml), 0644)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadYAMLConfig(configFile, func() *TestConfig {
			return &TestConfig{}
		})

		if err == nil {
			t.Error("ожидалась ошибка парсинга YAML")
		}

		if cfg != nil {
			t.Error("конфиг должен быть nil при ошибке парсинга")
		}
	})

	t.Run("пустой YAML файл - должны вернуться значения по умолчанию", func(t *testing.T) {
		configFile := tmpDir + "/empty-config.yaml"
		err := os.WriteFile(configFile, []byte(""), 0644)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadYAMLConfig(configFile, func() *TestConfig {
			return &TestConfig{
				Port:    1234,
				Host:    "default-host",
				Enabled: true,
			}
		})

		if err != nil {
			t.Errorf("не ожидалась ошибка: %v", err)
		}

		if cfg == nil {
			t.Fatal("конфиг не должен быть nil")
		}

		// Проверяем, что остались значения по умолчанию
		if cfg.Port != 1234 {
			t.Errorf("ожидался Port=1234, получено %d", cfg.Port)
		}
		if cfg.Host != "default-host" {
			t.Errorf("ожидался Host='default-host', получено %s", cfg.Host)
		}
		if !cfg.Enabled {
			t.Error("ожидался Enabled=true")
		}
	})

	t.Run("частичное заполнение конфига", func(t *testing.T) {
		yamlContent := `
port: 7777
# host не указан, должен остаться default
enabled: false
`
		configFile := tmpDir + "/partial-config.yaml"
		err := os.WriteFile(configFile, []byte(yamlContent), 0644)
		if err != nil {
			t.Fatal(err)
		}

		cfg, err := LoadYAMLConfig(configFile, func() *TestConfig {
			return &TestConfig{
				Port:    1111,
				Host:    "default",
				Enabled: true,
			}
		})

		if err != nil {
			t.Errorf("не ожидалась ошибка: %v", err)
		}

		if cfg == nil {
			t.Fatal("конфиг не должен быть nil")
		}

		if cfg.Port != 7777 {
			t.Errorf("ожидался Port=7777, получено %d", cfg.Port)
		}
		if cfg.Host != "default" {
			t.Errorf("ожидался Host='default', получено %s", cfg.Host)
		}
		if cfg.Enabled {
			t.Error("ожидался Enabled=false")
		}
	})
}
