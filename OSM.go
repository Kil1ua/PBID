// Licensed under the MIT License.
// Data source: OpenStreetMap (© OpenStreetMap contributors)

package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

type Coordinates struct {
	Lat string
	Lon string
}

type NominatimResponse []struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

type element struct {
	Tags map[string]string `json:"tags"`
}

type overpass struct {
	Elements []element `json:"elements"`
}

type BuildingInfo struct {
	Metro          string
	Bus            string
	Troll          string
	Tram           string
	Supermarket    string
	Outpost        string
	Bank           string
	Pharmacy       string
	Cafe           string
	Bar            string
	Park           string
	Clinic         string
	Kindergarten   string
	School         string
	BuildingLevels string
	BuildingYear   string
}

type ListingData struct {
	num             string
	BuildingAddress string
	BuildingYear    string
	BuildingLevels  string
	LocationCoords  string
	Metro           string
	Bus             string
	Troll           string
	Tram            string
	Supermarket     string
	Outpost         string
	Bank            string
	Pharmacy        string
	Clinic          string
	Cafe            string
	Bar             string
	Park            string
	Kindergarten    string
	School          string
	CenNear         string
	Now             string
}

var fileMutex sync.Mutex

func Adress(address string) (Coordinates, error) {
	baseURL := "https://nominatim.openstreetmap.org/search"
	u, err := url.Parse(baseURL)
	if err != nil {
		return Coordinates{}, err
	}

	q := u.Query()
	q.Set("q", address)
	q.Set("format", "json")
	q.Set("accept-language", "en-US")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return Coordinates{}, err
	}

	req.Header.Set("User-Agent", "GoRealEstateApp/1.0 (your-email@gmail.com)")

	time.Sleep(1 * time.Second)

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return Coordinates{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return Coordinates{}, fmt.Errorf("Nominatim API returned status %d for address '%s'", resp.StatusCode, address, string(bodyBytes))
	}

	var result NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Coordinates{}, err
	}

	if len(result) > 0 {
		return Coordinates{Lat: result[0].Lat, Lon: result[0].Lon}, nil
	}

	return Coordinates{}, fmt.Errorf("address not found")
}

