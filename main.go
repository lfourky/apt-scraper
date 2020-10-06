package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	maxConsecutiveEmptyMatches = 3
)

var (
	scrapeTarget = flag.String("s", "", "oglasi.rs;nekretnine.rs;4zida.rs")
)

func main() {
	flag.Parse()

	now := time.Now()
	defer func() {
		fmt.Println("time elapsed:", time.Since(now))
	}()

	switch *scrapeTarget {
	case "oglasi.rs":
		oglasiRS()
	case "nekretnine.rs":
		nekretnineRS()
	case "4zida.rs":
		cetiriZidaRS()
	}

	fmt.Println("scraping done")
}

func cetiriZidaRS() {
	scrape(
		"https://www.4zida.rs",
		"data/4zida.rs.txt",
		func(idx int) string {
			return fmt.Sprintf(`/prodaja-stanova?search_source=home&strana=%d`, idx+1)
		},
		regexp.MustCompile(`classified-title-and-price" href="(/prodaja/.*?)"`),
	)
}

func nekretnineRS() {
	scrape(
		"https://www.nekretnine.rs",
		"data/nekretnine.rs.txt",
		func(idx int) string {
			return fmt.Sprintf(`/stambeni-objekti/stanovi/izdavanje-prodaja/prodaja/lista/po-stranici/10/stranica/%d`, idx)
		},
		regexp.MustCompile(`<h2 class="offer-title.*?>\s+<a href="(.*?)"\s+`),
	)
}

func oglasiRS() {
	scrape(
		"https://www.oglasi.rs/oglasi",
		"data/oglasi.rs.txt",
		func(idx int) string {
			return fmt.Sprintf(`/nekretnine?p=%d`, idx)
		},
		regexp.MustCompile(`<a class="fpogl-list-image".*?href="(.*?)"`),
	)
}

func scrape(
	baseURL,
	targetFilePath string,
	nextURL func(int) string,
	rgxRent *regexp.Regexp,
) {
	f, err := os.OpenFile(targetFilePath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	panicOnError(err)

	emptyMatches := 0
	for pageIndex, rentIndex := 0, 0; ; pageIndex++ {
		if emptyMatches == maxConsecutiveEmptyMatches {
			fmt.Printf("%d max consecutive empty matches achieved - aborting", maxConsecutiveEmptyMatches)
			return
		}

		url := baseURL + nextURL(pageIndex)
		fmt.Println("Fetching page:", url)

		resp, err := http.Get(url)
		panicOnError(err)

		b, err := ioutil.ReadAll(resp.Body)
		panicOnError(err)

		matches := rgxRent.FindAllSubmatch(b, -1)
		if len(matches) == 0 {
			emptyMatches++
			continue
		}

		for i := 0; i < len(matches); i++ {
			if len(matches[i]) == 0 {
				continue
			}
			rentIndex++
			m := string(matches[i][1])
			rentPage := baseURL + strings.TrimSpace(m)
			fmt.Printf("[%d][%d] %s\n", pageIndex, rentIndex, rentPage)

			_, err := f.WriteString(rentPage + "\n")
			panicOnError(err)
		}
	}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
