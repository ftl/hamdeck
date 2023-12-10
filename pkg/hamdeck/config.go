package hamdeck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

const (
	ConfigDefaultFilename = "hamdeck.json"
	ConfigMainKey         = "hamdeck"
	ConfigStartPageID     = "start_page"
	ConfigPages           = "pages"
	ConfigButtons         = "buttons"
	ConfigType            = "type"
	ConfigIndex           = "index"
	ConfigConnections     = "connections"
)

func (d *HamDeck) ReadConfig(r io.Reader) error {
	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(r)
	if err != nil {
		return fmt.Errorf("cannot read the configuration: %w", err)
	}

	var rawData any
	err = json.Unmarshal(buffer.Bytes(), &rawData)
	if err != nil {
		return fmt.Errorf("cannot unmarshal the configuration: %w", err)
	}

	configuration, ok := rawData.(map[string]any)
	if !ok {
		return fmt.Errorf("configuration is of wrong type: %T", rawData)
	}
	effectiveConfiguration := findEffectiveConfiguration(configuration)

	d.buttonsPerFactory = make([]int, len(d.factories))
	d.connections = make(map[connectionKey]ConnectionConfig)
	d.pages = make(map[string]Page)

	connections, ok := (effectiveConfiguration[ConfigConnections]).(map[string]any)
	if ok {
		err = d.loadConnections(connections)
	}
	if err != nil {
		return err
	}

	d.startPageID, ok = effectiveConfiguration[ConfigStartPageID].(string)
	if !ok {
		d.startPageID = legacyPageID
	}
	pages, ok := effectiveConfiguration[ConfigPages].(map[string]any)
	if ok {
		err = d.loadPages(pages)
	}
	if err != nil {
		return err
	}

	buttons, ok := effectiveConfiguration[ConfigButtons].([]any)
	if ok {
		err = d.loadLegacyPage(buttons)
	} else if len(d.pages) == 0 {
		d.loadEmptyLegacyPage()
	}
	if err != nil {
		return err
	}

	return d.AttachPage(d.startPageID)
}

func findEffectiveConfiguration(configuration map[string]any) map[string]any {
	rawSubconfiguration, ok := configuration[ConfigMainKey]
	if !ok {
		return configuration
	}

	subconfiguration, ok := rawSubconfiguration.(map[string]any)
	if !ok {
		return configuration
	}
	return subconfiguration
}

func (d *HamDeck) loadConnections(configuration map[string]any) error {
	for name, config := range configuration {
		connection, ok := config.(map[string]any)
		if !ok {
			log.Printf("%s is not a valid connection configuration", name)
			continue
		}
		connectionType, ok := ToString(connection[ConfigType])
		if !ok {
			log.Printf("connection %s needs a type", name)
			continue
		}
		d.connections[connectionKey{name, connectionType}] = ConnectionConfig(connection)
	}
	return nil
}

func (d *HamDeck) loadPages(configuration map[string]any) error {
	for id, rawPage := range configuration {
		pageConfiguration, ok := rawPage.(map[string]any)
		if !ok {
			return fmt.Errorf("%s is not a valid page", id)
		}

		page, err := d.loadPage(id, pageConfiguration)
		if err != nil {
			return err
		}

		d.pages[id] = page
	}
	return nil
}

func (d *HamDeck) loadPage(id string, configuration map[string]any) (Page, error) {
	buttonsConfiguration, ok := configuration[ConfigButtons].([]any)
	if !ok {
		return Page{}, fmt.Errorf("page %s has no buttons defined", id)
	}

	buttons, err := d.loadButtons(buttonsConfiguration)
	if err != nil {
		return Page{}, err
	}

	return Page{
		buttons: buttons,
	}, nil
}

func (d *HamDeck) loadLegacyPage(configuration []any) error {
	buttons, err := d.loadButtons(configuration)
	if err != nil {
		return err
	}

	d.pages[legacyPageID] = Page{
		buttons: buttons,
	}
	return nil
}

func (d *HamDeck) loadEmptyLegacyPage() {
	d.pages[legacyPageID] = Page{
		buttons: make([]Button, len(d.buttons)),
	}
}

func (d *HamDeck) loadButtons(configuration []any) ([]Button, error) {
	result := make([]Button, len(d.buttons))
	for i, rawButtonConfig := range configuration {
		buttonConfig, ok := rawButtonConfig.(map[string]any)
		if !ok {
			log.Printf("buttons[%d] is not a button object", i)
			continue
		}

		buttonIndex, ok := ToInt(buttonConfig[ConfigIndex])
		if !ok {
			log.Printf("buttons[%d] has no valid index", i)
			continue
		}
		if buttonIndex < 0 || buttonIndex >= len(d.buttons) {
			log.Printf("%d is not a valid button index in [0, %d])", buttonIndex, len(d.buttons))
		}

		var button Button
		for j, factory := range d.factories {
			button = factory.CreateButton(buttonConfig)
			if button != nil {
				d.buttonsPerFactory[j] += 1
				break
			}
		}
		if button == nil {
			log.Printf("no factory found for buttons[%d]", i)
			continue
		}

		result[buttonIndex] = button
	}
	return result, nil
}

func (d *HamDeck) CloseUnusedFactories() {
	for i, factory := range d.factories {
		if d.buttonsPerFactory[i] == 0 {
			factory.Close()
		}
	}
}

func ToInt(raw any) (int, bool) {
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

func ToFloat(raw any) (float64, bool) {
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

func ToBool(raw any) (bool, bool) {
	if raw == nil {
		return false, false
	}
	switch s := raw.(type) {
	case int:
		return s != 0, true
	case float64:
		return s != 0, true
	case string:
		return strings.ToLower(s) == "true", true
	case bool:
		return s, true
	default:
		return false, false
	}
}

func ToString(raw any) (string, bool) {
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

func ToStringArray(raw any) ([]string, bool) {
	if raw == nil {
		return []string{}, false
	}
	rawValues, ok := raw.([]any)
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
