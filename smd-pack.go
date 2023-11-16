package modules

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
)

// Функция создания строки для вывода в лог ошибок
func CreateErrorLogParams(actionName string, s interface{}) (result string, err error) {
	if !IsStruct(s) {
		err = fmt.Errorf("передаваемый параметр должен быть структурой")
		return "", err
	}

	structType := reflect.TypeOf(s)
	structVal := reflect.ValueOf(s)
	fieldNum := structVal.NumField()

	for i := 0; i < fieldNum; i++ {
		fieldName := structType.Field(i).Name
		// Если значение IsZero(), то при вызове Interface() упадёт panic
		// С помощью этой конструкции избегаем ошибки, игнорируем IsZero поля
		if structVal.Field(i).IsZero() {
			result += fmt.Sprintf("%v: %v\n", fieldName, nil)
			continue
		}

		fieldValue := structVal.Field(i).Elem().Interface()
		result += fmt.Sprintf("%v: %v\n", fieldName, fieldValue)
	}

	return fmt.Sprintf("method: %s\n%s", actionName, result), nil
}

// Проверяет, является ли переданная переменная типом структура
func IsStruct(v interface{}) bool {
	vType := reflect.ValueOf(v)

	return vType.Kind() == reflect.Struct
}

// Функция проверки структуры на nil поля
func HaveStructNilField(s interface{}) bool {
	structVal := reflect.ValueOf(s)
	fieldNum := structVal.NumField()

	for i := 0; i < fieldNum; i++ {
		field := structVal.Field(i)

		// Стандартная проверка reflect на нулевые значения
		if field.IsZero() {
			return true
		}

		// Проверка строк на пробел
		if field.Elem().Kind() == reflect.String && len(field.Elem().String()) == 0 {
			return true
		}
	}
	return false
}

// Генерация случайной строки
func GenerateRandomString(keys int, length int) string {
	lettersAndNumbersString := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	numbersString := "0123456789"

	var keysString string

	if keys == 0 {
		keysString = lettersAndNumbersString
	} else if keys == 1 {
		keysString = numbersString
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = keysString[rand.Intn(len(keysString))]
	}

	return string(result)
}

// Заполнение структуры из []map[string]interface
func PopulateStructFromSelect(rows []map[string]interface{}, s interface{}) error {
	sValue := reflect.ValueOf(s)
	if sValue.Kind() != reflect.Ptr || sValue.IsNil() {
		return fmt.Errorf("populateStruct: expected a pointer to a struct, got %v", sValue.Type())
	}
	sType := sValue.Elem().Type()
	sValue = sValue.Elem()

	for _, row := range rows {
		for i := 0; i < sType.NumField(); i++ {
			field := sType.Field(i)
			if value, ok := row[field.Name]; ok {
				sField := sValue.FieldByName(field.Name)
				if sField.IsValid() && sField.CanSet() {
					sField.Set(reflect.ValueOf(value))
				}
			}
		}
	}

	return nil
}

func IsVersionActual(userVersion string, targetVersion string) (bool, error) {
	userVersionNum, err := versionStringToNumber(userVersion)
	if err != nil {
		return false, err
	}

	targetVersionNum, err := versionStringToNumber(targetVersion)
	if err != nil {
		return false, err
	}

	if userVersionNum >= targetVersionNum {
		return true, nil // Версия младше
	} else {
		return false, nil // Версия старше
	}
}

func versionStringToNumber(version string) (float64, error) {
	versionParts := strings.Split(version, ".")
	if len(versionParts) == 0 {
		return 0, fmt.Errorf("некорректный формат версии: %s", version)
	}

	versionNumber, err := strconv.ParseFloat(strings.Join(versionParts, ""), 64)
	if err != nil {
		return 0, fmt.Errorf("не получилось спарсить версию: %s", err)
	}

	return versionNumber, nil
}

// Конвертирует структуру в map[string]interface{}
func StructToMap(s interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Получаем тип входной структуры
	structType := reflect.TypeOf(s)

	// Убедимся, что входной параметр является структурой
	if structType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("входящий параметр не структура")
	}

	// Получаем значение структуры
	structValue := reflect.ValueOf(s)

	// Итерируемся по полям структуры
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		// Преобразовываем имя поля в строку
		fieldName := field.Tag.Get("json")

		// Если имя не указано в теге json, используем имя поля
		if fieldName == "" {
			fieldName = field.Name
		}

		// Преобразовываем значение поля в интерфейс и добавляем его в map
		result[fieldName] = fieldValue.Interface()
	}

	return result, nil
}
