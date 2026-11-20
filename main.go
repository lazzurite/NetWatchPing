package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/oschwald/geoip2-golang"
)

func LoadTargets(filename string) ([]Target, error) {
	// Read and decodethe JSON file
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var targets []Target
	if err := json.Unmarshal(file, &targets); err != nil {
		return nil, err
	}

	return targets, nil
}

func SaveTargets(filename string, targets []Target) error {
	// Encode the targets to JSON
	data, err := json.MarshalIndent(targets, "", "  ")
	if err != nil {
		return err
	}

	// Write the JSON data to the file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return err
	}
	return nil
}

func main() {
	Targets := []Target{{Name: "Google", Address: "google.com:443"}}
	loadedTargets, err := LoadTargets("targets.json")
	if err != nil {
		fmt.Println("Error loading targets: ", err)
	} else {
		Targets = loadedTargets
	}
	ticker := time.NewTicker(5 * time.Second)
	db, err := geoip2.Open("GeoLite2-Country.mmdb")
	if err != nil {
		log.Fatal(err)
	}

	//Building a Fyne window
	a := app.New()
	w := a.NewWindow("Net-watch")
	w.Resize(fyne.NewSize(1000, 600))
	statusLabel := widget.NewLabel("Checking servers...")
	statusLabel.TextStyle = fyne.TextStyle{Monospace: true}
	infoLabel := widget.NewLabel("")
	addLabel := widget.NewLabel("Add a server\n")
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Add a server name, for example Steam")
	addressEntry := widget.NewEntry()
	addressEntry.SetPlaceHolder("Add a server address with a port, for example store.steampowered.com:443")
	addButton := widget.NewButton("Add server", func() {
		//A button that adds a server with the nem and address you wrote below
		target := Target{
			Name:    nameEntry.Text,
			Address: addressEntry.Text,
		}
		if target.Name == "" || target.Address == "" {
			infoLabel.SetText("Please enter a valid server name and address with a port.")
		} else {
			Targets = append(Targets, target)
			SaveTargets("targets.json", Targets)
			infoLabel.SetText(fmt.Sprintf("Added server: %s (%s)", target.Name, target.Address))
			nameEntry.SetText("")
			addressEntry.SetText("")
		}
	})
	delLabel := widget.NewLabel("Delete a server\n")
	delEntry := widget.NewEntry()
	delEntry.SetPlaceHolder("Delete a server by name, for example Steam")
	delButton := widget.NewButton("Delete server", func() {
		//A button that deletes a server with the name you wrote below
		if len(Targets) > 0 {
			for i, target := range Targets {
				if target.Name == delEntry.Text {
					Targets = append(Targets[:i], Targets[i+1:]...)
					SaveTargets("targets.json", Targets)
					infoLabel.SetText(fmt.Sprintf("Deleted server: %s", delEntry.Text))
					delEntry.SetText("")
					return
				}
			}
		} else {
			infoLabel.SetText("No servers to delete.")
		}
	})
	w.SetContent(container.NewVBox(statusLabel, addLabel, nameEntry, addressEntry, addButton, delLabel, delEntry, delButton, infoLabel))

	go func() {
		// Initial check before the ticker starts
		check := CheckAll(Targets, db)
		statusLabel.SetText(check)
		for range ticker.C {
			// Creating a goroutine to check the servers every 5 seconds and update the label with the results
			check = CheckAll(Targets, db)
			fyne.Do(func() {
				statusLabel.SetText(check)
			})
		}
	}()

	w.ShowAndRun()
}

type Target struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}
type CheckResult struct {
	Target   Target
	Duration time.Duration
	Error    error
	Region   string
}

func ConnectTCP(network, address string, timeout time.Duration) (time.Duration, error) {
	// Measure the time taken to establish a TCP connection
	start := time.Now()

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return 0, err
	}
	duration := time.Since(start)

	defer conn.Close()

	return duration, nil
}

// Func for getting IP's region
func GetCountry(db *geoip2.Reader, addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	ip, err := net.LookupIP(host)
	if err != nil || len(ip) == 0 {
		return err.Error()
	}
	check, err := db.Country(ip[0])
	if err != nil {
		return "Error when checking the IP in database"
	}

	return check.Country.IsoCode

}

type Site struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Country string `json:"country"` // <-- Add this single line to your existing struct
}

func CheckAll(Targets []Target, db *geoip2.Reader) string {
	resultChan := make(chan CheckResult, len(Targets))
	for _, target := range Targets {
		go func(t Target) {
			duration, err := ConnectTCP("tcp", t.Address, 5*time.Second)

			region := GetCountry(db, t.Address)
			resultChan <- CheckResult{
				Target:   t,
				Duration: duration,
				Error:    err,
				Region:   region,
			}
		}(target)
	}
	var results []CheckResult

	for i := 0; i < len(Targets); i++ {
		results = append(results, <-resultChan)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Target.Name < results[j].Target.Name
	})

	var output string
	// Put the output into a string so Fyne can display it in a label
	output += fmt.Sprintf("Last check at: %s\n", time.Now().Format("15:04:05"))
	output += fmt.Sprintf("%-20s %-30s %-10s %-8s %-10s\n", "Name", "Address", "Status", "Region", "Duration")
	output += fmt.Sprintf("---------------------------------------------------------------------------------------------\n")
	for _, result := range results {
		if result.Error != nil {
			output += fmt.Sprintf("%-20s %-30s %-10s %-8s %-10s\n", result.Target.Name, result.Target.Address, "Offline", result.Region, result.Error)
		} else {
			output += fmt.Sprintf("%-20s %-30s %-10s %-8s %-10s\n", result.Target.Name, result.Target.Address, "Online", result.Region, result.Duration.Round(time.Millisecond))
		}
	}
	output += fmt.Sprintf("---------------------------------------------------------------------------------------------\n")

	return output
}
