package main
import(
    "fmt"
    "log"
    "io"
    "os"
    "net/http"
    "strings"
    "github.com/PuerkitoBio/goquery"
)

var mapUrl map[string]bool

func test(url string){
    res, err := goquery.NewDocument(url)
    if err != nil{
        log.Fatal(err)
    }
    count := 0
    res.Find("img").Each(func(_ int, s *goquery.Selection) {
         url, _ := s.Attr("src")
         l := strings.Split(strings.Split(url, "?")[0], "/")
         name := l[len(l) - 1]
         fmt.Printf("%v : %v\n", count, name)
         r, e := http.Get(url)
         if e != nil{
            log.Fatal(e)
         }
         defer r.Body.Close()
         file, e := os.Create("/tmp/rsrc/" + name)
         if e != nil{
            log.Fatal(e)
         }
         defer file.Close()
         io.Copy(file, r.Body)
         count++
    })
    fmt.Println("-----------------------------")
    res.Find("a").Each(func(_ int, s *goquery.Selection) {
         wkUrl, _ := s.Attr("href")
         getImageByUrl(wkUrl)
    })
}

func getImageByUrl(url string){
    _, ok := mapUrl[url]
    if ok {
        return
    }
    mapUrl[url] = true
    fmt.Println("-----------------------------")
    fmt.Println(url)
    res, err := goquery.NewDocument(url)
    if err != nil{
        fmt.Println(err)
        return
    }
    count := 0
    res.Find("img").Each(func(_ int, s *goquery.Selection) {
        url, _ := s.Attr("src")
        l := strings.Split(strings.Split(url, "?")[0], "/")
        name := l[len(l) - 1]
        fmt.Printf("%v : %v\n", count, name)
        r, e := http.Get(url)
        if e != nil{
           fmt.Println(err)
           return
        }
        defer r.Body.Close()
        file, e := os.Create("/tmp/rsrc/" + name)
        if e != nil{
           fmt.Println(err)
           return
        }
        defer file.Close()
        io.Copy(file, r.Body)
        count++
    })
    res.Find("a").Each(func(_ int, s *goquery.Selection) {
         wkUrl, _ := s.Attr("href")
         getImageByUrl(wkUrl)
    })
}

func main(){
    mapUrl = make(map[string]bool)
    url := "http://www.idea-webtools.com/2014/01/night-view-wallpaper.html"
    getImageByUrl(url)
}

