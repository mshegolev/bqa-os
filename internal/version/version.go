package version

import "fmt"

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

func Print() {
	fmt.Println("BQA-OS")
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Commit: %s\n", Commit)
	fmt.Printf("Built: %s\n", BuildDate)
}
