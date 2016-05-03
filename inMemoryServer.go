package main

import (
    "net/http"
    "html/template"
    "io/ioutil"
    "fmt"
    "log"
    "os"
    "time"
)

type inMemoryFile struct {
    FileName string
    Data []byte
    Downloads int64
}

const (
    fileDir = "www"
    port = ":9090"
    indexTemplateString = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>inMemoryServer.go</title>
        <style>
            body { background: #222; color: #CCC; font-family: monospace; }
            tr:first-child td { font-weight: bold; border-bottom: 1px solid #666; }
            td { min-width: 12em; }
            table { margin-bottom: 1.5em; }
            a { color: inherit; }
        </style>
	</head>
	<body>
        <h1>inMemoryServer.go</h1>
        <table>
            <tr>
                <td>file name</td>
                <td>size</td>
                <td>downloads</td>
            </tr>
		    {{range .Files}}
                <tr>
                    <td><a href="{{.FileName}}">{{.FileName}}</a></td>
                    <td>{{.Data | len | mb}}</td>
                    <td>{{.Downloads}}</td>
                </tr>
            {{else}}
                <tr><td colspan=3><strong>no rows</strong></tr>
            {{end}}
        </table>
        <i>this file server has been running since {{.StartupTime.Format "Mon, 02 Jan 2006 15:04:05 MST"}} ({{.Uptime}})</i><br>
        <i>it has served a total of {{.Served | mb}} since then (that is about {{.ServedPerDay | mb}} per day)</i><br>
        <i>feedback? let me know at mlinder314@gmail.com</i>
	</body>
</html>`
)

var (
    files = make(map[string]*inMemoryFile)
    totalBytesServed int = 0
    startupTime = time.Now()
    logger = log.New(os.Stdout, "", log.Lmicroseconds)
    funcMap = template.FuncMap{
    	"mb": formatSizeString,
    }
    indexTemplate = template.Must(template.New("index").Funcs(funcMap).Parse(indexTemplateString))
)

func formatSizeString(bytes int) string {
	b := float64(bytes)
	kb := b / 1024.0;
	mb := kb / 1024.0;
	gb := mb / 1024.0;
	tb := gb / 1024.0;
	
	if tb >= 1.0 {
		return fmt.Sprintf("%.1ftb", tb)
	}
	if gb >= 1.0 {
		return fmt.Sprintf("%.1fgb", gb)
	}
	if mb >= 1.0 {
		return fmt.Sprintf("%.1fmb", mb)
	}
	if kb >= 1.0 {
		return fmt.Sprintf("%.1fkb", kb)
	}
	return fmt.Sprintf("%.0fb", b)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Now().Sub(startupTime)
	bytesPerDay := int(float64(totalBytesServed) * 24.0 / (uptime.Hours()))
	templateData := &struct{
		Files map[string]*inMemoryFile
		Served int
		ServedPerDay int
		StartupTime time.Time
		Uptime time.Duration
	} { files, totalBytesServed, bytesPerDay, startupTime, uptime};
    err := indexTemplate.Execute(w, templateData)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
    logRequest := func(str string, args... interface{}) { logger.Printf("Request(%s) from %s: %s", r.URL.Path, r.RemoteAddr, fmt.Sprintf(str, args...)) }
    if r.URL.Path == "/" {
        // Serve Index
        logRequest("index")
        indexHandler(w, r)       
        return
    }
    
    // Serve file
    file := files[r.URL.Path[1:]]
    if file == nil {
        logRequest("not found")
        http.Error(w, "not found", http.StatusNotFound)
        return
    }
    
    logRequest("serving %d bytes", len(file.Data))
    file.Downloads = file.Downloads + 1
    totalBytesServed = totalBytesServed + len(file.Data)
    w.Header().Add("Content-Length", fmt.Sprintf("%d", len(file.Data)))
    _, err := w.Write(file.Data)
    if err != nil {
        logRequest("failed: %s", err.Error())
    }
}

func main() {
    // Load files
    logger.Printf("Loading files from %s/ into memory\n", fileDir)
    dir, _ := ioutil.ReadDir(fileDir)
    for _, file := range dir {
        if file.IsDir() {
            logger.Printf("Skipping dir %s\n", file.Name())
            continue
        }
        logger.Printf("- %s (%d bytes)\n", file.Name(), file.Size())
        data, _ := ioutil.ReadFile(fmt.Sprintf("%s/%s", fileDir, file.Name()))
        files[file.Name()] = &inMemoryFile{
            FileName: file.Name(),
            Data: data,
        }
    }
    logger.Printf("Serving %d files\n", len(files))
        
    // Spin up webserver
    logger.Printf("Listening on %s\n", port)
    http.HandleFunc("/", requestHandler)
    http.ListenAndServe(port, nil)
}