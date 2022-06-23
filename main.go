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
	//Success    []string
	choice string
	x      int
)

const HELP = `Options:
search <search string>

filters
	List filters

filters change
	Change filters
exit
	Stop the program

help
	Print this message
`

func mainShell() string {
	running, search := true, ""
	for running {
		out := mainReader()
		search, running = mainHandler(out)
	}
	return search
}

func mainReader() string {
	fmt.Printf("\n --> ")
	in := bufio.NewReader(os.Stdin)
	option, _ := in.ReadString('\n')
	return option
}

func mainHandler(option string) (string, bool) {
	if strings.HasPrefix(option, "search ") {
		return strings.Replace(strings.TrimSuffix(strings.TrimPrefix(option, "search"), "\n"), " ", "+", -1), false
	} else if option == "options\n" {
		fmt.Printf("%s %s %s", show1, show2, filter)
		return "", true
	} else if option == "options change\n" {
		filter = changeOptions()
		return "", true
	} else if option == "exit\n" {
		return "", false
	} else {
		fmt.Println(HELP)
		return "", true
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
	write := []byte(show1 + "\n" + show2 + "\n" + filter + "\n" + strconv.Itoa(perPage))
	ioutil.WriteFile("settings", write, 0644)

	return a + b
}

func get(search string) []string {
	a := 1
	loaded := []string{}
	result := []string{}

	running := true
	for running {
		tbody, err := getPage(search, a)
		if err != nil {
			fmt.Println(err)
		}

		result = processPage(tbody)
		loaded = append(loaded, result...)
		if len(result) != 4*75 {
			running = false
			break
		}
		a += 1
	}
	return loaded
}

func getPage(search string, a int) (*goquery.Selection, error) {

	link := "https://nyaa.si/?q=" + search + filter + "&p=" + strconv.Itoa(a)
	fmt.Println(link)
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc.Find("tbody"), nil
}

func processPage(target *goquery.Selection) []string {
	num := 1
	Success := []string{}

	target.Find("td").Each(func(index int, item *goquery.Selection) {
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
	return Success
}

func list(result []string, page int) {
	clear()
	leng := len(result) / 4

	for x = (page - 1) * perPage; x <= (page*perPage)-1; x += 1 {
		if x < leng {
			fmt.Printf(" %d == %s\n", x+1, result[(x)*4+1])
		} else {
			fmt.Println("THE END")
			break
		}

	}
	fmt.Println(" Page: ", page)
}

func input(result []string, page int, max int) {
	good := true
	for good {
		fmt.Printf(" MAX: %d\n <-- (back) (next) -->\n (mpv) --> ", max)
		in := bufio.NewReader(os.Stdin)
		choice, _ := in.ReadString('\n')

		if choice == "exit\n" {
			main()
		}

		if choice == "next\n" {
			if page+1 > max {
				list(result, page)
				fmt.Println("end")
			} else {
				page += 1
				list(result, page)
			}
		} else if choice == "back\n" {
			if page-1 <= 0 {
				list(result, page)
				fmt.Println("end")
			} else {
				page -= 1
				list(result, page)
			}
		} else if choice == "clear\n" {
			list(result, page)

		} else {
			choice, err := strconv.Atoi(strings.TrimSuffix(choice, "\n"))
			if err != nil {
				list(result, page)
				fmt.Println("Incorrect input")
				input(result, page, max)
			}

			if choice <= 0 || choice > len(result)/4 {
				list(result, page)
				fmt.Println("Out of range")
				input(result, page, max)
			}

			link := "https://nyaa.si" + result[(choice-1)*4+2]

			go play(link)
			list(result, page)
		}

	}
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
}

func main() {
	search := mainShell()

	loaded := get(search)
	max := len(loaded) / 120
	//fmt.Println("Max: ", max, len(loaded))
	list(loaded, 1)
	input(loaded, 1, max+1)
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
