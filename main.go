package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
)

type (
	Config struct {
		Modification string `json:"modification"`
		WebhookUrl   string `json:"webhookUrl"`
	}

	Modification struct {
		Id                   int          `json:"id"`
		Namespace            string       `json:"namespace"`
		Name                 string       `json:"name"`
		Featured             bool         `json:"featured"`
		Verified             bool         `json:"verified"`
		Organization         int          `json:"organization"`
		Author               string       `json:"author"`
		Downloads            int          `json:"downloads"`
		DownloadString       string       `json:"download_string"`
		ShortDescription     string       `json:"short_description"`
		Rating               Rating       `json:"rating"`
		Changelog            string       `json:"changelog"`
		RequiredLabymodBuild int          `json:"required_labymod_build"`
		Releases             int          `json:"releases"`
		LastUpdate           int          `json:"last_update"`
		Licence              string       `json:"licence"`
		VersionString        string       `json:"version_string"`
		Meta                 []string     `json:"meta"`
		Dependencies         []any        `json:"dependencies"`
		Permissions          []string     `json:"permissions"`
		SourceUrl            string       `json:"source_url"`
		BrandImages          []BrandImage `json:"brand_images"`
		Tags                 []int        `json:"tags"`
	}

	BrandImage struct {
		Type string `json:"type"`
		Hash string `json:"hash"`
	}

	Rating struct {
		Count  int     `json:"count"`
		Rating float64 `json:"rating"`
	}
)

var (
	config Config
)

func main() {
	log.SetPrefix("")
	log.SetFlags(0)
	var err error
	config, err := getConfig()
	if err != nil {
		log.Fatalln("Error loading config:", err)
		return
	}

	if len(config.Modification) < 1 {
		log.Fatalln("Please enter a modification")
		return
	}
	if len(config.WebhookUrl) < 1 {
		log.Fatalln("Please enter a webhook url")
		return
	}

	modification, err := fetchModification()
	if err != nil {
		log.Fatalln("Error fetching modification:", err)
		return
	}

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

func checkDifferent(modification *Modification, latest Modification) {
	modificationValue := reflect.ValueOf(*modification)
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
				Description: fmt.Sprintf("%v\n->\n%v", latestField.Interface(), newField.Interface()),
				Color:       6689010,
			}

			webhook := Webhook{
				Username:  "FlintMC Modification Tracker",
				AvatarUrl: "https://avatars.githubusercontent.com/u/76062092",
				Embeds:    []Embed{embed},
			}

			err := sendWebhook(config.WebhookUrl, webhook)
			if err != nil {
				log.Fatalln("Error sending webhook:", err)
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

func saveLatest(modification *Modification) error {

	jsonString, err := json.MarshalIndent(*modification, "", "  ")
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

func fetchModification() (*Modification, error) {
	resp, err := http.Get(fmt.Sprintf("https://flintmc.net/api/client-store/get-modification/%s", config.Modification))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var modification Modification
	err = json.Unmarshal(body, &modification)
	if err != nil {
		return nil, err
	}

	return &modification, nil
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

	if err = json.NewDecoder(file).Decode(&config); err != nil {
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
