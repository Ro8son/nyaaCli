package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	FILTER     []string
	CATEGORIES []string
	show1      string
	show2      string
	filter     string
	perPage    int
	state      int
	//Success    []string
	choice string
	x      int
)

const HELP = `Options:
search <search string>

options
	List filters

options change
	Change filters
exit
	Stop the program

help
	Print this message
`

func mainShell(listControler chan string, result []string, waitForList chan bool) string {
	running, search := true, ""
	for running {
		shell(waitForList)
		search, running = mainReader(listControler, result, waitForList)
	}
	return search
}

func shell(waitForList chan bool) {
	switch state {
	case 0:
		fmt.Printf("\n >> ")
	case 1:
		<-waitForList
		fmt.Printf("\n (mpv) >> ")
	}
}

func mainReader(listControler chan string, result []string, waitForList chan bool) (string, bool) {
	return mainHandler(reader(), listControler, result, waitForList)
}

func mainHandler(option string, listControler chan string, result []string, waitForList chan bool) (string, bool) {
	switch state {
	case 0:
		return normalHandler(option)
	case 1:
		mpvHandler(option, listControler, result, waitForList)
		return "", true
	}
	return "", false
}

func stringProcess(option string) string {
	return strings.Replace(strings.TrimPrefix(option, "search"), " ", "+", -1)
}

func normalHandler(option string) (string, bool) {
	if strings.HasPrefix(option, "search ") {
		state = 1
		return stringProcess(option), false
	} else if option == "options" {
		fmt.Printf("%s %s %s", show1, show2, filter)
		return "", true
	} else if option == "options change" {
		filter = changeOptions()
		return "", true
	} else if option == "exit" {
		os.Exit(0)
	}
	fmt.Println(HELP)
	return "", true

}

func mpvHandler(choice string, listControler chan string, result []string, waitForList chan bool) {
	if choice == "exit" {
		close(listControler)
		state = 0
		main()
	}

	if choice == "next" {
		listControler <- "next"
	} else if choice == "back" {
		listControler <- "back"
	} else if choice == "clear" {
		listControler <- "clear"
	} else {
		choice, err := strconv.Atoi(choice)
		if err != nil {
			listControler <- "wrong input"
			mainShell(listControler, result, waitForList)
		}

		if choice <= 0 || choice > len(result)/4 {
			listControler <- "wrong input"
			mainShell(listControler, result, waitForList)
		}

		link := "https://nyaa.si" + result[(choice-1)*4+2]

		go play(link)
		listControler <- "clear"
	}
}

func changeOptions() string {
	var cat int
	for x := 0; x < 3; x += 1 {
		fmt.Println(x+1, " == ", FILTER[x])
	}
	fmt.Scan(&cat)
	a := FILTER[cat+2]
	show1 = FILTER[cat-1]

	for x := 0; x < 24; x += 1 {
		fmt.Println(x+1, " == ", CATEGORIES[x])
	}
	fmt.Scan(&cat)
	b := CATEGORIES[cat+23]
	show2 = CATEGORIES[cat-1]
	write := []byte(show1 + "\n" + show2 + "\n" + a + b + "\n" + strconv.Itoa(perPage))
	ioutil.WriteFile("settings", write, 0644)

	return a + b
}

func get(search string, downloadChan chan *goquery.Selection, endChan chan bool) []string {

	go getPage(search, downloadChan, endChan)

	result := processPage(downloadChan, endChan)
	return result
}

//func errorCheck()

func getPage(search string, downloadChan chan *goquery.Selection, endChan chan bool) {
	running := true
	a := 1
	for running {
		link := "https://nyaa.si/?q=" + search + filter + "&p=" + strconv.Itoa(a)
		fmt.Println(link)
		resp, err := http.Get(link)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		if end := <-endChan; end == false {
			close(downloadChan)
			running = false
		} else {
			downloadChan <- doc.Find("tbody")
		}
		a += 1
	}
}

