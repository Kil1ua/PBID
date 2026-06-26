# PBID

PBID - *parsing of buildings and infrastructure data*

This is an __open-source project__ that is distributed under the __MIT__ license, so you can do whatever you want with this app and I won't be able to receive any complaints or issues :)

With this application, it is possible to collect information from open-sources (___Open Street Map and Overpass API___) for further analysis. This application will be very useful in academic and student activities.

---
# Features

- Enter up to **4 addresses** simultaneously (each is processed independently).
- Geocode addresses using **Nominatim** (OpenStreetMap).
- Get data about the building itself (number of floors, year of construction) and the nearby infrastructure (subways, buses, trolleybuses, trams, supermarkets, banks, pharmacies, cafes, bars, parks, schools, kindergartens, clinics, etc.).
- The result is saved in a **CSV file** for each address, including all the collected parameters.
- Simple GUI based on **Fyne** (light/dark theme switching is supported).

---

# Requirements

- **Go** version 1.16 or higher.
- Internet access to work with OpenStreetMap API.
- Operating system: Windows, macOS, Linux (supported by Fyne).

---

# Installation and build

1. Clone the repository (or create a file `OSM.go` with the provided code).
2. Install dependencies (the Fyne module):
   ```bash
   go mod init pbid
   go get fyne.io/fyne/v2
   ```
3. Build the application:
   ```bash
   go build -o pbid OSM.go
   ```
   Or run it immediately:
   ```bash
   go run OSM.gogo
   ```

---

# Usage

1. Launch the application.
2. Enter the addresses of the houses in the four input fields (if you need to process less, enter `No` or leave it blank).
3. Click the **"Start"** button.
4. Wait for the processing to complete - the status of each address will be displayed below the buttons.
5. The results will be saved in files with the following names:
   - `"address"_address1_info.csv`
   - `"address"_address2_info.csv`
   - etc.

The **«String cleaning»** button clears all input fields and statuses.  
The **«Change theme»** button switches the interface theme.

---

# Output data (CSV format)

| Field              | Description                          | Source |
| ----------------- | --------------------------------- | -------- |
| `num`             | Record number (always 1)           | -        |
| `BuildingAddress` | Building address                      | Go       |
| `BuildingYear`    | Year of construction (from OSM)            | OSM      |
| `BuildingLevels`  | Number of floors                 | OSM      |
| `LocationCoords`  | Coordinates (latitude, longitude)      | Overpass |
| `Metro`           | Availability of a metro/station nearby (1/0) | Overpass |
| `Bus`             | Availability of a bus stop      | Overpass |
| `Troll`           | Availability of a trolleybus stop   | Overpass |
| `Tram`            | Availability of a tram stop      | Overpass |
| `Supermarket`     | Availability of a supermarket              | Overpass |
| `Outpost` | Availability of a post office           | Overpass |
| `Bank`            | Availability of a bank                     | Overpass |
| `Pharmacy`        | Availability of a pharmacy                    | Overpass |
| `Clinic`          | Availability of a clinic                   | Overpass |
| `Cafe`            | Availability of a cafe                      | Overpass |
| `Bar`             | Availability of a bar                      | Overpass |
| `Park`            | Availability of a park or garden            | Overpass |
| `Kindergarten`    | Availability of a kindergarten             | Overpass |
| `School`          | Availability of a school                     | Overpass |
| `CenNear`         | (reserved, currently N/A)       | -        |
| `Now`             | (reserved, currently N/A)       | -        |

> **Note:** Values of `1` mean that there is an object within a 1-km radius of the building, but you can change this in the code itself.  The year of construction and the number of floors are only taken into account for the building itself.

---

# Data sources

- [OpenStreetMap](https://www.openstreetmap.org/) (© OpenStreetMap contributors)
- The data is provided under the [ODbL](https://opendatacommons.org/licenses/odbl/) license.
- Geocoding: **Nominatim** (limited to 1 request per second).
- Obtaining POI: **Overpass API** (follow the fair use rules).

---
# License

This software is distributed under the **MIT** license.  
For details, see the [LICENSE](LICENSE) file.

---
# Property Building Infrastructure Data Collector

## Data Sources
- OpenStreetMap (© OpenStreetMap contributors)
- Data licensed under ODbL (https://opendatacommons.org/licenses/odbl/)

## License
This software is licensed under the MIT License.
See LICENSE file for details.

## API Usage
- Nominatim: 1 request/second (compliant with usage policy)
- Overpass API: Fair usage

## Acknowledgements

- The project uses the [Fyne](https://fyne.io/) library for the graphical interface.
- The data is provided by the OpenStreetMap community.
