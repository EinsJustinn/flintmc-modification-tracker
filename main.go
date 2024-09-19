package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
)

type Config struct {
	Modification string `json:"modification"`
	WebhookUrl   string `json:"webhookUrl"`
}

type Modification struct {
	Id               int    `json:"id"`
	Namespace        string `json:"namespace"`
	Name             string `json:"name"`
	Featured         bool   `json:"featured"`
	Verified         bool   `json:"verified"`
	Organization     int    `json:"organization"`
	Author           string `json:"author"`
	Downloads        int    `json:"downloads"`
	DownloadString   string `json:"download_string"`
	ShortDescription string `json:"short_description"`
	Rating           struct {
		Count  int     `json:"count"`
		Rating float64 `json:"rating"`
	} `json:"rating"`
	Changelog            string        `json:"changelog"`
	RequiredLabymodBuild int           `json:"required_labymod_build"`
	Releases             int           `json:"releases"`
	LastUpdate           int           `json:"last_update"`
	Licence              string        `json:"licence"`
	VersionString        string        `json:"version_string"`
	Meta                 []string      `json:"meta"`
	Dependencies         []interface{} `json:"dependencies"`
	Permissions          []string      `json:"permissions"`
	SourceUrl            string        `json:"source_url"`
	BrandImages          []struct {
		Type string `json:"type"`
		Hash string `json:"hash"`
	} `json:"brand_images"`
	Tags []int `json:"tags"`
}

var (
	config Config
)

func main() {
	c, err := getConfig()
	if err != nil {
		fmt.Println("Error getting config:", err)
		os.Exit(2)
		return
	}
	config = c

	if config.Modification == "" {
		fmt.Println("Please enter a modification")
		os.Exit(2)
		return
	}
	if config.WebhookUrl == "" {
		fmt.Println("Please enter a webhook url")
		os.Exit(2)
		return
	}

	m, err := fetchModification()
	if err != nil {
		fmt.Println("Error fetching modification:", err)
		os.Exit(2)
		return
	}

	modification := m

	latest, err := getLatest()
	if err != nil {
		fmt.Println("Error getting latest modification:", err)
	} else {
		checkDifferent(modification, latest)
	}

	err = saveLatest(modification)
	if err != nil {
		fmt.Println(err)
	}
}

func checkDifferent(modification Modification, latest Modification) {
	modificationValue := reflect.ValueOf(modification)
	latestValue := reflect.ValueOf(latest)

	modificationType := modificationValue.Type()

	for i := 0; i < modificationValue.NumField(); i++ {
		field := modificationType.Field(i)
		newField := modificationValue.Field(i)
		latestField := latestValue.Field(i)

		if !reflect.DeepEqual(newField.Interface(), latestField.Interface()) {

			author := Author{
				Name: modification.Name,
				Url:  fmt.Sprintf("https://flintmc.net/modification/%d.%s", modification.Id, modification.Namespace),
			}

			embed := Embed{
				Author:      author,
				Title:       "Change: " + field.Name,
				Url:         "",
				Description: fmt.Sprintf("%v\n->\n%v", latestField.Interface(), newField.Interface()),
				Color:       6689010,
				Fields:      Field{},
				Thumbnail:   Thumbnail{},
				Image:       Image{},
				Footer:      Footer{},
			}

			webhook := Webhook{
				Username:  "FlintMC Modification Tracker",
				AvatarUrl: "https://avatars.githubusercontent.com/u/76062092",
				Content:   "",
				Embeds:    []Embed{embed},
			}

			err := sendWebhook(config.WebhookUrl, webhook)
			if err != nil {
				fmt.Println("Error sending webhook:", err)
				os.Exit(2)
				return
			}
		}
	}
}

func getLatest() (Modification, error) {
	var modification Modification
	file, _ := os.Open("latest.json")
	decoder := json.NewDecoder(file)
	err := decoder.Decode(&modification)
	if err != nil {
		return modification, err
	}
	return modification, nil
}

func saveLatest(modification Modification) error {

	jsonString, err := json.MarshalIndent(modification, "", "  ")
	if err != nil {
		return err
	}
	create, err := os.Create("latest.json")
	if err != nil {
		return err
	}
	defer create.Close()
	_, err = create.Write(jsonString)
	if err != nil {
		return err
	}

	return nil
}

func fetchModification() (Modification, error) {
	url := fmt.Sprintf("https://flintmc.net/api/client-store/get-modification/%s", config.Modification)

	resp, err := http.Get(url)
	if err != nil {
		return Modification{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Modification{}, err
	}
	var modification Modification
	err = json.Unmarshal(body, &modification)
	if err != nil {
		return Modification{}, err
	}

	return modification, nil
}

func getConfig() (Config, error) {
	var config Config

	file, err := os.Open("config.json")
	if os.IsNotExist(err) {
		defaultConfig := Config{
			Modification: "MODIFICATION_NAMESPACE",
			WebhookUrl:   "DISCORD_WEBHOOK_URL",
		}

		err := saveConfig(defaultConfig)
		if err != nil {
			return config, fmt.Errorf("failed to create config: %v", err)
		}

		return defaultConfig, nil
	} else if err != nil {
		return config, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return config, fmt.Errorf("failed to decode config file: %v", err)
	}

	return config, nil
}

func saveConfig(config Config) error {

	file, err := os.Create("config.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(config)
	if err != nil {
		return err
	}

	return nil
}
