package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"text/template"
	"time"
)

type (
	configuration struct {
		Codename string
		Password string
		OSPath   string
	}
)

var (
	//go:embed templates/bloodhound.json.tmpl
	jsonConfiguration embed.FS

	//go:embed templates/config.tmpl
	configurationTemplate embed.FS
)

var (
	customName                = flag.String("name", "", "Preferred name in lowercase")
	operatingSystemFolderBase = flag.String("path", ".", "Preferred filesystem parent path")
	passwordChoices           = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
)

func init() {
	rand.Seed(time.Now().UnixNano())

	flag.Parse()

	if len(*customName) == 0 {
		log.Fatal("You must provide a preferred name")
	}
}

func passwordGenerator() string {
	b := make([]rune, 32)
	for i := range b {
		b[i] = passwordChoices[rand.Intn(len(passwordChoices))]
	}
	return string(b)
}

func main() {
	if *operatingSystemFolderBase == "." {
		*operatingSystemFolderBase, _ = os.Getwd()

		if _, err := os.Stat(*operatingSystemFolderBase); os.IsNotExist(err) {
			log.Fatal(err.Error())
		}
	}

	var (
		bhJSONTemplate = template.Must(template.ParseFS(jsonConfiguration, "templates/bloodhound.json.tmpl"))
		configTemplate = template.Must(template.ParseFS(configurationTemplate, "templates/config.tmpl"))

		operationMetadata = configuration{
			Codename: *customName,
			Password: passwordGenerator(),
			OSPath:   fmt.Sprintf("%s/neo4j-inst-%s", *operatingSystemFolderBase, *customName),
		}
	)

	switch runtime.GOOS {
	case "windows":
		operationMetadata.OSPath = strings.ReplaceAll(operationMetadata.OSPath, "/", "\\")
	}

	if _, err := os.Stat(operationMetadata.OSPath); os.IsNotExist(err) {
		err = os.Mkdir(operationMetadata.OSPath, 0755)
		if err != nil {
			log.Fatal(err.Error())
		}

		if _, err := os.Stat(fmt.Sprintf("%s/bloodhound.json", operationMetadata.OSPath)); os.IsNotExist(err) {
			f, err := os.Create(fmt.Sprintf("%s/bloodhound.json", operationMetadata.OSPath))
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			_ = bhJSONTemplate.Execute(f, operationMetadata)
		}

		if _, err := os.Stat(fmt.Sprintf("%s/docker-compose.yml", operationMetadata.OSPath)); os.IsNotExist(err) {
			f, err := os.Create(fmt.Sprintf("%s/docker-compose.yml", operationMetadata.OSPath))
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			_ = configTemplate.Execute(f, operationMetadata)

			fmt.Printf("Current password for your operation is: %s\n", operationMetadata.Password)
			fmt.Printf("Go to the following directory:          %s\n", operationMetadata.OSPath)
			fmt.Printf("Run the following:                      %s\n", "docker compose up -d OR docker-compose up -d")
		}
	} else {
		log.Fatal(fmt.Sprintf("The path specified already exists: %s\nPlease remove it and re-run if desired.\n", operationMetadata.OSPath))
	}
}
