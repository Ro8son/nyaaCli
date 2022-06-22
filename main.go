package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	a          int
	num        int
	Success    []string
	hack       string
	choice     string
	running    bool
	x          int
	FILTER     []string
	CATEGORIES []string
	filter     string
	show1      string
	show2      string
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

func start(a int) {
	if a == 0 {
		show1 = "No filter"
		show2 = "All categories"
	}
	running = true
	for running {
		fmt.Printf("\n --> ")
		in := bufio.NewReader(os.Stdin)
		option, _ := in.ReadString('\n')
		// command history
		if strings.HasPrefix(option, "search ") {
			process(option, filter)
		} else if option == "exit\n" {
			running = false
		} else if option == "filters\n" {
			fmt.Println(show1, " ", show2, filter)
		} else if option == "filters change\n" {
			filter = change()
		} else {
			fmt.Println(HELP)
		}

	}
}

func change() string {
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

	return a + b
}

func get(c string, filter string, a int) (*goquery.Selection, string, error) {
	searchFor := strings.Replace(strings.TrimSuffix(c, "\n"), " ", "+", -1)

	searchString := "https://nyaa.si/?q=" + searchFor + filter + "&p=" + strconv.Itoa(a)

	fmt.Println(searchString)
	resp, err := http.Get(searchString)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, "", err
	}
	return doc.Find("tbody"), c, nil
}

func additionalFiltr(target *goquery.Selection, c string, a int) []string {
	num = 1
	Success = nil
	target.Find("td").Each(func(index int, item *goquery.Selection) {
		//fmt.Println(item.Text())
		if num == 9 {
			num = 1
		}
		switch num {
		case 1:
			wa, _ := item.Find("a").Attr("title")
			//fmt.Println(wa)
			Success = append(Success, wa)
		case 2:
			hack = ""
			item.Find("a").Each(func(i int, s *goquery.Selection) {
				wa, _ := s.Attr("title")
				hack = wa
			})
			//fmt.Println(hack)
			Success = append(Success, hack)
		case 3:
			item.Find("a").Each(func(i int, s *goquery.Selection) {
				wa, _ := s.Attr("href")
				//fmt.Println(wa)
				Success = append(Success, wa)
			})
		default:
			//fmt.Println(item.Text())
		}
		num += 1
	})
	//fmt.Println(Success)
	return Success
}

func process(search string, filter string) {
	running = true
	a := 1
	loaded := []string{}
	result := []string{}
	for running {
		target, c, err := get(strings.TrimPrefix(search, "search "), filter, a)
		if err != nil {
			fmt.Println(err)
		}

		result = additionalFiltr(target, c, 1)
		loaded = append(loaded, result...)
		//fmt.Println(loaded)
		a += 1
		if len(result) != 4*75 {
			running = false
			break
		}
	}
	max := len(loaded) / 120
	fmt.Println("Max: ", max, len(loaded))
	list(loaded, 1)
	input(loaded, 1, max+1)
}

func list(result []string, page int) {
	Clear()
	leng := len(result) / 4
	perPage := 30
	for x = (page - 1) * perPage; x <= (page*perPage)-1; x += 1 {
		if x < leng {
			fmt.Println(x+1, " == ", result[(x)*4+1])
		} else {
			fmt.Println("THE END")
			break
		}

	}
	fmt.Println("\nPage: ", page)
}

func input(result []string, page int, max int) {
	good := true
	for good {
		fmt.Println("MAX: ", max)
		fmt.Printf("<-- (back) (next) -->")
		fmt.Printf("\n (mpv) --> ")
		fmt.Scan(&choice)

		if choice != "exit" {
			if choice == "next" {
				if page+1 > max {
					list(result, page)
					fmt.Println("end")
				} else {
					page += 1
					list(result, page)
				}
			} else if choice == "back" {
				if page-1 <= 0 {
					list(result, page)
					fmt.Println("end")
				} else {
					page -= 1
					list(result, page)
				}
			} else if choice == "clear" {
				list(result, page)

			} else {
				choice, err := strconv.Atoi(choice)
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
		} else {
			good = false
			start(1)
		}

	}
}

func play(link string) {
	cmd := exec.Command("mpv", link)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func Clear() {
	//cmd := exec.Command("clear")
	//cmd.Stdout = os.Stdout
	//cmd.Run()
	fmt.Print("\033[H\033[2J")
}

func main() {
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
	filter = "&f=0&c=0_0"

	start(0)
}
