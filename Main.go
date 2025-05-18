package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/xyproto/unzip"
)

func main() {
	version := "1.28.0"
	_, err := os.Stat(fmt.Sprintf("./files/nginx-%v", version))
	if os.IsNotExist(err) {
		fmt.Println("Downloading nginx....")
		Filename := downloadFile(version)
		fmt.Println("nginx downloaded....")
		fmt.Println("extracting....")
		unzip.Extract(Filename, "./files")
		os.Remove(Filename)
		fmt.Printf("Extracted to ./files/nginx-%v\n", version)
		var fromport int
		fmt.Print("Enter the Port of ipv4 : ")
		fmt.Scan(&fromport)
		var toport int
		fmt.Print("Enter the Port of ipv6 : ")
		fmt.Scan(&toport)
		EditconfFile(fmt.Sprintf("./files/nginx-%v/conf/nginx.conf", version), fromport, toport)
	} else {
		fmt.Printf("Found Nginx version %v\n", version)
		ipv6 := getipv6()
		if ipv6 != "" {
			fmt.Println("Server started on address : " + ipv6)
		}
		cmd := exec.Command("./nginx.exe")
		cmd.Dir = fmt.Sprintf("./files/nginx-%v", version)
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Command finished with error: %v", err)
			return
		}

		go func() {
			// Wait for signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			<-sigChan
			fmt.Println("Exiting...")

			// Quit nginx
			// cmd.Process.Kill()
			cmd := exec.Command("./nginx.exe", "-s", "quit")
			cmd.Dir = fmt.Sprintf("./files/nginx-%v", version)
			cmd.Run() // Exit the program
			os.Exit(0)
		}()

		// Keep the program running
		select {}
	}
}

func getipv6() string {
	response, err := http.Get("http://ip6only.me/api/")
	if err != nil {
		fmt.Println("Check your internet Connection")
	}
	defer response.Body.Close()
	if err == nil {
		if response.StatusCode == 200 {
			body, _ := io.ReadAll(response.Body)
			ipv6 := strings.Split(string(body), ",")[1]
			return ipv6
		} else {
			fmt.Println("Check your internet Connection")
		}
	}
	return ""
}

func EditconfFile(conffile string, fromport int, toport int) {
	configuration := fmt.Sprintf(`
worker_processes  1;
events {
    worker_connections  1024;
}
stream{
	upstream backend{
	server localhost:%v;
	}
	server{
        listen [::]:%v;
        proxy_pass backend;
        proxy_timeout 10s;
        proxy_connect_timeout 1s;
    }
}
`, fromport, toport)
	os.Remove(conffile)
	newconffile, _ := os.Create(conffile)
	defer newconffile.Close()
	newconffile.WriteString(configuration)
}

func downloadFile(version string) string {
	response, error := http.Get("https://github.com/nginx/nginx/releases/download/release-" + version + "/nginx-" + version + ".zip")
	if error != nil {
		fmt.Printf("You got an error : %v \n", error)
	}
	defer response.Body.Close()

	body, _ := io.ReadAll(response.Body)
	nginxfile, _ := os.Create("nginx.zip")
	nginxfile.Write(body)
	defer nginxfile.Close()
	return nginxfile.Name()
}
