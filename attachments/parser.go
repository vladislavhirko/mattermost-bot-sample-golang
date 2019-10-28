package attachments

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func Parse() map[string]map[string]interface{} {
	if err := os.Chdir("attachments/src"); err != nil {
		log.Fatalln(err)
	}
	matches, err := filepath.Glob("*.json")
	if err != nil {
		log.Fatal(err)
	}
	result := make(map[string]map[string]interface{}, 1)
	for _, match := range matches {
		mapa := make(map[string]interface{}, 1)
		file, err := os.OpenFile(match, os.O_RDONLY, 0222)
		if err != nil {
			log.Fatalln(err)
		}
		fileDecoder := json.NewDecoder(file)
		if err = fileDecoder.Decode(&mapa); err != nil {
			log.Fatalln(err)
		}
		key := strings.TrimRight(match, ".json")
		result[key] = mapa
		file.Close()
	}
	return result
}
