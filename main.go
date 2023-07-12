package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	fmt.Println(`
██████╗░███████╗██╗░░░██╗███╗░░██╗██╗░░░██╗██╗░░░░░██╗░░░░░
██╔══██╗██╔════╝██║░░░██║████╗░██║██║░░░██║██║░░░░░██║░░░░░
██║░░██║█████╗░░╚██╗░██╔╝██╔██╗██║██║░░░██║██║░░░░░██║░░░░░
██║░░██║██╔══╝░░░╚████╔╝░██║╚████║██║░░░██║██║░░░░░██║░░░░░
██████╔╝███████╗░░╚██╔╝░░██║░╚███║╚██████╔╝███████╗███████╗
╚═════╝░╚══════╝░░░╚═╝░░░╚═╝░░╚══╝░╚═════╝░╚══════╝╚══════╝
	
It does literally nothing
`)

	log := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("msg=\"Request received\" method=%q URI=%q\n", r.Method, r.RequestURI)
		return
	})
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}