func processPage(downloadChan chan *goquery.Selection, endChan chan bool) []string {
	result := []string{}
	endChan <- true

	for tbody := range downloadChan {
		var Success []string
		num := 1
		tbody.Find("td").Each(func(index int, item *goquery.Selection) {
			if num == 9 {
				num = 1
			}
			switch num {
			case 1:
				wa, _ := item.Find("a").Attr("title")
				Success = append(Success, wa)
			case 2:
				hack := ""
				item.Find("a").Each(func(i int, s *goquery.Selection) {
					wa, _ := s.Attr("title")
					hack = wa
				})
				Success = append(Success, hack)
			case 3:
				item.Find("a").Each(func(i int, s *goquery.Selection) {
					wa, _ := s.Attr("href")
					Success = append(Success, wa)
				})
			default:
				//fmt.Println(item.Text())
			}
			num += 1
		})
		result = append(result, Success...)
		if len(Success) == 75*4 {
			endChan <- true
		} else {
			endChan <- false
		}
	}
	return result
}

func list(result []string, listControler chan string, waitForList chan bool, page int) {
	go once(listControler)
	max := (len(result) / 120) + 1
	var err bool

	for cappa := range listControler {
		page, err = controls(cappa, page, max)

		clear()

		leng := len(result) / 4

		for x := (page - 1) * perPage; x <= (page*perPage)-1; x += 1 {
			if x < leng {
				fmt.Printf(" %d == %s\n", x+1, result[x*4+1])
			} else {
				fmt.Println("THE END")
				break
			}

		}
		if err == true {
			fmt.Println("Wrong input")
		}
		fmt.Println(" Page: ", page, "MAX ", max)
		waitForList <- true
	}
}

func controls(cappa string, page int, max int) (int, bool) {
	switch cappa {
	case "next":
		if page+1 > max {
			return page, false
		} else {
			fmt.Println("Page si", page)
			page += 1
			fmt.Println("Page si afer", page)
			return page, false
		}
	case "back":
		if page-1 < 1 {
			return 1, false
		} else {
			page -= 1
			return page, false
		}
	case "clear":
		return page, false
	case "wrong input":
		return page, true
	}
	return 3, false
}

func load() {
	FILTER = []string{"No filter", "No remakes", "Trusted only", "&f=0", "&f=1", "&f=2"}
	CATEGORIES = []string{"All categories",
		"Anime",
		" - Anime Music Video",
		" - English-translated",
		" - Non-English-translated",
		" - Raw",
		"Audio",
		" - Lossless",
		" - Lossy",
		"Literature",
		" - English-translated",
		" - Non-English-translated",
		" - Raw",
		"Live Action",
		" - English-translated",
		" - Idol/Promotional Video",
		" - Non-English-translated",
		" - Raw",
		"Pictures",
		" - Graphics",
		" - Photos",
		"Software",
		" - Applications",
		" - Games",
		"&c=0_0", "&c=1_0", "&c=1_1", "&c=1_2", "&c=1_3", "&c=1_4",
		"&c=2_0", "&c=2_1", "&c=2_2", "&c=3_0", "&c=3_1", "&c=3_2",
		"&c=3_3", "&c=4_0", "&c=4_1", "&c=4_2", "&c=4_3", "&c=4_4",
		"&c=5_0", "&c=5_1", "&c=1_2", "&c=6_0", "&c=6_1", "&c=6_2"}
	content, err := ioutil.ReadFile("settings")
	if err != nil {
		fmt.Printf("%s - Using default options.", err)
		show1 = "No filter"
		show2 = "All categories"
		filter = "&f=0&c=0_0"
		perPage = 30
	} else {
		arr := strings.Split(string(content), "\n")
		show1 = arr[0]
		show2 = arr[1]
		filter = arr[2]
		perPage, _ = strconv.Atoi(arr[3])
		//it works™
	}
}
func init() {
	load()
	state = 0
}

func main() {
	listControler := make(chan string)
	loaded := []string{}
	waitForList := make(chan bool)
	for {
		search := mainShell(listControler, loaded, waitForList)

		downloadChan := make(chan *goquery.Selection)
		endChan := make(chan bool)
		loaded = get(search, downloadChan, endChan)

		go list(loaded, listControler, waitForList, 1)
	}
}

func once(listControler chan string) {
	listControler <- "clear"
}

func reader() string {
	in := bufio.NewReader(os.Stdin)
	choice, _ := in.ReadString('\n')
	return strings.TrimSuffix(choice, "\n")
}

func play(link string) {
	cmd := exec.Command("mpv", link)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func clear() {
	//cmd := exec.Command("clear")
	//cmd.Stdout = os.Stdout
	//cmd.Run()
	fmt.Print("\033[H\033[2J")
}
