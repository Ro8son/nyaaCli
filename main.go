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
	appState   int
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

func mainShell(shellOutput chan string, waitForMain chan bool) {
	for {
		<- waitForMain

		shellPrint()
		 

		shellOutput <- reader()
	}
}

func shellPrint() {
	switch appState {
	case 0:
		fmt.Printf("\n >> ")
	case 1:
		fmt.Printf("\n (mpv) >> ")
	}
}

func stringProcess(option string) string {
	return strings.Replace(strings.TrimPrefix(option, "search"), " ", "+", -1)
}

func mainHandler(shellOutput chan string, waitForMain chan bool) {
	waitForMain <- true
	result := []string{}
	page := 0

	for option := range shellOutput {
		if strings.HasPrefix(option, "search ") {
			appState = 1
			result = get(stringProcess(option))
			page = 1
			list(result, page)
		} else if option == "options" {
			options(1)
		} else if option == "options change" {
			options(2)
		} else if option == "exit" {
			os.Exit(0)
		} else if option == "next" {
			if page + 1 > (len(result) / 120) + 1 {
				list(result, page)
			} else {
				page+= 1
				list(result, page)
			}
		} else if option == "back" {
			if page - 1 < 1 {
			list(result, page)
			} else {
				page -= 1
				list(result, page)
			}
		} else if option == "clear" {
			list(result, page)
		} else {
			if appState == 1 {
				option, err := strconv.Atoi(choice)
				if err != nil {
					list(result, page)
				}

				if option <= 0 || option > len(result)/4 {
					list(result, page)
				}

				link := "https://nyaa.si" + result[(option-1)*4+2]

				go play(link)
				list(result, page)
			} else {
				fmt.Println(HELP)
			}
		}
		waitForMain <- true
	}
}

func options(order int) {
	switch order {
	case 1:
		fmt.Printf("%s %s %s %d", show1, show2, filter, perPage)
	case 2:
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

		writeToFile(show1 + "\n" + show2 + "\n" + a + b + "\n" + strconv.Itoa(perPage))

		filter = a + b

	}

}

func get(search string) []string {
	downloadChan := make(chan *goquery.Selection) 
	endChan := make(chan bool)
	
	go getPage(search, downloadChan, endChan)

	result := processPage(downloadChan, endChan)
	return result
}

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

func list(result []string, page int) {
	max := (len(result) / 120) + 1

	clearScreen()

	leng := len(result) / 4

	for x := (page - 1) * perPage; x <= (page*perPage)-1; x += 1 {
		if x < leng {
			fmt.Printf(" %d == %s\n", x+1, result[x*4+1])
		} else {
			fmt.Println("THE END")
			break
		}

	}
	fmt.Println(" Page: ", page, "MAX ", max)
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
	appState = 0
}

func main() {
	shellOutput := make(chan string)
	waitForMain := make(chan bool)

	go mainShell(shellOutput, waitForMain)

	mainHandler(shellOutput, waitForMain)
	
}

func writeToFile(writeString string) {
	write := []byte(writeString)
	ioutil.WriteFile("settings", write, 0644)
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

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
