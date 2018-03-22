package main
import(
    "encoding/json"
    "fmt"
    "flag"
    "log"
    "io"
    "os"
    "path/filepath"
    "net/http"
    "net/url"
    "strings"
    "sync"
    "github.com/PuerkitoBio/goquery"
)


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
         file, e := os.Create(config.StoragePath + name)
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
         recursiveGetImage(wkUrl, 0)
    })
}

func copy(name string, src io.Reader)(err error){
    file, err := os.Create(name)
    if err != nil{
        fmt.Println("os.Create err. : ", err)
        return
    }
    defer file.Close()

    _, err = io.Copy(file, src)
    if err != nil{
        fmt.Println("io.Copy err. : ", err)
        return
    }
    return
}
func saveImage(url, dir string)(err error){
    l := strings.Split(strings.Split(url, "?")[0], "/")
    name := l[len(l) - 1]
    r, err := http.Get(url)
    if err != nil{
       fmt.Println(err)
       return
    }
    defer r.Body.Close()

    path := filepath.Join(dir, name)
    err = os.MkdirAll(dir, 0755)
    if err != nil{
        fmt.Println("os.MkdirAll err. : ", err)
        return
    }

    err = copy(path, r.Body)
    if err != nil{
        fmt.Println(err)
        return
    }
    fmt.Printf("Done copy. %s to %s\n", url, path)
    return
}

func getDir(u *url.URL)string{
    list := strings.Split(strings.Trim(u.Path, "/"), "/")
    str := strings.Join(list[:len(list) - 1], "_")
    dir := filepath.Join(config.StoragePath, u.Host, str)
    return dir
}

func recursiveGetImage(urlStr string, count int){
    if count > config.Depth{
        return
    }

    _, ok := mapUrl[urlStr]
    if ok {
        return
    }
    mapUrl[urlStr] = true

    res, err := goquery.NewDocument(urlStr)
    if err != nil{
        return
    }

    netUrl, _ := url.Parse(urlStr)
    fmt.Println(netUrl)
    res.Find("img").Each(func(_ int, s *goquery.Selection) {
        urlStr, _ := s.Attr("src")
        workUrl, err := netUrl.Parse(urlStr)
        if err != nil{
            fmt.Println("URL.Parse err. : ", err)
            return
        }

        _, ok := mapUrl[workUrl.String()]
        if ok {
            return
        }

        if err := saveImage(workUrl.String(), getDir(workUrl)); err !=nil{
            fmt.Println("saveImage err. : ", err)
            return
        }

        mapUrl[workUrl.String()] = true
    })
    res.Find("a").Each(func(_ int, s *goquery.Selection) {
         wkUrl, _ := s.Attr("href")
         workUrl, _ := netUrl.Parse(wkUrl)
         recursiveGetImage(workUrl.String(), count + 1)
    })
}

// テキストファイルの取得
func getFileString(path string)(ret string, err error){
    file, err := os.Open(path)
    if err != nil{
        return
    }
    defer file.Close()
    buf := make([]byte, 256)
    for{
        n, err := file.Read(buf)
        if n == 0{
            break
        }
        if err != nil{
            break
        }
        ret += string(buf[:n])
    }
    return
}

// コンフィグ
type Config struct{
    Url string  `json:"url"`
    StoragePath string   `json:"storage_path"`
    Depth int   `json:"depth"`
}
func getConfigByString(str string)(ret Config){
    dec := json.NewDecoder(strings.NewReader(str))
    for{
        if err := dec.Decode(&ret); err == io.EOF{
            break
        } else if err != nil{
            return
        }
    }
    return
}
func getConfig(path string)(ret Config, err error){
    str, err := getFileString(path)
    if err != nil{
        return
    }
    ret = getConfigByString(str)
    return
}
const(
    chNum = 10
)
var(
    config Config
    mapUrl = make(map[string]bool)
    ch = make(chan struct{}, chNum)
    mutex sync.Mutex
)
func init(){
    var err error
    var path string

    flag.StringVar(&path, "path", "", "config.json path")
    flag.Parse()
    if path == ""{
        config, err = getConfig("config.json")
        if err != nil{
            log.Fatal(err)
        }
    }else{
        config, err = getConfig(path)
        if err != nil{
            log.Fatal(err)
        }
    }
}
func main(){
    recursiveGetImage(config.Url, 0)

    fmt.Println("end")
    for key, _ := range mapUrl{
        fmt.Println(key)
    }
}

