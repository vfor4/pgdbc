package main

import (
	"order/logging"
)

func main() {
	// http.HandleFunc("/search", handleSearch)
	// var svr http.Server
	// var wg sync.WaitGroup
	// svr.Addr = ":8081"
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	fmt.Println(svr.ListenAndServe())
	// }()
	//
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	fmt.Println("press any key to shut down service")
	// 	var s string
	// 	fmt.Scan(&s)
	// 	err := svr.Shutdown(context.Background())
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		return
	// 	}
	// }()
	// wg.Wait()
	logging.Logger.Printf("hellow %v", "vu")
}
