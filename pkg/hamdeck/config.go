package hamdeck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
)

const (
	ConfigDefaultFilename = "hamdeck.json"
	ConfigMainKey         = "hamdeck"
	ConfigButtons         = "buttons"
	ConfigType            = "type"
	ConfigIndex           = "index"
)

func (d *HamDeck) ReadConfig(r io.Reader) error {
	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(r)
	if err != nil {
		return fmt.Errorf("cannot read the configuration: %w", err)
	}

	var rawData interface{}
	err = json.Unmarshal(buffer.Bytes(), &rawData)
	if err != nil {
		return fmt.Errorf("cannot unmarshal the configuration: %w", err)
	}

	configuration, ok := rawData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("configuration is of wrong type: %T", rawData)
	}

	rawSubconfiguration, ok := configuration[ConfigMainKey]
	if !ok {
		return d.AttachConfiguredButtons(configuration)
	}

	subconfiguration, ok := rawSubconfiguration.(map[string]interface{})
	if !ok {
		return d.AttachConfiguredButtons(configuration)
	}
	return d.AttachConfiguredButtons(subconfiguration)
}

func (d *HamDeck) AttachConfiguredButtons(configuration map[string]interface{}) error {
	rawButtons, ok := configuration[ConfigButtons]
	if !ok {
		return fmt.Errorf("configuration contains no 'buttons' key")
	}

	buttons, ok := rawButtons.([]interface{})
	if !ok {
		return fmt.Errorf("'buttons' is not a list of button objects")
	}

	for i, rawButtonConfig := range buttons {
		buttonConfig, ok := rawButtonConfig.(map[string]interface{})
		if !ok {
			log.Printf("buttons[%d] is not a button object", i)
			continue
		}

		buttonIndex, ok := ToInt(buttonConfig[ConfigIndex])
		if !ok {
			log.Printf("buttons[%d] has no valid index", i)
			continue
		}

		var button Button
		for _, factory := range d.factories {
			button = factory.CreateButton(buttonConfig)
			if button != nil {
				break
			}
		}
		if button == nil {
			log.Printf("no factory found for buttons[%d]", i)
			continue
		}

		d.Attach(buttonIndex, button)
	}

	return nil
}

func ToInt(raw interface{}) (int, bool) {
	if raw == nil {
		return 0, false
	}
	switch i := raw.(type) {
	case int:
		return i, true
	case float64:
		return int(i), true
	case string:
		parsedI, err := strconv.Atoi(i)
		if err != nil {
			return 0, false
		}
		return parsedI, true
	default:
		return 0, false
	}
}

func ToFloat(raw interface{}) (float64, bool) {
	if raw == nil {
		return 0, false
	}
	switch f := raw.(type) {
	case int:
		return float64(f), true
	case float64:
		return f, true
	case string:
		parsedF, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return 0, false
		}
		return parsedF, true
	default:
		return 0, false
	}
}

func ToString(raw interface{}) (string, bool) {
	if raw == nil {
		return "", false
	}
	switch s := raw.(type) {
	case int:
		return fmt.Sprintf("%d", s), true
	case float64:
		return fmt.Sprintf("%f", s), true
	case string:
		return s, true
	default:
		return "", false
	}
}

func ToStringArray(raw interface{}) ([]string, bool) {
	if raw == nil {
		return []string{}, false
	}
	rawValues, ok := raw.([]interface{})
	if !ok {
		return []string{}, false
	}
	result := make([]string, len(rawValues))
	for i, rawValue := range rawValues {
		value, ok := ToString(rawValue)
		if ok {
			result[i] = value
		} else {
			return []string{}, false
		}
	}
	return result, true
}
