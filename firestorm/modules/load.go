package modules

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var api = "https://cloud.glowman554.de:3877"

func loadFileList(name string, version string) map[string]string {
	req, err := http.Get(fmt.Sprintf("%s/remote/info?name=%s&version=%s", api, name, version))
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()

	var fileIds map[string]int

	buffer, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(buffer, &fileIds)
	if err != nil {
		panic(err)
	}

	result := make(map[string]string)
	for filename, id := range fileIds {
		result[filename] = loadFile(name, version, filename, id)
	}
	return result
}

func loadFile(name string, version string, filename string, id int) string {
	// fmt.Printf("[LOADING] %s@%s:%s\n", name, version, filename)

	req, err := http.Get(fmt.Sprintf("%s/remote/get?id=%d", api, id))
	if err != nil {
		panic(err)
	}

	buffer, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}

	return string(buffer)
}
