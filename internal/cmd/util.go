package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/pflag"
)

func GetStringParam(flag *pflag.Flag) (param string) {
	param = ""
	if flag != nil {
		param = flag.Value.String()
	}
	return param
}

// IsValidResourceFormat tests if a given string is of the format
// "projects/{some non-empty string}" or "folders/{some non-empty string}".
// It returns true if the format is valid, and false otherwise.
func IsValidResourceFormat(s string) bool {
	// Check if the string starts with "projects/"
	if strings.HasPrefix(s, "projects/") {
		// Ensure there is content after the prefix.
		// len("projects/") is 9.
		return len(s) > 9
	}

	// Check if the string starts with "folders/"
	if strings.HasPrefix(s, "folders/") {
		// Ensure there is content after the prefix.
		// len("folders/") is 8.
		return len(s) > 8
	}

	// If neither prefix matches, return false.
	return false
}

// GetProjectID extracts the project identifier from a string of the format "projects/{id}".
// It returns the identifier if the format is valid, otherwise it returns an empty
// string and an error.
func GetProjectID(s string) (string, error) {
	if !strings.HasPrefix(s, "projects/") {
		return "", fmt.Errorf("invalid format: string does not have 'projects/' prefix")
	}

	// Use TrimPrefix to get everything after "projects/"
	projectID := strings.TrimPrefix(s, "projects/")

	if projectID == "" {
		return "", fmt.Errorf("invalid format: missing project ID after 'projects/'")
	}
	return projectID, nil
}

func IsFolder(s string) bool {
	if !strings.HasPrefix(s, "folders/") {
		return false
	}
	return true
}

func PrintGeneratedApplication(generatedApplications map[string][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
	defer w.Flush()

	for appName, generatedAppValues := range generatedApplications {
		fmt.Fprintln(w, "APP NAME\tDISCOVERED UUID\tAPP HUB TYPE\tRESOURCE URI")
		fmt.Fprintln(w, "--------\t---------------\t-------------\t-----------")
		// Loop through the slice with the index (i) and value
		fmt.Fprintf(w, "%s\t", appName)
		for i, value := range generatedAppValues {
			// Print the item followed by a tab character
			fmt.Fprintf(w, "%s\t", value)
			if (i+1)%3 == 0 {
				fmt.Fprintf(w, "\n\t")
			}
		}
		fmt.Fprintln(w, "")
		//fmt.Fprintln(w, "APP NAME\tDISCOVERED UUID\tAPP HUB TYPE\tRESOURCE URI")
		//fmt.Fprintln(w, "--------\t---------------\t-------------\t-----------")
	}
}
