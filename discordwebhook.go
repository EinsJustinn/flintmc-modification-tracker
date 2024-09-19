package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Webhook struct {
	Username  string  `json:"username,omitempty"`
	AvatarUrl string  `json:"avatar_url,omitempty"`
	Content   string  `json:"content,omitempty"`
	Embeds    []Embed `json:"embeds,omitempty"`
}

type Embed struct {
	Author      Author    `json:"author,omitempty"`
	Title       string    `json:"title,omitempty"`
	Url         string    `json:"url,omitempty"`
	Description string    `json:"description,omitempty"`
	Color       int       `json:"color,omitempty"`
	Fields      Field     `json:"fields,omitempty"`
	Thumbnail   Thumbnail `json:"thumbnail,omitempty"`
	Image       Image     `json:"image,omitempty"`
	Footer      Footer    `json:"footer,omitempty"`
}

type Author struct {
	Name    string `json:"name,omitempty"`
	Url     string `json:"url,omitempty"`
	IconUrl string `json:"icon_url,omitempty"`
}

type Field struct {
	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
	Inline bool   `json:"inline,omitempty"`
}

type Thumbnail struct {
	Url string `json:"url,omitempty"`
}

type Image struct {
	Url string `json:"url,omitempty"`
}

type Footer struct {
	Text    string `json:"text,omitempty"`
	IconUrl string `json:"icon_url,omitempty"`
}

func sendWebhook(url string, webhook Webhook) error {

	buffer := &bytes.Buffer{}

	err := json.NewEncoder(buffer).Encode(webhook)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", buffer)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("webhook returned %s", resp.Status)
	}

	fmt.Println(resp.StatusCode)
	return nil
}
