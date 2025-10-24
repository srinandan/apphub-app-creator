// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"internal/cmd"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra/doc"
)

const ENABLED = "true"

var samples = `# apphub-app-creator command Samples

The following table contains some examples of apphub-app-creator.

| Operations | Command |
|---|---|
| generate | ` + getSingleLine(cmd.GetGenAppExample(0)) + `|
| generate | ` + getSingleLine(cmd.GetGenAppExample(1)) + `|
| generate | ` + getSingleLine(cmd.GetGenAppExample(2)) + `|
| generate | ` + getSingleLine(cmd.GetGenAppExample(3)) + `|
| generate | ` + getSingleLine(cmd.GetGenAppExample(4)) + `|
| generate | ` + getSingleLine(cmd.GetGenAppExample(5)) + `|
| generate | ` + getSingleLine(cmd.GetGenAppExample(6)) + `|
| generate | ` + getSingleLine(cmd.GetGenAppExample(7)) + `|
| delete   | ` + getSingleLine(cmd.GetDelAppExample(0)) + `|
| delete   | ` + getSingleLine(cmd.GetDelAppExample(1)) + `|


NOTE: This file is auto-generated during a release. Do not modify.`

func main() {
	var err error
	var docFiles []string

	if os.Getenv("CLI_SKIP_DOCS") != ENABLED {

		if docFiles, err = filepath.Glob("./docs/apphub-app-creator*.md"); err != nil {
			log.Fatal(err)
		}

		for _, docFile := range docFiles {
			if err = os.Remove(docFile); err != nil {
				log.Fatal(err)
			}
		}

		if err = doc.GenMarkdownTree(cmd.RootCmd, "./docs"); err != nil {
			log.Fatal(err)
		}
	}

	_ = writeByteArrayToFile("./samples/README.md", false, []byte(samples))
}

func getSingleLine(s string) string {
	return "`" + strings.ReplaceAll(strings.ReplaceAll(s, "\\", ""), "\n", "") + "`"
}

// writeByteArrayToFile accepts []bytes and writes to a file
func writeByteArrayToFile(exportFile string, fileAppend bool, payload []byte) error {
	fileFlags := os.O_CREATE | os.O_WRONLY

	if fileAppend {
		fileFlags |= os.O_APPEND
	} else {
		fileFlags |= os.O_TRUNC
	}

	f, err := os.OpenFile(exportFile, fileFlags, 0o644)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(payload)
	if err != nil {
		fmt.Println("error writing to file: ", err)
		return err
	}
	return nil
}
