package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const blackBin = "BLACK RUBBISH WHEELIE BIN"
const blueBin = "BLUE RECYCLING WHEELIE BIN"
const foodBox = "FOOD BOX"

type AddressesResonse struct {
	// Param1 bool `json:"param1"` // Unused
	Addresess []struct {
		// Disabled bool   `json:"Disabled"` // Unused
		// Group    any    `json:"Group"` // Unused
		// Selected bool   `json:"Selected"` // Unused
		Address string `json:"Text"`
		UPRN    string `json:"Value"`
	} `json:"param2"`
	// Param3 bool `json:"param3"` // Unused
}

func getAddresses(postcode string) (AddressesResonse, error) {
	payload := url.Values{}
	payload.Set("Postcode", postcode)
	req, err := http.NewRequest(http.MethodPost,
		"https://www.ealing.gov.uk/site/custom_scripts/WasteCollectionWS/home/GetAddress",
		strings.NewReader(payload.Encode()),
	)
	if err != nil {
		return AddressesResonse{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return AddressesResonse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return AddressesResonse{}, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	var addresses AddressesResonse

	err = json.NewDecoder(resp.Body).Decode(&addresses)
	if err != nil {
		return AddressesResonse{}, err
	}

	return addresses, nil
}

type CollectionResponse struct {
	// CalendarCode string `json:"param1"` // Sometimes returns a bool
	Collections []struct {
		Service              string   `json:"Service"`
		CollectionDate       []string `json:"collectionDate"`
		CollectionDateString string   `json:"collectionDateString"`
		// CollectionSchedule   string   `json:"collectionSchedule"` // Sometimes not included
	} `json:"param2"`
	// Param3 bool `json:"param3"` // Unused
}

func getCollection(uprn string) (CollectionResponse, error) {
	payload := url.Values{}
	payload.Set("UPRN", uprn)
	req, err := http.NewRequest(http.MethodPost,
		"https://www.ealing.gov.uk/site/custom_scripts/WasteCollectionWS/home/FindCollection",
		strings.NewReader(payload.Encode()),
	)
	if err != nil {
		return CollectionResponse{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CollectionResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CollectionResponse{}, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	var collection CollectionResponse

	err = json.NewDecoder(resp.Body).Decode(&collection)
	if err != nil {
		return CollectionResponse{}, err
	}

	return collection, nil
}

func nextBin(cr CollectionResponse) (string, string, error) {
	if len(cr.Collections) == 0 {
		return "", "", fmt.Errorf("no collections")
	}

	bin := cr.Collections[0].Service
	dateString := cr.Collections[0].CollectionDate[0]
	date, err := time.Parse("02/01/2006", dateString)
	if err != nil {
		return "", "", fmt.Errorf("unable to parse date %s: %v", dateString, err)
	}

	for _, c := range cr.Collections[1:] {
		if c.Service == blackBin || c.Service == blueBin {
			d, err := time.Parse("02/01/2006", c.CollectionDate[0])
			if err != nil {
				return "", "", fmt.Errorf("unable to parse date %s: %v", c.CollectionDate[0], err)
			}
			if !date.Before(d) {
				bin = c.Service
				dateString = c.CollectionDate[0]
				date = d
			}
		}
	}
	return bin, dateString, nil
}

func assetForBin(bin string) string {
	switch bin {
	case blackBin:
		return "assets/black.png"
	case blueBin:
		return "assets/blue.png"
	case foodBox:
		return "assets/food.png"
	default:
		return ""
	}
}

type CalendarCodeResponse struct {
	Code string `json:"param1"`
	// Param2 string `json:"param2"` // Unused
	// Param3 string `json:"param3"` // Unused
}

func getCalendarCode(uprn string) (CalendarCodeResponse, error) {
	payload := url.Values{}
	payload.Set("UPRN", uprn)
	req, err := http.NewRequest(http.MethodPost,
		"https://www.ealing.gov.uk/site/custom_scripts/WasteCollectionWS/home/GetCalendarCode",
		strings.NewReader(payload.Encode()),
	)
	if err != nil {
		return CalendarCodeResponse{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return CalendarCodeResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return CalendarCodeResponse{}, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	var ccr CalendarCodeResponse

	err = json.NewDecoder(resp.Body).Decode(&ccr)
	if err != nil {
		return CalendarCodeResponse{}, err
	}

	return ccr, nil
}

func handlerForUPRN(uprn string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		col, err := getCollection(uprn)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Unable to get collection: %v", err)
			return
		}

		bin, date, err := nextBin(col)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Unable to calculate next bin: %v", err)
			return
		}

		// This API is unreliable, regularly returns 404
		var calendarLine string
		cc, err := getCalendarCode(uprn)
		if err == nil {
			calendarLink := fmt.Sprintf("https://www.ealing.gov.uk/site/custom_scripts/WasteCollectionWS/docs/WC_Calendar_1222_%s.pdf", cc.Code)
			calendarLine = fmt.Sprintf(`<p><center><a href="%s">Collection Calendar</a></center></p>`, calendarLink)
		}

		tmpl := `<!DOCTYPE html>
			<html>
			<body>
				<p><center><img src="%s" height="500"></center></p>
				<p><center>%s</center></p>
				<p><center>%s</center></p>
				%s
			</body>
			</html>
		`
		asset := assetForBin(bin)
		fmt.Fprintf(w, tmpl, asset, bin, date, calendarLine)
	}
}

func main() {
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	serveUPRN := serveCmd.String("uprn", "", "UPRN")

	if len(os.Args) < 2 {
		fmt.Println("expected 'addresses' or 'serve' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "addresses":
		postcode := strings.Join(os.Args[2:], " ")
		ar, err := getAddresses(postcode)
		if err != nil {
			log.Fatal(err)
		}
		for _, a := range ar.Addresess {
			fmt.Printf("%s: %s\n", a.UPRN, strings.TrimSpace(a.Address))
		}

	case "serve":
		uprn := os.Getenv("EALING_BIN_UPRN")
		err := serveCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
		if *serveUPRN != "" {
			uprn = *serveUPRN
		}

		http.HandleFunc("/", handlerForUPRN(uprn))
		http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
		log.Fatal(http.ListenAndServe(":8080", nil))
	default:
		fmt.Println("expected 'addresses' or 'serve' subcommands")
		os.Exit(1)
	}
}
