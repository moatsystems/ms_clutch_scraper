package main

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
)

// structure for an entry about the agency
type entry struct {
	name           string
	logo           string
	rating         string
	tagline        string
	locality       string
	clutch_profile string
	agency_website string
}

// converts raw info into an array of entries and returns it
func convertDataToEntries(data []string, urls []string) []entry {
	entries := make([]entry, len(urls))
	for i := 0; i < len(urls)-1; i++ {
		e := entry{
			name:           data[i*6],
			logo:           data[i*6+1],
			rating:         data[i*6+2],
			tagline:        data[i*6+3],
			locality:       data[i*6+4],
			clutch_profile: data[i*6+5],
			agency_website: urls[i],
		}
		entries[i] = e
	}
	return entries
}

// actual scraping of data
// two arrays are being returned:
// first one contains all the data of the agency
// second one contains the links to the agency sites
// this is done because there are two separate
// containers that needs to be scraped
// to get all the data
func collectData() ([]string, []string) {
	var data []string
	var urls []string

	c := colly.NewCollector(
		colly.AllowedDomains("clutch.co", "www.clutch.co"),                                                                           // Allow requests only to clutch.co
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36"), // User Agent
		colly.IgnoreRobotsTxt(), // Ignore the target machine `robots.txt` statement
	)

	c.OnHTML(".provider-info", func(e *colly.HTMLElement) {

		if len(e.ChildAttr("a[class='company_logotype'] > img", "src")) > 0 {
			data = append(data, e.ChildText(".company_info"))
			data = append(data, e.ChildAttr("a[class='company_logotype'] > img", "src")) // we take src attribute if the image is non-lazy loaded
			data = append(data, e.ChildText(".sg-rating__number"))
			data = append(data, e.ChildText(".tagline"))
			data = append(data, e.ChildText(".locality"))
			data = append(data, "https://clutch.co"+e.ChildAttr("a", "href"))
		} else {
			data = append(data, e.ChildText(".company_info"))
			data = append(data, e.ChildAttr("a[class='company_logotype'] > img", "data-src")) // we take data-src attribute if the image is lazy loaded
			data = append(data, e.ChildText(".sg-rating__number"))
			data = append(data, e.ChildText(".tagline"))
			data = append(data, e.ChildText(".locality"))
			data = append(data, "https://clutch.co"+e.ChildAttr("a", "href"))
		}
	})

	c.OnHTML(".provider-detail", func(e *colly.HTMLElement) { // provider-detail class contains elements with the websites
		urls = append(urls, e.ChildAttr("a[class='website-link__item']", "href"))
	})

	c.OnHTML("a.page-link", func(e *colly.HTMLElement) {
		nextPage := e.Request.AbsoluteURL(e.Attr("href"))
		c.Visit(nextPage)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	c.Visit("https://clutch.co/uk/agencies/digital-marketing/london")
	return data, urls
}

// writing the raw data into the csv file
// parameters:
// w - Writer object for csv file
// data - string array containing data of the agencies
// urls - string array containing websites of the agencies
func writeToCSV(w io.Writer, data []string, urls []string) {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"name", "logo", "rating", "tagline", "locality", "clutch_profile", "agency_website"})

	//merge data arrays for printing to csv
	var fullDataArray = make([][7]string, len(urls))
	index := 0
	for i := 0; i < len(urls)-1; i++ {
		for j := 0; j < 6; j++ {
			fullDataArray[i][j] = data[index]
			index++
		}
	}
	for i, u := range urls {
		fullDataArray[i][6] = u
	}

	for _, l := range fullDataArray {
		writer.Write([]string{strings.Join(l[:], ",")})
	}

}

func main() {
	fName := "./data/london_digital_agencies.csv"

	file, err := os.Create(fName)

	if err != nil {
		log.Fatalf("Cannot create file %q: %s\n", fName, err)
		return
	}

	defer file.Close()

	data, urls := collectData()

	writeToCSV(file, data, urls)

	log.Printf("Scraping finished, check file %q for results\n", fName)

}
