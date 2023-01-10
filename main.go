package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/gocolly/colly"
)

const (
	enGuardiaURL = "https://www.ccma.cat/catradio/alacarta/en-guardia/ultims-programes/?pagina="
	URLprefix    = "https://www.ccma.cat"
	TotalPages   = 68
	DataDir      = "capitols"
)

type Chapter struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Image       string `json:"image"`
	File        string `json:"file"`
	jsonFile    string `json:"-"`
}

func main() {
	action := flag.String("action", "all", "all, scrap or serve")
	dataDir := flag.String("dataDir", DataDir, "data directory")
	pages := flag.Int("pages", TotalPages, "number of pages to scrap")
	flag.Parse()

	if err := os.MkdirAll(*dataDir, 0o755); err != nil {
		panic(err)
	}
	switch *action {
	case "scrap":
		scrap(*dataDir, *pages)
	case "serve", "all":
		if *action == "all" {
			scrap(*dataDir, *pages)
		}
		chapters, err := readChapters(*dataDir)
		if err != nil {
			log.Fatal(err)
		}
		if err := serveWebPage(chapters, *dataDir); err != nil {
			log.Fatal(err)
		}
	default:
		panic("invalid action")
	}
}

func readChapters(dir string) ([]Chapter, error) {
	var chapters []Chapter

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".json") {
			return nil
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		var chapter Chapter
		if err := json.Unmarshal(data, &chapter); err != nil {
			log.Printf("cannot unmarshal json file %s: %v\n", path, err)
			return nil
		}

		chapters = append(chapters, chapter)

		return nil
	})

	if err != nil {
		return nil, err
	}

	// sort chapters by number
	sort.Slice(chapters, func(i, j int) bool {
		numI, errI := getChapterNumber(chapters[i])
		numJ, errJ := getChapterNumber(chapters[j])
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}
		return numI < numJ
	})

	return chapters, nil
}

// getChapterNumber extracts the chapter number from the Title field of the Chapter.
// If the number cannot be extracted, it returns an error.
func getChapterNumber(c Chapter) (int, error) {
	// Try to extract number from title
	var num int
	_, err := fmt.Sscanf(c.Title, "%d", &num)
	if num > 0 && err == nil {
		return num, nil
	}
	// Try to extract the number from the description
	re := regexp.MustCompile(`Capítol (\d+)`)
	matchesList := re.FindStringSubmatch(c.Description)
	if len(matchesList) < 2 {
		return 0, fmt.Errorf("cannot extract number from chapter title or description")
	}

	num, err = strconv.Atoi(matchesList[1])
	if err != nil {
		return 0, err
	}
	return num, nil
}

func serveWebPage(chapters []Chapter, dataDir string) error {
	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		return err
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		if err := tmpl.Execute(w, chapters); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// Serve the local files
	http.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(dataDir))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return http.ListenAndServe(":8080", nil)
}

func scrap(dataDir string, pages int) {
	c := colly.NewCollector()

	data := []Chapter{}
	index := 0

	c.OnHTML(".R-operatiu a", func(e *colly.HTMLElement) {
		if len(data[index].Link) == 0 {
			data[index].Link = "https:" + strings.TrimSpace(e.Attr("href"))
			// Build fileName from fullPath
			fileURL, err := url.Parse(data[index].Link)
			if err != nil {
				log.Printf("error: file name cannot be extracted: %v\n", err)
				return
			}
			path := fileURL.Path
			segments := strings.Split(path, "/")
			if len(segments) == 0 {
				log.Println("error: file name cannot be extracted")
				return
			}
			data[index].File = segments[len(segments)-1]
			data[index].jsonFile = strings.Split(data[index].File, ".")[0] + ".json"
		}
	})

	c.OnHTML(".entradeta", func(e *colly.HTMLElement) {
		if len(data[index].Description) == 0 {
			data[index].Description = e.Text
		}
	})

	c.OnHTML("h1", func(e *colly.HTMLElement) {
		if len(data[index].Title) == 0 {
			data[index].Title = e.Text
		}
	})

	type capitol struct {
		link  string
		image string
	}

	capitols := []capitol{}
	c2 := colly.NewCollector()
	c2.OnHTML(".F-capsaImatge", func(e *colly.HTMLElement) {
		capitols = append(capitols, capitol{
			link:  fmt.Sprintf("%s%s", URLprefix, e.Attr("href")),
			image: e.ChildAttr("img", "src"),
		})
	})

	for page := 1; page <= pages; page++ {
		log.Printf("Page %d of %d\n", page, TotalPages)
		capitols = []capitol{}
		if err := c2.Visit(fmt.Sprintf("%s%d", enGuardiaURL, page)); err != nil {
			log.Printf("error: could not scrap page %d %v\n", page, err)
			continue
		}
		for i, cap := range capitols {
			log.Printf("[%d/%d] scrapping %s\n", i, len(capitols), cap.link)
			data = append(data, Chapter{Image: cap.image})
			if err := c.Visit(cap.link); err != nil {
				log.Printf("error: %v\n", err)
				continue
			}
			saveCapitol(data[index], dataDir)
			index++
		}
	}
}

func saveCapitol(data Chapter, dataDir string) {
	bytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(dataDir, data.jsonFile), bytes, 0o664); err != nil {
		log.Fatal(err)
	}
	download(data.Link, filepath.Join(dataDir, data.File))
}

func download(link, fileName string) {
	inf, err := os.Stat(fileName)
	if err == nil && inf.Size() > 0 {
		log.Printf("file %s already exist, skipping\n", fileName)
		return
	}

	// Create blank file
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	client := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	// Put content on file
	resp, err := client.Get(link)
	if err != nil {
		log.Printf("error downloading file %s: %v\n", fileName, err)
		return
	}
	defer resp.Body.Close()
	size, err := io.Copy(file, resp.Body)
	if err != nil {
		log.Printf("error copying file %s: %v\n", fileName, err)
		return
	}
	defer file.Close()

	fmt.Printf("=> Downloaded file %s with size %d\n", fileName, size)
}