func GetBuildingInfo(location string, inf *widget.Label) BuildingInfo {

	query := fmt.Sprintf(`
[out:json][timeout:25];
(
  node(around:50,%s)["building"];
  way(around:50,%s)["building"];
  relation(around:50,%s)["building"];

  node["amenity"~"bank|pharmacy|cafe|bar|clinic|kindergarten|school"](around:1000,%s);
  node["shop"~"supermarket|outpost"](around:1000,%s);
  node["railway"~"station|subway_entrance"](around:1000,%s);
  node["highway"="bus_stop"](around:1000,%s);
  node["public_transport"="stop_position"]["trolleybus"="yes"](around:1000,%s);
  node["railway"="tram_stop"](around:1000,%s);
  node["building"="hospital"](around:1000,%s);
  node["leisure"~"park|garden"](around:1000,%s);
);
out body;
`, location, location, location, location, location, location, location, location, location, location, location)

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", "https://overpass-api.de/api/interpreter", bytes.NewBufferString("data="+query))
	if err != nil {
		inf.SetText(fmt.Sprintf("Error creating request to Overpass API for '%s': %v\n", location, err))
		return BuildingInfo{}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "GoRealEstateApp/1.0 (your-email@gmail.com)")

	resp, err := client.Do(req)
	if err != nil {
		inf.SetText(fmt.Sprintf("Overpass API request failed for '%s': %v\n", location, err))
		return BuildingInfo{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		inf.SetText(fmt.Sprintf("HTTP error from Overpass API for '%s': статус %d. Тело: %s\n", location, resp.StatusCode, string(bodyBytes)))
		return BuildingInfo{}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		inf.SetText(fmt.Sprintf("Error reading Overpass API response for '%s': %v\n", location, err))
		return BuildingInfo{}
	}

	var parsed overpass
	if err := json.Unmarshal(body, &parsed); err != nil {
		inf.SetText(fmt.Sprintf("JSON parsing error from Overpass API for '%s': %v\n", location, err))
		return BuildingInfo{}
	}

	buildingInfo := BuildingInfo{
		Metro: "0", Bus: "0", Troll: "0", Tram: "0", Supermarket: "0",
		Outpost: "0", Bank: "0", Pharmacy: "0", Cafe: "0", Bar: "0",
		Clinic: "0", Kindergarten: "0", Park: "0", School: "0",
		BuildingLevels: "N/A", BuildingYear: "N/A",
	}

	for _, el := range parsed.Elements {
		if _, ok := el.Tags["building"]; ok {
			if levels, ok := el.Tags["building:levels"]; ok {
				buildingInfo.BuildingLevels = levels
			}
			if startDate, ok := el.Tags["start_date"]; ok {
				buildingInfo.BuildingYear = startDate
			} else if yearOfConstruction, ok := el.Tags["year_of_construction"]; ok {
				buildingInfo.BuildingYear = yearOfConstruction
			}
		}

		if amenity, ok := el.Tags["amenity"]; ok {
			switch amenity {
			case "bank":
				buildingInfo.Bank = "1"
			case "pharmacy":
				buildingInfo.Pharmacy = "1"
			case "cafe":
				buildingInfo.Cafe = "1"
			case "bar":
				buildingInfo.Bar = "1"
			case "clinic":
				buildingInfo.Clinic = "1"
			case "kindergarten":
				buildingInfo.Kindergarten = "1"
			case "school":
				buildingInfo.School = "1"
			}
		}
		if shop, ok := el.Tags["shop"]; ok {
			switch shop {
			case "supermarket":
				buildingInfo.Supermarket = "1"
			case "outpost":
				buildingInfo.Outpost = "1"
			}
		}
		if railway, ok := el.Tags["railway"]; ok {
			if railway == "station" || railway == "subway_entrance" {
				buildingInfo.Metro = "1"
			}
			if railway == "tram_stop" {
				buildingInfo.Tram = "1"
			}
		}
		if highway, ok := el.Tags["highway"]; ok {
			if highway == "bus_stop" {
				buildingInfo.Bus = "1"
			}
		}
		if publicTransport, ok := el.Tags["public_transport"]; ok {
			if publicTransport == "stop_position" {
				if trolleybus, ok := el.Tags["trolleybus"]; ok && trolleybus == "yes" {
					buildingInfo.Troll = "1"
				}
			}
		}

		if leisure, ok := el.Tags["leisure"]; ok {
			switch leisure {
			case "park", "garden":
				buildingInfo.Park = "1"
			}
		}
	}
	return buildingInfo
}

func scrapeData(fileNameSuffix string, buildingAddress string, coordsStr string, infoData BuildingInfo, statusLabel, errorLabel *widget.Label) {

	csvFileName := fmt.Sprintf("%s_%s.csv", strings.ReplaceAll(buildingAddress, "/", "_"), fileNameSuffix)
	csvFileName = strings.ReplaceAll(csvFileName, "\\", "_")
	csvFileName = strings.ReplaceAll(csvFileName, ":", "_")

	fileMutex.Lock()
	file, err := os.OpenFile(csvFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		errorLabel.SetText(fmt.Sprintf("Failed to open/create the file %s: %v", csvFileName, err))
		fileMutex.Unlock()
		return
	}

	writer := csv.NewWriter(file)

	stat, err := file.Stat()
	if err == nil && stat.Size() == 0 {
		headers := []string{
			"num", "BuildingAddress", "BuildingYear", "BuildingLevels", "LocationCoords",
			"Metro", "Bus", "Troll", "Tram", "Supermarket", "Outpost",
			"Bank", "Pharmacy", "Clinic", "Cafe", "Bar", "Park",
			"Kindergarten", "School", "CenNear", "Now",
		}
		if err := writer.Write(headers); err != nil {
			errorLabel.SetText(fmt.Sprintf("Error writing headers to %s: %v", csvFileName, err))
		}
		writer.Flush()
	}
	fileMutex.Unlock()

	listing := ListingData{
		num:             "1",
		BuildingAddress: buildingAddress,
		BuildingYear:    infoData.BuildingYear,
		BuildingLevels:  infoData.BuildingLevels,
		LocationCoords:  coordsStr,
		Metro:           infoData.Metro,
		Bus:             infoData.Bus,
		Troll:           infoData.Troll,
		Tram:            infoData.Tram,
		Supermarket:     infoData.Supermarket,
		Outpost:         infoData.Outpost,
		Bank:            infoData.Bank,
		Pharmacy:        infoData.Pharmacy,
		Clinic:          infoData.Clinic,
		Cafe:            infoData.Cafe,
		Bar:             infoData.Bar,
		Park:            infoData.Park,
		Kindergarten:    infoData.Kindergarten,
		School:          infoData.School,
		CenNear:         "N/A",
		Now:             "N/A",
	}

	fileMutex.Lock()
	record := []string{
		listing.num, listing.BuildingAddress, listing.BuildingYear, listing.BuildingLevels, listing.LocationCoords,
		listing.Metro, listing.Bus, listing.Troll, listing.Tram, listing.Supermarket, listing.Outpost,
		listing.Bank, listing.Pharmacy, listing.Clinic, listing.Cafe, listing.Bar, listing.Park,
		listing.Kindergarten, listing.School, listing.CenNear, listing.Now,
	}
	if err := writer.Write(record); err != nil {
		dff := fmt.Sprintf("Error writing to CSV: %v", err)
		errorLabel.SetText(dff)
	}
	fileMutex.Unlock()

	file.Close()
	statusLabel.SetText(fmt.Sprintf("Address processing %s Completed! Infrastructure and building OSM data saved in %s.", buildingAddress, csvFileName))
}

func processAddress(address string, statusLabel, errorLabel *widget.Label, suffix string) {
	if address == "No" || address == "" {
		statusLabel.SetText("You did not collect data for this address.")
		return
	}

	coords, err := Adress(address)
	if err != nil {
		errorLabel.SetText(fmt.Sprintf("Geocoding error for '%s': %v", address, err))
		return
	}
	coordsStr := coords.Lat + "," + coords.Lon

	buildingData := GetBuildingInfo(coordsStr, errorLabel)

	canonicalAddress := address

	go scrapeData(suffix, canonicalAddress, coordsStr, buildingData, statusLabel, errorLabel)
}

func main() {
	a := app.New()
	w := a.NewWindow("PBID")

	labelAddress1 := widget.NewLabel("Enter the address of the first house. If not needed, write 'No'")
	entryAddress1 := widget.NewEntry()
	entryAddress1.SetPlaceHolder("Example: 1600 Amphitheatre Parkway, Mountain View, CA")
	entryAddress1.Wrapping = fyne.TextWrapOff

	labelAddress2 := widget.NewLabel("Enter the address of the second house. If not applicable, write 'No'")
	entryAddress2 := widget.NewEntry()
	entryAddress2.SetPlaceHolder("Example: 1 Infinite Loop, Cupertino, CA")
	entryAddress2.Wrapping = fyne.TextWrapOff

	labelAddress3 := widget.NewLabel("Enter the address of the third house. If not needed, write 'No'")
	entryAddress3 := widget.NewEntry()
	entryAddress3.SetPlaceHolder("Example: 350 5th Ave, New York, NY")
	entryAddress3.Wrapping = fyne.TextWrapOff

	labelAddress4 := widget.NewLabel("Enter the address of the fourth house. If not applicable, write 'No'")
	entryAddress4 := widget.NewEntry()
	entryAddress4.SetPlaceHolder("Example: 221B Baker St, London, UK")
	entryAddress4.Wrapping = fyne.TextWrapOff

	inf := widget.NewLabel("")
	status1 := widget.NewLabel("")
	status2 := widget.NewLabel("")
	status3 := widget.NewLabel("")
	status4 := widget.NewLabel("")

	start := widget.NewButton("Start", func() {
		inf.SetText("Processing...")
		status1.SetText("")
		status2.SetText("")
		status3.SetText("")
		status4.SetText("")

		go processAddress(entryAddress1.Text, status1, inf, "address1_info")
		go processAddress(entryAddress2.Text, status2, inf, "address2_info")
		go processAddress(entryAddress3.Text, status3, inf, "address3_info")
		go processAddress(entryAddress4.Text, status4, inf, "address4_info")
	})

	cls := widget.NewButton("String cleaning", func() {
		inf.SetText("")
		entryAddress1.SetText("")
		entryAddress2.SetText("")
		entryAddress3.SetText("")
		entryAddress4.SetText("")
		status1.SetText("")
		status2.SetText("")
		status3.SetText("")
		status4.SetText("")
	})

	th := true
	changeTheme := widget.NewButton("Change theme", func() {
		if th {
			a.Settings().SetTheme(theme.LightTheme())
			th = false
		} else {
			a.Settings().SetTheme(theme.DarkTheme())
			th = true
		}
	})

	scroll1 := container.NewHScroll(entryAddress1)
	scroll2 := container.NewHScroll(entryAddress2)
	scroll3 := container.NewHScroll(entryAddress3)
	scroll4 := container.NewHScroll(entryAddress4)

	w.SetContent(container.NewVBox(
		labelAddress1,
		scroll1,
		labelAddress2,
		scroll2,
		labelAddress3,
		scroll3,
		labelAddress4,
		scroll4,
		changeTheme,
		cls,
		start,
		inf,
		widget.NewSeparator(),
		status1,
		status2,
		status3,
		status4,
	))

	w.ShowAndRun()
}
